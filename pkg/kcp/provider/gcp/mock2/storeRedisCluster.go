package mock2

import (
	"context"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func (s *store) CreateRedisCluster(ctx context.Context, req *clusterpb.CreateClusterRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*clusterpb.Cluster], error) {
	panic("implement me")
}

func (s *store) GetRedisCluster(ctx context.Context, req *clusterpb.GetClusterRequest, _ ...gax.CallOption) (*clusterpb.Cluster, error) {
	panic("implement me")
}

func (s *store) GetRedisClusterCertificateAuthority(ctx context.Context, req *clusterpb.GetClusterCertificateAuthorityRequest, _ ...gax.CallOption) (*clusterpb.CertificateAuthority, error) {
	panic("implement me")
}

func (s *store) UpdateRedisCluster(ctx context.Context, req *clusterpb.UpdateClusterRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*clusterpb.Cluster], error) {
	panic("implement me")
}

func (s *store) DeleteRedisCluster(ctx context.Context, req *clusterpb.DeleteClusterRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	panic("implement me")
}
