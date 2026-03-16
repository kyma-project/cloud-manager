package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type ComputeGlobalOperationsClient interface {
	GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error)
	ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, opts ...gax.CallOption) Iterator[*computepb.Operation]
}

var _ ComputeGlobalOperationsClient = (*computeGlobalOperationsClient)(nil)

type computeGlobalOperationsClient struct {
	inner *compute.GlobalOperationsClient
}

func (c *computeGlobalOperationsClient) GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *computeGlobalOperationsClient) ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, opts ...gax.CallOption) Iterator[*computepb.Operation] {
	return c.inner.List(ctx, req, opts...)
}
