package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	Region() string

	DescribeDhcpOptions(ctx context.Context, name string) ([]ec2types.DhcpOptions, error)
	CreateDhcpOptions(ctx context.Context, name, domainName string, tags []ec2types.Tag) (*ec2types.DhcpOptions, error)
	AssociateDhcpOptions(ctx context.Context, vpcId string, dhcpOptionsId string) error
	DeleteDhcpOptions(ctx context.Context, dhcpOptionsId string) error

	CreateInternetGateway(ctx context.Context, name string) (*ec2types.InternetGateway, error)
	AttachInternetGateway(ctx context.Context, vpcId, internetGatewayId string) error
	DescribeInternetGateways(ctx context.Context, name string) ([]ec2types.InternetGateway, error)
	DeleteInternetGateway(ctx context.Context, internetGatewayId string) error

	CreateVpc(ctx context.Context, name, cidr string, tags []ec2types.Tag) (*ec2types.Vpc, error)
	DeleteVpc(ctx context.Context, vpcId string) error
	DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error)
	DescribeVpc(ctx context.Context, vpcId string) (*ec2types.Vpc, error)
	AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2types.VpcCidrBlockAssociation, error)
	DisassociateVpcCidrBlockInput(ctx context.Context, associationId string) error

	DescribeSecurityGroups(ctx context.Context, filters []ec2types.Filter, groupIds []string) ([]ec2types.SecurityGroup, error)
	CreateSecurityGroup(ctx context.Context, vpcId, name string, tags []ec2types.Tag) (string, error)
	DeleteSecurityGroup(ctx context.Context, id string) error
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(region, awsclient.NewEc2Client(ec2.NewFromConfig(cfg))), nil
	}
}

func newClient(region string, ec2Client awsclient.Ec2Client) Client {
	return &client{
		Ec2Client: ec2Client,
		region:    region,
	}
}

var _ Client = (*client)(nil)

type client struct {
	awsclient.Ec2Client

	region string
}

func (c *client) Region() string {
	return c.region
}
