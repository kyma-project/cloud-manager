# Pattern 3: Consistent Field Naming for Shared Concepts

When the same user decision exists across all providers, the field name must be identical across all provider-specific resources. Value formats that differ between providers are mapped by the controller, not exposed to the user.

This is distinct from Pattern 1 (base+extensions structure) — it applies specifically within provider-specific SKR resources where the field *name* diverges even though the *concept* is the same.

---

## Where This Pattern Comes From

### Crossplane Provider AWS — engineVersion normalization
**Repo:** [crossplane-contrib/provider-aws](https://github.com/crossplane-contrib/provider-aws)
**Key file:** [`apis/cache/v1beta1/replicationgroup_types.go`](https://github.com/crossplane-contrib/provider-aws/blob/master/apis/cache/v1beta1/replicationgroup_types.go)

In Crossplane, a Composition maps a neutral Claim field to a provider Managed Resource. The *mapping* — including value transformation — lives in the Composition, not in the CRD field name. The user writing `engineVersion: "7.0"` always writes the same field name regardless of whether the underlying provider calls it `engineVersion`, `redisVersion`, or `REDIS_7_0`.

```yaml
# User writes once (Claim):
spec:
  engineVersion: "7.0"          # always this field name

# Composition maps to GCP (internal):
# engineVersion: "7.0" → redisVersion: "REDIS_7_0"

# Composition maps to AWS (internal):
# engineVersion: "7.0" → engineVersion: "7.0"  (passthrough)
```

**Lesson for CM:** The field name the user writes is the contract. The mapping to the provider's vocabulary is the controller's job. CM's Go action pipelines can do this mapping as well as Crossplane's YAML patches — the difference is that CM currently does not consistently apply it at the SKR level.

---

### Kubernetes API conventions — consistent field naming
**Repo:** [kubernetes/community](https://github.com/kubernetes/community)
**Key file:** [`contributors/devel/sig-architecture/api-conventions.md`](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#naming-conventions)

The Kubernetes API conventions require that fields representing the same concept use the same name across all resource types. `spec.replicas` means replicas everywhere. `spec.selector` means selector everywhere. This is the baseline convention CM's SKR API should follow for cross-provider fields.

**Lesson for CM:** `replicasPerShard` should mean the same thing everywhere. Using `replicasPerPrimary` on Azure for the same concept breaks this convention without any justification.

---

## How Cloud Manager Uses This Pattern

**Redis status normalization — already correct (do not change):**

```yaml
kind: GcpRedisInstance
spec:
  redisTier: "P1"               # GCP-specific — correct, user must choose this

status:
  memorySizeGb: 6               # neutral — same field on all providers
  replicaCount: 1               # neutral — same field on all providers
  primaryEndpoint: "10.0.0.5:6379"   # neutral — same field on all providers
```

The pattern is already applied correctly in status. The gap is in spec field naming.

---

## Applying to CM CRD Families

### Redis — engineVersion field name

The version concept is identical across all providers — which Redis engine version to run. The field name and value format differ across providers for no reason.

**Before:**
```yaml
kind: AwsRedisInstance
spec:
  engineVersion: "7.0"          # dotted numeric

---
kind: GcpRedisInstance
spec:
  redisVersion: "REDIS_7_0"     # GCP enum constant — different field name, different format

---
kind: AzureRedisInstance
spec:
  redisVersion: "6.0"           # dotted numeric — different field name from AWS
```

**After:**
```yaml
kind: AwsRedisInstance
spec:
  engineVersion: "7.0"          # unchanged

---
kind: GcpRedisInstance
spec:
  engineVersion: "7.0"          # was: redisVersion: "REDIS_7_0"
                                 # controller maps "7.0" → "REDIS_7_0" when calling GCP API

---
kind: AzureRedisInstance
spec:
  engineVersion: "7.0"          # was: redisVersion: "6.0"
```

---

### Redis Cluster — replicasPerShard field name

Same concept — read replicas per shard — with a different field name on Azure only.

**Before:**
```yaml
kind: AwsRedisCluster
spec:
  replicasPerShard: 1           # AWS, GCP, Alicloud

---
kind: AzureRedisCluster
spec:
  replicasPerPrimary: 1         # Azure — same concept, different name
```

**After:**
```yaml
kind: AzureRedisCluster
spec:
  replicasPerShard: 1           # was: replicasPerPrimary
                                 # controller maps to Azure API's replicasPerPrimary internally
```

---

### NfsVolume — status capacity field missing

NfsVolume has no normalized `capacity` field in status. Redis normalizes `memorySizeGb` — the same discipline should apply to NFS.

**Before (no capacity in status):**
```yaml
kind: GcpNfsVolume
status:
  id: "projects/my-project/locations/us-east1/instances/my-nfs"
  state: Ready
  # no capacity field — consumer cannot read back what was provisioned
```

**After:**
```yaml
kind: GcpNfsVolume
status:
  id: "projects/my-project/locations/us-east1/instances/my-nfs"
  state: Ready
  capacity: "1Ti"           # normalized k8s Quantity — same field on all providers
                            # controller converts from provider's integer GiB value
