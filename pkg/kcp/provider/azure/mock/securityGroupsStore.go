package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ SecurityGroupsClient = &securityGroupsStore{}

func newSecurityGroupsStore(subscription string) *securityGroupsStore {
	return &securityGroupsStore{
		subscription: subscription,
		items:        map[string]map[string]*armnetwork.SecurityGroup{},
	}
}

type securityGroupsStore struct {
	m sync.Mutex

	subscription string

	// items are resourceGroupName => securityGroupName => SecurityGroup
	items map[string]map[string]*armnetwork.SecurityGroup
}

func (s *securityGroupsStore) GetSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) (*armnetwork.SecurityGroup, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	return s.getSecurityGroupNoLock(resourceGroupName, networkSecurityGroupName)
}

func (s *securityGroupsStore) getSecurityGroupNoLock(resourceGroupName, networkSecurityGroupName string) (*armnetwork.SecurityGroup, error) {
	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*armnetwork.SecurityGroup{}
	}
	sg, ok := s.items[resourceGroupName][networkSecurityGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	return sg, nil
}

func (s *securityGroupsStore) CreateSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName, location string, securityRules []*armnetwork.SecurityRule, tags map[string]string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getSecurityGroupNoLock(resourceGroupName, networkSecurityGroupName)
	if err == nil {
		return fmt.Errorf("security group %s/%s already exists", resourceGroupName, networkSecurityGroupName)
	}
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return err
	}

	id := azureutil.NewNetworkSecurityGroupResourceId(s.subscription, resourceGroupName, networkSecurityGroupName)

	sg := &armnetwork.SecurityGroup{
		ID:       ptr.To(id.String()),
		Name:     ptr.To(networkSecurityGroupName),
		Location: ptr.To(location),
		Tags:     azureutil.AzureTags(tags),
		Properties: &armnetwork.SecurityGroupPropertiesFormat{
			SecurityRules:     securityRules,
			ProvisioningState: ptr.To(armnetwork.ProvisioningStateSucceeded),
		},
	}

	s.items[resourceGroupName][networkSecurityGroupName] = sg

	return nil
}

func (s *securityGroupsStore) DeleteSecurityGroup(ctx context.Context, resourceGroupName, networkSecurityGroupName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getSecurityGroupNoLock(resourceGroupName, networkSecurityGroupName)
	if err != nil {
		return err
	}

	delete(s.items[resourceGroupName], networkSecurityGroupName)

	return nil
}
