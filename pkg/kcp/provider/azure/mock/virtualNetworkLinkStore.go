package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"sync"
)

var _ VirtualNetworkLinkClient = &virtualNetworkLinkStore{}

func newVirtualNetworkLinkStore(subscription string) *virtualNetworkLinkStore {
	return &virtualNetworkLinkStore{
		subscription: subscription,
		items:        map[string]map[string]*armprivatedns.PrivateZone{},
	}
}

type virtualNetworkLinkStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => securityGroupName => SecurityGroup
	items map[string]map[string]*armprivatedns.PrivateZone
}

func (s *virtualNetworkLinkStore) GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error) {
	return nil, nil
}

func (s *virtualNetworkLinkStore) CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string, parameters armprivatedns.VirtualNetworkLink) error {
	return nil
}

func (s *virtualNetworkLinkStore) DeleteVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) error {
	return nil
}
