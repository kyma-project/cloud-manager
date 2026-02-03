package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"k8s.io/utils/ptr"
)

type NetworkClient interface {
	ListNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error)
	ListExternalNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error)
	GetNetwork(ctx context.Context, id string) (*networks.Network, error)
	CreateNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error)
	CreateExternalNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error)
	DeleteNetwork(ctx context.Context, id string) error

	ListInternalNetworksByName(ctx context.Context, name string) ([]networks.Network, error)
	GetNetworkByName(ctx context.Context, name string) (*networks.Network, error)
}

func NewNetworkClient() SapClientProvider[NetworkClient] {
	return func(ctx context.Context, pp ProviderParams) (NetworkClient, error) {
		pi, err := NewProviderClient(ctx, pp)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sap provider client: %v", err)
		}

		netSvc, err := openstack.NewNetworkV2(pi.ProviderClient, pi.EndpointOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create network v2 client: %v", err)
		}

		return &networkClient{
			netSvc: netSvc,
		}, nil
	}
}

var _ NetworkClient = (*networkClient)(nil)

type networkClient struct {
	netSvc *gophercloud.ServiceClient
}

// low level methods matching the gophercloud api ================================

func (c *networkClient) ListNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error) {
	page, err := networks.List(c.netSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *networkClient) ListExternalNetworks(ctx context.Context, opts networks.ListOpts) ([]networks.Network, error) {
	extOpts := external.ListOptsExt{
		ListOptsBuilder: opts,
		External:        ptr.To(true),
	}
	page, err := networks.List(c.netSvc, extOpts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *networkClient) GetNetwork(ctx context.Context, id string) (*networks.Network, error) {
	network, err := networks.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (c *networkClient) CreateExternalNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error) {
	return c.CreateNetwork(ctx, external.CreateOptsExt{
		CreateOptsBuilder: opts,
		External:          ptr.To(true),
	})
}

func (c *networkClient) CreateNetwork(ctx context.Context, opts networks.CreateOptsBuilder) (*networks.Network, error) {
	network, err := networks.Create(ctx, c.netSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (c *networkClient) DeleteNetwork(ctx context.Context, id string) error {
	err := networks.Delete(ctx, c.netSvc, id).ExtractErr()
	return err
}

// high level derived methods ============================================================

func (c *networkClient) ListInternalNetworksByName(ctx context.Context, name string) ([]networks.Network, error) {
	return c.ListNetworks(ctx, networks.ListOpts{
		Name: name,
	})
}

func (c *networkClient) GetNetworkByName(ctx context.Context, name string) (*networks.Network, error) {
	arr, err := c.ListNetworks(ctx, networks.ListOpts{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	if len(arr) > 0 {
		return &arr[0], nil
	}
	return nil, nil
}
