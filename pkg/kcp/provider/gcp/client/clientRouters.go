package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type RoutersClient interface {
	ListRouters(ctx context.Context, req *computepb.ListRoutersRequest, opts ...gax.CallOption) IteratorWithAll[*computepb.Router]
}

var _ RoutersClient = (*routersClient)(nil)

type routersClient struct {
	inner *compute.RoutersClient
}

func (c *routersClient) ListRouters(ctx context.Context, req *computepb.ListRoutersRequest, opts ...gax.CallOption) IteratorWithAll[*computepb.Router] {
	return c.inner.List(ctx, req, opts...)
}
