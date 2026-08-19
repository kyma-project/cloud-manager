# AliCloud Redis Tier Cross-Check

Cross-provider comparison of Redis tier memory sizes. Used to verify AliCloud tier
alignment with AWS, GCP, and Azure equivalents.

---

## Redis Instance — S Tiers (Standard HA, no read replica)

| Kyma Tier | AliCloud (GB) | AWS (GiB) | GCP (GB) | Azure (GiB) |
|-----------|--------------|-----------|----------|-------------|
| S1        | 1            | 1.37      | 1        | 1           |
| S2        | 2            | 3.09      | 3        | 2.5         |
| S3        | 4            | 6.38      | 6        | 6           |
| S4        | 8            | 12.93     | 12       | 13          |
| S5        | 16           | 26.04     | 24       | 26          |
| S6        | —            | 52.26     | 48       | —           |
| S7        | —            | 103.68    | 101      | —           |
| S8        | —            | 209.55    | 200      | —           |

AliCloud S tiers map to `redis.master.*.default` classes (80k QPS, engine 5.0).

---

## Redis Instance — P Tiers (Premium HA, +1 read replica)

| Kyma Tier | AliCloud (GB) | AWS (GiB) | GCP (GB) | Azure (GiB) |
|-----------|--------------|-----------|----------|-------------|
| P1        | 4            | 6.38      | 5        | 6           |
| P2        | 8            | 12.93     | 12       | 13          |
| P3        | 16           | 26.04     | 24       | 26          |
| P4        | 32           | 52.26     | 48       | 53          |
| P5        | 64           | 103.68    | 101      | 120         |
| P6        | —            | 209.55    | 200      | —           |

AliCloud P tiers map to `redis.amber.master.*.multithread` classes (240k QPS, engine 5.0,
+1 read-only replica). P1 starts at 4 GB (`stand` class) — closest available class to the
cross-provider P1 baseline of ~6 GB. No AliCloud amber.master class exists between 4 GB and 8 GB.

---

## Redis Cluster — C Tiers (per-shard memory)

| Kyma Tier | AliCloud (GB/shard) | AWS (GiB/shard) | GCP (GB/shard) | Azure (GiB/shard) |
|-----------|--------------------|-----------------|-----------------|--------------------|
| C1        | —                  | 1.37            | ~1 (nano)       | —                  |
| C2        | —                  | 3.09            | —               | —                  |
| C3        | 4                  | 6.38            | 6.5             | 6                  |
| C4        | 8                  | 12.93           | 13              | 13                 |
| C5        | 16                 | 26.04           | —               | 26                 |
| C6        | 32                 | 52.26           | 58              | 53                 |
| C7        | 64                 | 103.68          | —               | 160                |
| C8        | —                  | 209.55          | —               | —                  |

AliCloud cluster tiers map to `redis.logic.sharding.{N}g.{shards}db.0rodb.{proxy}proxy.default`
proxy-based classes. C3 starts at 4 GB/shard — closest available class to the cross-provider
C3 baseline of ~6 GB. No proxy-based class exists between 4 GB and 8 GB.

Total cluster capacity = per-shard memory × shardCount (min 1, max 32).

---

## Notes

- AliCloud memory sizes are nominal GB (1 GB = 1000 MB in AliCloud marketing); AWS/GCP/Azure
  report GiB (1 GiB = 1024 MB). Differences of ~7% are expected between providers.
- AliCloud does not expose S6–S8 or C1–C2 equivalents in the tier set. Users needing
  >16 GB single-instance should use cluster mode.
- AliCloud P tiers include a read-only replica (`ReadOnlyCount=1`) backed by
  `redis.amber.master.*` (enterprise-grade, 240k QPS). P4/P5 class names
  (`4xlarge`=32 GB, `8xlarge`=64 GB) follow the AliCloud naming pattern and have
  not been independently verified via `DescribeAvailableResource` — confirm before GA.
