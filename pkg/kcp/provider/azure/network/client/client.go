package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type ResourceGroupClient interface {
	GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error)
	CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error)
	DeleteResourceGroup(ctx context.Context, name string) error
}

type NetworkClient interface {
	CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
	DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) error
}

type Client interface {
	ResourceGroupClient
	NetworkClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			return nil, err
		}

		networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		resourceGroupFactory, err := armresources.NewClientFactory(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		return newClient(resourceGroupFactory.NewResourceGroupsClient(), networkClientFactory.NewVirtualNetworksClient()), nil
	}
}

type client struct {
	resourceGroup *armresources.ResourceGroupsClient
	network       *armnetwork.VirtualNetworksClient
}

func newClient(resourceGroup *armresources.ResourceGroupsClient, networkClient *armnetwork.VirtualNetworksClient) *client {
	return &client{
		resourceGroup: resourceGroup,
		network:       networkClient,
	}
}

func (c *client) CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error {
	var azureTags map[string]*string
	if tags != nil {
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	_, err := c.network.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, armnetwork.VirtualNetwork{
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

func (c *client) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	resp, err := c.network.Get(ctx, resourceGroupName, virtualNetworkName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func (c *client) DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) error {
	_, err := c.network.BeginDelete(ctx, resourceGroupName, virtualNetworkName, nil)
	return err
}

func (c *client) GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error) {
	resp, err := c.resourceGroup.Get(ctx, name, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroup, nil
}

func (c *client) CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error) {
	var azureTags map[string]*string
	if tags != nil {
		azureTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			azureTags[k] = ptr.To(v)
		}
	}
	resp, err := c.resourceGroup.CreateOrUpdate(ctx, name, armresources.ResourceGroup{
		Location: ptr.To(location),
		Tags:     azureTags,
	}, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ResourceGroup, nil
}

func (c *client) DeleteResourceGroup(ctx context.Context, name string) error {
	_, err := c.resourceGroup.BeginDelete(ctx, name, nil)
	return err
}
