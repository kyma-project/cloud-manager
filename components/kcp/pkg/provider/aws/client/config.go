package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func NewGardenConfig(ctx context.Context, region, key, secret string) (cfg aws.Config, err error) {
	cfg, err = config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
	)
	return
}

func NewSkrConfig(ctx context.Context, region, key, secret, assumeRole string) (cfg aws.Config, err error) {
	var assumeCfg aws.Config
	assumeCfg, err = config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
	)
	if err != nil {
		return
	}
	stsCli := sts.NewFromConfig(assumeCfg)
	cfg, err = config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			stscreds.NewAssumeRoleProvider(
				stsCli,
				assumeRole,
			)),
		),
	)
	return
}
