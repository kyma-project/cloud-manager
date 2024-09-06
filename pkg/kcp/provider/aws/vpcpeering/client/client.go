package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	DescribeVpc(ctx context.Context, vpcId string) (*types.Vpc, error)
	DescribeVpcs(ctx context.Context, name string) ([]types.Vpc, error)
	CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tag []types.Tag) (*types.VpcPeeringConnection, error)
	DescribeVpcPeeringConnections(ctx context.Context) ([]types.VpcPeeringConnection, error)
	AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*types.VpcPeeringConnection, error)
	DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error
	DescribeRouteTables(ctc context.Context, vpcId string) ([]types.RouteTable, error)
	CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error
	DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error
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

func (c *client) DescribeVpc(ctx context.Context, vpcId string) (*types.Vpc, error) {
	out, err := c.svc.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Vpcs) > 0 {
		return &out.Vpcs[0], nil
	}
	return nil, nil
}

func (c *client) DescribeVpcs(ctx context.Context, name string) ([]types.Vpc, error) {
	in := &ec2.DescribeVpcsInput{}
	if name != "" {
		in.Filters = []types.Filter{
			{
				Name:   ptr.To("tag:Name"),
				Values: []string{name},
			},
		}
	}
	out, err := c.svc.DescribeVpcs(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.Vpcs, nil
}

func (c *client) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tags []types.Tag) (*types.VpcPeeringConnection, error) {
	out, err := c.svc.CreateVpcPeeringConnection(ctx, &ec2.CreateVpcPeeringConnectionInput{
		VpcId:       vpcId,
		PeerVpcId:   remoteVpcId,
		PeerRegion:  remoteRegion,
		PeerOwnerId: remoteAccountId,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeVpcPeeringConnection,
				Tags:         tags,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return out.VpcPeeringConnection, nil
}

func (c *client) DescribeVpcPeeringConnections(ctx context.Context) ([]types.VpcPeeringConnection, error) {
	out, err := c.svc.DescribeVpcPeeringConnections(ctx, &ec2.DescribeVpcPeeringConnectionsInput{})
	if err != nil {
		return nil, err
	}
	return out.VpcPeeringConnections, err
}

func (c *client) AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*types.VpcPeeringConnection, error) {
	out, err := c.svc.AcceptVpcPeeringConnection(ctx, &ec2.AcceptVpcPeeringConnectionInput{
		VpcPeeringConnectionId: connectionId,
	})

	if err != nil {
		return nil, err
	}

	return out.VpcPeeringConnection, nil
}

func (c *client) DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error {
	_, err := c.svc.DeleteVpcPeeringConnection(ctx, &ec2.DeleteVpcPeeringConnectionInput{
		VpcPeeringConnectionId: connectionId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *client) DescribeRouteTables(ctx context.Context, vpcId string) ([]types.RouteTable, error) {
	out, err := c.svc.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return out.RouteTables, nil
}

func (c *client) CreateRoute(ctx context.Context, routeTableId, destinationCidrBlock, vpcPeeringConnectionId *string) error {
	_, err := c.svc.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:           routeTableId,
		DestinationCidrBlock:   destinationCidrBlock,
		VpcPeeringConnectionId: vpcPeeringConnectionId,
	})

	return err
}

func (c *client) DeleteRoute(ctx context.Context, routeTableId, destinationCidrBlock *string) error {
	_, err := c.svc.DeleteRoute(ctx, &ec2.DeleteRouteInput{
		RouteTableId:         routeTableId,
		DestinationCidrBlock: destinationCidrBlock,
	})
	return err
}
