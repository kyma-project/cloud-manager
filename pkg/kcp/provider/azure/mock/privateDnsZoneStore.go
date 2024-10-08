package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ PrivateDnsZoneClient = &privateDnsZoneStore{}

func newPrivateDnsZoneStore(subscription string) *privateDnsZoneStore {
	return &privateDnsZoneStore{
		subscription: subscription,
		items:        map[string]map[string]*armprivatedns.PrivateZone{},
	}
}

type privateDnsZoneStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => privateDnsZoneName => armprivatedns.PrivateZone
	items map[string]map[string]*armprivatedns.PrivateZone
}

func (s *privateDnsZoneStore) GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) (*armprivatedns.PrivateZone, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	privateDnsZone, err := s.getPrivateZoneNonLocking(resourceGroupName, privateDnsZoneName)
	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(privateDnsZone)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *privateDnsZoneStore) getPrivateZoneNonLocking(resourceGroupName, privateDnsZoneName string) (*armprivatedns.PrivateZone, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	info, ok := group[privateDnsZoneName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	return info, nil
}

func (s *privateDnsZoneStore) CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*armprivatedns.PrivateZone{}
	}
	_, ok = s.items[resourceGroupName][privateDnsZoneName]
	if ok {
		return fmt.Errorf("dns zone %s already exist", privateDnsZoneName)
	}

	props := &armprivatedns.PrivateZoneProperties{}
	props.ProvisioningState = ptr.To(armprivatedns.ProvisioningStateSucceeded)

	item := &armprivatedns.PrivateZone{
		Location:   to.Ptr("global"),
		Properties: props,
		Name:       to.Ptr(privateDnsZoneName),
	}

	s.items[resourceGroupName][privateDnsZoneName] = item

	return nil
}

func (s *privateDnsZoneStore) DeletePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getPrivateZoneNonLocking(resourceGroupName, privateDnsZoneName)
	if err != nil {
		return err
	}

	s.items[resourceGroupName][privateDnsZoneName] = nil

	return nil
}
