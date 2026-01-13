package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	DescribeVpc(ctx context.Context, vpcId string) (*ec2types.Vpc, error)
	DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error)
	AssociateVpcCidrBlock(ctx context.Context, vpcId, cidr string) (*ec2types.VpcCidrBlockAssociation, error)
	DisassociateVpcCidrBlockInput(ctx context.Context, associationId string) error
	DescribeSubnets(ctx context.Context, vpcId string) ([]ec2types.Subnet, error)
	CreateSubnet(ctx context.Context, vpcId, az, cidr string, tags []ec2types.Tag) (*ec2types.Subnet, error)
	DeleteSubnet(ctx context.Context, subnetId string) error
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(awsclient.NewEc2Client(ec2.NewFromConfig(cfg))), nil
	}
}

func newClient(ec2Client awsclient.Ec2Client) Client {
	return &client{Ec2Client: ec2Client}
}

var _ Client = (*client)(nil)

type client struct {
	awsclient.Ec2Client
}
