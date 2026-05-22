package v1beta1

// AzureManagedRedisInstanceTier defines the Kyma service tier for an
// AzureManagedRedisInstance. The tier letter encodes the underlying Azure
// Managed Redis SKU, high-availability, and clustering policy:
//
//	S1-S5 → Balanced family,        non-HA, EnterpriseCluster (dev workloads)
//	P1-P5 → ComputeOptimized family, HA,    EnterpriseCluster (production workloads)
//
// The mapping is owned by the SKR controller (see pkg/skr/azuremanagedredisinstance/util.go).
//
// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;P1;P2;P3;P4;P5
type AzureManagedRedisInstanceTier string

const (
	AzureManagedRedisInstanceTierS1 AzureManagedRedisInstanceTier = "S1"
	AzureManagedRedisInstanceTierS2 AzureManagedRedisInstanceTier = "S2"
	AzureManagedRedisInstanceTierS3 AzureManagedRedisInstanceTier = "S3"
	AzureManagedRedisInstanceTierS4 AzureManagedRedisInstanceTier = "S4"
	AzureManagedRedisInstanceTierS5 AzureManagedRedisInstanceTier = "S5"

	AzureManagedRedisInstanceTierP1 AzureManagedRedisInstanceTier = "P1"
	AzureManagedRedisInstanceTierP2 AzureManagedRedisInstanceTier = "P2"
	AzureManagedRedisInstanceTierP3 AzureManagedRedisInstanceTier = "P3"
	AzureManagedRedisInstanceTierP4 AzureManagedRedisInstanceTier = "P4"
	AzureManagedRedisInstanceTierP5 AzureManagedRedisInstanceTier = "P5"
)
