package mock

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"k8s.io/utils/ptr"
	"sync"
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
	Tags []ec2Types.Tag
}

type VpcConfig interface {
	AddVpc(id, cidr string, tags []ec2Types.Tag, subnets []VpcSubnet) *ec2Types.Vpc
	SetVpcError(id string, err error)
}

type vpcEntry struct {
	vpc     ec2Types.Vpc
	subnets []ec2Types.Subnet
}

type vpcStore struct {
	m     sync.Mutex
	items []*vpcEntry

	errorMap map[string]error
}

func newVpcStore() *vpcStore {
	return &vpcStore{
		errorMap: make(map[string]error),
	}
}

func (s *vpcStore) itemByVpcId(vpcId string) (*vpcEntry, error) {
	idx := pie.FindFirstUsing(s.items, func(e *vpcEntry) bool {
		return ptr.Deref(e.vpc.VpcId, "") == vpcId
	})
	if idx == -1 {
		return nil, fmt.Errorf("vpc with id %s does not exist", vpcId)
	}
	return s.items[idx], nil
}

// Config implementation =======================================

func (s *vpcStore) AddVpc(id, cidr string, tags []ec2Types.Tag, subnets []VpcSubnet) *ec2Types.Vpc {
	s.m.Lock()
	defer s.m.Unlock()
	existinIndex := pie.FindFirstUsing(s.items, func(value *vpcEntry) bool {
		return ptr.Deref(value.vpc.VpcId, "xxx") == id
	})
	if existinIndex > -1 {
		return &s.items[existinIndex].vpc
	}

	item := &vpcEntry{
		vpc: ec2Types.Vpc{
			VpcId:     ptr.To(id),
			CidrBlock: ptr.To(cidr),
			Tags:      tags,
		},
		subnets: pie.Map(subnets, func(x VpcSubnet) ec2Types.Subnet {
			return ec2Types.Subnet{
				AvailabilityZone:   ptr.To(x.AZ),
				AvailabilityZoneId: ptr.To(x.AZ),
				CidrBlock:          ptr.To(x.Cidr),
				State:              ec2Types.SubnetStateAvailable,
				SubnetId:           ptr.To(uuid.NewString()),
				Tags:               append(make([]ec2Types.Tag, 0, len(tags)), x.Tags...),
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

func (s *vpcStore) DescribeVpc(ctx context.Context, vpcId string) (*ec2Types.Vpc, error) {
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

func (s *vpcStore) DescribeVpcs(ctx context.Context, name string) ([]ec2Types.Vpc, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	all := pie.Map(s.items, func(e *vpcEntry) ec2Types.Vpc {
		return e.vpc
	})
	if name == "" {
		return all, nil
	}
	return pie.Filter(all, func(vpc ec2Types.Vpc) bool {
		return awsutil.NameEc2TagEquals(vpc.Tags, name)
	}), nil
}

func (s *vpcStore) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2Types.VpcCidrBlockAssociation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	a := ec2Types.VpcCidrBlockAssociation{
		AssociationId: ptr.To(uuid.NewString()),
		CidrBlock:     ptr.To(cidr),
		CidrBlockState: &ec2Types.VpcCidrBlockState{
			State:         ec2Types.VpcCidrBlockStateCodeAssociated,
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
		return &smithy.GenericAPIError{
			Code:    "404",
			Message: "Not found",
		}
	}

	theItem.vpc.CidrBlockAssociationSet = pie.Delete(theItem.vpc.CidrBlockAssociationSet, theIndex)

	return nil
}

func (s *vpcStore) DescribeSubnets(ctx context.Context, vpcId string) ([]ec2Types.Subnet, error) {
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

func (s *vpcStore) DescribeSubnet(ctx context.Context, subnetId string) (*ec2Types.Subnet, error) {
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

func (s *vpcStore) CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []ec2Types.Tag) (*ec2Types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	item, err := s.itemByVpcId(vpcId)
	if err != nil {
		return nil, err
	}
	subnet := ec2Types.Subnet{
		AvailabilityZone:   ptr.To(az),
		AvailabilityZoneId: ptr.To(az),
		CidrBlock:          ptr.To(cidr),
		State:              ec2Types.SubnetStateAvailable,
		SubnetId:           ptr.To(uuid.NewString()),
		Tags:               append(make([]ec2Types.Tag, 0, len(tags)), tags...),
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
	return &smithy.GenericAPIError{
		Code:    "404",
		Message: fmt.Sprintf("subnet %s does not exist", subnetId),
	}
}
