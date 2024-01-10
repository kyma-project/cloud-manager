package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-resources-manager/components/kcp/pkg/provider/aws/client"
	"k8s.io/utils/pointer"
)

type Client interface {
	DescribeVpcs(ctx context.Context) ([]types.Vpc, error)
	AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*types.VpcCidrBlockAssociation, error)
	DescribeSubnets(ctx context.Context, vpcId string) ([]types.Subnet, error)
	CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []types.Tag) (*types.Subnet, error)
	DeleteSubnet(ctx context.Context, subnetId string) error
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return awsclient.NewCachedSkrClientProvider(
		func(ctx context.Context, region, key, secret, role string) (Client, error) {
			cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
			if err != nil {
				return nil, err
			}
			return newClient(ec2.NewFromConfig(cfg)), nil
		},
	)
}

func newClient(svc *ec2.Client) Client {
	return &client{svc: svc}
}

type client struct {
	svc *ec2.Client
}

func (c *client) DescribeVpcs(ctx context.Context) ([]types.Vpc, error) {
	out, err := c.svc.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}
	return out.Vpcs, nil
}

func (c *client) AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*types.VpcCidrBlockAssociation, error) {
	in := &ec2.AssociateVpcCidrBlockInput{
		VpcId:     aws.String(vpcId),
		CidrBlock: aws.String(cidr),
	}
	out, err := c.svc.AssociateVpcCidrBlock(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.CidrBlockAssociation, nil
}

func (c *client) DescribeSubnets(ctx context.Context, vpcId string) ([]types.Subnet, error) {
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

func (c *client) CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []types.Tag) (*types.Subnet, error) {
	in := &ec2.CreateSubnetInput{
		VpcId:            pointer.String(vpcId),
		AvailabilityZone: pointer.String(az),
		CidrBlock:        pointer.String(cidr),
	}
	if len(tags) > 0 {
		in.TagSpecifications = []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSubnet,
				Tags:         tags,
			},
		}
	}
	out, err := c.svc.CreateSubnet(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.Subnet, nil
}

func (c *client) DeleteSubnet(ctx context.Context, subnetId string) error {
	_, err := c.svc.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: pointer.String(subnetId),
	})
	if err != nil {
		return err
	}
	return nil
}
