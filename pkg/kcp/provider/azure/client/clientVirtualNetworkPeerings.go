package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"k8s.io/utils/ptr"
)

type VirtualNetworkPeeringClient interface {
	CreateOrUpdatePeering(ctx context.Context,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		remoteVnetId string,
		allowVnetAccess bool,
		useVpcGateway bool,
		allowGatewayTransit bool) error

	ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error)
	GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error)
	DeletePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) error
}

func NewVirtualNetworkPeeringClient(svc *armnetwork.VirtualNetworkPeeringsClient) VirtualNetworkPeeringClient {
	return &virtualNetworkPeeringClient{svc: svc}
}

var _ VirtualNetworkPeeringClient = &virtualNetworkPeeringClient{}

type virtualNetworkPeeringClient struct {
	svc *armnetwork.VirtualNetworkPeeringsClient
}

func (c *virtualNetworkPeeringClient) CreateOrUpdatePeering(
	ctx context.Context,
	resourceGroupName,
	virtualNetworkName,
	virtualNetworkPeeringName,
	remoteVnetId string,
	allowVnetAccess bool,
	useRemoteGateway bool,
	allowGatewayTransit bool,
) error {
	_, err := c.svc.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName,
		armnetwork.VirtualNetworkPeering{
			Properties: &armnetwork.VirtualNetworkPeeringPropertiesFormat{
				AllowForwardedTraffic:     ptr.To(true),
				AllowGatewayTransit:       ptr.To(allowGatewayTransit),
				AllowVirtualNetworkAccess: ptr.To(allowVnetAccess),
				UseRemoteGateways:         ptr.To(useRemoteGateway),
				RemoteVirtualNetwork: &armnetwork.SubResource{
					ID: ptr.To(remoteVnetId),
				},
			},
		},
		&armnetwork.VirtualNetworkPeeringsClientBeginCreateOrUpdateOptions{
			SyncRemoteAddressSpace: to.Ptr(armnetwork.SyncRemoteAddressSpaceTrue),
		},
	)

	return err
}

func (c *virtualNetworkPeeringClient) ListPeerings(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {
	pager := c.svc.NewListPager(resourceGroupName, virtualNetworkName, nil)

	var items []*armnetwork.VirtualNetworkPeering

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		items = append(items, page.Value...)
	}

	return items, nil
}

func (c *virtualNetworkPeeringClient) GetPeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	response, err := c.svc.Get(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, nil)
	if err != nil {
		return nil, err
	}

	return &response.VirtualNetworkPeering, nil
}

func (c *virtualNetworkPeeringClient) DeletePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, nil)
	return err
}
