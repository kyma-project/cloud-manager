package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type GlobalOperationsClient interface {
	GetGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error)
	Wait(ctx context.Context, req *computepb.WaitGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error)
}

var _ GlobalOperationsClient = (*globalOperationsClient)(nil)

type globalOperationsClient struct {
	inner *compute.GlobalOperationsClient
}

func (c *globalOperationsClient) GetGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *globalOperationsClient) Wait(ctx context.Context, req *computepb.WaitGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	return c.inner.Wait(ctx, req, opts...)
}
