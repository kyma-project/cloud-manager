package client

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
)

type SubnetClient interface {
	ListSubnets(ctx context.Context, opts subnets.ListOpts) ([]subnets.Subnet, error)
	GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error)
	CreateSubnet(ctx context.Context, opts subnets.CreateOpts) (*subnets.Subnet, error)
	DeleteSubnet(ctx context.Context, id string) error

	ListSubnetsByNetworkId(ctx context.Context, networkId string) ([]subnets.Subnet, error)
	GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error)
	CreateSubnetOp(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error)
}

var _ SubnetClient = (*subnetClient)(nil)

type subnetClient struct {
	netSvc *gophercloud.ServiceClient
}

// low level methods matching the gophercloud api ================================

func (c *subnetClient) ListSubnets(ctx context.Context, opts subnets.ListOpts) ([]subnets.Subnet, error) {
	page, err := subnets.List(c.netSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := subnets.ExtractSubnets(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *subnetClient) GetSubnet(ctx context.Context, id string) (*subnets.Subnet, error) {
	subnet, err := subnets.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (c *subnetClient) CreateSubnet(ctx context.Context, opts subnets.CreateOpts) (*subnets.Subnet, error) {
	subnet, err := subnets.Create(ctx, c.netSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return subnet, nil
}

func (c *subnetClient) DeleteSubnet(ctx context.Context, id string) error {
	err := subnets.Delete(ctx, c.netSvc, id).ExtractErr()
	return err
}

// high level derived methods ============================================================

func (c *subnetClient) GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error) {
	arr, err := c.ListSubnets(ctx, subnets.ListOpts{
		NetworkID: networkId,
		Name:      subnetName,
	})
	if err != nil {
		return nil, err
	}
	if len(arr) > 0 {
		return &arr[0], nil
	}
	return nil, nil
}

func (c *subnetClient) ListSubnetsByNetworkId(ctx context.Context, networkId string) ([]subnets.Subnet, error) {
	return c.ListSubnets(ctx, subnets.ListOpts{
		NetworkID: networkId,
	})
}

func (c *subnetClient) CreateSubnetOp(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error) {
	return c.CreateSubnet(ctx, subnets.CreateOpts{
		NetworkID: networkId,
		CIDR:      cidr,
		Name:      subnetName,
		IPVersion: 4,
	})
}
