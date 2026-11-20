package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	awsclient.CloudWatchLogsClient
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(
			awsclient.NewCloudWatchLogsClient(cloudwatchlogs.NewFromConfig(cfg)),
		), nil
	}
}

func newClient(cwLogs awsclient.CloudWatchLogsClient) Client {
	return &client{
		CloudWatchLogsClient: cwLogs,
	}
}

var _ Client = (*client)(nil)

type client struct {
	awsclient.CloudWatchLogsClient
}
