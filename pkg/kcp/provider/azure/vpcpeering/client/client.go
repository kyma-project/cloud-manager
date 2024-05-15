package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/pointer"
)

type Client interface {
	BeginCreateOrUpdate(ctx context.Context,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		remoteVnetId string,
		allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error)

	List(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error)
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

		return newClient(clientFactory.NewVirtualNetworkPeeringsClient()), nil
	}
}

type client struct {
	svc *armnetwork.VirtualNetworkPeeringsClient
}

func newClient(svc *armnetwork.VirtualNetworkPeeringsClient) Client {
	return &client{svc: svc}
}

func (c *client) BeginCreateOrUpdate(
	ctx context.Context,
	resourceGroupName,
	virtualNetworkName,
	virtualNetworkPeeringName,
	remoteVnetId string,
	allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error) {
	poller, err := c.svc.BeginCreateOrUpdate(
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

func (c *client) List(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {

	pager := c.svc.NewListPager(resourceGroupName, virtualNetworkName, nil)

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
