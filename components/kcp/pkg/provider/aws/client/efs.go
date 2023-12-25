package client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

type EfsClient interface {
}

func EfsClientFactory(cfg aws.Config) EfsClient {
	return &efsClient{svc: efs.NewFromConfig(cfg)}
}

type efsClient struct {
	svc *efs.Client
}
