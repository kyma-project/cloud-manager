package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/3th1nk/cidr"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func VpcSubnetsFromScope(scope *cloudcontrolv1beta1.Scope) []VpcSubnet {
	cnt := 0
	result := pie.Flat(pie.Map(scope.Spec.Scope.Aws.Network.Zones, func(z cloudcontrolv1beta1.AwsZone) []VpcSubnet {
		result := []VpcSubnet{
			{
				AZ:   z.Name,
				Cidr: z.Workers,
				Tags: awsutil.Ec2Tags("Name", fmt.Sprintf("%s-%d", scope.Spec.Scope.Aws.VpcNetwork, cnt)),
			},
			{
				AZ:   z.Name,
				Cidr: z.Public,
				Tags: awsutil.Ec2Tags("Name", fmt.Sprintf("%s-public-%d", scope.Spec.Scope.Aws.VpcNetwork, cnt)),
			},
			{
				AZ:   z.Name,
				Cidr: z.Internal,
				Tags: awsutil.Ec2Tags("Name", fmt.Sprintf("%s-internal-%d", scope.Spec.Scope.Aws.VpcNetwork, cnt)),
			},
		}
		cnt++
		return result
	}))
	return result
}

type VpcSubnet struct {
	AZ   string
	Cidr string
	Tags []ec2types.Tag
}

type VpcConfig interface {
	AddVpc(id, cidr string, tags []ec2types.Tag, subnets []VpcSubnet) *ec2types.Vpc
	SetVpcError(id string, err error)
	AddNatGateway(vpcId string, subnetId string) (*ec2types.NatGateway, error)
}

type vpcEntry struct {
	vpc         ec2types.Vpc
	subnets     []ec2types.Subnet
	natGateways []*ec2types.NatGateway
}

var _ VpcConfig = &vpcStore{}

type vpcStore struct {
	m sync.Mutex

	items []*vpcEntry

	addressRange iprangeallocate.AddressSpace

	errorMap map[string]error

	internetGateways []*ec2types.InternetGateway
	dhcpOptions      []*ec2types.DhcpOptions
}

func newVpcStore() *vpcStore {
	return &vpcStore{
		addressRange: iprangeallocate.NewAddressSpace(),
		errorMap:     make(map[string]error),
	}
}

func (s *vpcStore) itemByVpcId(vpcId string) (*vpcEntry, error) {
	idx := pie.FindFirstUsing(s.items, func(e *vpcEntry) bool {
		return ptr.Deref(e.vpc.VpcId, "") == vpcId
	})
	if idx == -1 {
		err := &smithy.GenericAPIError{
			Code:    "InvalidVpcID.NotFound",
			Message: fmt.Sprintf("vpc with id %s does not exist", vpcId),
		}
		return nil, err
	}
	return s.items[idx], nil
}

// Config implementation =======================================

func (s *vpcStore) AddNatGateway(vpcId string, subnetId string) (*ec2types.NatGateway, error) {
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	var subnet *ec2types.Subnet
	for _, sbnt := range item.subnets {
		if ptr.Deref(sbnt.SubnetId, "") == subnetId {
			subnet = &sbnt
			break
		}
	}
	if subnet == nil {
		return nil, fmt.Errorf("subnet with id %s does not exist", subnetId)
	}

	rng := s.addressRange.MustAllocate(32)
	cdr := cidr.ParseNoError(rng)

	gw := &ec2types.NatGateway{
		NatGatewayAddresses: []ec2types.NatGatewayAddress{
			{
				AllocationId: ptr.To(uuid.NewString()),
				IsPrimary:    ptr.To(true),
				PublicIp:     ptr.To(cdr.IP().String()),
			},
		},
	}

	item.natGateways = append(item.natGateways, gw)

	return gw, err
}

func (s *vpcStore) AddVpc(id, cidrVal string, tags []ec2types.Tag, subnets []VpcSubnet) *ec2types.Vpc {
	s.m.Lock()
	defer s.m.Unlock()
	existingIndex := pie.FindFirstUsing(s.items, func(value *vpcEntry) bool {
		return ptr.Deref(value.vpc.VpcId, "xxx") == id
	})
	if existingIndex > -1 {
		return &s.items[existingIndex].vpc
	}

	item := &vpcEntry{
		vpc: ec2types.Vpc{
			VpcId:     ptr.To(id),
			CidrBlock: ptr.To(cidrVal),
			Tags:      tags,
			CidrBlockAssociationSet: []ec2types.VpcCidrBlockAssociation{
				{
					AssociationId: ptr.To(uuid.NewString()),
					CidrBlock:     ptr.To(cidrVal),
					CidrBlockState: &ec2types.VpcCidrBlockState{
						State:         ec2types.VpcCidrBlockStateCodeAssociated,
						StatusMessage: ptr.To("Associated"),
					},
				},
			},
		},
		subnets: pie.Map(subnets, func(x VpcSubnet) ec2types.Subnet {
			return ec2types.Subnet{
				AvailabilityZone:   ptr.To(x.AZ),
				AvailabilityZoneId: ptr.To(x.AZ),
				CidrBlock:          ptr.To(x.Cidr),
				State:              ec2types.SubnetStateAvailable,
				SubnetId:           ptr.To(uuid.NewString()),
				Tags:               append(make([]ec2types.Tag, 0, len(tags)), x.Tags...),
				VpcId:              ptr.To(id),
			}
		}),
	}
	s.items = append(s.items, item)

	return &item.vpc
}

func (s *vpcStore) SetVpcError(id string, err error) {
	s.errorMap[id] = err
}

// Client implementation ========================================

func (s *vpcStore) DescribeDhcpOptions(ctx context.Context, name string) ([]ec2types.DhcpOptions, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var result []ec2types.DhcpOptions
	for _, x := range s.dhcpOptions {
		if name == "" || awsutil.NameEc2TagEquals(x.Tags, name) {
			cpy, err := util.JsonClone(x)
			if err != nil {
				return nil, err
			}
			result = append(result, *cpy)
		}
	}

	return result, nil
}

func (s *vpcStore) CreateDhcpOptions(ctx context.Context, name string, domainName string, tags []ec2types.Tag) (*ec2types.DhcpOptions, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, x := range s.dhcpOptions {
		nm := awsutil.GetEc2TagValue(x.Tags, "Name")
		if nm == name {
			return nil, fmt.Errorf("already have dhcp option named %q", name)
		}
	}

	tags = append(tags, ec2types.Tag{
		Key:   ptr.To("Name"),
		Value: ptr.To(name),
	})

	o := &ec2types.DhcpOptions{
		DhcpConfigurations: []ec2types.DhcpConfiguration{
			{
				Key:    ptr.To("domain-name"),
				Values: []ec2types.AttributeValue{{Value: ptr.To(domainName)}},
			},
			{
				Key:    ptr.To("domain-name-servers"),
				Values: []ec2types.AttributeValue{{Value: ptr.To("AmazonProvidedDNS")}},
			},
		},
		DhcpOptionsId: ptr.To(uuid.NewString()),
		Tags:          tags,
	}
	s.dhcpOptions = append(s.dhcpOptions, o)

	return o, nil
}

func (s *vpcStore) AssociateDhcpOptions(ctx context.Context, vpcId string, dhcpOptionsId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return err
	}

	item.vpc.DhcpOptionsId = ptr.To(dhcpOptionsId)

	return nil
}

func (s *vpcStore) DeleteDhcpOptions(ctx context.Context, dhcpOptionsId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, item := range s.items {
		if ptr.Deref(item.vpc.DhcpOptionsId, "") == dhcpOptionsId {
			return fmt.Errorf("dhcp options %q can not be deleted since they are associated with vpc %q", dhcpOptionsId, ptr.Deref(item.vpc.VpcId, ""))
		}
	}

	found := false
	s.dhcpOptions = pie.FilterNot(s.dhcpOptions, func(x *ec2types.DhcpOptions) bool {
		match := ptr.Deref(x.DhcpOptionsId, "") == dhcpOptionsId
		if match {
			found = true
		}
		return match
	})
	if !found {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("dhcp options %q not found", dhcpOptionsId))
	}

	return nil
}

func (s *vpcStore) CreateInternetGateway(ctx context.Context, name string) (*ec2types.InternetGateway, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	iwg := &ec2types.InternetGateway{
		InternetGatewayId: ptr.To(uuid.New().String()),
		Tags: []ec2types.Tag{
			{
				Key:   ptr.To("Name"),
				Value: ptr.To(name),
			},
		},
	}

	s.internetGateways = append(s.internetGateways, iwg)

	return iwg, nil
}

func (s *vpcStore) AttachInternetGateway(ctx context.Context, vpcId, internetGatewayId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var igw *ec2types.InternetGateway
	for _, x := range s.internetGateways {
		if ptr.Deref(x.InternetGatewayId, "") == internetGatewayId {
			igw = x
			break
		}
	}
	if igw == nil {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("internet gateway %q not found", internetGatewayId))
	}

	var vpc *ec2types.Vpc
	for _, item := range s.items {
		if ptr.Deref(item.vpc.VpcId, "") == vpcId {
			vpc = &item.vpc
			break
		}
	}
	if vpc == nil {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("vpc %q not found", vpcId))
	}

	for _, att := range igw.Attachments {
		if ptr.Deref(att.VpcId, "") == vpcId {
			return fmt.Errorf("vpc %q is already attached to internet gateway %q", vpcId, internetGatewayId)
		}
	}
	igw.Attachments = append(igw.Attachments, ec2types.InternetGatewayAttachment{
		State: ec2types.AttachmentStatusAttached,
		VpcId: ptr.To(vpcId),
	})

	return nil
}

func (s *vpcStore) DescribeInternetGateways(ctx context.Context, name string) ([]ec2types.InternetGateway, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var result []ec2types.InternetGateway
	for _, igw := range s.internetGateways {
		if name == "" || awsutil.NameEc2TagEquals(igw.Tags, name) {
			cpy, err := util.JsonClone(igw)
			if err != nil {
				return nil, err
			}
			result = append(result, *cpy)
		}
	}

	return result, nil
}

func (s *vpcStore) DeleteInternetGateway(ctx context.Context, internetGatewayId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	found := false
	s.internetGateways = pie.FilterNot(s.internetGateways, func(igw *ec2types.InternetGateway) bool {
		match := ptr.Deref(igw.InternetGatewayId, "") == internetGatewayId
		if match {
			found = true
		}
		return match
	})
	if !found {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("internet gateway %q not found", internetGatewayId))
	}
	return nil
}

func (s *vpcStore) CreateVpc(ctx context.Context, name, cidr string, tags []ec2types.Tag) (*ec2types.Vpc, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	tags = append([]ec2types.Tag{
		{
			Key:   ptr.To("Name"),
			Value: ptr.To(name),
		},
	})

	item := &vpcEntry{
		vpc: ec2types.Vpc{
			VpcId:     ptr.To(uuid.NewString()),
			CidrBlock: ptr.To(cidr),
			Tags:      tags,
			State:     ec2types.VpcStateAvailable,
			CidrBlockAssociationSet: []ec2types.VpcCidrBlockAssociation{
				{
					AssociationId: ptr.To(uuid.NewString()),
					CidrBlock:     ptr.To(cidr),
					CidrBlockState: &ec2types.VpcCidrBlockState{
						State:         ec2types.VpcCidrBlockStateCodeAssociated,
						StatusMessage: ptr.To("Associated"),
					},
				},
			},
		},
	}
	s.items = append(s.items, item)

	return &item.vpc, nil
}

func (s *vpcStore) DeleteVpc(ctx context.Context, vpcId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	found := false
	s.items = pie.FilterNot(s.items, func(entry *vpcEntry) bool {
		match := ptr.Deref(entry.vpc.VpcId, "xxx") == vpcId
		if match {
			found = true
		}
		return match
	})

	if !found {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("vpc %s not found", vpcId))
	}

	return nil
}

func (s *vpcStore) DescribeVpc(ctx context.Context, vpcId string) (*ec2types.Vpc, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[vpcId]; ok && err != nil {
		return nil, err
	}

	for _, item := range s.items {
		if ptr.Deref(item.vpc.VpcId, "") == vpcId {
			return &item.vpc, nil
		}
	}

	return nil, nil
}

func (s *vpcStore) DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	all := pie.Map(s.items, func(e *vpcEntry) ec2types.Vpc {
		return e.vpc
	})
	if name == "" {
		return all, nil
	}
	return pie.Filter(all, func(vpc ec2types.Vpc) bool {
		return awsutil.NameEc2TagEquals(vpc.Tags, name)
	}), nil
}

func (s *vpcStore) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidrVal string) (*ec2types.VpcCidrBlockAssociation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	a := ec2types.VpcCidrBlockAssociation{
		AssociationId: ptr.To(uuid.NewString()),
		CidrBlock:     ptr.To(cidrVal),
		CidrBlockState: &ec2types.VpcCidrBlockState{
			State:         ec2types.VpcCidrBlockStateCodeAssociated,
			StatusMessage: ptr.To("Associated"),
		},
	}
	item.vpc.CidrBlockAssociationSet = append(item.vpc.CidrBlockAssociationSet, a)
	return &a, nil
}

func (s *vpcStore) DisassociateVpcCidrBlockInput(ctx context.Context, associationId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var theItem *vpcEntry
	theIndex := -1
	for _, item := range s.items {
		for idx, cidrBlock := range item.vpc.CidrBlockAssociationSet {
			if ptr.Deref(cidrBlock.AssociationId, "") == associationId {
				theItem = item
				theIndex = idx
				break
			}
		}
	}
	if theItem == nil || theIndex == -1 {
		return awsmeta.NewHttpNotFoundError(fmt.Errorf("not found"))
	}

	theItem.vpc.CidrBlockAssociationSet = pie.Delete(theItem.vpc.CidrBlockAssociationSet, theIndex)

	return nil
}

func (s *vpcStore) DescribeSubnets(ctx context.Context, vpcId string) ([]ec2types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	return item.subnets, nil
}

func (s *vpcStore) DescribeSubnet(ctx context.Context, subnetId string) (*ec2types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	for _, item := range s.items {
		for _, subnet := range item.subnets {
			if ptr.Deref(subnet.SubnetId, "") == subnetId {
				return &subnet, nil
			}
		}
	}

	return nil, nil
}

func (s *vpcStore) CreateSubnet(ctx context.Context, vpcId, az, cidrVal string, tags []ec2types.Tag) (*ec2types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	subnet := ec2types.Subnet{
		AvailabilityZone:   ptr.To(az),
		AvailabilityZoneId: ptr.To(az),
		CidrBlock:          ptr.To(cidrVal),
		State:              ec2types.SubnetStateAvailable,
		SubnetId:           ptr.To(uuid.NewString()),
		Tags:               append(make([]ec2types.Tag, 0, len(tags)), tags...),
		VpcId:              ptr.To(vpcId),
	}
	item.subnets = append(item.subnets, subnet)
	return &subnet, nil
}

func (s *vpcStore) DeleteSubnet(ctx context.Context, subnetId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	for _, item := range s.items {
		idx := -1
		for i, subnet := range item.subnets {
			if ptr.Deref(subnet.SubnetId, "") == subnetId {
				idx = i
				break
			}
		}
		if idx > -1 {
			item.subnets = pie.Delete(item.subnets, idx)
			return nil
		}
	}
	return awsmeta.NewHttpNotFoundError(fmt.Errorf("subnet %s does not exist", subnetId))
}

func (s *vpcStore) DescribeNatGateway(ctx context.Context, vpcId string) ([]ec2types.NatGateway, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}

	var result []ec2types.NatGateway
	for _, gw := range item.natGateways {
		cln, err := util.JsonClone(gw)
		if err != nil {
			return nil, err
		}
		result = append(result, *cln)
	}
	return result, nil
}
