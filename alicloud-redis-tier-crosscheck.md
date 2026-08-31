# AliCloud Redis Tier Cross-Check

Cross-provider comparison of Redis tier memory sizes.

**AliCloud class family:** `redis.master.*.cloud` (cloud-disk, engines 5.0/6.0/7.0)
**Size naming:** small=1GB, mid=2GB, stand=4GB, large=8GB, 2xlarge=16GB, 4xlarge=32GB, 8xlarge=64GB
**Max confirmed:** AliCloud standard HA supports up to 64 GB (8xlarge). 16xlarge (128GB) is **unverified**.

The `stand` (4 GB) class has no cross-provider equivalent and is skipped in S tiers.

---

## Redis Instance — S Tiers (Standard HA, no read replica)

| Kyma Tier | AliCloud class            | AliCloud (GB) | AWS (GiB) | GCP (GB) | Azure (GiB) |
|-----------|--------------------------|--------------|-----------|----------|-------------|
| S1        | redis.master.small.cloud |  1           | 1.37      | 1        | 1           |
| S2        | redis.master.mid.cloud   |  2           | 3.09      | 3        | 2.5         |
| S3        | redis.master.large.cloud |  8           | 6.38      | 6        | 6           |
| S4        | redis.master.2xlarge.cloud| 16          | 12.93     | 12       | 13          |
| S5        | redis.master.4xlarge.cloud| 32          | 26.04     | 24       | 26          |
| S6        | —                        |  —           | 52.26     | 48       | —           |
| S7        | —                        |  —           | 103.68    | 101      | —           |
| S8        | —                        |  —           | 209.55    | 200      | —           |

---

## Redis Instance — P Tiers (Standard HA, +1 read-only replica, ReadOnlyCount=1)

| Kyma Tier | AliCloud class             | AliCloud (GB) | AWS (GiB) | GCP (GB) | Azure (GiB) |
|-----------|---------------------------|--------------|-----------|----------|-------------|
| P1        | redis.master.large.cloud  |  8           | 6.38      | 5        | 6           |
| P2        | redis.master.2xlarge.cloud| 16           | 12.93     | 12       | 13          |
| P3        | redis.master.4xlarge.cloud| 32           | 26.04     | 24       | 26          |
| P4        | redis.master.8xlarge.cloud| 64           | 52.26     | 48       | 53          |
| P5        | redis.master.16xlarge.cloud| 128 ⚠️      | 103.68    | 101      | 120         |
| P6        | —                         | —            | 209.55    | 200      | —           |

⚠️ P5 (`16xlarge`, 128 GB) — **not verified** via `DescribeAvailableResource`. AliCloud standard HA
documentation states max 64 GB. Remove or replace P5 if 16xlarge.cloud is not available.

---

## Redis Cluster — C Tiers (per-shard memory, proxy-based)

| Kyma Tier | AliCloud (GB/shard) | AWS (GiB/shard) | GCP (GB/shard) | Azure (GiB/shard) |
|-----------|--------------------|-----------------|-----------------|--------------------|
| C3        | 4                  | 6.38            | 6.5             | 6                  |
| C4        | 8                  | 12.93           | 13              | 13                 |
| C5        | 16                 | 26.04           | —               | 26                 |
| C6        | 32                 | 52.26           | 58              | 53                 |
| C7        | 64                 | 103.68          | —               | 160                |

Total cluster capacity = per-shard memory × shardCount (min 1, max 32).

---

## Notes

- AliCloud GB vs GiB: ~7% difference expected vs AWS/Azure reporting GiB.
- `stand` (4 GB) skipped in S tiers — no cross-provider equivalent.
- S3-S5 and P1-P4 are close to but not exact matches with GCP/AWS/Azure baselines.
- All instance classes need final confirmation via `DescribeAvailableResource` with admin credentials before GA.
