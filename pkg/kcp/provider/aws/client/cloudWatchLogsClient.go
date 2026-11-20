package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type CloudWatchLogsClient interface {
	CreateLogGroup(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error)
	DeleteLogGroup(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error)
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error)
	ListTagsForResource(ctx context.Context, params *cloudwatchlogs.ListTagsForResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, params *cloudwatchlogs.TagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *cloudwatchlogs.UntagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.UntagResourceOutput, error)
}

func NewCloudWatchLogsClient(svc *cloudwatchlogs.Client) CloudWatchLogsClient {
	return &cloudWatchLogsClient{
		svc: svc,
	}
}

var _ CloudWatchLogsClient = (*cloudWatchLogsClient)(nil)

type cloudWatchLogsClient struct {
	svc *cloudwatchlogs.Client
}

func (c *cloudWatchLogsClient) CreateLogGroup(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	return c.svc.CreateLogGroup(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) DeleteLogGroup(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
	return c.svc.DeleteLogGroup(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return c.svc.DescribeLogGroups(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	return c.svc.PutRetentionPolicy(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) ListTagsForResource(ctx context.Context, params *cloudwatchlogs.ListTagsForResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListTagsForResourceOutput, error) {
	return c.svc.ListTagsForResource(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) TagResource(ctx context.Context, params *cloudwatchlogs.TagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.TagResourceOutput, error) {
	return c.svc.TagResource(ctx, params, optFns...)
}

func (c *cloudWatchLogsClient) UntagResource(ctx context.Context, params *cloudwatchlogs.UntagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.UntagResourceOutput, error) {
	return c.svc.UntagResource(ctx, params, optFns...)
}
