package client

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"k8s.io/utils/ptr"
)

type RoutersClient interface {
	ListRouters(ctx context.Context, req *computepb.ListRoutersRequest, opts ...gax.CallOption) Iterator[*computepb.Router]
	GetRouter(ctx context.Context, req *computepb.GetRouterRequest, opts ...gax.CallOption) (*computepb.Router, error)
	InsertRouter(ctx context.Context, req *computepb.InsertRouterRequest, opts ...gax.CallOption) (VoidOperation, error)
	DeleteRouter(ctx context.Context, req *computepb.DeleteRouterRequest, opts ...gax.CallOption) (VoidOperation, error)

	// Higher level methods

	GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error)
}

var _ RoutersClient = (*routersClient)(nil)

type routersClient struct {
	inner *compute.RoutersClient
}

func (c *routersClient) InsertRouter(ctx context.Context, req *computepb.InsertRouterRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Insert(ctx, req, opts...)
}

func (c *routersClient) GetRouter(ctx context.Context, req *computepb.GetRouterRequest, opts ...gax.CallOption) (*computepb.Router, error) {
	return c.inner.Get(ctx, req, opts...)
}

func (c *routersClient) ListRouters(ctx context.Context, req *computepb.ListRoutersRequest, opts ...gax.CallOption) Iterator[*computepb.Router] {
	return c.inner.List(ctx, req, opts...)
}

func (c *routersClient) DeleteRouter(ctx context.Context, req *computepb.DeleteRouterRequest, opts ...gax.CallOption) (VoidOperation, error) {
	return c.inner.Delete(ctx, req, opts...)
}

// Higher level methods =======================================================================

func (c *routersClient) GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error) {
	it := c.ListRouters(ctx, &computepb.ListRoutersRequest{
		Filter:  ptr.To(fmt.Sprintf(`network eq .*%s`, vpcName)),
		Project: project,
		Region:  region,
	}).All()
	var results []*computepb.Router
	for r, err := range it {
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
