package mock

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	armResources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/utils/ptr"
	"sync"
)

type RedisCacheClientFakeUtils interface {
	GetRedisCacheByResourceGroupName(resourceGroupName string) *armRedis.ResourceInfo
	DeleteRedisCacheByResourceGroupName(resourceGroupName string)
}

type redisCacheClientFake struct {
	mutex               sync.Mutex
	redisResourceGroups map[string]redisResourceGroup
}

type redisResourceGroup struct {
	redisGroupInstance *armResources.ResourceGroup
	redisInstance      *armRedis.ResourceInfo
}

func (redisCacheClientFake *redisCacheClientFake) GetResourceGroup(ctx context.Context, name string) (*armResources.ResourceGroupsClientGetResponse, error) {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	resourceGroup, resourceGroupPresent := redisCacheClientFake.redisResourceGroups[name]
	if resourceGroupPresent {
		resourceGroupsClientGetResponse := armResources.ResourceGroupsClientGetResponse{
			ResourceGroup: *resourceGroup.redisGroupInstance,
		}

		return &resourceGroupsClientGetResponse, nil
	}

	responseErr := azcore.ResponseError{
		StatusCode: 404,
	}
	return nil, errors.Join(&responseErr)
}

func (redisCacheClientFake *redisCacheClientFake) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.CreateParameters) error {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	if resourceGroup, ok := redisCacheClientFake.redisResourceGroups[resourceGroupName]; ok {

		provState := armRedis.ProvisioningStateSucceeded
		redisInstanceCreated := armRedis.ResourceInfo{
			Location: ptr.To("US east"),
			Name:     &redisInstanceName,
			Properties: &armRedis.Properties{
				ProvisioningState: &provState,
				HostName:          ptr.To("tch.redis-host.com"),
				Port:              to.Ptr[int32](6379),
				AccessKeys: &armRedis.AccessKeys{
					PrimaryKey: ptr.To("primary-key"),
				},
				SKU: &armRedis.SKU{
					Capacity: to.Ptr[int32](1),
				},
			},
		}
		resourceGroup.redisInstance = &redisInstanceCreated
		redisCacheClientFake.redisResourceGroups[resourceGroupName] = resourceGroup

		return nil
	}

	return errors.New("failed to create FAKE Azure Redis instance")
}

func (redisCacheClientFake *redisCacheClientFake) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armRedis.ResourceInfo, error) {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	resourceGroup, resourceGroupPresent := redisCacheClientFake.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		return resourceGroup.redisInstance, nil
	}

	return nil, errors.New("failed to get FAKE Azure Redis instance")
}

func (redisCacheClientFake *redisCacheClientFake) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	resourceGroup, resourceGroupPresent := redisCacheClientFake.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		resourceGroup.redisInstance.Properties.ProvisioningState = to.Ptr(armRedis.ProvisioningStateDeleting)
		return nil
	}

	return errors.New("failed to delete FAKE Azure Redis instance")
}

func (redisCacheClientFake *redisCacheClientFake) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) (string, error) {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	resourceGroup, resourceGroupPresent := redisCacheClientFake.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		return *resourceGroup.redisInstance.Properties.AccessKeys.PrimaryKey, nil
	}

	return "", errors.New("failed to get FAKE Azure Redis credentials")
}

func (redisCacheClientFake *redisCacheClientFake) CreateResourceGroup(ctx context.Context, name string, location string) error {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	redisResourceGroup := redisResourceGroup{
		redisGroupInstance: &armResources.ResourceGroup{
			Location: &location,
			Name:     &name,
		},
	}
	redisCacheClientFake.redisResourceGroups[name] = redisResourceGroup

	return nil
}

func (redisCacheClientFake *redisCacheClientFake) DeleteResourceGroup(ctx context.Context, name string) error {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	delete(redisCacheClientFake.redisResourceGroups, name)

	return nil
}

func (redisCacheClientFake *redisCacheClientFake) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armRedis.UpdateParameters) error {
	redisCacheClientFake.mutex.Lock()
	defer redisCacheClientFake.mutex.Unlock()

	resourceGroup, resourceGroupPresent := redisCacheClientFake.redisResourceGroups[resourceGroupName]
	if resourceGroupPresent {
		resourceGroup.redisInstance.Properties.SKU = parameters.Properties.SKU
	}

	return errors.New("failed to update FAKE Azure Redis instance")
}
