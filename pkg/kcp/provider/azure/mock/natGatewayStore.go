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

func newNatGatewayStore(subscription string) *natGatewayStore {
	return &natGatewayStore{
		subscription: subscription,
		items:        make(map[string]map[string]*natGatewayEntry),
	}
}

type natGatewayEntry struct {
	gw *armnetwork.NatGateway
}

var _ NatGatewayClient = &natGatewayStore{}

type natGatewayStore struct {
	m sync.Mutex

	subscription string

	// items are map resourceGroup => natGatewayName => natGatewayEntry
	items map[string]map[string]*natGatewayEntry
}

func (s *natGatewayStore) CreateNatGateway(ctx context.Context, resourceGroupName, natGatewayName, location, zone string, subnetIds, publicIpAddressIds []string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*natGatewayEntry{}
	}

	_, ok = s.items[resourceGroupName][natGatewayName]
	if ok {
		return fmt.Errorf("nat gateway %s/%s/%s already exists", s.subscription, resourceGroupName, natGatewayName)
	}

	natGateway := &armnetwork.NatGateway{
		ID:       ptr.To(azureutil.NewNatGatewayResourceId(s.subscription, resourceGroupName, natGatewayName).String()),
		Name:     ptr.To(natGatewayName),
		Location: ptr.To(location),
		Zones:    []*string{ptr.To(zone)},
		SKU: &armnetwork.NatGatewaySKU{
			Name: ptr.To(armnetwork.NatGatewaySKUNameStandard),
		},
		Type: ptr.To("Microsoft.Network/natGateways"),
		Properties: &armnetwork.NatGatewayPropertiesFormat{
			ProvisioningState: ptr.To(armnetwork.ProvisioningStateSucceeded),
		},
	}
	if subnetIds != nil {
		natGateway.Properties.Subnets = pie.Map(subnetIds, func(subnetId string) *armnetwork.SubResource {
			return &armnetwork.SubResource{ID: ptr.To(subnetId)}
		})
	}
	if publicIpAddressIds != nil {
		natGateway.Properties.PublicIPAddresses = pie.Map(publicIpAddressIds, func(publicIpAddressId string) *armnetwork.SubResource {
			return &armnetwork.SubResource{ID: ptr.To(publicIpAddressId)}
		})
	}

	s.items[resourceGroupName][natGatewayName] = &natGatewayEntry{gw: natGateway}

	return nil
}

func (s *natGatewayStore) GetNatGateway(ctx context.Context, resourceGroupName, natGatewayName string) (*armnetwork.NatGateway, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getNatGatewayEntryNoLock(resourceGroupName, natGatewayName)
	if err != nil {
		return nil, err
	}
	return util.JsonClone(entry.gw)
}

func (s *natGatewayStore) getNatGatewayEntryNoLock(resourceGroupName, natGatewayName string) (*natGatewayEntry, error) {
	_, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	_, ok = s.items[resourceGroupName][natGatewayName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return s.items[resourceGroupName][natGatewayName], nil
}
