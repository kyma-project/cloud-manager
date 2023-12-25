package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Factory[T any] func(cfg aws.Config) T

type ConfigBuilder interface {
	WithRegion(region string) ConfigBuilder
	WithCredentials(key, secret string) ConfigBuilder
	WithAssumeRole(role string) ConfigBuilder
	Build(ctx context.Context) (aws.Config, error)
}

func NewConfigBuilder() ConfigBuilder {
	return &configBuilder{}
}

type configBuilder struct {
	assumeRole string
	region     string
	key        string
	secret     string
}

func (b *configBuilder) WithRegion(region string) ConfigBuilder {
	b.region = region
	return b
}

func (b *configBuilder) WithCredentials(key, secret string) ConfigBuilder {
	b.key = key
	b.secret = secret
	return b
}

func (b *configBuilder) WithAssumeRole(role string) ConfigBuilder {
	b.assumeRole = role
	return b
}

func (b *configBuilder) Build(ctx context.Context) (aws.Config, error) {
	var cfg aws.Config
	var err error
	if len(b.assumeRole) > 0 {
		assumeCfg, e := config.LoadDefaultConfig(
			ctx,
			config.WithRegion(b.region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(b.key, b.secret, "")),
		)
		if e != nil {
			return cfg, e
		}
		stsCli := sts.NewFromConfig(assumeCfg)
		cfg, err = config.LoadDefaultConfig(
			ctx,
			config.WithRegion(b.region),
			config.WithCredentialsProvider(aws.NewCredentialsCache(
				stscreds.NewAssumeRoleProvider(
					stsCli,
					b.assumeRole,
				)),
			),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(
			ctx,
			config.WithRegion(b.region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(b.key, b.secret, "")),
		)
	}
	return cfg, err
}
