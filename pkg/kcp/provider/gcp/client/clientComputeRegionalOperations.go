package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type ComputeRegionalOperationsClient interface {
	GetComputeRegionalOperation(ctx context.Context, req *computepb.GetRegionOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error)
	ListComputeRegionalOperations(ctx context.Context, req *computepb.ListRegionOperationsRequest, opts ...gax.CallOption) Iterator[*computepb.Operation]
}

var _ ComputeRegionalOperationsClient = (*computeRegionalOperationsClient)(nil)

type computeRegionalOperationsClient struct {
	inner *compute.RegionOperationsClient
}

func (c *computeRegionalOperationsClient) GetComputeRegionalOperation(ctx context.Context, req *computepb.GetRegionOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *computeRegionalOperationsClient) ListComputeRegionalOperations(ctx context.Context, req *computepb.ListRegionOperationsRequest, opts ...gax.CallOption) Iterator[*computepb.Operation] {
	return c.inner.List(ctx, req, opts...)
}
