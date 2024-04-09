package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	DescribeVpcs(ctx context.Context) ([]types.Vpc, error)
	CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string) (*types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnections(ctx context.Context) ([]types.VpcPeeringConnection, error)
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(ec2.NewFromConfig(cfg)), nil
	}
}

func newClient(svc *ec2.Client) Client { return &client{svc: svc} }

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

func (c *client) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string) (*types.VpcPeeringConnection, error) {
	//out, err := c.svc.CreateVpcPeeringConnection(ctx, &ec2.CreateVpcPeeringConnectionInput{
	//	VpcId:       vpcId,
	//	PeerVpcId:   remoteVpcId,
	//	PeerRegion:  remoteRegion,
	//	PeerOwnerId: remoteAccountId,
	//})
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//return out.VpcPeeringConnection, nil

	return nil, nil
}

func (c *client) DescribeVpcPeeringConnections(ctx context.Context) ([]types.VpcPeeringConnection, error) {
	out, err := c.svc.DescribeVpcPeeringConnections(ctx, &ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		return nil, err
	}
	return out.VpcPeeringConnections, err
}
