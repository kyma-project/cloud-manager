package v1beta1

// AlicloudRedisClusterTier defines the per-shard capacity tier for an
// AlicloudRedisCluster. AliCloud regions use a proxy-based sharded architecture
// (redis.logic.sharding.*); the instance class is derived from the tier and
// shardCount at creation time:
//
//	C3 →  4 GB per shard  (redis.logic.sharding.4g.{N}db.0rodb.{P}proxy.default)
//	C4 →  8 GB per shard  (redis.logic.sharding.8g.{N}db.0rodb.{P}proxy.default)
//	C5 → 16 GB per shard  (redis.logic.sharding.16g.{N}db.0rodb.{P}proxy.default)
//	C6 → 32 GB per shard  (redis.logic.sharding.32g.{N}db.0rodb.{P}proxy.default)
//	C7 → 64 GB per shard  (redis.logic.sharding.64g.{N}db.0rodb.{P}proxy.default)
//
// where N = shardCount and P = max(4, shardCount).
//
// Total cluster capacity = per-shard memory × shardCount.
//
// C tiers start at C3 (4 GB/shard) to align with the minimum useful cluster
// size, matching the Azure Managed Redis convention where C tiers start above
// the smallest shard size.
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
