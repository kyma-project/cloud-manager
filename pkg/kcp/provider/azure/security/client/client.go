package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	SubscriptionId() string

	azureclient.ResourceGroupClient
	azureclient.SecurityClient
	azureclient.NetworkFlowLogsClient
	azureclient.OperationalInsightsClient
	azureclient.StorageAccountClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		clientOptions := azureclient.NewClientOptionsBuilder().Build()

		resourcesFactory, err := armresources.NewClientFactory(subscriptionId, cred, clientOptions)
		if err != nil {
			return nil, err
		}
		securityClientFactory, err := armsecurity.NewClientFactory(subscriptionId, cred, clientOptions)
		if err != nil {
			return nil, err
		}

		networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, clientOptions)
		if err != nil {
			return nil, err
		}

		operationalInsightsFactory, err := armoperationalinsights.NewClientFactory(subscriptionId, cred, clientOptions)

		storageFactory, err := armstorage.NewClientFactory(subscriptionId, cred, clientOptions)
		if err != nil {
			return nil, err
		}

		return newClient(
			subscriptionId,
			azureclient.NewResourceGroupClient(resourcesFactory.NewResourceGroupsClient()),
			azureclient.NewSecurityClient(securityClientFactory.NewPricingsClient()),
			azureclient.NewOperationalInsightsClient(operationalInsightsFactory.NewWorkspacesClient()),
			azureclient.NewNetworkFlowLogsClient(networkClientFactory.NewWatchersClient(), networkClientFactory.NewFlowLogsClient()),
			azureclient.NewStorageAccountClient(storageFactory.NewAccountsClient()),
		), nil
	}
}

func newClient(
	subscriptionID string,
	resourceGroupClient azureclient.ResourceGroupClient,
	securityClient azureclient.SecurityClient,
	operationalInsightsClient azureclient.OperationalInsightsClient,
	networkFlowLogsClient azureclient.NetworkFlowLogsClient,
	storageAccountsClient azureclient.StorageAccountClient,
) *client {
	return &client{
		subscriptionID:            subscriptionID,
		ResourceGroupClient:       resourceGroupClient,
		SecurityClient:            securityClient,
		OperationalInsightsClient: operationalInsightsClient,
		NetworkFlowLogsClient:     networkFlowLogsClient,
		StorageAccountClient:      storageAccountsClient,
	}
}

type client struct {
	azureclient.ResourceGroupClient
	azureclient.SecurityClient
	azureclient.OperationalInsightsClient
	azureclient.NetworkFlowLogsClient
	azureclient.StorageAccountClient

	subscriptionID string
}

func (c *client) SubscriptionId() string {
	return c.subscriptionID
}
