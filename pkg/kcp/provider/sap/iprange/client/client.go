package client

import (
	"context"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
	sapclient.NetworkClient
	sapclient.SubnetClient
	sapclient.RouterClient
	sapclient.PortClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
		f := sapclient.NewClientFactory(pp)
		nc, err := f.NetworkClient(ctx)
		if err != nil {
			return nil, err
		}
		sc, err := f.SubnetClient(ctx)
		if err != nil {
			return nil, err
		}
		rc, err := f.RouterClient(ctx)
		if err != nil {
			return nil, err
		}
		pc, err := f.PortClient(ctx)
		if err != nil {
			return nil, err
		}
		return &client{
			NetworkClient: nc,
			SubnetClient:  sc,
			RouterClient:  rc,
			PortClient:    pc,
		}, nil
	}
}

type client struct {
	sapclient.NetworkClient
	sapclient.SubnetClient
	sapclient.RouterClient
	sapclient.PortClient
}

//
//func (c *client) GetNetworkByName(ctx context.Context, networkName string) (*networks.Network, error) {
//	page, err := networks.List(c.netSvc, networks.ListOpts{
//		Name: networkName,
//	}).AllPages(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to list openstack networks: %v", err)
//	}
//	arr, err := networks.ExtractNetworks(page)
//	if err != nil {
//		return nil, fmt.Errorf("failed to extract networks: %v", err)
//	}
//
//	if len(arr) > 0 {
//		return &arr[0], nil
//	}
//
//	return nil, nil
//}
//
//func (c *client) GetSubnetByName(ctx context.Context, networkId string, subnetName string) (*subnets.Subnet, error) {
//	page, err := subnets.List(c.netSvc, subnets.ListOpts{
//		NetworkID: networkId,
//		Name:      subnetName,
//	}).AllPages(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to list openstack subnets: %v", err)
//	}
//	arr, err := subnets.ExtractSubnets(page)
//	if err != nil {
//		return nil, fmt.Errorf("failed to extract subnets: %v", err)
//	}
//
//	if len(arr) > 0 {
//		return &arr[0], nil
//	}
//
//	return nil, nil
//}
//
//func (c *client) CreateSubnetOp(ctx context.Context, networkId, cidr, subnetName string) (*subnets.Subnet, error) {
//	subnet, err := subnets.Create(ctx, c.netSvc, subnets.CreateOpts{
//		NetworkID: networkId,
//		CIDR:      cidr,
//		Name:      subnetName,
//		IPVersion: 4,
//	}).Extract()
//	if err != nil {
//		return nil, fmt.Errorf("failed to create subnet: %v", err)
//	}
//
//	return subnet, nil
//}
//
//func (c *client) DeleteSubnet(ctx context.Context, subnetId string) error {
//	err := subnets.Delete(ctx, c.netSvc, subnetId).ExtractErr()
//	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
//		return nil
//	}
//	if err != nil {
//		return fmt.Errorf("error deleting subnet: %w", err)
//	}
//	return nil
//}
//
//func (c *client) GetRouterByName(ctx context.Context, routerName string) (*routers.Router, error) {
//	page, err := routers.List(c.netSvc, routers.ListOpts{
//		Name: routerName,
//	}).AllPages(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to list openstack routers: %v", err)
//	}
//	arr, err := routers.ExtractRouters(page)
//	if err != nil {
//		return nil, fmt.Errorf("failed to extract routers: %v", err)
//	}
//	if len(arr) > 0 {
//		return &arr[0], nil
//	}
//	return nil, nil
//}
//
//func (c *client) ListRouterSubnetInterfaces(ctx context.Context, routerId string) ([]sapclient.RouterSubnetInterfaceInfo, error) {
//	page, err := ports.List(c.netSvc, ports.ListOpts{
//		DeviceID: routerId,
//	}).AllPages(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to list openstack ports: %v", err)
//	}
//	arrPorts, err := ports.ExtractPorts(page)
//	if err != nil {
//		return nil, fmt.Errorf("failed to extract ports: %v", err)
//	}
//	var result []sapclient.RouterSubnetInterfaceInfo
//	for _, port := range arrPorts {
//		if port.DeviceOwner != "network:router_gateway" {
//			for _, ipSpec := range port.FixedIPs {
//				result = append(result, sapclient.RouterSubnetInterfaceInfo{
//					PortID:    port.ID,
//					IpAddress: ipSpec.IPAddress,
//					SubnetID:  ipSpec.SubnetID,
//				})
//			}
//		}
//	}
//	return result, nil
//}
//
//func (c *client) AddSubnetToRouter(ctx context.Context, routerId string, subnetId string) (*routers.InterfaceInfo, error) {
//	ii, err := routers.AddInterface(ctx, c.netSvc, routerId, routers.AddInterfaceOpts{
//		SubnetID: subnetId,
//	}).Extract()
//	if err != nil {
//		return nil, fmt.Errorf("error adding subnet to router: %v", err)
//	}
//	return ii, nil
//}
//
//func (c *client) RemoveSubnetFromRouter(ctx context.Context, routerId string, subnetId string) error {
//	_, err := routers.RemoveInterface(ctx, c.netSvc, routerId, routers.RemoveInterfaceOpts{
//		SubnetID: subnetId,
//	}).Extract()
//	if err != nil {
//		return fmt.Errorf("error removing subnet from router: %v", err)
//	}
//	return nil
//}
