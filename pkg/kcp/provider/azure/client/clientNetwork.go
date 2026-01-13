package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type NetworkClient interface {
	CreateNetworkOld(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error
	CreateNetwork(ctx context.Context, resourceGroupName string, virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error)

	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)

	DeleteNetworkOld(ctx context.Context, resourceGroupName, virtualNetworkName string) error
	DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (Poller[armnetwork.VirtualNetworksClientDeleteResponse], error)
}

func NewNetworkClient(svc *armnetwork.VirtualNetworksClient) NetworkClient {
	return &networkClient{svc: svc}
}

// helper functions ===================================================================

func NewVirtualNetwork(location, addressSpace string, tags map[string]string) armnetwork.VirtualNetwork {
	var netTags map[string]*string
	if tags != nil {
		netTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			netTags[k] = ptr.To(v)
		}
	}

	return armnetwork.VirtualNetwork{
		Location: ptr.To(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{ptr.To(addressSpace)},
			},
		},
		Tags: netTags,
	}
}

// networkClient impl ===================================================================

var _ NetworkClient = &networkClient{}

type networkClient struct {
	svc *armnetwork.VirtualNetworksClient
}

func (c *networkClient) CreateNetwork(ctx context.Context, resourceGroupName string, virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	return c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, parameters, options)
}

func (c *networkClient) CreateNetworkOld(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error {
	poller, err := c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, NewVirtualNetwork(location, addressSpace, tags), nil)
	if err != nil {
		return err
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *networkClient) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	resp, err := c.svc.Get(ctx, resourceGroupName, virtualNetworkName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func (c *networkClient) DeleteNetworkOld(ctx context.Context, resourceGroupName, virtualNetworkName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, nil)
	return err
}

func (c *networkClient) DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	return c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, options)
}
