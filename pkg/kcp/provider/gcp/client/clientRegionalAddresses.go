package client

import (
	"context"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"k8s.io/utils/ptr"
)

type RegionalAddressesClient interface {
	ListAddresses(ctx context.Context, req *computepb.ListAddressesRequest, opts ...gax.CallOption) Iterator[*computepb.Address]

	// Higher level functions

	GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error)
}

var _ RegionalAddressesClient = (*regionalAddressesClient)(nil)

type regionalAddressesClient struct {
	inner *compute.AddressesClient
}

func (c *regionalAddressesClient) ListAddresses(ctx context.Context, req *computepb.ListAddressesRequest, opts ...gax.CallOption) Iterator[*computepb.Address] {
	return c.inner.List(ctx, req, opts...)
}

// Higher level functions =======================================================================

func (c *regionalAddressesClient) GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error) {
	it := c.ListAddresses(ctx, &computepb.ListAddressesRequest{
		Project: project,
		Region:  region,
		Filter:  ptr.To(`purpose="NAT_AUTO"`), // the API does not work with users filter, so have to do this
	}).All()
	var results []*computepb.Address
	for x, err := range it {
		if err != nil {
			return nil, err
		}
		for _, usr := range x.Users {
			if strings.HasSuffix(usr, "/"+routerName) {
				results = append(results, x)
				break
			}
		}
	}
	return results, nil
}
