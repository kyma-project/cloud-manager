package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type StsClient interface {
	GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}

func StsClientFactory(cfg aws.Config) StsClient {
	return &stsClient{svc: sts.NewFromConfig(cfg)}
}

type stsClient struct {
	svc *sts.Client
}

func (c *stsClient) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	return c.svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}
