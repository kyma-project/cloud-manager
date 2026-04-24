package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
	CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*wafv2types.WebACL, string, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
	// Logging
	PutLoggingConfiguration(ctx context.Context, input *wafv2.PutLoggingConfigurationInput) error
	GetLoggingConfiguration(ctx context.Context, resourceArn string) (*wafv2types.LoggingConfiguration, error)
	DeleteLoggingConfiguration(ctx context.Context, resourceArn string) error
	// CloudWatch Logs
	CreateLogGroup(ctx context.Context, logGroupName string) error
	GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error)
	DeleteLogGroup(ctx context.Context, logGroupName string) error
	DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error)
	PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error
	TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(
			awsclient.NewWafv2Client(wafv2.NewFromConfig(cfg)),
			awsclient.NewLogsClient(cloudwatchlogs.NewFromConfig(cfg)),
		), nil
	}
}

func newClient(wafv2Client awsclient.Wafv2Client, logsClient awsclient.LogsClient) Client {
	return &client{wafv2Client: wafv2Client, logsClient: logsClient}
}

var _ Client = (*client)(nil)

type client struct {
	wafv2Client awsclient.Wafv2Client
	logsClient  awsclient.LogsClient
}

func (c *client) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*wafv2types.WebACL, string, error) {
	return c.wafv2Client.CreateWebACL(ctx, input)
}

func (c *client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	return c.wafv2Client.GetWebACL(ctx, name, id, scope)
}

func (c *client) UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error {
	return c.wafv2Client.UpdateWebACL(ctx, input)
}

func (c *client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	return c.wafv2Client.DeleteWebACL(ctx, name, id, scope, lockToken)
}

func (c *client) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	return c.wafv2Client.ListWebACLs(ctx, scope)
}

func (c *client) PutLoggingConfiguration(ctx context.Context, input *wafv2.PutLoggingConfigurationInput) error {
	return c.wafv2Client.PutLoggingConfiguration(ctx, input)
}

func (c *client) GetLoggingConfiguration(ctx context.Context, resourceArn string) (*wafv2types.LoggingConfiguration, error) {
	return c.wafv2Client.GetLoggingConfiguration(ctx, resourceArn)
}

func (c *client) DeleteLoggingConfiguration(ctx context.Context, resourceArn string) error {
	return c.wafv2Client.DeleteLoggingConfiguration(ctx, resourceArn)
}

func (c *client) CreateLogGroup(ctx context.Context, logGroupName string) error {
	return c.logsClient.CreateLogGroup(ctx, logGroupName)
}

func (c *client) GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error) {
	return c.logsClient.GetLogGroup(ctx, logGroupName)
}

func (c *client) DeleteLogGroup(ctx context.Context, logGroupName string) error {
	return c.logsClient.DeleteLogGroup(ctx, logGroupName)
}

func (c *client) DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error) {
	return c.logsClient.DescribeLogGroups(ctx, prefix)
}

func (c *client) PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error {
	return c.logsClient.PutRetentionPolicy(ctx, logGroupName, retentionDays)
}

func (c *client) TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error {
	return c.logsClient.TagLogGroup(ctx, logGroupName, tags)
}
