package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	client2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	DescribeVpcs(ctx context.Context) ([]types.Vpc, error)
}

func NewClientProvider() client2.SkrClientProvider[Client] {
	return func(ctx context.Context, region, key, secret, role string) (Client, error) {
		cfg, err := client2.NewSkrConfig(ctx, region, key, secret, role)
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
