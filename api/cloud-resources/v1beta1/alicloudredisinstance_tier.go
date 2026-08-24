package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// All classes use the cloud-disk family (redis.master.*.cloud) which supports
// engine versions 5.0, 6.0, and 7.0.
//
// The `stand` (4 GB) AliCloud class has no cross-provider equivalent and is
// skipped so that S3-S5 align with GCP/AWS/Azure baselines:
//
//	S1 →   1 GB  redis.master.small.cloud
//	S2 →   2 GB  redis.master.mid.cloud
//	S3 →   8 GB  redis.master.large.cloud
//	S4 →  16 GB  redis.master.2xlarge.cloud
//	S5 →  32 GB  redis.master.4xlarge.cloud
//	             standard HA (master+replica), no read-only replica
//
//	P1 →   8 GB  redis.master.large.cloud
//	P2 →  16 GB  redis.master.2xlarge.cloud
//	P3 →  32 GB  redis.master.4xlarge.cloud
//	P4 →  64 GB  redis.master.8xlarge.cloud
//	P5 → 128 GB  redis.master.16xlarge.cloud
//	             standard HA + 1 read-only replica
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
