package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type NetworkClient interface {
	CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
	DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) error
}

func NewNetworkClient(svc *armnetwork.VirtualNetworksClient) NetworkClient {
	return &networkClient{svc: svc}
}

var _ NetworkClient = &networkClient{}

type networkClient struct {
	svc *armnetwork.VirtualNetworksClient
}

func (c *networkClient) CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error {
	var azureTags map[string]*string
	if tags != nil {
		azureTags := make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	_, err := c.svc.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, armnetwork.VirtualNetwork{
		Location: ptr.To(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{ptr.To(addressSpace)},
			},
		},
		Tags: azureTags,
	}, nil)
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

func (c *networkClient) DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, nil)
	return err
}
