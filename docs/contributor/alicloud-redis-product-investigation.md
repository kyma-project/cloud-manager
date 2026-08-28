# AliCloud Redis Product Investigation

**Date:** 2026-08-27  
**Author:** Stefan Kikic  
**Purpose:** Determine the correct instance class families and engine versions for `AlicloudRedisInstance` (HA) and `AlicloudRedisCluster` (sharded) resources across AliCloud regions.

All findings are based on live AliCloud SDK calls (`r-kvstore` API) — actual `CreateInstance` attempts, `DescribeInstanceAttribute` responses, and error codes. No documentation assumptions.

---

## Decision Record

### Decision 1 — Use `tair.rdb.*g` for `AlicloudRedisInstance`

**Chosen:** `tair.rdb.1g` / `tair.rdb.2g` / `tair.rdb.4g` / `tair.rdb.8g` / `tair.rdb.16g` / `tair.rdb.32g` / `tair.rdb.64g`  
**Rejected:** `redis.master.*.cloud`, `redis.master.*.default`  
**Engine:** 7.0

**Why:**
- `redis.master.*.cloud` and `redis.master.*.default` are **local-disk** classes despite the `.cloud` suffix. AliCloud returns `EngineVersion.NotSupportOnLocalDisk` for engine 7.0 on all local-disk classes in every tested region. Max supported engine on these classes is **5.0**.
- `tair.rdb.*` is AliCloud's cloud-disk HA family. Confirmed available in `ap-northeast-1` and `ap-southeast-1` with engine 7.0 via successful `CreateInstance` calls. Specs verified via `DescribeInstanceAttribute`: standard architecture, double node (HA), 1 GB–32 GB range.
- Engine 7.0 is the current production default across other cloud providers (AWS, GCP, Azure) in cloud-manager. AliCloud must be consistent.

**Evidence:** `CreateInstance` with `redis.master.small.cloud` + engine `7.0` → `ErrorCode: EngineVersion.NotSupportOnLocalDisk` in `ap-northeast-1`, `ap-southeast-1`, `eu-central-1`. Same with `redis.master.small.default`. `CreateInstance` with `tair.rdb.1g` + engine `7.0` → `InstanceId` returned successfully in `ap-northeast-1` and `ap-southeast-1`.

---

### Decision 2 — Use `redis.shard.*.ce` for `AlicloudRedisCluster`

**Chosen:** `redis.shard.small.ce` / `redis.shard.mid.ce` / `redis.shard.large.ce` / `redis.shard.2xlarge.ce` / `redis.shard.4xlarge.ce`  
**Rejected:** `redis.logic.sharding.*proxy.*default`  
**Engine:** 7.0

**Why:**
- `redis.logic.sharding.*proxy.*default` is a **local-disk proxy cluster**. AliCloud returns `EngineVersion.NotSupportOnLocalDisk` for engine 7.0. Max supported engine is **5.0**.
- `redis.shard.*.ce` is the cloud-native cluster family. Confirmed available in `ap-northeast-1` and `ap-southeast-1` with engine 7.0 via successful `CreateInstance` calls. `ShardCount` is passed as a separate parameter — the class name does not encode shard count, making it simpler and less fragile than the proxy class format.
- Internally, `redis.shard.*.ce` instances return `ArchitectureType: cluster` and `redis.cluster.sharding.common.ce` as the resolved class name. `ShardCount` and per-shard memory are both reflected correctly in `DescribeInstanceAttribute`.

**Evidence:** `CreateInstance` with `redis.logic.sharding.4g.4db.0rodb.4proxy.default` + engine `7.0` → `ErrorCode: EngineVersion.NotSupportOnLocalDisk` in all regions. `CreateInstance` with `redis.shard.small.ce` + engine `7.0` + `ShardCount=3` → `InstanceId` returned successfully in `ap-northeast-1` and `ap-southeast-1`.

---

### Decision 3 — Use `ap-northeast-1` (Tokyo) as the e2e test region

**Chosen:** `ap-northeast-1`  
**Rejected for e2e:** `eu-central-1`

**Why:**
- `eu-central-1` is the Gardener seed region (`ali-ha-eu1`). Both `tair.rdb.*` and `redis.shard.*.ce` return `InsufficientResourceCapacity` there — AliCloud has no available nodes for these class families in Frankfurt.
- `ap-northeast-1` has both `tair.rdb.*` and `redis.shard.*.ce` available with engine 7.0.
- `eu-central-1` should not be used for e2e because the modern class families are supply-constrained there. The Gardener seed managing this region happens to be in Frankfurt, but the shoot region for tests must be somewhere with actual capacity.

---

## Summary

| Resource | Recommended Class Family | Engine | Notes |
|----------|--------------------------|--------|-------|
| `AlicloudRedisInstance` (HA) | `tair.rdb.*g` | 7.0 | Cloud-disk, HA (double node), available in ap-northeast-1 and ap-southeast-1 |
| `AlicloudRedisCluster` (sharded) | `redis.shard.*.ce` | 7.0 | Cloud-native cluster, all regions tested |

---

## 1. Regions Tested

| Region | Location | Notes |
|--------|----------|-------|
| `ap-northeast-1` | Tokyo | Older region, local-disk only for standard HA |
| `ap-southeast-1` | Singapore | Newer region, cloud-disk (tair.rdb) available |
| `eu-central-1` | Frankfurt | Gardener seed region; tair.rdb + redis.shard.ce have no capacity |

---

## 2. Instance Class Families — Standard HA

### 2.1 `redis.master.*.default` and `redis.master.*.cloud`

Despite the `.cloud` suffix, both behave identically — local disk, max engine **5.0**.

- `EngineVersion.NotSupportOnLocalDisk` on 7.0 in all tested regions
- `redis.master.*.cloud` additionally returns `InvalidCapacity.NotFound` in most regions

**Size variants:** small (1 GB), mid (2 GB), large (4 GB), 2xlarge (8 GB), 4xlarge (16 GB)

**Verdict:** ❌ Max engine 5.0 only. Do not use if 7.0 is required.

---

### 2.2 `tair.rdb.*g` — Cloud-Disk HA ✅ RECOMMENDED

AliCloud's modern cloud-disk based Redis-compatible HA product. Supports engine 7.0.

**Availability:**
- `ap-northeast-1` ✅ engine 5.0 + 7.0 confirmed
- `ap-southeast-1` ✅ engine 5.0 + 7.0 confirmed
- `eu-central-1` ⚠️ `InsufficientResourceCapacity` (class exists, no supply in Frankfurt)

**Specs (per instance, engine 7.0, `ap-southeast-1`, confirmed via `DescribeInstanceAttribute`):**

| Class | Memory | QPS | Connections | Bandwidth | Architecture | Node Type |
|-------|--------|-----|-------------|-----------|--------------|-----------|
| `tair.rdb.1g` | 1,024 MB | 300,000 | 30,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.2g` | 2,048 MB | 300,000 | 30,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.4g` | 4,096 MB | 300,000 | 40,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.8g` | 8,192 MB | 300,000 | 40,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.16g` | 16,384 MB | 300,000 | 40,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.32g` | 32,768 MB | 300,000 | 50,000 | 96 MB/s | standard | double (HA) |
| `tair.rdb.64g` | 65,536 MB | 300,000 | 50,000 | 96 MB/s | standard | double (HA) |

`ArchitectureType: standard`, `NodeType: double` — confirmed primary + replica HA.  
`ShardCount: 1`, `ReplicaCount: 1`.

**Verdict:** ✅ Use for all S-tier and P-tier HA instances.

---

### 2.3 `r.single.*`, `r.ha.*`

All returned `InvalidCapacity.NotFound` in all tested regions.

**Verdict:** ❌ Not available.

---

## 3. Cluster Class Families

### 3.1 `redis.shard.*.ce` — Cloud-Native Cluster ✅ RECOMMENDED

**Important:** `redis.shard.small.ce` / `redis.shard.large.ce` etc. are **input aliases**. The API returns the actual class as `redis.cluster.sharding.common.ce`. The per-shard memory is determined by the input alias, ShardCount is passed separately.

**Availability:**
- `ap-northeast-1` ✅ engine 5.0 + 7.0
- `ap-southeast-1` ✅ engine 5.0 + 7.0
- `eu-central-1` ⚠️ `InsufficientResourceCapacity`

**Specs (3 shards, engine 7.0, `ap-southeast-1`, confirmed via `DescribeInstanceAttribute`):**

| Input Class | Memory/shard | Total (3 shards) | QPS | Connections | Bandwidth | Architecture |
|-------------|-------------|-----------------|-----|-------------|-----------|--------------|
| `redis.shard.small.ce` | 1,024 MB | 3,072 MB | 300,000 | 30,000 | 144 MB/s | cluster |
| `redis.shard.mid.ce` | 2,048 MB | 6,144 MB | 300,000 | 30,000 | 288 MB/s | cluster |
| `redis.shard.large.ce` | 4,096 MB | 12,288 MB | 300,000 | 60,000 | 288 MB/s | cluster |

All resolve to `redis.cluster.sharding.common.ce` internally.

**Verdict:** ✅ Use for all cluster tiers. Supports engine 7.0. Simpler than proxy classes.

---

### 3.2 `redis.logic.sharding.*proxy.*default` — Local-Disk Proxy Cluster

Proxy-based sharded cluster, local disk, max engine **5.0**.

**Availability:** `ap-northeast-1` ✅, `ap-southeast-1` ✅, `eu-central-1` ✅ — available everywhere.

Class name format: `redis.logic.sharding.{mem}g.{N}db.0rodb.{P}proxy.default`  
Example: `redis.logic.sharding.4g.4db.0rodb.4proxy.default`

**Verdict:** Use only if engine 5.0 is acceptable. For 7.0, use `redis.shard.*.ce`.

---

### 3.3 `tair.rdb.cluster.*`

`InvalidCapacity.NotFound` in all tested regions.

**Verdict:** ❌ Not available.

---

## 4. Region Availability Matrix

| Class Family | ap-northeast-1 | ap-southeast-1 | eu-central-1 |
|---|---|---|---|
| `redis.master.*.default` | ✅ ≤5.0 | ✅ ≤5.0 | ✅ ≤5.0 |
| `redis.master.*.cloud` | ❌ NotFound | ❌ NotFound | ❌ NotFound |
| `tair.rdb.*g` (HA) | ✅ 5.0+7.0 | ✅ 5.0+7.0 | ⚠️ No capacity |
| `redis.shard.*.ce` (cluster) | ✅ 5.0+7.0 | ✅ 5.0+7.0 | ⚠️ No capacity |
| `redis.logic.sharding.*` (proxy) | ✅ ≤5.0 | ✅ ≤5.0 | ✅ ≤5.0 |
| `tair.rdb.cluster.*` | ❌ | ❌ | ❌ |

**eu-central-1 note:** `InsufficientResourceCapacity` = class is defined but AliCloud has no available nodes in Frankfurt. Supply constraint, not a product limitation.

---

## 5. Implications for Code

### 5.1 `AlicloudRedisInstance` tier map — `pkg/skr/alicloudredisinstance/util.go`

**Change:** `redis.master.*.cloud` → `tair.rdb.*g`

| Tier | Old Class | New Class | Memory |
|------|-----------|-----------|--------|
| S1 | `redis.master.small.cloud` | `tair.rdb.1g` | 1 GB |
| S2 | `redis.master.mid.cloud` | `tair.rdb.2g` | 2 GB |
| S3 | `redis.master.large.cloud` | `tair.rdb.4g` | 4 GB |
| S4 | `redis.master.2xlarge.cloud` | `tair.rdb.8g` | 8 GB |
| S5 | `redis.master.4xlarge.cloud` | `tair.rdb.16g` | 16 GB |
| P1 | `redis.master.large.cloud` + RO=1 | `tair.rdb.4g` + RO=1 | 4 GB |
| P2 | `redis.master.2xlarge.cloud` + RO=1 | `tair.rdb.8g` + RO=1 | 8 GB |
| P3 | `redis.master.4xlarge.cloud` + RO=1 | `tair.rdb.16g` + RO=1 | 16 GB |
| P4 | `redis.master.8xlarge.cloud` + RO=1 | `tair.rdb.32g` + RO=1 | 32 GB |
| P5 | `redis.master.16xlarge.cloud` + RO=1 | `tair.rdb.64g` + RO=1 | 64 GB |

**`ReadOnlyCount` support on `tair.rdb.*`:** needs verification via CreateInstance with `ReadOnlyCount=1`. The standard HA structure (double node) suggests replica is built in; an additional ReadOnly replica may require a separate API parameter.

### 5.2 `AlicloudRedisCluster` tier map — `pkg/skr/alicloudrediscluster/util.go`

**Current:** `redis.logic.sharding.*proxy.*default` — max engine 5.0  
**If engine 7.0 required:** Switch to `redis.shard.*.ce`

New class format: `redis.shard.{size}.ce` + `ShardCount` parameter — simpler than encoding shard count in the class name.

| Tier | Cluster Class | Per-shard Memory |
|------|--------------|-----------------|
| C3 | `redis.shard.small.ce` | 1 GB/shard |
| C4 | `redis.shard.mid.ce` | 2 GB/shard |
| C5 | `redis.shard.large.ce` | 4 GB/shard |
| C6 | `redis.shard.2xlarge.ce` | 8 GB/shard |
| C7 | `redis.shard.4xlarge.ce` | 16 GB/shard |

### 5.3 e2e Feature Files

**`skr-shared-redis-alicloud.feature`:** After tier map update to `tair.rdb.1g`, change `engineVersion` from `"5.0"` to `"7.0"`.

**`skr-shared-rediscluster-alicloud.feature`:** Already uses `redis.shard.large.ce` which supports 7.0. No change needed.

### 5.4 e2e Test Region

Use `ap-northeast-1` (Tokyo) or `ap-southeast-1` (Singapore).  
Do **not** use `eu-central-1` — both `tair.rdb.*` and `redis.shard.*.ce` have `InsufficientResourceCapacity` there.

---

## 6. Key Findings

1. **`.cloud` suffix does not mean cloud-disk.** `redis.master.*.cloud` uses local disk, max engine 5.0 — identical to `.default`.

2. **Engine 7.0 for HA requires `tair.rdb.*`.** Only cloud-disk HA family confirmed available in Tokyo and Singapore.

3. **`redis.shard.*.ce` is the correct cluster class for engine 7.0.** Internally maps to `redis.cluster.sharding.common.ce`. Input class name controls per-shard memory size; `ShardCount` is separate.

4. **`DescribeAvailableResource` is unreliable** — returns empty in all tested regions regardless of policy. Direct `CreateInstance` probing with error code inspection is the only reliable discovery method.

5. **eu-central-1 supply constraint.** Modern class families exist but have no available nodes. `redis.logic.sharding.*` (local disk, ≤5.0) is the only thing that works reliably in Frankfurt.

6. **All `redis.shard.*.ce` sizes map to the same internal class** (`redis.cluster.sharding.common.ce`). The input alias controls per-shard memory; total memory = shards × per-shard memory.

---

## 7. Addendum — `redis.shard.*.ce` Class Alias Behaviour (discovered during e2e)

**Finding:** All `redis.shard.*.ce` input aliases resolve to the single canonical class name `redis.cluster.sharding.common.ce` in `DescribeInstanceAttribute`. The size (per-shard memory) is stored internally by AliCloud and reflected in the `Capacity` field, not in `InstanceClass`.

| Input alias | Observed InstanceClass | Capacity (3 shards) |
|-------------|----------------------|---------------------|
| `redis.shard.small.ce` | `redis.cluster.sharding.common.ce` | 3,072 MB (1GB/shard) |
| `redis.shard.large.ce` | `redis.cluster.sharding.common.ce` | 12,288 MB (4GB/shard) |

**ModifyInstanceSpec with size aliases works:** Calling `ModifyInstanceSpec` with `redis.shard.large.ce` on a `redis.shard.small.ce` instance successfully changes the per-shard memory (confirmed: 2,048 MB → 8,192 MB for 2-shard instance). The `InstanceClass` field remains `redis.cluster.sharding.common.ce` after the modification.

**Bug fixed:** `modifyInstanceClass` in the KCP cluster reconciler compared `redis.shard.small.ce` (desired, stored in KCP spec) against `redis.cluster.sharding.common.ce` (observed, from API) and always detected drift, causing `InstanceClassDoesNotChange` errors every reconcile. Fixed by adding `ceClusterClassKey()` normalisation that maps any `redis.shard.*.ce` → `redis.cluster.sharding.common.ce` before comparison.
