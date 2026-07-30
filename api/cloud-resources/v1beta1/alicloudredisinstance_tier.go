package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// The tier letter+number encodes the underlying AliCloud r-kvstore instance
// class and read-only replica count:
//
//	S1-S5 → redis.master.*.default, ReadOnlyCount=0
//	        standard HA (master+replica), 80k QPS, engine 5.0
//	P1-P5 → redis.amber.master.*.multithread, ReadOnlyCount=1
//	        enterprise HA (master+replica+read-only replica), 240k QPS, engine 5.0
//
// Both class families are available in all AliCloud international regions.
// Engine version is constrained to "5.0" - local-disk and amber-multithread
// classes do not support 6.0 or 7.0 in international regions.
//
// Both letter (S↔P) and number (1..5) are mutable via ModifyInstanceSpec; no
// recreation is required. EngineVersion is immutable after creation.
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
