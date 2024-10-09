package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ PrivateDnsZoneGroupClient = &privateDnsZoneGroupStore{}

func newPrivateDnsZoneGroupStore(subscription string) *privateDnsZoneGroupStore {
	return &privateDnsZoneGroupStore{
		subscription: subscription,
		items:        map[string]map[string]map[string]*armnetwork.PrivateDNSZoneGroup{},
	}
}

type privateDnsZoneGroupStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => privateEndPointName => privateDnsZoneGroupName => armnetwork.PrivateDNSZoneGroup
	items map[string]map[string]map[string]*armnetwork.PrivateDNSZoneGroup
}

func (s *privateDnsZoneGroupStore) GetPrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) (*armnetwork.PrivateDNSZoneGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	privateDnsZoneGroup, err := s.getPrivateDnsZoneGroupNonLocking(resourceGroupName, privateEndPointName, privateDnsZoneGroupName)
	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(privateDnsZoneGroup)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *privateDnsZoneGroupStore) getPrivateDnsZoneGroupNonLocking(resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) (*armnetwork.PrivateDNSZoneGroup, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	pep, ok := group[privateEndPointName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	privateDnsZoneGroup, ok := pep[privateDnsZoneGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}

	return privateDnsZoneGroup, nil
}

func (s *privateDnsZoneGroupStore) CreatePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string, parameters armnetwork.PrivateDNSZoneGroup) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]map[string]*armnetwork.PrivateDNSZoneGroup{}
	}
	_, ok = s.items[resourceGroupName][privateEndPointName]
	if !ok {
		s.items[resourceGroupName][privateEndPointName] = map[string]*armnetwork.PrivateDNSZoneGroup{}
	}
	_, ok = s.items[resourceGroupName][privateEndPointName][privateDnsZoneGroupName]
	if ok {
		return fmt.Errorf("private dns zone group %s already exist", privateDnsZoneGroupName)
	}

	props := &armnetwork.PrivateDNSZoneGroupPropertiesFormat{}
	props.ProvisioningState = ptr.To(armnetwork.ProvisioningStateSucceeded)

	item := &armnetwork.PrivateDNSZoneGroup{
		Properties: props,
		Name:       to.Ptr(privateDnsZoneGroupName),
	}

	s.items[resourceGroupName][privateEndPointName][privateDnsZoneGroupName] = item

	return nil
}

func (s *privateDnsZoneGroupStore) DeletePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getPrivateDnsZoneGroupNonLocking(resourceGroupName, privateEndPointName, privateDnsZoneGroupName)
	if err != nil {
		return err
	}

	s.items[resourceGroupName][privateEndPointName][privateDnsZoneGroupName] = nil

	return nil
}
