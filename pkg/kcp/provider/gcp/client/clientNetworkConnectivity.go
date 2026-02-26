package client

import (
	"context"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	networkconnectivity "cloud.google.com/go/networkconnectivity/apiv1"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"github.com/googleapis/gax-go/v2"
)

type NetworkConnectivityClient interface {
	CreateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.CreateServiceConnectionPolicyRequest, opts ...gax.CallOption) (ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error)
	UpdateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.UpdateServiceConnectionPolicyRequest, opts ...gax.CallOption) (ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error)
	GetServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.GetServiceConnectionPolicyRequest, opts ...gax.CallOption) (*networkconnectivitypb.ServiceConnectionPolicy, error)
	DeleteServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.DeleteServiceConnectionPolicyRequest, opts ...gax.CallOption) (VoidOperation, error)

	GetNetworkConnectivityOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error)
	ListNetworkConnectivityOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation]
}

var _ NetworkConnectivityClient = (*networkConnectivityClient)(nil)

type networkConnectivityClient struct {
	inner *networkconnectivity.CrossNetworkAutomationClient
}

func (c *networkConnectivityClient) CreateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.CreateServiceConnectionPolicyRequest, opts ...gax.CallOption) (ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	return c.inner.CreateServiceConnectionPolicy(ctx, req, opts...)
}

func (c *networkConnectivityClient) UpdateServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.UpdateServiceConnectionPolicyRequest, opts ...gax.CallOption) (ResultOperation[*networkconnectivitypb.ServiceConnectionPolicy], error) {
	return c.inner.UpdateServiceConnectionPolicy(ctx, req, opts...)
}

func (c *networkConnectivityClient) GetServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.GetServiceConnectionPolicyRequest, opts ...gax.CallOption) (*networkconnectivitypb.ServiceConnectionPolicy, error) {
	return c.inner.GetServiceConnectionPolicy(ctx, req, opts...)
}

func (c *networkConnectivityClient) DeleteServiceConnectionPolicy(ctx context.Context, req *networkconnectivitypb.DeleteServiceConnectionPolicyRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.DeleteServiceConnectionPolicy(ctx, req, opts...)
}

func (c *networkConnectivityClient) GetNetworkConnectivityOperation(ctx context.Context, req *longrunningpb.GetOperationRequest, opts ...gax.CallOption) (*longrunningpb.Operation, error) {
	return c.inner.GetOperation(ctx, req, opts...)
}

func (c *networkConnectivityClient) ListNetworkConnectivityOperations(ctx context.Context, req *longrunningpb.ListOperationsRequest, opts ...gax.CallOption) Iterator[*longrunningpb.Operation] {
	return c.inner.ListOperations(ctx, req, opts...)
}
