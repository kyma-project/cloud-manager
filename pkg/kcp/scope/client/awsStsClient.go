package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/aws/client"
)

func NewAwsStsGardenClientProvider() client.GardenClientProvider[AwsStsClient] {
	return client.NewCachedGardenClientProvider(
		func(ctx context.Context, region, key, secret string) (AwsStsClient, error) {
			cfg, err := client.NewGardenConfig(ctx, region, key, secret)
			if err != nil {
				return nil, err
			}
			return newAwsStsClient(sts.NewFromConfig(cfg)), nil
		},
	)
}

type AwsStsClient interface {
	GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}

func newAwsStsClient(svc *sts.Client) AwsStsClient {
	return &awsStsClient{svc: svc}
}

type awsStsClient struct {
	svc *sts.Client
}

func (c *awsStsClient) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	return c.svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}
