package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	DescribeVpc(ctx context.Context, vpcId string) (*types.Vpc, error)
	DescribeVpcs(ctx context.Context, name string) ([]types.Vpc, error)
	CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tag []types.Tag) (*types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnection(ctx context.Context, vpcPeeringConnectionId string) (*types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnections(ctx context.Context) ([]types.VpcPeeringConnection, error)
	AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*types.VpcPeeringConnection, error)
	DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error
	DescribeRouteTables(ctc context.Context, vpcId string) ([]types.RouteTable, error)
	CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error
	DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error
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

func newClient(ec2Client awsclient.Ec2Client) Client { return &client{Ec2Client: ec2Client} }

var _ Client = (*client)(nil)

type client struct {
	awsclient.Ec2Client
}
