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
var _ SubnetsClient = &networkStore{}
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

	errorMap map[string]error
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

func (s *networkStore) SetPeeringError(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string, err error) {
	if isContextCanceled(ctx) {
		return
	}
	s.m.Lock()
	defer s.m.Unlock()

	if s.errorMap == nil {
		s.errorMap = make(map[string]error)
	}

	resourceId := azureutil.NewVirtualNetworkPeeringResourceId(s.subscription, resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
	s.errorMap[resourceId.String()] = err
}

// NetworkClient ===================================================

func (s *networkStore) CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*networkEntry{}
	}

	_, ok = s.items[resourceGroupName][virtualNetworkName]
	if ok {
		return fmt.Errorf("virtual network %s/%s/%s already exists", s.subscription, resourceGroupName, virtualNetworkName)
	}

	var netTags map[string]*string
	if tags != nil {
		netTags = make(map[string]*string, len(tags))
		for k, v := range tags {
			netTags[k] = ptr.To(v)
		}
	}

	net := &armnetwork.VirtualNetwork{
		ID:       ptr.To(azureutil.NewVirtualNetworkResourceId(s.subscription, resourceGroupName, virtualNetworkName).String()),
		Name:     ptr.To(virtualNetworkName),
		Location: ptr.To(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{ptr.To(addressSpace)},
			},
		},
		Tags: netTags,
	}
	s.items[resourceGroupName][virtualNetworkName] = &networkEntry{network: net}

	return nil
}

func (s *networkStore) GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return nil, err
	}
	return util.JsonClone(entry.network)
}

func (s *networkStore) getNetworkEntryNoLock(resourceGroupName, virtualNetworkName string) (*networkEntry, error) {
	_, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	_, ok = s.items[resourceGroupName][virtualNetworkName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return s.items[resourceGroupName][virtualNetworkName], nil
}

func (s *networkStore) DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return err
	}

	delete(s.items[resourceGroupName], virtualNetworkName)

	return nil
}

// SubnetsClient =========

func (s *networkStore) GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	return s.getSubnetNoLock(resourceGroupName, virtualNetworkName, subnetName)
}

func (s *networkStore) getSubnetNoLock(resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.Subnet, error) {
	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return nil, err
	}

	if entry.network.Properties == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	for _, subnet := range entry.network.Properties.Subnets {
		if ptr.Deref(subnet.Name, "") == subnetName {
			return subnet, nil
		}
	}

	return nil, azuremeta.NewAzureNotFoundError()
}

func (s *networkStore) CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, addressPrefix, securityGroupId, natGatewayId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return fmt.Errorf("azure network does not exist: %w", err)
	}

	if entry.network.Properties == nil {
		entry.network.Properties = &armnetwork.VirtualNetworkPropertiesFormat{}
	}

	id := azureutil.NewSubnetResourceId(s.subscription, resourceGroupName, virtualNetworkName, subnetName)

	for _, subnet := range entry.network.Properties.Subnets {
		if ptr.Deref(subnet.Name, "") == subnetName {
			return fmt.Errorf("subnet %s already exists", id)
		}
	}

	subnet := &armnetwork.Subnet{
		ID:   ptr.To(id.String()),
		Name: ptr.To(subnetName),
		Properties: &armnetwork.SubnetPropertiesFormat{
			AddressPrefix:         ptr.To(addressPrefix),
			DefaultOutboundAccess: ptr.To(false),
			ProvisioningState:     ptr.To(armnetwork.ProvisioningStateSucceeded),
		},
	}
	if securityGroupId != "" {
		subnet.Properties.NetworkSecurityGroup = &armnetwork.SecurityGroup{
			ID: ptr.To(securityGroupId),
		}
	}
	if natGatewayId != "" {
		subnet.Properties.NatGateway = &armnetwork.SubResource{
			ID: ptr.To(natGatewayId),
		}
	}

	entry.network.Properties.Subnets = append(entry.network.Properties.Subnets, subnet)

	return nil
}

func (s *networkStore) DeleteSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return err
	}

	_, err = s.getSubnetNoLock(resourceGroupName, virtualNetworkName, subnetName)
	if err != nil {
		return err
	}

	entry.network.Properties.Subnets = pie.Filter(entry.network.Properties.Subnets, func(x *armnetwork.Subnet) bool {
		return ptr.Deref(x.Name, "") != subnetName
	})

	return nil
}

// VpcPeeringClient ==============================================

func (s *networkStore) CreatePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, remoteVnetId string, allowVnetAccess bool) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	err := s.getError(resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	if err != nil {
		return err
	}

	id := azureutil.NewVirtualNetworkPeeringResourceId(s.subscription, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName).String()

	_, err = s.getPeeringNoLock(resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		// errors like network not found
		return err
	}
	if err == nil {
		// peering already exists
		return fmt.Errorf("vpc peering %s already exists", id)
	}

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return err
	}

	if entry.network.Properties == nil {
		entry.network.Properties = &armnetwork.VirtualNetworkPropertiesFormat{}
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

	entry.network.Properties.VirtualNetworkPeerings = append(entry.network.Properties.VirtualNetworkPeerings, peering)

	return nil
}

func (s *networkStore) ListPeerings(ctx context.Context, resourceGroup string, virtualNetworkName string) ([]*armnetwork.VirtualNetworkPeering, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return nil, err
	}
	if entry.network.Properties == nil {
		return nil, nil
	}

	result := make([]*armnetwork.VirtualNetworkPeering, 0, len(entry.network.Properties.VirtualNetworkPeerings))
	for _, originalPeering := range entry.network.Properties.VirtualNetworkPeerings {
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

	err := s.getError(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)

	if err != nil {
		return nil, err
	}

	peering, err := s.getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
	if err != nil {
		return nil, err
	}

	return util.JsonClone(peering)
}

func (s *networkStore) getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) (*armnetwork.VirtualNetworkPeering, error) {
	entry, err := s.getNetworkEntryNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return nil, err
	}
	if entry.network.Properties == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	for _, peering := range entry.network.Properties.VirtualNetworkPeerings {
		if ptr.Deref(peering.Name, "") == virtualNetworkPeeringName {
			return peering, nil
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

	err := s.getError(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)

	if err != nil {
		return err
	}

	_, err = s.getPeeringNoLock(resourceGroup, virtualNetworkName, virtualNetworkPeeringName)
	if err != nil {
		// errors like network or peering not found
		return err
	}
	entry, err := s.getNetworkEntryNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return err
	}

	entry.network.Properties.VirtualNetworkPeerings = pie.Filter(entry.network.Properties.VirtualNetworkPeerings, func(item *armnetwork.VirtualNetworkPeering) bool {
		return ptr.Deref(item.Name, "") != virtualNetworkPeeringName
	})

	return nil
}

func (s *networkStore) getError(resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error {
	resourceId := azureutil.NewVirtualNetworkPeeringResourceId(s.subscription, resourceGroup, virtualNetworkName, virtualNetworkPeeringName)

	if err, errorExists := s.errorMap[resourceId.String()]; errorExists {
		return err
	}

	return nil
}
