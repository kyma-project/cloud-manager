package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"math/rand"
	"sync"
)

func newPublicIpAddressStore(subscription string) *publicIpAddressStore {
	return &publicIpAddressStore{
		subscription: subscription,
		items:        make(map[string]map[string]*publicIpAddressesEntry),
	}
}

type publicIpAddressesEntry struct {
	ipAddress *armnetwork.PublicIPAddress
}

var _ PublicIpAddressesClient = &publicIpAddressStore{}

type publicIpAddressStore struct {
	m sync.Mutex

	subscription string

	// items are map resourceGroup => publicIpAddressName => publicIpAddressesEntry
	items map[string]map[string]*publicIpAddressesEntry
}

func (s *publicIpAddressStore) CreatePublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName, location, zone string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*publicIpAddressesEntry{}
	}

	_, ok = s.items[resourceGroupName][publicIpAddressName]
	if ok {
		return fmt.Errorf("public ip address %s/%s/%s already exists", s.subscription, resourceGroupName, publicIpAddressName)
	}

	ipAddress := &armnetwork.PublicIPAddress{
		ID:       ptr.To(azureutil.NewPublicIpAddressResourceId(s.subscription, resourceGroupName, publicIpAddressName).String()),
		Name:     ptr.To(publicIpAddressName),
		Location: ptr.To(location),
		SKU: &armnetwork.PublicIPAddressSKU{
			Name: ptr.To(armnetwork.PublicIPAddressSKUNameStandard),
			Tier: ptr.To(armnetwork.PublicIPAddressSKUTierRegional),
		},
		Zones: []*string{ptr.To(zone)},
		Type:  ptr.To("Microsoft.Network/publicIPAddresses"),
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			ProvisioningState:      ptr.To(armnetwork.ProvisioningStateSucceeded),
			IPAddress:              ptr.To(fmt.Sprintf("33.%d.%d.%d", rand.Intn(250)+1, rand.Intn(250)+1, rand.Intn(250)+1)),
			PublicIPAddressVersion: ptr.To(armnetwork.IPVersionIPv4),
		},
	}

	s.items[resourceGroupName][publicIpAddressName] = &publicIpAddressesEntry{ipAddress: ipAddress}

	return nil
}

func (s *publicIpAddressStore) GetPublicIpAddress(ctx context.Context, resourceGroupName, publicIpAddressName string) (*armnetwork.PublicIPAddress, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	entry, err := s.getPublicIpAddressEntryNoLock(resourceGroupName, publicIpAddressName)
	if err != nil {
		return nil, err
	}

	return util.JsonClone(entry.ipAddress)
}

func (s *publicIpAddressStore) getPublicIpAddressEntryNoLock(resourceGroupName, publicIpAddressName string) (*publicIpAddressesEntry, error) {
	_, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	_, ok = s.items[resourceGroupName][publicIpAddressName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return s.items[resourceGroupName][publicIpAddressName], nil
}
