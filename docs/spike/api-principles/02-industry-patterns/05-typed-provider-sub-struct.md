# Pattern 5: Typed Provider Sub-Struct

**Origin:**
- Crossplane Composite Resources · [crossplane/crossplane — composite-resources.md](https://github.com/crossplane/crossplane/blob/main/docs/concepts/composite-resources.md)
- Cluster API infrastructure split · [kubernetes-sigs/cluster-api](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go)

**What it solves:** When the resource envelope is shared (lifecycle, status contract, common references) but some configuration is genuinely provider-specific. A typed `spec.instance.<provider>` sub-struct keeps the resource tidy — users only see fields relevant to their provider — while CRD-level validation enforces exactly one sub-struct is set.

---

## Reference examples

**Crossplane Composition — neutral Claim, typed provider mapping:**
```yaml
# User writes (Claim — neutral):
kind: XPostgreSQLInstance
spec:
  parameters:
    storageGB: 20
    version: "14"
    tier: standard

# Composition maps tier to provider-specific values:
# tier=standard → dbInstanceClass=db.t3.small (AWS)
# tier=standard → tier=db-n1-standard-2 (GCP)
```

**Cluster API — base Cluster references typed provider object:**
```yaml
kind: Cluster
spec:
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
    kind: AWSCluster      # typed — not unstructured
    name: my-cluster
```

---

## Cloud Manager — KCP Redis (already correct)

CM's KCP layer already uses this pattern correctly with `MaxProperties=1` enforcing exactly one provider sub-struct:

```yaml
kind: RedisInstance      # cloud-control API
spec:
  scope:
    name: my-scope       # base — same on all providers
  ipRange:
    name: my-iprange     # base — same on all providers
  instance:              # exactly one of the following
    gcp:
      memorySizeGb: 6
      tier: "STANDARD_HA"
      redisVersion: "REDIS_7_0"
    # aws: / azure: / alicloud: are mutually exclusive
```

No change needed at the KCP layer.

---

## Cloud Manager — Redis config field naming at SKR layer (needs fix)

The config map concept is the same on all providers — freeform Redis configuration key-value pairs. Azure's internal struct shape differs (typed vs. freeform), but the field *name* should be consistent. Having `parameters`, `redisConfigs`, and `redisConfiguration` for the same concept forces users to re-learn the field name per provider.

**Before:**
```yaml
kind: AwsRedisInstance
spec:
  parameters:                       # freeform map
    maxmemory-policy: volatile-lru

---
kind: GcpRedisInstance
spec:
  redisConfigs:                     # freeform map — different name
    maxmemory-policy: volatile-lru

---
kind: AzureRedisInstance
spec:
  redisConfiguration:               # typed struct — different name again
    maxmemory-policy: "volatile-lru"
    maxclients: "8"
```

**After:**
```yaml
kind: AwsRedisInstance
spec:
  parameters:                       # unchanged
    maxmemory-policy: volatile-lru

---
kind: GcpRedisInstance
spec:
  parameters:                       # was: redisConfigs
    maxmemory-policy: volatile-lru

---
kind: AzureRedisInstance
spec:
  parameters:                       # was: redisConfiguration
    maxmemory-policy: "volatile-lru"
    maxclients: "8"
    # struct shape (typed fields) preserved — only the field name changes
```

The controller maps `parameters` to the provider's internal representation. Azure's typed validation stays in the action pipeline, not visible to the user as a different field name.
