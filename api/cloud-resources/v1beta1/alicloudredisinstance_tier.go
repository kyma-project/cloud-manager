package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// The tier letter+number encodes the underlying AliCloud r-kvstore instance
// class and read-only replica count. All classes use the cloud-disk family
// (redis.master.*.cloud) which supports engine versions 5.0, 6.0, and 7.0.
//
//	S1 →  1 GB  redis.master.small.cloud,    ReadOnlyCount=0
//	S2 →  2 GB  redis.master.mid.cloud,      ReadOnlyCount=0
//	S3 →  4 GB  redis.master.stand.cloud,    ReadOnlyCount=0
//	S4 →  8 GB  redis.master.large.cloud,    ReadOnlyCount=0
//	S5 → 16 GB  redis.master.2xlarge.cloud,  ReadOnlyCount=0
//	            standard HA (master+replica), no read-only replica
//
//	P1 →  4 GB  redis.master.stand.cloud,    ReadOnlyCount=1
//	P2 →  8 GB  redis.master.large.cloud,    ReadOnlyCount=1
//	P3 → 16 GB  redis.master.2xlarge.cloud,  ReadOnlyCount=1
//	P4 → 32 GB  redis.master.4xlarge.cloud,  ReadOnlyCount=1
//	P5 → 64 GB  redis.master.8xlarge.cloud,  ReadOnlyCount=1
//	            standard HA + 1 read-only replica
//
// P tiers start at 4 GB to align with the cross-provider P tier baseline:
// AWS P1=6.38 GB, GCP P1=5 GB, Azure P1=6 GB.
//
// Only the capacity number is mutable; the service letter (S↔P) is immutable
// after creation. EngineVersion is immutable after creation.
//
// The tier→InstanceClass mapping lives in pkg/skr/alicloudredisinstance/util.go.
//
// +kubebuilder:validation:Enum=S1;S2;S3;S4;S5;P1;P2;P3;P4;P5
type AlicloudRedisTier string

const (
	// S - Standard HA, master + replica, no read-only replica.
	AlicloudRedisTierS1 AlicloudRedisTier = "S1"
	AlicloudRedisTierS2 AlicloudRedisTier = "S2"
	AlicloudRedisTierS3 AlicloudRedisTier = "S3"
	AlicloudRedisTierS4 AlicloudRedisTier = "S4"
	AlicloudRedisTierS5 AlicloudRedisTier = "S5"

	// P - Premium HA, master + replica + one read-only replica.
	AlicloudRedisTierP1 AlicloudRedisTier = "P1"
	AlicloudRedisTierP2 AlicloudRedisTier = "P2"
	AlicloudRedisTierP3 AlicloudRedisTier = "P3"
	AlicloudRedisTierP4 AlicloudRedisTier = "P4"
	AlicloudRedisTierP5 AlicloudRedisTier = "P5"
)
