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

var _ PrivateEndPointsClient = &privateEndPointsStore{}

func newPrivateEndPointsStore(subscription string) *privateEndPointsStore {
	return &privateEndPointsStore{
		subscription: subscription,
		items:        map[string]map[string]*armnetwork.PrivateEndpoint{},
	}
}

type privateEndPointsStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => privateEndPointName => armnetwork.PrivateEndpoint
	items map[string]map[string]*armnetwork.PrivateEndpoint
}

func (s *privateEndPointsStore) GetPrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	privateEndPoint, err := s.getPrivateEndPointNonLocking(resourceGroupName, privateEndPointName)
	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(privateEndPoint)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *privateEndPointsStore) getPrivateEndPointNonLocking(resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	pep, ok := group[privateEndPointName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	return pep, nil
}

func (s *privateEndPointsStore) CreatePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string, parameters armnetwork.PrivateEndpoint) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*armnetwork.PrivateEndpoint{}
	}
	_, ok = s.items[resourceGroupName][privateEndPointName]
	if ok {
		return fmt.Errorf("private end point %s already exist", privateEndPointName)
	}

	props := &armnetwork.PrivateEndpointProperties{}
	props.ProvisioningState = ptr.To(armnetwork.ProvisioningStateSucceeded)

	item := &armnetwork.PrivateEndpoint{
		Properties: props,
		Name:       to.Ptr(privateEndPointName),
	}

	s.items[resourceGroupName][privateEndPointName] = item

	return nil
}

func (s *privateEndPointsStore) DeletePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getPrivateEndPointNonLocking(resourceGroupName, privateEndPointName)
	if err != nil {
		return err
	}

	delete(s.items[resourceGroupName], privateEndPointName)

	return nil
}
