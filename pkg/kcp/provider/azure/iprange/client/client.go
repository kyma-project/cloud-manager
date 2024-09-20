package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	azureclient.SubnetsClient
	azureclient.SecurityGroupsClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			return nil, err
		}

		resourcesClientFactory, err := armresources.NewClientFactory(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewResourceGroupClient(resourcesClientFactory.NewResourceGroupsClient()),
			azureclient.NewSecurityGroupsClient(networkClientFactory.NewSecurityGroupsClient()),
			azureclient.NewSubnetsClient(networkClientFactory.NewSubnetsClient()),
		), nil
	}
}

var _ Client = &client{}

type client struct {
	azureclient.ResourceGroupClient
	azureclient.SecurityGroupsClient
	azureclient.SubnetsClient
}

func newClient(
	resourceGroupsClient azureclient.ResourceGroupClient,
	securityGroupsClient azureclient.SecurityGroupsClient,
	subnetsClient azureclient.SubnetsClient,
) *client {
	return &client{
		ResourceGroupClient:  resourceGroupsClient,
		SecurityGroupsClient: securityGroupsClient,
		SubnetsClient:        subnetsClient,
	}
}
