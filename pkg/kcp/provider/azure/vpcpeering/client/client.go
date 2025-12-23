package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	CreateOrUpdatePeering(ctx context.Context,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		remoteVnetId string,
		allowVnetAccess bool,
		useRemoteGateway bool,
		allowGatewayTransit bool,
	) error

	ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error)
	GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error)
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
	DeletePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) error
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().WithAnyTenant().Build())

		if err != nil {
			return nil, err
		}

		clientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().WithAuxiliaryTenants(auxiliaryTenants).Build())

		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewVirtualNetworkPeeringClient(clientFactory.NewVirtualNetworkPeeringsClient()),
			azureclient.NewNetworkClient(clientFactory.NewVirtualNetworksClient()),
		), nil
	}
}

type client struct {
	azureclient.NetworkClient
	azureclient.VirtualNetworkPeeringClient
}

func newClient(peeringClient azureclient.VirtualNetworkPeeringClient, networkClient azureclient.NetworkClient) Client {
	return &client{
		NetworkClient:               networkClient,
		VirtualNetworkPeeringClient: peeringClient,
	}
}
