package mock2

import (
	"context"

	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func (s *store) CreateRedisInstance(ctx context.Context, req *redispb.CreateInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*redispb.Instance], error) {
	panic("implement me")
}

func (s *store) GetRedisInstance(ctx context.Context, req *redispb.GetInstanceRequest, _ ...gax.CallOption) (*redispb.Instance, error) {
	panic("implement me")
}

func (s *store) UpdateRedisInstance(ctx context.Context, req *redispb.UpdateInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*redispb.Instance], error) {
	panic("implement me")
}

func (s *store) UpgradeRedisInstance(ctx context.Context, req *redispb.UpgradeInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*redispb.Instance], error) {
	panic("implement me")
}

func (s *store) DeleteRedisInstance(ctx context.Context, req *redispb.DeleteInstanceRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	panic("implement me")
}
