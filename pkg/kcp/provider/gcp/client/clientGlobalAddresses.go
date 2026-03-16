package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type GlobalAddressesClient interface {
	GetGlobalAddress(ctx context.Context, req *computepb.GetGlobalAddressRequest, opts ...gax.CallOption) (*computepb.Address, error)
	DeleteGlobalAddress(ctx context.Context, req *computepb.DeleteGlobalAddressRequest, opts ...gax.CallOption) (VoidOperation, error)
	InsertGlobalAddress(ctx context.Context, req *computepb.InsertGlobalAddressRequest, opts ...gax.CallOption) (VoidOperation, error)
	ListGlobalAddresses(ctx context.Context, req *computepb.ListGlobalAddressesRequest, opts ...gax.CallOption) Iterator[*computepb.Address]
}

var _ GlobalAddressesClient = (*globalAddressesClient)(nil)

type globalAddressesClient struct {
	inner compute.GlobalAddressesClient
}

func (c *globalAddressesClient) GetGlobalAddress(ctx context.Context, req *computepb.GetGlobalAddressRequest, opts ...gax.CallOption) (*computepb.Address, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *globalAddressesClient) DeleteGlobalAddress(ctx context.Context, req *computepb.DeleteGlobalAddressRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Delete(ctx, req, opts...)
}

func (c *globalAddressesClient) InsertGlobalAddress(ctx context.Context, req *computepb.InsertGlobalAddressRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Insert(ctx, req, opts...)
}

func (c *globalAddressesClient) ListGlobalAddresses(ctx context.Context, req *computepb.ListGlobalAddressesRequest, opts ...gax.CallOption) Iterator[*computepb.Address] {
	return c.inner.List(ctx, req, opts...)
}
