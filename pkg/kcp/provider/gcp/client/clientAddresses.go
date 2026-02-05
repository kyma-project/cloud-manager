package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type AddressesClient interface {
	ListAddresses(ctx context.Context, req *computepb.ListAddressesRequest, opts ...gax.CallOption) IteratorWithAll[*computepb.Address]
}

var _ AddressesClient = (*addressesClient)(nil)

type addressesClient struct {
	inner *compute.AddressesClient
}

func (c *addressesClient) ListAddresses(ctx context.Context, req *computepb.ListAddressesRequest, opts ...gax.CallOption) IteratorWithAll[*computepb.Address] {
	return c.inner.List(ctx, req, opts...)
}
