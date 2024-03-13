package client

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	"github.com/kyma-project/cloud-manager/pkg/metrics"
)

func NewGardenConfig(ctx context.Context, region, key, secret string) (cfg aws.Config, err error) {
	if key == "" || secret == "" {
		err = errors.New("AWS credentials are not specified")
		return
	}
	cfg, err = config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")),
		config.WithRetryer(
			func() aws.Retryer {
				// https://github.com/aws/aws-sdk-go-v2/issues/1744
				return retry.NewStandard()
			},
		),
	)
	cfg.APIOptions = append(cfg.APIOptions, func(stack *smithymiddleware.Stack) error {
		return stack.Deserialize.Add(metrics.AwsReportMetricsMiddleware(), smithymiddleware.After)
	})
	return
}

func NewSkrConfig(ctx context.Context, region, key, secret, assumeRole string) (cfg aws.Config, err error) {
	if key == "" || secret == "" || assumeRole == "" {
		err = errors.New("AWS credentials are not specified")
		return
	}
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
		config.WithRetryer(
			func() aws.Retryer {
				// https://github.com/aws/aws-sdk-go-v2/issues/1744
				return retry.NewStandard()
			},
		),
	)
	cfg.APIOptions = append(cfg.APIOptions, func(stack *smithymiddleware.Stack) error {
		return stack.Deserialize.Add(metrics.AwsReportMetricsMiddleware(), smithymiddleware.After)
	})
	return
}
