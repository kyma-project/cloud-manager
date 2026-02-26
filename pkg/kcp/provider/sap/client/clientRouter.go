package client

import (
	"context"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
)

type RouterClient interface {
	ListRouters(ctx context.Context, opts routers.ListOpts) ([]routers.Router, error)
	GetRouter(ctx context.Context, id string) (*routers.Router, error)
	CreateRouter(ctx context.Context, opts routers.CreateOpts) (*routers.Router, error)
	UpdateRouter(ctx context.Context, routerId string, opts routers.UpdateOpts) (*routers.Router, error)
	DeleteRouter(ctx context.Context, id string) error
	AddRouterInterface(ctx context.Context, routerId string, opts routers.AddInterfaceOpts) (*routers.InterfaceInfo, error)
	RemoveRouterInterface(ctx context.Context, routerId string, opts routers.RemoveInterfaceOpts) (*routers.InterfaceInfo, error)

	GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error)
	AddSubnetToRouter(ctx context.Context, routerId string, subnetId string) (*routers.InterfaceInfo, error)
	RemoveSubnetFromRouter(ctx context.Context, routerId string, subnetId string) error
}

var _ RouterClient = (*routerClient)(nil)

type routerClient struct {
	netSvc *gophercloud.ServiceClient
}

// low level methods matching the gophercloud api ================================

func (c *routerClient) ListRouters(ctx context.Context, opts routers.ListOpts) ([]routers.Router, error) {
	page, err := routers.List(c.netSvc, opts).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	arr, err := routers.ExtractRouters(page)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (c *routerClient) GetRouter(ctx context.Context, id string) (*routers.Router, error) {
	router, err := routers.Get(ctx, c.netSvc, id).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return router, nil
}

func (c *routerClient) CreateRouter(ctx context.Context, opts routers.CreateOpts) (*routers.Router, error) {
	router, err := routers.Create(ctx, c.netSvc, opts).Extract()
	if err != nil {
		return nil, err
	}
	return router, nil
}

func (c *routerClient) UpdateRouter(ctx context.Context, routerId string, opts routers.UpdateOpts) (*routers.Router, error) {
	router, err := routers.Update(ctx, c.netSvc, routerId, opts).Extract()
	if err != nil {
		return nil, err
	}
	return router, nil
}

func (c *routerClient) DeleteRouter(ctx context.Context, id string) error {
	err := routers.Delete(ctx, c.netSvc, id).ExtractErr()
	return err
}

func (c *routerClient) AddRouterInterface(ctx context.Context, routerId string, opts routers.AddInterfaceOpts) (*routers.InterfaceInfo, error) {
	ii, err := routers.AddInterface(ctx, c.netSvc, routerId, opts).Extract()
	if err != nil {
		return nil, err
	}
	return ii, nil
}

func (c *routerClient) RemoveRouterInterface(ctx context.Context, routerId string, opts routers.RemoveInterfaceOpts) (*routers.InterfaceInfo, error) {
	ii, err := routers.RemoveInterface(ctx, c.netSvc, routerId, opts).Extract()
	if err != nil {
		return nil, err
	}
	return ii, nil
}

// high level derived methods ============================================================

func (c *routerClient) GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error) {
	arr, err := c.ListRouters(ctx, routers.ListOpts{Name: routerName})
	if err != nil {
		return nil, err
	}
	if len(arr) > 0 {
		return &arr[0], nil
	}
	return nil, nil
}

func (c *routerClient) AddSubnetToRouter(ctx context.Context, routerId string, subnetId string) (*routers.InterfaceInfo, error) {
	return c.AddRouterInterface(ctx, routerId, routers.AddInterfaceOpts{
		SubnetID: subnetId,
	})
}

func (c *routerClient) RemoveSubnetFromRouter(ctx context.Context, routerId string, subnetId string) error {
	_, err := c.RemoveRouterInterface(ctx, routerId, routers.RemoveInterfaceOpts{
		SubnetID: subnetId,
	})
	return err
}
