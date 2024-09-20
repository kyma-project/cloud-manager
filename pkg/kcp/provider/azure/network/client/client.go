package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	azureclient.ResourceGroupClient
	azureclient.NetworkClient
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

		return newClient(
			azureclient.NewResourceGroupClient(resourceGroupFactory.NewResourceGroupsClient()),
			azureclient.NewNetworkClient(networkClientFactory.NewVirtualNetworksClient()),
		), nil
	}
}

type client struct {
	azureclient.ResourceGroupClient
	azureclient.NetworkClient
}

func newClient(resourceGroupClient azureclient.ResourceGroupClient, networkClient azureclient.NetworkClient) *client {
	return &client{
		ResourceGroupClient: resourceGroupClient,
		NetworkClient:       networkClient,
	}
}
