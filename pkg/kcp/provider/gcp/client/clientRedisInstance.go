package client

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	redis "cloud.google.com/go/redis/apiv1"
	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/googleapis/gax-go/v2"
)

type RedisInstanceClient interface {
	CreateRedisInstance(ctx context.Context, req *redispb.CreateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error)
	GetRedisInstance(ctx context.Context, req *redispb.GetInstanceRequest, opts ...gax.CallOption) (*redispb.Instance, error)
	UpdateRedisInstance(ctx context.Context, req *redispb.UpdateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error)
	UpgradeRedisInstance(ctx context.Context, req *redispb.UpgradeInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error)
	DeleteRedisInstance(ctx context.Context, req *redispb.DeleteInstanceRequest, opts ...gax.CallOption) (VoidOperation, error)

	GetRedisInstanceOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error)
	ListRedisInstanceOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation]
}

var _ RedisInstanceClient = (*redisInstanceClient)(nil)

type redisInstanceClient struct {
	inner *redis.CloudRedisClient
}

func (c *redisInstanceClient) CreateRedisInstance(ctx context.Context, req *redispb.CreateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error) {
	return c.inner.CreateInstance(ctx, req)
}

func (c *redisInstanceClient) GetRedisInstance(ctx context.Context, req *redispb.GetInstanceRequest, opts ...gax.CallOption) (*redispb.Instance, error) {
	return c.inner.GetInstance(ctx, req)
}

func (c *redisInstanceClient) UpdateRedisInstance(ctx context.Context, req *redispb.UpdateInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error) {
	return c.inner.UpdateInstance(ctx, req)
}

func (c *redisInstanceClient) UpgradeRedisInstance(ctx context.Context, req *redispb.UpgradeInstanceRequest, opts ...gax.CallOption) (ResultOperation[*redispb.Instance], error) {
	return c.inner.UpgradeInstance(ctx, req, opts...)
}

func (c *redisInstanceClient) DeleteRedisInstance(ctx context.Context, req *redispb.DeleteInstanceRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.DeleteInstance(ctx, req, opts...)
}

func (c *redisInstanceClient) GetRedisInstanceOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error) {
	return c.inner.GetOperation(ctx, req, opts...)
}

func (c *redisInstanceClient) ListRedisInstanceOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation] {
	return c.inner.ListOperations(ctx, req, opts...)
}
