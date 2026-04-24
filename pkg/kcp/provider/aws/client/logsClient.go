package client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	logstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"k8s.io/utils/ptr"
)

type LogsClient interface {
	CreateLogGroup(ctx context.Context, logGroupName string) error
	GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error)
	DeleteLogGroup(ctx context.Context, logGroupName string) error
	DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error)
	PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error
	TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error
}

func NewLogsClient(svc *cloudwatchlogs.Client) LogsClient {
	return &logsClient{
		svc: svc,
	}
}

var _ LogsClient = (*logsClient)(nil)

type logsClient struct {
	svc *cloudwatchlogs.Client
}

func (c *logsClient) CreateLogGroup(ctx context.Context, logGroupName string) error {
	_, err := c.svc.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: ptr.To(logGroupName),
	})
	if err != nil {
		var alreadyExists *logstypes.ResourceAlreadyExistsException
		if errors.As(err, &alreadyExists) {
			return nil // Idempotent
		}
		return err
	}
	return nil
}

func (c *logsClient) GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error) {
	out, err := c.svc.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: ptr.To(logGroupName),
		Limit:              ptr.To(int32(1)),
	})
	if err != nil {
		return nil, err
	}
	if len(out.LogGroups) == 0 {
		return nil, &logstypes.ResourceNotFoundException{
			Message: ptr.To("Log group not found"),
		}
	}
	// Ensure exact match (prefix search might return partial matches)
	if *out.LogGroups[0].LogGroupName != logGroupName {
		return nil, &logstypes.ResourceNotFoundException{
			Message: ptr.To("Log group not found"),
		}
	}
	return &out.LogGroups[0], nil
}

func (c *logsClient) DeleteLogGroup(ctx context.Context, logGroupName string) error {
	_, err := c.svc.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: ptr.To(logGroupName),
	})
	if err != nil {
		var notFound *logstypes.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil // Idempotent
		}
		return err
	}
	return nil
}

func (c *logsClient) DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error) {
	out, err := c.svc.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: ptr.To(prefix),
	})
	if err != nil {
		return nil, err
	}
	return out.LogGroups, nil
}

func (c *logsClient) PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error {
	_, err := c.svc.PutRetentionPolicy(ctx, &cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    ptr.To(logGroupName),
		RetentionInDays: ptr.To(retentionDays),
	})
	return err
}

func (c *logsClient) TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error {
	if len(tags) == 0 {
		return nil
	}
	_, err := c.svc.TagLogGroup(ctx, &cloudwatchlogs.TagLogGroupInput{
		LogGroupName: ptr.To(logGroupName),
		Tags:         tags,
	})
	return err
}
