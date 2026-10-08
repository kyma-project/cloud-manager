package v1beta1

// AlicloudRedisClusterTier defines the per-shard capacity tier for an
// AlicloudRedisCluster. AliCloud uses the cloud-native CE cluster family
// (redis.shard.*.ce); ShardCount is a separate spec field:
//
//	C3 →  1 GB per shard  (redis.shard.small.ce)
//	C4 →  2 GB per shard  (redis.shard.mid.ce)
//	C5 →  4 GB per shard  (redis.shard.large.ce)
//	C6 →  8 GB per shard  (redis.shard.2xlarge.ce)
//	C7 → 16 GB per shard  (redis.shard.4xlarge.ce)
//
// Total cluster capacity = per-shard memory × shardCount.
//
// C tiers start at C3 to align with the minimum useful cluster size,
//
// +kubebuilder:validation:Enum=C3;C4;C5;C6;C7
type AlicloudRedisClusterTier string

const (
	AlicloudRedisClusterTierC3 AlicloudRedisClusterTier = "C3"
	AlicloudRedisClusterTierC4 AlicloudRedisClusterTier = "C4"
	AlicloudRedisClusterTierC5 AlicloudRedisClusterTier = "C5"
	AlicloudRedisClusterTierC6 AlicloudRedisClusterTier = "C6"
	AlicloudRedisClusterTierC7 AlicloudRedisClusterTier = "C7"
)
