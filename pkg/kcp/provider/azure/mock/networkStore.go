package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
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

	remoteSubscription *TenantSubscription
}

// Config ===================================================

func (s *networkStore) SetPeeringConnectedFullInSync(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error {
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
		peering.Properties = &armnetwork.VirtualNetworkPeeringPropertiesFormat{}
	}

	peering.Properties.PeeringState = ptr.To(armnetwork.VirtualNetworkPeeringStateConnected)

	peering.Properties.PeeringSyncLevel = ptr.To(armnetwork.VirtualNetworkPeeringLevelFullyInSync)

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

func (s *networkStore) SetPeeringSyncLevel(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string, peeringLevel armnetwork.VirtualNetworkPeeringLevel) error {
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
		peering.Properties = &armnetwork.VirtualNetworkPeeringPropertiesFormat{}
	}

	peering.Properties.PeeringSyncLevel = ptr.To(peeringLevel)

	return nil
}

func (s *networkStore) SetNetworkAddressSpace(ctx context.Context, resourceGroup, virtualNetworkName, addressSpace string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroup, virtualNetworkName)
	if err != nil {
		return err
	}

	entry.network.Properties.AddressSpace = &armnetwork.AddressSpace{
		AddressPrefixes: []*string{ptr.To(addressSpace)},
	}

	return nil
}

func (s *networkStore) AddRemoteSubscription(ctx context.Context, remoteSubscription *TenantSubscription) {
	if isContextCanceled(ctx) {
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.remoteSubscription = remoteSubscription
}

// NetworkClient ===================================================

func (s *networkStore) CreateNetwork(ctx context.Context, resourceGroupName string, virtualNetworkName string, parameters armnetwork.VirtualNetwork, options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (azureclient.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	if options != nil && options.ResumeToken != "" {
		ret, err := s.GetNetwork(ctx, resourceGroupName, virtualNetworkName)
		return NewPollerMock(armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
			VirtualNetwork: ptr.Deref(ret, armnetwork.VirtualNetwork{}),
		}, err, ""), err
	}

	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*networkEntry{}
	}

	item, ok := s.items[resourceGroupName][virtualNetworkName]
	if ok {
		if options != nil && options.ResumeToken != "" {
			return NewPollerMock(armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
				VirtualNetwork: *item.network,
			}, nil, fmt.Sprintf("resumeToken/%s/%s", resourceGroupName, virtualNetworkName)), nil
		}
		//TODO: return correct Azure error type, check azuremeta for details, first find out what actual Azure returns in this case
		return nil, fmt.Errorf("virtual network %s/%s/%s already exists", s.subscription, resourceGroupName, virtualNetworkName)
	}
	if parameters.Properties == nil || parameters.Properties.AddressSpace == nil || len(parameters.Properties.AddressSpace.AddressPrefixes) == 0 {
		//TODO: return correct Azure error type, check azuremeta for details, first find out what actual Azure returns in this case
		return nil, fmt.Errorf("invalid parameters: address space must be specified")
	}
	parameters.Properties.ProvisioningState = ptr.To(armnetwork.ProvisioningStateSucceeded)

	cpy := parameters
	net := &cpy
	net.ID = ptr.To(azureutil.NewVirtualNetworkResourceId(s.subscription, resourceGroupName, virtualNetworkName).String())
	net.Name = ptr.To(virtualNetworkName)

	s.items[resourceGroupName][virtualNetworkName] = &networkEntry{network: net}

	ret, err := util.JsonClone(net)
	if err != nil {
		return nil, err
	}

	return NewPollerMock(armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
		VirtualNetwork: *ret,
	}, nil, fmt.Sprintf("resumeToken/%s/%s", resourceGroupName, virtualNetworkName)), nil
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

func (s *networkStore) DeleteNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string, options *armnetwork.VirtualNetworksClientBeginDeleteOptions) (azureclient.Poller[armnetwork.VirtualNetworksClientDeleteResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)

	if azuremeta.IsNotFound(err) && options != nil && options.ResumeToken != "" {
		return NewPollerMock(armnetwork.VirtualNetworksClientDeleteResponse{}, nil, fmt.Sprintf("resumeToken/%s/%s", resourceGroupName, virtualNetworkName)), nil
	}
	if err != nil {
		return nil, err
	}

	delete(s.items[resourceGroupName], virtualNetworkName)

	return NewPollerMock(armnetwork.VirtualNetworksClientDeleteResponse{}, nil, fmt.Sprintf("resumeToken/%s/%s", resourceGroupName, virtualNetworkName)), nil
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

func (s *networkStore) CreateSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, subnetParameters armnetwork.Subnet, options *armnetwork.SubnetsClientBeginCreateOrUpdateOptions) (azureclient.Poller[armnetwork.SubnetsClientCreateOrUpdateResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return nil, fmt.Errorf("azure network does not exist: %w", err)
	}

	if entry.network.Properties == nil {
		entry.network.Properties = &armnetwork.VirtualNetworkPropertiesFormat{}
	}

	id := azureutil.NewSubnetResourceId(s.subscription, resourceGroupName, virtualNetworkName, subnetName)

	for _, subnet := range entry.network.Properties.Subnets {
		if ptr.Deref(subnet.Name, "") == subnetName {
			if options != nil && options.ResumeToken != "" {
				return NewPollerMock(armnetwork.SubnetsClientCreateOrUpdateResponse{
					Subnet: *subnet,
				}, nil, fmt.Sprintf("resumeToken/%s/%s/%s", resourceGroupName, virtualNetworkName, subnetName)), nil
			}
			return nil, fmt.Errorf("subnet %s already exists: %w", id, azuremeta.NewAzureConflictError())
		}
	}
	if options != nil && options.ResumeToken != "" {
		return nil, fmt.Errorf("subnet %s does not exist for resume: %w", id, azuremeta.NewAzureNotFoundError())
	}

	subnetParameters.ID = ptr.To(id.String())
	subnetParameters.Name = ptr.To(subnetName)
	if subnetParameters.Properties == nil {
		subnetParameters.Properties = &armnetwork.SubnetPropertiesFormat{}
	}
	subnetParameters.Properties.ProvisioningState = ptr.To(armnetwork.ProvisioningStateSucceeded)

	subnet, err := util.JsonClone(&subnetParameters)
	if err != nil {
		return nil, err
	}

	entry.network.Properties.Subnets = append(entry.network.Properties.Subnets, subnet)

	return NewPollerMock(armnetwork.SubnetsClientCreateOrUpdateResponse{
		Subnet: *subnet,
	}, nil, fmt.Sprintf("resumeToken/%s/%s/%s", resourceGroupName, virtualNetworkName, subnetName)), nil
}

func (s *networkStore) DeleteSubnet(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string, options *armnetwork.SubnetsClientBeginDeleteOptions) (azureclient.Poller[armnetwork.SubnetsClientDeleteResponse], error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return nil, err
	}

	_, err = s.getSubnetNoLock(resourceGroupName, virtualNetworkName, subnetName)
	if azuremeta.IsNotFound(err) && options != nil && options.ResumeToken != "" {
		return NewPollerMock(armnetwork.SubnetsClientDeleteResponse{}, nil, fmt.Sprintf("resumeToken/%s/%s/%s", resourceGroupName, virtualNetworkName, subnetName)), nil
	}
	if err != nil {
		return nil, err
	}
	if options != nil && options.ResumeToken != "" {
		return nil, fmt.Errorf("subnet %s still exists after resume: %w", subnetName, azuremeta.NewAzureConflictError())
	}

	entry.network.Properties.Subnets = pie.Filter(entry.network.Properties.Subnets, func(x *armnetwork.Subnet) bool {
		return ptr.Deref(x.Name, "") != subnetName
	})

	return NewPollerMock(armnetwork.SubnetsClientDeleteResponse{}, nil, fmt.Sprintf("resumeToken/%s/%s/%s", resourceGroupName, virtualNetworkName, subnetName)), nil
}

// VpcPeeringClient ==============================================

func (s *networkStore) CreateOrUpdatePeering(ctx context.Context, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName, remoteVnetId string, allowVnetAccess bool, useRemoteGateway bool, allowGatewayTransit bool) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	var err error
	err = s.getError(resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)

	if err != nil {
		return err
	}

	id := azureutil.NewVirtualNetworkPeeringResourceId(s.subscription, resourceGroupName, virtualNetworkName, virtualNetworkPeeringName).String()

	peering, err := s.getPeeringNoLock(resourceGroupName, virtualNetworkName, virtualNetworkPeeringName)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		// errors like network not found
		return err
	}

	entry, err := s.getNetworkEntryNoLock(resourceGroupName, virtualNetworkName)
	if err != nil {
		return err
	}

	if entry.network.Properties == nil {
		entry.network.Properties = &armnetwork.VirtualNetworkPropertiesFormat{}
	}

	if peering == nil {
		peering = &armnetwork.VirtualNetworkPeering{
			ID:   ptr.To(id),
			Name: ptr.To(virtualNetworkPeeringName),
		}

		entry.network.Properties.VirtualNetworkPeerings = append(entry.network.Properties.VirtualNetworkPeerings, peering)
	}

	resourceId, err := azureutil.ParseResourceID(remoteVnetId)

	if err != nil {
		return err
	}

	var remoteAddressSpace *string
	if s.remoteSubscription != nil {

		remoteVnet, err := (*s.remoteSubscription).GetNetwork(ctx, resourceId.ResourceGroup, resourceId.ResourceName)
		if err != nil {
			return err
		}
		remoteAddressSpace = remoteVnet.Properties.AddressSpace.AddressPrefixes[0]
	}
	peering.Properties = &armnetwork.VirtualNetworkPeeringPropertiesFormat{
		AllowForwardedTraffic:     ptr.To(true),
		AllowGatewayTransit:       ptr.To(allowGatewayTransit),
		AllowVirtualNetworkAccess: ptr.To(allowVnetAccess),
		UseRemoteGateways:         ptr.To(useRemoteGateway),
		RemoteVirtualNetwork: &armnetwork.SubResource{
			ID: ptr.To(remoteVnetId),
		},
		PeeringState:     ptr.To(armnetwork.VirtualNetworkPeeringStateInitiated),   // check if this is needed
		PeeringSyncLevel: ptr.To(armnetwork.VirtualNetworkPeeringLevelFullyInSync), // check if this is needed
		LocalAddressSpace: &armnetwork.AddressSpace{
			AddressPrefixes: []*string{
				entry.network.Properties.AddressSpace.AddressPrefixes[0],
			},
		},
		RemoteAddressSpace: &armnetwork.AddressSpace{
			AddressPrefixes: []*string{
				remoteAddressSpace,
			},
		},
	}

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
