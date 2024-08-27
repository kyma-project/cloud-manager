package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
)

type storeSubscriptionContext struct {
	peeringStore *peeringStore
	networkStore *networkStore
	subscription string
}

func (c *storeSubscriptionContext) CreatePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, remoteVnetId string, allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	c.peeringStore.m.Lock()
	defer c.peeringStore.m.Unlock()

	id := util.VirtualNetworkPeeringResourceId(c.subscription, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	item := &peeringEntry{
		resourceGroupName:  resourceGroupName,
		virtualNetworkName: virtualNetworkName,
		peering: armnetwork.VirtualNetworkPeering{
			ID:   ptr.To(id),
			Name: ptr.To(virtualNetworkPeeringName),
			Properties: &armnetwork.VirtualNetworkPeeringPropertiesFormat{
				AllowForwardedTraffic:     ptr.To(true),
				AllowGatewayTransit:       ptr.To(false),
				AllowVirtualNetworkAccess: ptr.To(allowVnetAccess),
				UseRemoteGateways:         ptr.To(false),
				RemoteVirtualNetwork: &armnetwork.SubResource{
					ID: ptr.To(remoteVnetId),
				},
				PeeringState: ptr.To(armnetwork.VirtualNetworkPeeringStateInitiated),
			},
		},
	}

	c.peeringStore.items = append(c.peeringStore.items, item)

	return &item.peering, nil
}

func (c *storeSubscriptionContext) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	for _, x := range c.networkStore.items {
		if virtualNetworkName == ptr.Deref(x.network.Name, "") &&
			resourceGroupName == x.resourceGroup {
			return &x.network, nil
		}
	}
	return nil, nil
}

func (c *storeSubscriptionContext) ListPeerings(ctx context.Context, resourceGroup string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {
	items := pie.Filter(c.peeringStore.items, func(e *peeringEntry) bool {
		return e.resourceGroupName == resourceGroup && e.virtualNetworkName == virtualNetworkName
	})
	return pie.Map(items, func(e *peeringEntry) *armnetwork.VirtualNetworkPeering {
		return &e.peering
	}), nil
}

func (c *storeSubscriptionContext) GetPeering(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	for _, x := range c.peeringStore.items {
		if virtualNetworkPeeringName == ptr.Deref(x.peering.Name, "") &&
			resourceGroup == x.resourceGroupName &&
			virtualNetworkName == x.virtualNetworkName {
			return &x.peering, nil
		}
	}
	return nil, nil
}

func (c *storeSubscriptionContext) DeletePeering(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error {
	c.peeringStore.items = pie.Filter(c.peeringStore.items, func(x *peeringEntry) bool {
		return !(x.resourceGroupName == resourceGroup &&
			x.virtualNetworkName == virtualNetworkName &&
			virtualNetworkPeeringName == ptr.Deref(x.peering.Name, ""))
	})

	return nil
}
