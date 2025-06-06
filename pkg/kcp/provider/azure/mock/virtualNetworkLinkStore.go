package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ VirtualNetworkLinkClient = &virtualNetworkLinkStore{}

func newVirtualNetworkLinkStore(subscription string) *virtualNetworkLinkStore {
	return &virtualNetworkLinkStore{
		subscription: subscription,
		items:        map[string]map[string]map[string]*armprivatedns.VirtualNetworkLink{},
	}
}

type virtualNetworkLinkStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => privateDnsZoneName => virtualNetworkLinkName => *armprivatedns.VirtualNetworkLink
	items map[string]map[string]map[string]*armprivatedns.VirtualNetworkLink
}

func (s *virtualNetworkLinkStore) GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	virtualNetworkLink, err := s.getVirtualNetworkLinkNonLocking(resourceGroupName, privateZoneName, virtualNetworkLinkName)
	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(virtualNetworkLink)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *virtualNetworkLinkStore) getVirtualNetworkLinkNonLocking(resourceGroupName, privateDnsZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	dnsGroup, ok := group[privateDnsZoneName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	virtualNetworkLink, ok := dnsGroup[virtualNetworkLinkName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	if virtualNetworkLink == nil {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return virtualNetworkLink, nil
}

func (s *virtualNetworkLinkStore) CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName, vnetId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]map[string]*armprivatedns.VirtualNetworkLink{}
	}
	_, ok = s.items[resourceGroupName][privateDnsZoneName]
	if !ok {
		s.items[resourceGroupName][privateDnsZoneName] = map[string]*armprivatedns.VirtualNetworkLink{}
	}
	_, ok = s.items[resourceGroupName][privateDnsZoneName][virtualNetworkLinkName]
	if ok {
		return fmt.Errorf("virtual link %s already exist", virtualNetworkLinkName)
	}

	props := &armprivatedns.VirtualNetworkLinkProperties{}
	props.ProvisioningState = ptr.To(armprivatedns.ProvisioningStateSucceeded)
	props.VirtualNetworkLinkState = ptr.To(armprivatedns.VirtualNetworkLinkStateCompleted)

	item := &armprivatedns.VirtualNetworkLink{
		ID:         ptr.To(azureutil.NewVirtualNetworkLinkResourceId(s.subscription, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName).String()),
		Location:   to.Ptr("global"),
		Properties: props,
		Name:       to.Ptr(virtualNetworkLinkName),
	}

	s.items[resourceGroupName][privateDnsZoneName][virtualNetworkLinkName] = item

	return nil
}

func (s *virtualNetworkLinkStore) DeleteVirtualNetworkLink(ctx context.Context, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getVirtualNetworkLinkNonLocking(resourceGroupName, privateDnsZoneName, virtualNetworkLinkName)
	if err != nil {
		return err
	}

	s.items[resourceGroupName][privateDnsZoneName][virtualNetworkLinkName] = nil

	return nil
}
