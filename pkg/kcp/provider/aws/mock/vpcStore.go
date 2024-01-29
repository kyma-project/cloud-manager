package mock

import (
	"context"
	"fmt"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"k8s.io/utils/pointer"
)

type VpcSubnet struct {
	AZ   string
	Cidr string
	Tags []ec2Types.Tag
}

type VpcConfig interface {
	AddVpc(id, cidr string, tags []ec2Types.Tag, subnets []VpcSubnet)
}

type vpcEntry struct {
	vpc     ec2Types.Vpc
	subnets []ec2Types.Subnet
}

type vpcStore struct {
	items []vpcEntry
}

func (s *vpcStore) itemByVpcId(vpcId string) (vpcEntry, error) {
	idx := pie.FindFirstUsing(s.items, func(e vpcEntry) bool {
		return pointer.StringDeref(e.vpc.VpcId, "") == vpcId
	})
	if idx == -1 {
		return vpcEntry{}, fmt.Errorf("vpc with id %s does not exist", vpcId)
	}
	return s.items[idx], nil
}

// Config implementation =======================================

func (s *vpcStore) AddVpc(id, cidr string, tags []ec2Types.Tag, subnets []VpcSubnet) {
	s.items = append(s.items, vpcEntry{
		vpc: ec2Types.Vpc{
			VpcId:     pointer.String(id),
			CidrBlock: pointer.String(cidr),
			Tags:      tags,
		},
		subnets: pie.Map(subnets, func(x VpcSubnet) ec2Types.Subnet {
			return ec2Types.Subnet{
				AvailabilityZone:   pointer.String(x.AZ),
				AvailabilityZoneId: pointer.String(x.AZ),
				CidrBlock:          pointer.String(x.Cidr),
				State:              ec2Types.SubnetStateAvailable,
				SubnetId:           pointer.String(uuid.NewString()),
				Tags:               append(make([]ec2Types.Tag, 0, len(tags)), tags...),
				VpcId:              pointer.String(id),
			}
		}),
	})
}

// Client implementation ========================================

func (s *vpcStore) DescribeVpcs(ctx context.Context) ([]ec2Types.Vpc, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	return pie.Map(s.items, func(e vpcEntry) ec2Types.Vpc {
		return e.vpc
	}), nil
}

func (s *vpcStore) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2Types.VpcCidrBlockAssociation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	item, err := s.itemByVpcId(vpcId)
	if err == nil {
		return nil, err
	}
	a := ec2Types.VpcCidrBlockAssociation{
		AssociationId: pointer.String(uuid.NewString()),
		CidrBlock:     pointer.String(cidr),
		CidrBlockState: &ec2Types.VpcCidrBlockState{
			State:         ec2Types.VpcCidrBlockStateCodeAssociated,
			StatusMessage: pointer.String("Associated"),
		},
	}
	item.vpc.CidrBlockAssociationSet = append(item.vpc.CidrBlockAssociationSet, a)
	return &a, nil
}

func (s *vpcStore) DescribeSubnets(ctx context.Context, vpcId string) ([]ec2Types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	item, err := s.itemByVpcId(vpcId)
	if err == nil {
		return nil, err
	}
	return item.subnets, nil
}

func (s *vpcStore) CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []ec2Types.Tag) (*ec2Types.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	item, err := s.itemByVpcId(vpcId)
	if err == nil {
		return nil, err
	}
	subnet := ec2Types.Subnet{
		AvailabilityZone:   pointer.String(az),
		AvailabilityZoneId: pointer.String(az),
		CidrBlock:          pointer.String(cidr),
		State:              ec2Types.SubnetStateAvailable,
		SubnetId:           pointer.String(uuid.NewString()),
		Tags:               append(make([]ec2Types.Tag, 0, len(tags)), tags...),
		VpcId:              pointer.String(vpcId),
	}
	item.subnets = append(item.subnets, subnet)
	return &subnet, nil
}

func (s *vpcStore) DeleteSubnet(ctx context.Context, subnetId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	for _, item := range s.items {
		idx := -1
		for i, subnet := range item.subnets {
			if pointer.StringDeref(subnet.SubnetId, "") == subnetId {
				idx = i
			}
		}
		if idx > -1 {
			item.subnets = pie.Delete(item.subnets, idx)
			return nil
		}
	}
	return fmt.Errorf("subnet with id %s", subnetId)
}
