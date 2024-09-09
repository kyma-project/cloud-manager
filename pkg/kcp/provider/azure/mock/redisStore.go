package mock

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/utils/ptr"
	"sync"
)

var _ RedisInstanceClient = &redisStore{}
var _ RedisConfig = &redisStore{}

func newRedisStore(subscription string) *redisStore {
	return &redisStore{
		subscription:        subscription,
		redisResourceGroups: map[string]redisResourceGroup{},
	}
}

type redisStore struct {
	m                   sync.Mutex
	subscription        string
	redisResourceGroups map[string]redisResourceGroup
}

type redisResourceGroup struct {
	redisGroupInstance *armresources.ResourceGroup
	redisInstance      *armredis.ResourceInfo
}

// Config =================================================================================================

func (s *redisStore) DeleteRedisCacheByResourceGroupName(resourceGroupName string) error {
	panic("implement me")
}

// RedisInstanceClient ====================================================================================

func (s *redisStore) GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroupsClientGetResponse, error) {
	s.m.Lock()
	defer s.m.Unlock()

	resourceGroup, resourceGroupPresent := s.redisResourceGroups[name]
	if resourceGroupPresent {
		resourceGroupsClientGetResponse := armresources.ResourceGroupsClientGetResponse{
			ResourceGroup: *resourceGroup.redisGroupInstance,
		}

		return &resourceGroupsClientGetResponse, nil
	}

	responseErr := azcore.ResponseError{
		StatusCode: 404,
	}
	return nil, errors.Join(&responseErr)
}

func (s *redisStore) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error {
	s.m.Lock()
	defer s.m.Unlock()

	if resourceGroup, ok := s.redisResourceGroups[resourceGroupName]; ok {

		provState := armredis.ProvisioningStateSucceeded
		redisInstanceCreated := armredis.ResourceInfo{
			Location: ptr.To("US east"),
			Name:     &redisInstanceName,
			Properties: &armredis.Properties{
				ProvisioningState: &provState,
				HostName:          ptr.To("tch.redis-host.com"),
				Port:              to.Ptr[int32](6379),
				AccessKeys: &armredis.AccessKeys{
					PrimaryKey: ptr.To("primary-key"),
				},
				SKU: &armredis.SKU{
					Capacity: to.Ptr[int32](1),
				},
			},
		}
		resourceGroup.redisInstance = &redisInstanceCreated
		s.redisResourceGroups[resourceGroupName] = resourceGroup

		return nil
	}

	return nil
}

func (s *redisStore) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error) {
	s.m.Lock()
	defer s.m.Unlock()

	resourceGroup, resourceGroupPresent := s.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		return resourceGroup.redisInstance, nil
	}

	return nil, nil
}

func (s *redisStore) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	s.m.Lock()
	defer s.m.Unlock()

	resourceGroup, resourceGroupPresent := s.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		resourceGroup.redisInstance.Properties.ProvisioningState = to.Ptr(armredis.ProvisioningStateDeleting)
		return nil
	}

	return nil
}

func (s *redisStore) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	resourceGroup, resourceGroupPresent := s.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		return *resourceGroup.redisInstance.Properties.AccessKeys.PrimaryKey, nil
	}

	return "", nil
}

func (s *redisStore) CreateResourceGroup(ctx context.Context, name string, location string) error {
	s.m.Lock()
	defer s.m.Unlock()

	redisResourceGroup := redisResourceGroup{
		redisGroupInstance: &armresources.ResourceGroup{
			Location: &location,
			Name:     &name,
		},
	}
	s.redisResourceGroups[name] = redisResourceGroup

	return nil
}

func (s *redisStore) DeleteResourceGroup(ctx context.Context, name string) error {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.redisResourceGroups, name)

	return nil
}

func (s *redisStore) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error {
	s.m.Lock()
	defer s.m.Unlock()

	resourceGroup, resourceGroupPresent := s.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		resourceGroup.redisInstance.Properties.SKU = parameters.Properties.SKU
	}

	return nil
}
