package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ NetworkClient = &networkStore{}
var _ NetworkConfig = &networkStore{}
var _ VpcPeeringClient = &networkStore{}

type networkEntry struct {
	network *armnetwork.VirtualNetwork
}

func newNetworkStore(subscription string) *networkStore {
	return &networkStore{
		subscription: subscription,
		items:        map[string]map[string]*networkEntry{},
	}
}

type networkStore struct {
	m sync.Mutex

	subscription string

	// items are map resouceGroup => networkName => networkEntry
	items map[string]map[string]*networkEntry
}

// Config ===================================================

func (s *networkStore) SetPeeringStateConnected(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	peering, err := s.getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
	if err != nil {
		return err
	}

	if peering.Properties == nil {
		peering.Properties = &armnetwork.VirtualNetworkPeeringPropertiesFormat{
			PeeringState: ptr.To(armnetwork.VirtualNetworkPeeringStateConnected),
		}
	} else {
		peering.Properties.PeeringState = ptr.To(armnetwork.VirtualNetworkPeeringStateConnected)
	}

	return nil
}

// NetworkClient ===================================================

func (s *networkStore) CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) (*armnetwork.VirtualNetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*networkEntry{}
	}

	_, ok = s.items[resourceGroupName][virtualNetworkName]
	if ok {
		return nil, fmt.Errorf("virtual network %s/%s/%s already exists", s.subscription, resourceGroupName, virtualNetworkName)
	}

	netTags := make(map[string]*string, len(tags))
	for k, v := range tags {
		netTags[k] = ptr.To(v)
	}

	net := &armnetwork.VirtualNetwork{
		ID:       ptr.To(azureutil.VirtualNetworkResourceId(s.subscription, resourceGroupName, virtualNetworkName)),
		Location: ptr.To(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{ptr.To(addressSpace)},
			},
		},
		Tags: netTags,
	}
	s.items[resourceGroupName][virtualNetworkName] = &networkEntry{network: net}

	return util.JsonClone(net)
}

func (s *networkStore) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	return s.getNetworkNoLock(resourceGroupName, virtualNetworkName)
}

func (s *networkStore) getNetworkNoLock(resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	_, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	_, ok = s.items[resourceGroupName][virtualNetworkName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return util.JsonClone(s.items[resourceGroupName][virtualNetworkName].network)
}

// VpcPeeringClient ==============================================

func (s *networkStore) CreatePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, remoteVnetId string, allowVnetAccess bool) (*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	id := azureutil.VirtualNetworkPeeringResourceId(s.subscription, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	_, err := s.getPeeringNoLock(resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		// errors like network not found
		return nil, err
	}
	if err == nil {
		// peering already exists
		return nil, fmt.Errorf("vpc peering %s already exists", id)
	}

	net, err := s.getNetworkNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return nil, err
	}

	if net.Properties == nil {
		net.Properties = &armnetwork.VirtualNetworkPropertiesFormat{}
	}

	peering := &armnetwork.VirtualNetworkPeering{
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
	}

	net.Properties.VirtualNetworkPeerings = append(net.Properties.VirtualNetworkPeerings, peering)

	return util.JsonClone(peering)
}

func (s *networkStore) ListPeerings(ctx context.Context, resourceGroup string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	net, err := s.getNetworkNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return nil, err
	}
	if net.Properties == nil {
		return nil, nil
	}

	result := make([]*armnetwork.VirtualNetworkPeering, 0, len(net.Properties.VirtualNetworkPeerings))
	for _, originalPeering := range net.Properties.VirtualNetworkPeerings {
		copyPeering, err := util.JsonClone(originalPeering)
		if err != nil {
			return nil, err
		}
		result = append(result, copyPeering)
	}

	return result, nil
}

func (s *networkStore) GetPeering(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	return s.getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
}

func (s *networkStore) getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	net, err := s.getNetworkNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return nil, err
	}
	if net.Properties == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	for _, peering := range net.Properties.VirtualNetworkPeerings {
		if ptr.Deref(peering.Name, "") == virtualNetworkPeeringName {
			return util.JsonClone(peering)
		}
	}

	return nil, azuremeta.NewAzureNotFoundError()
}

func (s *networkStore) DeletePeering(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
	if err != nil {
		// errors like network or peering not found
		return err
	}
	net, err := s.getNetworkNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return err
	}

	net.Properties.VirtualNetworkPeerings = pie.Filter(net.Properties.VirtualNetworkPeerings, func(item *armnetwork.VirtualNetworkPeering) bool {
		return ptr.Deref(item.Name, "") != virtualNetworkPeeringName
	})

	return nil
}
