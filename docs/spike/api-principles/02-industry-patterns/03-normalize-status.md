# Pattern 3: Normalize Status, Keep Provider Spec

**Origin:**
- Crossplane Managed Resources · [crossplane/crossplane — managed-resources.md](https://github.com/crossplane/crossplane/blob/main/docs/concepts/managed-resources.md)
- Kubernetes StorageClass `parameters` → PV `spec` (provisioner normalizes)

**What it solves:** Provider-specific sizing vocabulary in spec is correct — the user must choose the provider's tier or instance type. But the controller normalizes output to provider-neutral status fields so workloads can consume connection details without provider knowledge.

---

## Reference example

**Crossplane RDS — provider-specific spec, neutral status:**
```yaml
# User spec: AWS-specific
kind: RDSInstance
spec:
  forProvider:
    dbInstanceClass: db.t3.micro    # AWS instance class — user must know this

# Status: normalized by controller
status:
  atProvider:
    endpoint: "my-db.abc123.us-east-1.rds.amazonaws.com"
    port: 5432
    # workload reads endpoint+port — same fields regardless of AWS/GCP/Azure
```

---

## Cloud Manager — Redis status normalization (already correct)

CM already does this correctly for Redis tiers and sizing:

```yaml
kind: GcpRedisInstance
spec:
  redisTier: "P1"               # GCP-specific — user must choose this

status:
  memorySizeGb: 6               # neutral — same field on all providers
  replicaCount: 1               # neutral
  primaryEndpoint: "10.0.0.5:6379"  # neutral
  authString: "secret"          # neutral
```

No change needed here.

---

## Cloud Manager — Redis version field naming (needs fix)

The version concept is the same user decision on every provider — which Redis engine version to run. The field name and value format differ across providers for no reason.

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
  redisVersion: "6.0"           # dotted numeric — same format as AWS but different field name
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

## Cloud Manager — replicasPerShard vs replicasPerPrimary (needs fix)

Same concept — how many read replicas per shard — with different field names on Azure.

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
                                 # controller maps to Azure's replicasPerPrimary internally
```
