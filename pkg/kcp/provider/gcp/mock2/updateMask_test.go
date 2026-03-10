package mock2

import (
	"testing"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestUpdateMask(t *testing.T) {

	t.Run("UpdateFilestoreInstance", func(t *testing.T) {
		instance := &filestorepb.Instance{
			Name:        "instance-name",
			Description: "instance description",
			State:       filestorepb.Instance_READY,
			Tier:        filestorepb.Instance_PREMIUM,
			Labels: map[string]string{
				"label-key": "label-value",
			},
			FileShares: []*filestorepb.FileShareConfig{
				{
					Name:       "share-name",
					CapacityGb: 1024,
				},
			},
			Networks: []*filestorepb.NetworkConfig{
				{
					Network:         "network-name",
					ReservedIpRange: "some-ip-range",
					ConnectMode:     filestorepb.NetworkConfig_DIRECT_PEERING,
				},
			},
		}

		update := &filestorepb.Instance{
			Name:        "instance-name-2",
			Description: "instance description-2",
			FileShares: []*filestorepb.FileShareConfig{
				{
					CapacityGb: 2048,
				},
			},
		}

		updateMask := &fieldmaskpb.FieldMask{
			Paths: []string{"name", "file_shares"},
		}

		assert.NoError(t, UpdateMask(instance, update, updateMask))

		assert.Equal(t, "instance-name-2", instance.Name)
		// description must remain the same, since it's not listed in the updateMask
		assert.Equal(t, "instance description", instance.Description)
		assert.Equal(t, int64(2048), instance.FileShares[0].CapacityGb)
	})

	t.Run("UpdateRedisInstance", func(t *testing.T) {
		instance := &redispb.Instance{
			Name:         "redis-instance-name",
			DisplayName:  "redis-display-name",
			AuthEnabled:  false,
			MemorySizeGb: 4,
			ReplicaCount: 1,
			RedisConfigs: map[string]string{
				"maxmemory-policy": "allkeys-lru",
			},
			MaintenancePolicy: &redispb.MaintenancePolicy{
				Description: "original policy",
			},
		}

		update := &redispb.Instance{
			Name:         "redis-instance-name-updated",
			DisplayName:  "redis-display-name-updated",
			AuthEnabled:  true,
			MemorySizeGb: 8,
			ReplicaCount: 3,
			RedisConfigs: map[string]string{
				"maxmemory-policy":       "volatile-lru",
				"notify-keyspace-events": "Ex",
			},
			MaintenancePolicy: &redispb.MaintenancePolicy{
				Description: "updated policy",
			},
		}

		updateMask := &fieldmaskpb.FieldMask{
			Paths: []string{"auth_enabled", "maintenance_policy", "memory_size_gb", "redis_configs", "replica_count"},
		}

		assert.NoError(t, UpdateMask(instance, update, updateMask))

		// These fields should be updated (listed in updateMask)
		assert.Equal(t, true, instance.AuthEnabled)
		assert.Equal(t, int32(8), instance.MemorySizeGb)
		assert.Equal(t, int32(3), instance.ReplicaCount)
		assert.Equal(t, "volatile-lru", instance.RedisConfigs["maxmemory-policy"])
		assert.Equal(t, "Ex", instance.RedisConfigs["notify-keyspace-events"])
		assert.Equal(t, "updated policy", instance.MaintenancePolicy.Description)

		// These fields should remain unchanged (not listed in updateMask)
		assert.Equal(t, "redis-instance-name", instance.Name)
		assert.Equal(t, "redis-display-name", instance.DisplayName)
	})

	t.Run("UpdateRedisCluster", func(t *testing.T) {
		cluster := &clusterpb.Cluster{
			Name:         "redis-cluster-name",
			NodeType:     clusterpb.NodeType_REDIS_STANDARD_SMALL,
			ReplicaCount: proto.Int32(1),
			ShardCount:   proto.Int32(3),
			RedisConfigs: map[string]string{
				"maxmemory-policy": "allkeys-lru",
			},
		}

		update := &clusterpb.Cluster{
			Name:         "redis-cluster-name-updated",
			NodeType:     clusterpb.NodeType_REDIS_HIGHMEM_MEDIUM,
			ReplicaCount: proto.Int32(2),
			ShardCount:   proto.Int32(5),
			RedisConfigs: map[string]string{
				"maxmemory-policy":       "volatile-lru",
				"notify-keyspace-events": "Ex",
			},
		}

		updateMask := &fieldmaskpb.FieldMask{
			Paths: []string{"node_type", "redis_configs", "replica_count", "shard_count"},
		}

		assert.NoError(t, UpdateMask(cluster, update, updateMask))

		// These fields should be updated (listed in updateMask)
		assert.Equal(t, clusterpb.NodeType_REDIS_HIGHMEM_MEDIUM, cluster.NodeType)
		assert.Equal(t, int32(2), cluster.GetReplicaCount())
		assert.Equal(t, int32(5), cluster.GetShardCount())
		assert.Equal(t, "volatile-lru", cluster.RedisConfigs["maxmemory-policy"])
		assert.Equal(t, "Ex", cluster.RedisConfigs["notify-keyspace-events"])

		// These fields should remain unchanged (not listed in updateMask)
		assert.Equal(t, "redis-cluster-name", cluster.Name)
	})

}
