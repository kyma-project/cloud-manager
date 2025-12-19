package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	azureclient.NetworkClient
	azureclient.SubnetsClient
	azureclient.NatGatewayClient
	azureclient.PublicIPAddressesClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, _ ...string) (Client, error) {
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptions().Build())
		if err != nil {
			return nil, err
		}

		networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptions().Build())
		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewNetworkClient(networkClientFactory.NewVirtualNetworksClient()),
			azureclient.NewSubnetsClient(networkClientFactory.NewSubnetsClient()),
			azureclient.NewNatGatewayClient(networkClientFactory.NewNatGatewaysClient()),
			azureclient.NewPublicIPAddressesClient(networkClientFactory.NewPublicIPAddressesClient()),
		), nil
	}
}

func newClient(
	networkClient azureclient.NetworkClient,
	subnetsClient azureclient.SubnetsClient,
	natGatewayClient azureclient.NatGatewayClient,
	publicIpAddressClient azureclient.PublicIPAddressesClient,
) *client {
	return &client{
		NetworkClient:           networkClient,
		NatGatewayClient:        natGatewayClient,
		SubnetsClient:           subnetsClient,
		PublicIPAddressesClient: publicIpAddressClient,
	}
}

type client struct {
	azureclient.NetworkClient
	azureclient.SubnetsClient
	azureclient.NatGatewayClient
	azureclient.PublicIPAddressesClient
}
