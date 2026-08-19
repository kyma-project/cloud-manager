package v1beta1

// AlicloudRedisTier defines the Kyma service tier for an AlicloudRedisInstance.
// The tier letter+number encodes the underlying AliCloud r-kvstore instance
// class and read-only replica count:
//
//	S1 →  1 GB  redis.master.small.default,           ReadOnlyCount=0
//	S2 →  2 GB  redis.master.mid.default,              ReadOnlyCount=0
//	S3 →  4 GB  redis.master.stand.default,            ReadOnlyCount=0
//	S4 →  8 GB  redis.master.large.default,            ReadOnlyCount=0
//	S5 → 16 GB  redis.master.2xlarge.default,          ReadOnlyCount=0
//	            standard HA (master+replica), 80k QPS, engine 5.0
//
//	P1 →  1 GB  redis.amber.master.small.multithread,  ReadOnlyCount=1
//	P2 →  2 GB  redis.amber.master.mid.multithread,    ReadOnlyCount=1
//	P3 →  4 GB  redis.amber.master.stand.multithread,  ReadOnlyCount=1
//	P4 →  8 GB  redis.amber.master.large.multithread,  ReadOnlyCount=1
//	P5 → 16 GB  redis.amber.master.2xlarge.multithread,ReadOnlyCount=1
//	            enterprise HA (master+replica+read-only replica), 240k QPS, engine 5.0
//
// Note: AliCloud P tiers start at 1 GB, whereas on AWS (P1=6 GB), GCP (P1=5 GB),
// and Azure (P1=6 GB) the P tier starts at a higher memory baseline. This is an
// AliCloud platform constraint — the smallest amber.master class available in
// international regions is 1 GB.
//
// Both class families are available in all AliCloud international regions.
// Engine version is constrained to "5.0" — local-disk and amber-multithread
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
