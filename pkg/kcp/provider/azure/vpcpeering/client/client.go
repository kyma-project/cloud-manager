package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/pointer"
)

type Client interface {
	CreatePeering(ctx context.Context,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		remoteVnetId string,
		allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error)

	ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error)
	GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error)
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
}

func NewClientProvider() azureclient.SkrClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		clientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		clientFactory.NewVirtualNetworksClient()
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
	allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error) {
	poller, err := c.peering.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		armnetwork.VirtualNetworkPeering{
			Properties: &armnetwork.VirtualNetworkPeeringPropertiesFormat{
				AllowForwardedTraffic:     pointer.Bool(true),
				AllowGatewayTransit:       pointer.Bool(false),
				AllowVirtualNetworkAccess: pointer.Bool(allowVnetAccess),
				UseRemoteGateways:         pointer.Bool(false),
				RemoteVirtualNetwork: &armnetwork.SubResource{
					ID: pointer.String(remoteVnetId),
				},
			},
		},
		nil,
	)

	if err != nil {
		return nil, err
	}

	res, err := poller.PollUntilDone(ctx, nil)

	if err != nil {
		return nil, err
	}

	return &res.VirtualNetworkPeering, nil
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

		for _, v := range page.Value {
			items = append(items, v)
		}
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

	// options := armnetwork.VirtualNetworksClientGetOptions{Expand: nil}
	response, err := c.network.Get(ctx, resourceGroupName, virtualNetworkName, nil)

	if err != nil {
		return nil, err
	}

	return &response.VirtualNetwork, nil
}
