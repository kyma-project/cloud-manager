package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error)
	CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName, vnetId string) error
	DeleteVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) error
	GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) (*armprivatedns.PrivateZone, error)
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{
			AdditionallyAllowedTenants: []string{"*"},
		})

		if err != nil {
			return nil, err
		}

		clientFactory, err := armprivatedns.NewClientFactory(subscriptionId, cred, &arm.ClientOptions{
			AuxiliaryTenants: auxiliaryTenants,
		})

		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewVirtualNetworkLinkClient(
				clientFactory.NewVirtualNetworkLinksClient(),
			),
			azureclient.NewPrivateDnsZoneClient(
				clientFactory.NewPrivateZonesClient(),
			),
		), nil
	}
}

type client struct {
	azureclient.VirtualNetworkLinkClient
	azureclient.PrivateDnsZoneClient
}

func newClient(
	virtualNetworkLinkClient azureclient.VirtualNetworkLinkClient,
	privateDnzZoneClient azureclient.PrivateDnsZoneClient) Client {
	return &client{
		VirtualNetworkLinkClient: virtualNetworkLinkClient,
		PrivateDnsZoneClient:     privateDnzZoneClient,
	}
}
