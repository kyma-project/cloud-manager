package client

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
)

type PortClient interface {
	ListPorts(ctx context.Context, opts ports.ListOpts) ([]ports.Port, error)
	GetPort(ctx context.Context, id string) (*ports.Port, error)
	CreatePort(ctx context.Context, opts ports.CreateOpts) (*ports.Port, error)
	DeletePort(ctx context.Context, id string) error

	ListRouterSubnetInterfaces(ctx context.Context, routerId string) ([]RouterSubnetInterfaceInfo, error)
}

type RouterSubnetInterfaceInfo struct {
	PortID    string
	IpAddress string
	SubnetID  string
}

var _ PortClient = (*portClient)(nil)

type portClient struct {
	netSvc *gophercloud.ServiceClient
}

func (c *portClient) ListPorts(ctx context.Context, opts ports.ListOpts) ([]ports.Port, error) {
	arr, err := ports.List(c.netSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	result, err := ports.ExtractPorts(arr)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *portClient) GetPort(ctx context.Context, id string) (*ports.Port, error) {
	port, err := ports.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return port, nil
}

func (c *portClient) CreatePort(ctx context.Context, opts ports.CreateOpts) (*ports.Port, error) {
	port, err := ports.Create(ctx, c.netSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return port, nil
}

func (c *portClient) DeletePort(ctx context.Context, id string) error {
	err := ports.Delete(ctx, c.netSvc, id).ExtractErr()
	if err != nil {
		return err
	}
	return nil
}

func (c *portClient) ListRouterSubnetInterfaces(ctx context.Context, routerId string) ([]RouterSubnetInterfaceInfo, error) {
	arrPorts, err := c.ListPorts(ctx, ports.ListOpts{
		DeviceID: routerId,
	})
	if err != nil {
		return nil, err
	}
	var result []RouterSubnetInterfaceInfo
	for _, port := range arrPorts {
		if port.DeviceOwner != "network:router_gateway" {
			for _, ipSpec := range port.FixedIPs {
				result = append(result, RouterSubnetInterfaceInfo{
					PortID:    port.ID,
					IpAddress: ipSpec.IPAddress,
					SubnetID:  ipSpec.SubnetID,
				})
			}
		}
	}
	return result, nil
}
