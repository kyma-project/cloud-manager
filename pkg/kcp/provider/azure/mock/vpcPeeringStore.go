package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	"k8s.io/utils/pointer"
	"sync"
)

type VpcPeeringConfig interface {
	SetSubscription(subscriptionId string)
}

type vpcPeeringEntry struct {
	subscription       string
	resourceGroupName  string
	virtualNetworkName string
	peering            armnetwork.VirtualNetworkPeering
}
type vpcPeeringStore struct {
	m              sync.Mutex
	items          []*vpcPeeringEntry
	subscriptionId string
}

func (s *vpcPeeringStore) SetSubscription(subscriptionId string) {
	s.subscriptionId = subscriptionId
}

func (s *vpcPeeringStore) BeginCreateOrUpdate(
	ctx context.Context,
	resourceGroupName,
	virtualNetworkName,
	virtualNetworkPeeringName,
	remoteVnetId string,
	allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/virtualNetworkPeerings/%s",
		s.subscriptionId,
		resourceGroupName,
		virtualNetworkName,
		virtualNetworkPeeringName)

	item := &vpcPeeringEntry{
		subscription:       s.subscriptionId,
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

	s.items = append(s.items, item)

	return &item.peering, nil
}

func (s *vpcPeeringStore) List(ctx context.Context, resourceGroupName string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {
	items := pie.Filter(s.items, func(e *vpcPeeringEntry) bool {
		return e.resourceGroupName == resourceGroupName && e.virtualNetworkName == virtualNetworkName
	})
	return pie.Map(items, func(e *vpcPeeringEntry) *armnetwork.VirtualNetworkPeering {
		return &e.peering
	}), nil
}

func (s *vpcPeeringStore) Get(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	for _, x := range s.items {
		if virtualNetworkPeeringName == pointer.StringDeref(x.peering.Name, "") &&
			resourceGroupName == x.resourceGroupName &&
			virtualNetworkName == x.virtualNetworkName {
			return &x.peering, nil
		}
	}
	return nil, nil
}
