package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	"google.golang.org/api/googleapi"
	"k8s.io/utils/ptr"
)

type MemoryStoreClusterClientFakeUtils interface {
	GetMemoryStoreRedisClusterByName(name string) *clusterpb.Cluster
	SetMemoryStoreRedisClusterLifeCycleState(name string, state clusterpb.Cluster_State)
	DeleteMemorStoreRedisClusterByName(name string)
}

type memoryStoreClusterClientFake struct {
	mutex         sync.Mutex
	redisClusters map[string]*clusterpb.Cluster
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) GetMemoryStoreRedisClusterByName(name string) *clusterpb.Cluster {
	return memoryStoreClusterClientFake.redisClusters[name]
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) SetMemoryStoreRedisClusterLifeCycleState(name string, state clusterpb.Cluster_State) {
	if instance, ok := memoryStoreClusterClientFake.redisClusters[name]; ok {
		instance.State = state
	}
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) DeleteMemorStoreRedisClusterByName(name string) {
	delete(memoryStoreClusterClientFake.redisClusters, name)
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) CreateRedisCluster(ctx context.Context, projectId, locationId, clusterId string, options gcpredisclusterclient.CreateRedisClusterOptions) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	memoryStoreClusterClientFake.mutex.Lock()
	defer memoryStoreClusterClientFake.mutex.Unlock()

	name := gcpredisclusterclient.GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)
	redisCluster := &clusterpb.Cluster{
		Name:         name,
		State:        clusterpb.Cluster_CREATING,
		RedisConfigs: options.RedisConfigs,
		ReplicaCount: ptr.To(options.ReplicaCount),
		ShardCount:   ptr.To(options.ShardCount),
		NodeType:     clusterpb.NodeType(clusterpb.NodeType_value[options.NodeType]),

		DiscoveryEndpoints: []*clusterpb.DiscoveryEndpoint{
			{
				Address: "127.0.0.3",
				Port:    6879,
			},
		},

		PscConfigs: []*clusterpb.PscConfig{{
			Network: options.VPCNetworkFullName,
		}},
		PersistenceConfig:      &clusterpb.ClusterPersistenceConfig{Mode: clusterpb.ClusterPersistenceConfig_DISABLED},
		AuthorizationMode:      clusterpb.AuthorizationMode_AUTH_MODE_DISABLED,
		TransitEncryptionMode:  clusterpb.TransitEncryptionMode_TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION,
		ZoneDistributionConfig: &clusterpb.ZoneDistributionConfig{Mode: clusterpb.ZoneDistributionConfig_MULTI_ZONE},

		DeletionProtectionEnabled: ptr.To(false),
	}
	memoryStoreClusterClientFake.redisClusters[name] = redisCluster

	return nil
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) UpdateRedisCluster(ctx context.Context, redisCluster *clusterpb.Cluster, updateMask []string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	memoryStoreClusterClientFake.mutex.Lock()
	defer memoryStoreClusterClientFake.mutex.Unlock()

	if instance, ok := memoryStoreClusterClientFake.redisClusters[redisCluster.Name]; ok {
		instance.State = clusterpb.Cluster_UPDATING

		instance.NodeType = redisCluster.NodeType
		instance.ReplicaCount = redisCluster.ReplicaCount
		instance.ShardCount = redisCluster.ShardCount
		instance.RedisConfigs = redisCluster.RedisConfigs
		instance.MaintenancePolicy = redisCluster.MaintenancePolicy
	}

	return nil
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) DeleteRedisCluster(ctx context.Context, projectId, locationId, clusterId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}

	memoryStoreClusterClientFake.mutex.Lock()
	defer memoryStoreClusterClientFake.mutex.Unlock()

	name := gcpredisclusterclient.GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)

	if instance, ok := memoryStoreClusterClientFake.redisClusters[name]; ok {
		instance.State = clusterpb.Cluster_DELETING
		return nil
	}

	return &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}
}

func (memoryStoreClusterClientFake *memoryStoreClusterClientFake) GetRedisCluster(ctx context.Context, projectId, locationId, clusterId string) (*clusterpb.Cluster, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	memoryStoreClusterClientFake.mutex.Lock()
	defer memoryStoreClusterClientFake.mutex.Unlock()

	name := gcpredisclusterclient.GetGcpMemoryStoreRedisClusterName(projectId, locationId, clusterId)

	if instance, ok := memoryStoreClusterClientFake.redisClusters[name]; ok {
		return instance, nil
	}

	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not Found",
	}

}
