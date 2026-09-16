# Pattern 3: Normalize Status, Keep Provider Spec

Provider-specific sizing vocabulary in spec is correct — the user must choose the provider's tier or instance type. The controller normalizes output to provider-neutral status fields so workloads can consume connection details without provider knowledge.

---

## Where This Pattern Comes From

### Crossplane Managed Resources
**Repo:** [crossplane/crossplane](https://github.com/crossplane/crossplane)
**Key file:** [`docs/concepts/managed-resources.md`](https://github.com/crossplane/crossplane/blob/main/docs/concepts/managed-resources.md)

Managed Resources mirror the cloud API in spec. The controller populates `status.atProvider` with normalized output fields (endpoint, port, ARN) that workloads consume without knowing the provider.

```yaml
kind: RDSInstance
spec:
  forProvider:
    dbInstanceClass: db.t3.micro    # AWS-specific — user must choose this
    engine: postgres
    engineVersion: "14"

status:
  atProvider:
    endpoint: "my-db.abc123.us-east-1.rds.amazonaws.com"
    port: 5432                       # same field name regardless of AWS/GCP/Azure
```

**Lesson for CM:** Keep provider-specific sizing in spec (user must choose). Map to neutral output fields in status. CM already does this correctly for Redis tiers — `status.memorySizeGb`, `status.replicaCount`, `status.primaryEndpoint` are normalized across all providers.

---

### Kubernetes StorageClass → PersistentVolume
**Repo:** [kubernetes/api](https://github.com/kubernetes/api)
**Key file:** [`storage/v1/types.go`](https://github.com/kubernetes/api/blob/master/storage/v1/types.go)

`StorageClass.parameters` carries provider-specific config (IOPS, encryption, disk type). The provisioned `PV` exposes neutral fields (`capacity`, `accessModes`, `volumeMode`) regardless of which provisioner created it.

```yaml
# StorageClass — provider-specific input
kind: StorageClass
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"              # AWS-specific

---
# PV — normalized output, same shape for all provisioners
kind: PersistentVolume
spec:
  capacity:
    storage: 10Gi           # neutral
  accessModes: [ReadWriteOnce]
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: "vol-0a1b2c3d"
```

**Lesson for CM:** The same normalization applies to Redis: provider-specific tier in spec, neutral `memorySizeGb` in status.

---

## How Cloud Manager Uses This Pattern

**Redis status normalization — already correct:**

```yaml
kind: GcpRedisInstance
spec:
  redisTier: "P1"               # GCP-specific — user must choose this

status:
  memorySizeGb: 6               # neutral — same field on all providers
  replicaCount: 1               # neutral
  primaryEndpoint: "10.0.0.5:6379"   # neutral
  authString: "secret"          # neutral
```

No change needed here.

---

## Applying to CM CRD Families

### Redis — engineVersion field name

The version concept is the same user decision on every provider. The field name and value format differ across providers for no reason.

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
                                 # controller maps to Azure's replicasPerPrimary internally
```
