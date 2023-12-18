package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/utils/pointer"
)

type networkClient struct {
	svc *ec2.Client
}

func (c *networkClient) DescribeVpcs(ctx context.Context) ([]types.Vpc, error) {
	out, err := c.svc.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return out.Vpcs, nil
}

func (c *networkClient) DescribeSubnets(ctx context.Context, vpcId string) ([]types.Subnet, error) {
	out, err := c.svc.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   pointer.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return out.Subnets, nil
}

func (c *networkClient) CreateSubnet(ctx context.Context, vpcId, az, cidr string) (*types.Subnet, error) {
	out, err := c.svc.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:            pointer.String(vpcId),
		AvailabilityZone: pointer.String(az),
		CidrBlock:        pointer.String(cidr),
	})
	if err != nil {
		return nil, err
	}
	return out.Subnet, nil
}

func (c *networkClient) DeleteSubnet(ctx context.Context, subnetId string) error {
	_, err := c.svc.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: pointer.String(subnetId),
	})
	if err != nil {
		return err
	}
	return nil
}
