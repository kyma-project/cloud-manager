package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type SubnetClient interface {
	InsertSubnet(ctx context.Context, req *computepb.InsertSubnetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
	GetSubnet(ctx context.Context, req *computepb.GetSubnetworkRequest, opts ...gax.CallOption) (*computepb.Subnetwork, error)
	DeleteSubnet(ctx context.Context, req *computepb.DeleteSubnetworkRequest, opts ...gax.CallOption) (VoidOperation, error)
}

var _ SubnetClient = (*subnetClient)(nil)

type subnetClient struct {
	inner *compute.SubnetworksClient
}

func (c *subnetClient) InsertSubnet(ctx context.Context, req *computepb.InsertSubnetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Insert(ctx, req, opts...)
}

func (c *subnetClient) GetSubnet(ctx context.Context, req *computepb.GetSubnetworkRequest, opts ...gax.CallOption) (*computepb.Subnetwork, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *subnetClient) DeleteSubnet(ctx context.Context, req *computepb.DeleteSubnetworkRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Delete(ctx, req, opts...)
}
