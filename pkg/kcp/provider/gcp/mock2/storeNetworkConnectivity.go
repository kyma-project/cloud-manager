package mock2

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

func (s *store) CreateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.CreateServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	panic("implement me")
}

func (s *store) UpdateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.UpdateServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.WaitableOperationWithResult[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	panic("implement me")
}

func (s *store) GetServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.GetServiceConnectionPolicyRequest, _ ...gax.CallOption) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	panic("implement me")
}

func (s *store) DeleteServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.DeleteServiceConnectionPolicyRequest, _ ...gax.CallOption) (gcpclient.WaitableVoidOperation, error) {
	panic("implement me")
}

// Operations ======================================

func (s *store) GetNetworkConnectivityOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, _ ...gax.CallOption) (*longrunningpb.Operation, error) {
	panic("implement me")
}

func (s *store) ListNetworkConnectivityOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*longrunningpb.Operation] {
	panic("implement me")
}
