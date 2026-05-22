package v1beta1

// AzureManagedRedisClusterTier defines the Kyma service tier for an
// AzureManagedRedisCluster. All cluster tiers are HA OSSCluster on the
// ComputeOptimized family; the number indicates the size:
//
//	C3 → ComputeOptimized_X5    (smallest)
//	C7 → ComputeOptimized_X100  (largest)
//
// The mapping is owned by the SKR controller (see pkg/skr/azuremanagedrediscluster/util.go).
//
// +kubebuilder:validation:Enum=C3;C4;C5;C6;C7
type AzureManagedRedisClusterTier string

const (
	AzureManagedRedisClusterTierC3 AzureManagedRedisClusterTier = "C3"
	AzureManagedRedisClusterTierC4 AzureManagedRedisClusterTier = "C4"
	AzureManagedRedisClusterTierC5 AzureManagedRedisClusterTier = "C5"
	AzureManagedRedisClusterTierC6 AzureManagedRedisClusterTier = "C6"
	AzureManagedRedisClusterTierC7 AzureManagedRedisClusterTier = "C7"
)
