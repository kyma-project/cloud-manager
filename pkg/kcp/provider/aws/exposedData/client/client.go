package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error)
	DescribeNatGateways(ctx context.Context, vpcId string) ([]ec2types.NatGateway, error)
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(ec2.NewFromConfig(cfg)), nil
	}
}

func newClient(svc *ec2.Client) Client {
	return &client{svc: svc}
}

type client struct {
	svc *ec2.Client
}

func (c *client) DescribeVpcs(ctx context.Context, name string) ([]ec2types.Vpc, error) {
	in := &ec2.DescribeVpcsInput{}
	if name != "" {
		in.Filters = []ec2types.Filter{
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

func (c *client) DescribeNatGateways(ctx context.Context, vpcId string) ([]ec2types.NatGateway, error) {
	in := &ec2.DescribeNatGatewaysInput{
		Filter: []ec2types.Filter{
			{
				Name:   ptr.To("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}
	resp, err := c.svc.DescribeNatGateways(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp.NatGateways, nil
}
