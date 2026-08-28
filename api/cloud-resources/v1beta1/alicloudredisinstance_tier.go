package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// All classes use the Tair cloud-disk family (tair.rdb.*g) which supports
// engine versions 5.0, 6.0, and 7.0.
//
// The `stand` (4 GB) AliCloud class has no cross-provider equivalent and is
// skipped so that S3-S5 align with GCP/AWS/Azure baselines:
//
//	S1 →   1 GB  tair.rdb.1g
//	S2 →   2 GB  tair.rdb.2g
//	S3 →   4 GB  tair.rdb.4g
//	S4 →   8 GB  tair.rdb.8g
//	S5 →  16 GB  tair.rdb.16g
//	             standard HA (master+replica), no read-only replica
//
//	P1 →   4 GB  tair.rdb.4g
//	P2 →   8 GB  tair.rdb.8g
//	P3 →  16 GB  tair.rdb.16g
//	P4 →  32 GB  tair.rdb.32g
//	P5 →  64 GB  tair.rdb.64g
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
