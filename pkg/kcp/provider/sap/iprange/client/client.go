package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
	GetNetworkByName(ctx context.Context, networkName string) (*networks.Network, error)
	GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error)
	CreateSubnet(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error)
	DeleteSubnet(ctx context.Context, subnetId string) error
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
			netSvc: netSvc,
		}, nil
	}
}

type client struct {
	netSvc *gophercloud.ServiceClient
}

func (c *client) GetNetworkByName(ctx context.Context, networkName string) (*networks.Network, error) {
	page, err := networks.List(c.netSvc, networks.ListOpts{
		Name: networkName,
	}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list openstack networks: %v", err)
	}
	arr, err := networks.ExtractNetworks(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract networks: %v", err)
	}

	if len(arr) > 0 {
		return &arr[0], nil
	}

	return nil, nil
}

func (c *client) GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error) {
	page, err := subnets.List(c.netSvc, subnets.ListOpts{
		NetworkID: networkId,
		Name:      subnetName,
	}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list openstack subnets: %v", err)
	}
	arr, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract subnets: %v", err)
	}

	if len(arr) > 0 {
		return &arr[0], nil
	}

	return nil, nil
}

func (c *client) CreateSubnet(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error) {
	subnet, err := subnets.Create(ctx, c.netSvc, subnets.CreateOpts{
		NetworkID: networkId,
		CIDR:      cidr,
		Name:      subnetName,
		IPVersion: 4,
	}).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %v", err)
	}

	return subnet, nil
}

func (c *client) DeleteSubnet(ctx context.Context, subnetId string) error {
	err := subnets.Delete(ctx, c.netSvc, subnetId).ExtractErr()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting subnet: %w", err)
	}
	return nil
}
