package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/pointer"
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
			ID:   pointer.String(id),
			Name: pointer.String(virtualNetworkPeeringName),
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
	}

	c.peeringStore.items = append(c.peeringStore.items, item)

	return &item.peering, nil
}

func (c *storeSubscriptionContext) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	for _, x := range c.networkStore.items {
		if virtualNetworkName == pointer.StringDeref(x.network.Name, "") &&
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
		if virtualNetworkPeeringName == pointer.StringDeref(x.peering.Name, "") &&
			resourceGroup == x.resourceGroupName &&
			virtualNetworkName == x.virtualNetworkName {
			return &x.peering, nil
		}
	}
	return nil, nil
}
