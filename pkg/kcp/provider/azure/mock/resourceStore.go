package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ ResourceGroupsClient = &resourceStore{}

func newResourceStore(subscription string) *resourceStore {
	return &resourceStore{
		subscription: subscription,
		items:        map[string]*armresources.ResourceGroup{},
	}
}

type resourceStore struct {
	m            sync.Mutex
	subscription string

	items map[string]*armresources.ResourceGroup
}

func (s *resourceStore) GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	return s.getResourceGroupNoLock(name)
}

func (s *resourceStore) getResourceGroupNoLock(name string) (*armresources.ResourceGroup, error) {
	rg, ok := s.items[name]
	if !ok {
		return nil, azureMeta.NewAzureNotFoundError()
	}

	return util.JsonClone(rg)
}

func (s *resourceStore) CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getResourceGroupNoLock(name)
	if azureMeta.IgnoreNotFoundError(err) != nil {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("resource group '%s' already exists", name)
	}

	rgTags := make(map[string]*string, len(tags))
	for k, v := range tags {
		rgTags[k] = ptr.To(v)
	}
	rg := &armresources.ResourceGroup{
		Location: ptr.To(location),
		Tags:     rgTags,
		ID:       ptr.To(azureutil.NewResourceGroupResourceId(s.subscription, name).String()),
		Name:     ptr.To(name),
	}

	s.items[name] = rg

	return rg, nil
}

func (s *resourceStore) DeleteResourceGroup(ctx context.Context, name string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getResourceGroupNoLock(name)
	if err != nil {
		return err
	}

	delete(s.items, name)

	return nil
}
