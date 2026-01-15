package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	SubscriptionID() string

	azureclient.ResourceGroupClient
	azureclient.NetworkClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		resourcesClientFactory, err := armresources.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		return newClient(
			subscriptionId,
			azureclient.NewResourceGroupClient(resourcesClientFactory.NewResourceGroupsClient()),
			azureclient.NewSecurityGroupsClient(networkClientFactory.NewSecurityGroupsClient()),
			azureclient.NewNetworkClient(networkClientFactory.NewVirtualNetworksClient()),
		), nil
	}
}

func newClient(
	subscriptionID string,
	resourceGroupClient azureclient.ResourceGroupClient,
	securityGroupsClient azureclient.SecurityGroupsClient,
	networkClient azureclient.NetworkClient,
) *client {
	return &client{
		ResourceGroupClient:  resourceGroupClient,
		SecurityGroupsClient: securityGroupsClient,
		NetworkClient:        networkClient,

		subscriptionID: subscriptionID,
	}
}

type client struct {
	azureclient.ResourceGroupClient
	azureclient.SecurityGroupsClient
	azureclient.NetworkClient

	subscriptionID string
}

func (c *client) SubscriptionID() string {
	return c.subscriptionID
}
