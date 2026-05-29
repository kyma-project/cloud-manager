package v1beta1

// AzureManagedRedisTier defines the Kyma service tier for an AzureManagedRedis.
// The tier letter+number encodes the underlying Azure Managed Redis SKU,
// high-availability flag, and clustering policy:
//
//	S1-S5 → Balanced family,        non-HA, EnterpriseCluster (dev workloads)
//	P1-P5 → ComputeOptimized family, HA,    EnterpriseCluster (production HA)
//	C3-C7 → ComputeOptimized family, HA,    OSSCluster        (sharded cluster)
//
// The ComputeOptimized SKUs in P and C share identical pricing and capacity;
// the only difference is the clustering policy applied at the database level.
//
// The mapping is owned by the SKR controller (see pkg/skr/azuremanagedredis/util.go).
//
// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;P1;P2;P3;P4;P5;C3;C4;C5;C6;C7
type AzureManagedRedisTier string

const (
	// S — Balanced, non-HA, EnterpriseCluster (dev/test).
	AzureManagedRedisTierS1 AzureManagedRedisTier = "S1"
	AzureManagedRedisTierS2 AzureManagedRedisTier = "S2"
	AzureManagedRedisTierS3 AzureManagedRedisTier = "S3"
	AzureManagedRedisTierS4 AzureManagedRedisTier = "S4"
	AzureManagedRedisTierS5 AzureManagedRedisTier = "S5"

	// P — ComputeOptimized, HA, EnterpriseCluster (production HA, no sharding).
	AzureManagedRedisTierP1 AzureManagedRedisTier = "P1"
	AzureManagedRedisTierP2 AzureManagedRedisTier = "P2"
	AzureManagedRedisTierP3 AzureManagedRedisTier = "P3"
	AzureManagedRedisTierP4 AzureManagedRedisTier = "P4"
	AzureManagedRedisTierP5 AzureManagedRedisTier = "P5"

	// C — ComputeOptimized, HA, OSSCluster (production HA + Redis Cluster sharding).
	AzureManagedRedisTierC3 AzureManagedRedisTier = "C3"
	AzureManagedRedisTierC4 AzureManagedRedisTier = "C4"
	AzureManagedRedisTierC5 AzureManagedRedisTier = "C5"
	AzureManagedRedisTierC6 AzureManagedRedisTier = "C6"
	AzureManagedRedisTierC7 AzureManagedRedisTier = "C7"
)
