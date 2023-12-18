package client

import (
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

type efsClient struct {
	svc efs.Client
}
