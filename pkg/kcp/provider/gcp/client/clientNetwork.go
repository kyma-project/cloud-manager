package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type NetworkClient interface {
	GetNetwork(ctx context.Context, req *computepb.GetNetworkRequest, opts ...gax.CallOption) (*computepb.Network, error)
	InsertNetwork(ctx context.Context, req *computepb.InsertNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
	ListNetworks(ctx context.Context, req *computepb.ListNetworksRequest, opts ...gax.CallOption) Iterator[*computepb.Network]
	DeleteNetwork(ctx context.Context, req *computepb.DeleteNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
}

var _ NetworkClient = (*networkClient)(nil)

type networkClient struct {
	inner *compute.NetworksClient
}

func (c *networkClient) GetNetwork(ctx context.Context, req *computepb.GetNetworkRequest, opts ...gax.CallOption) (*computepb.Network, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *networkClient) InsertNetwork(ctx context.Context, req *computepb.InsertNetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Insert(ctx, req, opts...)
}

func (c *networkClient) ListNetworks(ctx context.Context, req *computepb.ListNetworksRequest, opts ...gax.CallOption) Iterator[*computepb.Network] {
	return c.inner.List(ctx, req, opts...)
}

func (c *networkClient) DeleteNetwork(ctx context.Context, req *computepb.DeleteNetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Delete(ctx, req, opts...)
}
