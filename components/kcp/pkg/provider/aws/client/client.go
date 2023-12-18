package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

func NewConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	return cfg, err
}

func NewEfsClient(ctx context.Context) (*efs.Client, error) {
	cfg, err := NewConfig(ctx)
	if err != nil {
		return nil, err
	}
	svc := efs.NewFromConfig(cfg)
	return svc, nil
}
