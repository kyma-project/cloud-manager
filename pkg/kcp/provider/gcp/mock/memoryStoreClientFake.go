package mock

import (
	"context"
	"sync"

	redis "cloud.google.com/go/redis/apiv1"
	"cloud.google.com/go/redis/apiv1/redispb"
	memoryStoreClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
)

type MemoryStoreClientFakeUtils interface {
	GetMemoryStoreRedisByName(name string) *redispb.Instance
	SetMemoryStoreRedisLifeCycleState(name string, state redispb.Instance_State)
	DeleteMemorStoreRedisByName(name string)
}

type memoryStoreClientFake struct {
	mutex          sync.Mutex
	redisInstances map[string]*redispb.Instance
}

func (memoryStoreClientFake *memoryStoreClientFake) GetMemoryStoreRedisByName(name string) *redispb.Instance {
	return memoryStoreClientFake.redisInstances[name]
}

func (memoryStoreClientFake *memoryStoreClientFake) SetMemoryStoreRedisLifeCycleState(name string, state redispb.Instance_State) {
	if instance, ok := memoryStoreClientFake.redisInstances[name]; ok {
		instance.State = state
	}
}

func (memoryStoreClientFake *memoryStoreClientFake) DeleteMemorStoreRedisByName(name string) {
	delete(memoryStoreClientFake.redisInstances, name)
}

func (memoryStoreClientFake *memoryStoreClientFake) CreateRedisInstance(ctx context.Context, projectId string, locationId string, instanceId string, options memoryStoreClient.CreateRedisInstanceOptions) (*redis.CreateInstanceOperation, error) {
	memoryStoreClientFake.mutex.Lock()
	defer memoryStoreClientFake.mutex.Unlock()

	name := memoryStoreClient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)
	redisInstance := &redispb.Instance{
		Name:             name,
		State:            redispb.Instance_CREATING,
		Host:             "192.168.0.1",
		Port:             6093,
		ReadEndpoint:     "192.168.24.1",
		ReadEndpointPort: 5093,
	}
	memoryStoreClientFake.redisInstances[name] = redisInstance

	return &redis.CreateInstanceOperation{}, nil // redis.CreateInstanceOperation is not used in actual code, so empty object is returned
}

func (memoryStoreClientFake *memoryStoreClientFake) UpdateRedisInstance(ctx context.Context, redisInstance *redispb.Instance, updateMask []string) error {
	memoryStoreClientFake.mutex.Lock()
	defer memoryStoreClientFake.mutex.Unlock()

	if instance, ok := memoryStoreClientFake.redisInstances[redisInstance.Name]; ok {
		instance.State = redispb.Instance_UPDATING
	}

	return nil
}

func (memoryStoreClientFake *memoryStoreClientFake) DeleteRedisInstance(ctx context.Context, projectId string, locationId string, instanceId string) error {
	memoryStoreClientFake.mutex.Lock()
	defer memoryStoreClientFake.mutex.Unlock()

	name := memoryStoreClient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)

	if instance, ok := memoryStoreClientFake.redisInstances[name]; ok {
		instance.State = redispb.Instance_DELETING
	}

	return nil
}

func (memoryStoreClientFake *memoryStoreClientFake) GetRedisInstance(ctx context.Context, projectId string, locationId string, instanceId string) (*redispb.Instance, *redispb.InstanceAuthString, error) {
	memoryStoreClientFake.mutex.Lock()
	defer memoryStoreClientFake.mutex.Unlock()

	name := memoryStoreClient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)

	instance := memoryStoreClientFake.redisInstances[name]

	return instance, &redispb.InstanceAuthString{AuthString: "0df0aea4-2cd6-4b9a-900f-a650661e1740"}, nil
}
