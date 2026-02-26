package client

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	cluster "cloud.google.com/go/redis/cluster/apiv1"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/googleapis/gax-go/v2"
)

type RedisClusterClient interface {
	CreateRedisCluster(ctx context.Context, req *clusterpb.CreateClusterRequest, opts ...gax.CallOption) (ResultOperation[*clusterpb.Cluster], error)
	GetRedisCluster(ctx context.Context, req *clusterpb.GetClusterRequest, opts ...gax.CallOption) (*clusterpb.Cluster, error)
	GetRedisClusterCertificateAuthority(ctx context.Context, req *clusterpb.GetClusterCertificateAuthorityRequest, opts ...gax.CallOption) (*clusterpb.CertificateAuthority, error)
	UpdateRedisCluster(ctx context.Context, req *clusterpb.UpdateClusterRequest, opts ...gax.CallOption) (ResultOperation[*clusterpb.Cluster], error)
	DeleteRedisCluster(ctx context.Context, req *clusterpb.DeleteClusterRequest, opts ...gax.CallOption) (VoidOperation, error)

	GetRedisClusterOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error)
	ListRedisClusterOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation]
}

var _ RedisClusterClient = (*redisClusterClient)(nil)

type redisClusterClient struct {
	inner *cluster.CloudRedisClusterClient
}

func (c *redisClusterClient) CreateRedisCluster(ctx context.Context, req *clusterpb.CreateClusterRequest, opts ...gax.CallOption) (ResultOperation[*clusterpb.Cluster], error) {
	return c.inner.CreateCluster(ctx, req, opts...)
}

func (c *redisClusterClient) GetRedisCluster(ctx context.Context, req *clusterpb.GetClusterRequest, opts ...gax.CallOption) (*clusterpb.Cluster, error) {
	return c.inner.GetCluster(ctx, req, opts...)
}

func (c *redisClusterClient) GetRedisClusterCertificateAuthority(ctx context.Context, req *clusterpb.GetClusterCertificateAuthorityRequest, opts ...gax.CallOption) (*clusterpb.CertificateAuthority, error) {
	return c.inner.GetClusterCertificateAuthority(ctx, req, opts...)
}

func (c *redisClusterClient) UpdateRedisCluster(ctx context.Context, req *clusterpb.UpdateClusterRequest, opts ...gax.CallOption) (ResultOperation[*clusterpb.Cluster], error) {
	return c.inner.UpdateCluster(ctx, req, opts...)
}

func (c *redisClusterClient) DeleteRedisCluster(ctx context.Context, req *clusterpb.DeleteClusterRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.DeleteCluster(ctx, req, opts...)
}

func (c *redisClusterClient) GetRedisClusterOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error) {
	return c.inner.GetOperation(ctx, req, opts...)
}

func (c *redisClusterClient) ListRedisClusterOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation] {
	return c.inner.ListOperations(ctx, req, opts...)
}
