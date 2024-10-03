package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
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

	// items are resourceGroupName => securityGroupName => SecurityGroup
	items map[string]map[string]*armprivatedns.PrivateZone
}

func (s *privateDnsZoneStore) GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneGroupName string) (*armprivatedns.PrivateZone, error) {
	return nil, nil
}

func (s *privateDnsZoneStore) CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error {
	return nil
}

func (s *privateDnsZoneStore) DeletePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneGroupName string) error {
	return nil
}
