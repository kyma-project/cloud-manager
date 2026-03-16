package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type VpcNetworkClient interface {
	GetNetwork(ctx context.Context, req *computepb.GetNetworkRequest, opts ...gax.CallOption) (*computepb.Network, error)
	InsertNetwork(ctx context.Context, req *computepb.InsertNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
	ListNetworks(ctx context.Context, req *computepb.ListNetworksRequest, opts ...gax.CallOption) Iterator[*computepb.Network]
	DeleteNetwork(ctx context.Context, req *computepb.DeleteNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)

	AddPeering(ctx context.Context, req *computepb.AddPeeringNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
	RemovePeering(ctx context.Context, req *computepb.RemovePeeringNetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
}

var _ VpcNetworkClient = (*networkClient)(nil)

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

func (c *networkClient) AddPeering(ctx context.Context, req *computepb.AddPeeringNetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.AddPeering(ctx, req, opts...)
}

func (c *networkClient) RemovePeering(ctx context.Context, req *computepb.RemovePeeringNetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.RemovePeering(ctx, req, opts...)
}
