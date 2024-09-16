package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	CreatePeering(ctx context.Context,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		remoteVnetId string,
		allowVnetAccess bool) error

	ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error)
	GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error)
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
	DeletePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) error
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		clientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(clientFactory.NewVirtualNetworkPeeringsClient(), clientFactory.NewVirtualNetworksClient()), nil
	}
}

type client struct {
	peering *armnetwork.VirtualNetworkPeeringsClient
	network *armnetwork.VirtualNetworksClient
}

func newClient(peeringClient *armnetwork.VirtualNetworkPeeringsClient, networkClient *armnetwork.VirtualNetworksClient) Client {
	return &client{
		peering: peeringClient,
		network: networkClient,
	}
}

func (c *client) CreatePeering(
	ctx context.Context,
	resourceGroupName,
	virtualNetworkName,
	virtualNetworkPeeringName,
	remoteVnetId string,
	allowVnetAccess bool) error {
	_, err := c.peering.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		armnetwork.VirtualNetworkPeering{
			Properties: &armnetwork.VirtualNetworkPeeringPropertiesFormat{
				AllowForwardedTraffic:     ptr.To(true),
				AllowGatewayTransit:       ptr.To(false),
				AllowVirtualNetworkAccess: ptr.To(allowVnetAccess),
				UseRemoteGateways:         ptr.To(false),
				RemoteVirtualNetwork: &armnetwork.SubResource{
					ID: ptr.To(remoteVnetId),
				},
			},
		},
		nil,
	)

	return err
}

func (c *client) ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {

	pager := c.peering.NewListPager(resourceGroupName, virtualNetworkName, nil)

	var items []*armnetwork.VirtualNetworkPeering

	for pager.More() {
		page, err := pager.NextPage(ctx)

		// there are more pages but getting next page failed
		if err != nil {
			return nil, err
		}

		items = append(items, page.Value...)
	}

	return items, nil
}

func (c *client) GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {

	response, err := c.peering.Get(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, nil)

	if err != nil {
		return nil, err
	}

	return &response.VirtualNetworkPeering, nil
}

func (c *client) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {

	response, err := c.network.Get(ctx, resourceGroupName, virtualNetworkName, nil)

	if err != nil {
		return nil, err
	}

	return &response.VirtualNetwork, nil
}

func (c *client) DeletePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) error {
	_, err := c.peering.BeginDelete(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, nil)
	return err
}
