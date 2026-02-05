package client

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	"k8s.io/utils/ptr"

	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

type Client interface {
	GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error)
	GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error)
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
	return func() Client {
		return &client{
			routersClient: gcpClients.RoutersWrapped(),
			addressClient: gcpClients.AddressesWrapped(),
		}
	}
}

var _ Client = (*client)(nil)

type client struct {
	addressClient gcpclient.AddressesClient
	routersClient gcpclient.RoutersClient
}

func (c *client) GetVpcRouters(ctx context.Context, project string, region string, vpcName string) ([]*computepb.Router, error) {
	it := c.routersClient.ListRouters(ctx, &computepb.ListRoutersRequest{
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

func (c *client) GetRouterIpAddresses(ctx context.Context, project string, region string, routerName string) ([]*computepb.Address, error) {
	it := c.addressClient.ListAddresses(ctx, &computepb.ListAddressesRequest{
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
