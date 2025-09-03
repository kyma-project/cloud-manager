package client

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
	GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error)
}

var _ Client = &client{}

type client struct {
	netSvc *gophercloud.ServiceClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
		pi, err := sapclient.NewProviderClient(ctx, pp)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sap provider client: %v", err)
		}
		netSvc, err := openstack.NewNetworkV2(pi.ProviderClient, pi.EndpointOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create network v2 client: %v", err)
		}
		return &client{
			netSvc:   netSvc,
		}, nil
	}
}

func (c *client) GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error) {
	pg, err := routers.List(c.netSvc, routers.ListOpts{
		Name: routerName,
	}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing routers: %w", err)
	}
	allRouters, err := routers.ExtractRouters(pg)
	if err != nil {
		return nil, fmt.Errorf("error extracting routers: %w", err)
	}

	if len(allRouters) == 0 {
		return nil, nil
	}

	return &allRouters[0], nil
}
