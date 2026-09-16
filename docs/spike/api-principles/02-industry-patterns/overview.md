# Industry Patterns — Applied to Cloud Manager CRD API

Each pattern includes: where it comes from, what it solves, and a concrete before/after for a Cloud Manager CRD.

## Patterns

| # | Pattern | Applies to |
|---|---------|------------|
| [1](./01-base-with-extensions.md) | **Common base + provider-specific extensions** | All CM CRD families |
| 2 | Neutral intent resource | IpRange (already correct), NfsVolume capacity |
| 3 | Normalize status, keep provider spec | Redis version/replicas field naming |
| 4 | Provider-specific resource (ACK pattern) | VpcPeering (correct as-is) |
| 5 | Typed provider sub-struct | Redis config map field naming |
| 6 | Portable container with unstructured payload | Future WafPolicy rules |

---

## Pattern 2: Neutral Intent Resource

**Origin:** Kubernetes Storage (`PersistentVolumeClaim`) · [k8s.io/api/core/v1](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L533)
and Gardener DNS (`DNSEntry`) · [gardener/external-dns-management](https://github.com/gardener/external-dns-management/blob/master/pkg/apis/dns/v1alpha1/dnsentry.go)

**What it solves:** User expresses *what* they need in provider-neutral vocabulary. The controller routes to the right backend at runtime. The user never writes a provider name.

**PVC example (Kubernetes):**
```yaml
# User writes this — no mention of AWS EBS, GCP PD, or Azure Disk
kind: PersistentVolumeClaim
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

**Gardener DNSEntry example:**
```yaml
# User writes this — no mention of Route53, Cloud DNS, or Azure DNS
kind: DNSEntry
spec:
  dnsName: "my-service.example.com"
  ttl: 120
  targets: ["1.2.3.4"]
```

**Cloud Manager today — `IpRange` already follows this pattern correctly:**
```yaml
# Works on AWS, GCP, Azure, Alicloud, OpenStack — user writes nothing provider-specific
kind: IpRange
spec:
  cidr: "10.250.0.0/22"
```

**Cloud Manager today — `NfsVolume` partially follows it but breaks on capacity units:**

*Before (inconsistent):*
```yaml
# GCP and SAP use integer capacityGb
kind: GcpNfsVolume
spec:
  capacityGb: 1024        # integer, GCP API unit

---
# AWS and Alicloud use Kubernetes Quantity
kind: AwsNfsVolume
spec:
  capacity: "1Ti"         # Kubernetes resource.Quantity
```

*After (pattern applied — Kubernetes Quantity everywhere, controller converts):*
```yaml
kind: GcpNfsVolume
spec:
  capacity: "1Ti"         # same field name, same type as every other K8s storage resource

kind: AwsNfsVolume
spec:
  capacity: "1Ti"         # unchanged — already correct

kind: SapNfsVolume
spec:
  capacity: "100Gi"       # was: capacityGb: 100
```

The controller converts `capacity` to the integer GiB value the cloud API expects. Users learn one field name from PVC documentation and it works everywhere.

---

## Pattern 2: Normalize Status, Keep Provider Spec

**Origin:** Crossplane Managed Resources · [crossplane/crossplane](https://github.com/crossplane/crossplane/blob/main/docs/concepts/managed-resources.md)
and Kubernetes StorageClass `parameters` → PV `spec`

**What it solves:** Provider-specific sizing vocabulary in spec (the user must choose the provider's unit), but the controller maps it to neutral output fields in status that workloads can consume without provider knowledge.

**Crossplane example — user writes provider-specific, status is normalized:**
```yaml
# User spec: provider-specific
kind: RDSInstance
spec:
  forProvider:
    dbInstanceClass: db.t3.micro     # AWS-specific instance class

# Status: normalized by controller
status:
  atProvider:
    endpoint: "my-db.abc123.us-east-1.rds.amazonaws.com"
    port: 5432
```

**Cloud Manager today — Redis already does this correctly:**
```yaml
# User spec: provider-specific tier (correct — user must choose)
kind: GcpRedisInstance
spec:
  redisTier: "P1"          # GCP tier name — user needs to know this

# Status: normalized (correct — workload reads neutral fields)
status:
  memorySizeGb: 6
  replicaCount: 1
  primaryEndpoint: "10.0.0.5:6379"
  authString: "secret"
```

**Where Cloud Manager breaks this pattern — version field naming:**

*Before (inconsistent — same user decision, three different field names and value formats):*
```yaml
kind: AwsRedisInstance
spec:
  engineVersion: "7.0"        # dotted numeric

kind: GcpRedisInstance
spec:
  redisVersion: "REDIS_7_0"   # GCP enum constant

kind: AzureRedisInstance
spec:
  redisVersion: "6.0"         # dotted numeric, different field name from AWS
```

*After (pattern applied — one field name, controller maps to provider format):*
```yaml
kind: AwsRedisInstance
spec:
  engineVersion: "7.0"        # unchanged

kind: GcpRedisInstance
spec:
  engineVersion: "7.0"        # was: redisVersion: "REDIS_7_0"
                               # controller maps "7.0" → "REDIS_7_0" internally

kind: AzureRedisInstance
spec:
  engineVersion: "7.0"        # was: redisVersion: "6.0"
```

**Where Cloud Manager breaks this pattern — replica field naming in clusters:**

*Before (same concept, different names):*
```yaml
kind: AwsRedisCluster
spec:
  replicasPerShard: 1      # AWS, GCP, Alicloud

kind: AzureRedisCluster
spec:
  replicasPerPrimary: 1    # Azure only — same concept
```

*After:*
```yaml
kind: AzureRedisCluster
spec:
  replicasPerShard: 1      # unified — controller maps to Azure's replicasPerPrimary
```

---

## Pattern 3: Provider-Specific Resource (ACK / ASO pattern)

**Origin:** AWS Controllers for Kubernetes (ACK) · [aws-controllers-k8s/community](https://github.com/aws-controllers-k8s/community)
and Azure Service Operator (ASO) · [Azure/azure-service-operator](https://github.com/Azure/azure-service-operator)

**What it solves:** When concepts differ structurally across providers — not just in vocabulary — forcing a portable wrapper either drops fidelity or produces a union type where most fields are meaningless for any given provider. Provider-specific resources are the honest answer.

**ACK example — no abstraction, CRD mirrors cloud API:**
```yaml
kind: ElasticacheReplicationGroup    # AWS-specific kind
spec:
  replicationGroupID: my-redis
  replicationGroupDescription: "my cluster"
  cacheNodeType: cache.r6g.large     # AWS instance type
  numNodeGroups: 3                   # AWS term for shards
  replicasPerNodeGroup: 1
```

**Cloud Manager — VpcPeering correctly uses this pattern:**

```yaml
# Three separate resources — not a design flaw, a deliberate choice
kind: AwsVpcPeering
spec:
  remoteVpcId: "vpc-0a1b2c3d"
  remoteRegion: "us-east-1"
  remoteAccountId: "123456789012"     # cross-account identity — no GCP/Azure equivalent
  remoteRouteTableUpdateStrategy: "AUTO"  # AWS route propagation concept

kind: GcpVpcPeering
spec:
  remoteVpc: "my-remote-vpc"
  remoteProject: "my-gcp-project"     # GCP project — no AWS/Azure equivalent
  importCustomRoutes: false

kind: AzureVpcPeering
spec:
  remoteVnet: "/subscriptions/.../virtualNetworks/my-vnet"   # ARM resource ID
  remoteTenant: "00000000-..."        # cross-tenant AAD — no AWS/GCP equivalent
  useRemoteGateway: false
```

These are correct as-is. A hypothetical `VpcPeering` with a `provider` field and a union spec would produce a resource where `remoteAccountId` only makes sense on AWS, `remoteProject` only on GCP, and `remoteTenant` only on Azure — a worse API with more confusion, not less.

**When to apply this pattern vs. Pattern 1:**
- Apply Pattern 1 when the user decision is the same across providers (capacity, version, auth).
- Apply Pattern 3 when the user decision is inherently provider-specific (peering identity model, network attachment type, routing strategy).

---

## Pattern 4: Typed Provider Sub-Struct (Crossplane Composition pattern)

**Origin:** Crossplane Composite Resources · [crossplane/crossplane](https://github.com/crossplane/crossplane/blob/main/docs/concepts/composite-resources.md)
and Cluster API infrastructure provider split · [kubernetes-sigs/cluster-api](https://github.com/kubernetes-sigs/cluster-api/blob/main/docs/book/src/developer/providers/cluster-infrastructure.md)

**What it solves:** When the resource envelope is shared (create, reference, lifecycle, status contract) but some configuration is provider-specific, a typed `spec.instance.<provider>` sub-struct keeps the user-facing resource tidy while surfacing only relevant fields per provider.

**Crossplane example — neutral Claim, provider detail in Composition:**
```yaml
# User writes (Claim — neutral):
kind: XPostgreSQLInstance
spec:
  parameters:
    storageGB: 20
    version: "14"

# Composition maps to provider-specific Managed Resource (internal, not user-facing):
# storageGB: 20 → spec.forProvider.allocatedStorage: 20 (AWS RDS)
# version: "14" → spec.forProvider.engineVersion: "14.9" (AWS RDS)
```

**Cloud Manager — KCP already uses this pattern correctly:**
```yaml
# KCP RedisInstance — envelope is shared, provider detail is in sub-struct
kind: RedisInstance
spec:
  instance:
    gcp:
      memorySizeGb: 6
      tier: "STANDARD_HA"
      redisVersion: "REDIS_7_0"
    # aws:, azure:, alicloud: are mutually exclusive alternatives
```

**Cloud Manager — where this pattern should be extended to config maps:**

Redis config is freeform on AWS/GCP but typed on Azure. Currently Azure `redisConfiguration` is a typed struct — which is correct — but the field *name* differs from `parameters` (AWS) and `redisConfigs` (GCP).

*Before:*
```yaml
kind: AwsRedisInstance
spec:
  parameters:
    maxmemory-policy: volatile-lru    # freeform map

kind: GcpRedisInstance
spec:
  redisConfigs:
    maxmemory-policy: volatile-lru    # freeform map, different name

kind: AzureRedisInstance
spec:
  redisConfiguration:
    maxmemory-policy: "volatile-lru"  # typed struct, different name again
```

*After (unified field name, structural difference stays):*
```yaml
kind: AwsRedisInstance
spec:
  parameters:                         # unchanged
    maxmemory-policy: volatile-lru

kind: GcpRedisInstance
spec:
  parameters:                         # was: redisConfigs
    maxmemory-policy: volatile-lru

kind: AzureRedisInstance
spec:
  parameters:                         # was: redisConfiguration — struct shape preserved,
    maxmemory-policy: "volatile-lru"  # only field name changes
```

---

## Pattern 5: Portable Container with Unstructured Payload

**Origin:** Kubernetes `ConfigMap` / `Secret` (`data: map[string]string`) and
StorageClass `parameters: map[string]string` · [k8s.io/api/storage/v1](https://github.com/kubernetes/api/blob/master/storage/v1/types.go)

**What it solves:** When the resource envelope is provider-neutral (create, reference, lifecycle) but the payload content — rule sets, policies, scripts — has no meaningful cross-provider representation. An unstructured `data` field hands off validation to the provider while keeping the CM resource shape consistent.

**StorageClass example:**
```yaml
kind: StorageClass
provisioner: ebs.csi.aws.com
parameters:                    # opaque — provisioner validates, not Kubernetes
  type: gp3
  iops: "3000"
  encrypted: "true"
```

**Cloud Manager — future WAF resource (clean test case):**

WAF rule sets are provider-native and content-rich. An AWS `AWSManagedRulesCommonRuleSet` has no direct structural equivalent on GCP or Azure. Forcing a portable schema would require either a lowest-common-denominator set of toggles or a large union with mostly-empty fields.

*Proposed shape using this pattern:*
```yaml
kind: WafPolicy
spec:
  targetRef:
    name: my-gateway            # what to protect (provider-neutral reference)
  # Common portable toggles (intersection of all providers):
  rules:
    owaspTop10: true
    rateLimit:
      requestsPerMinute: 1000
  # Provider-specific payload — content validated by cloud provider, not CM:
  instance:
    aws:
      managedRuleGroups:
        - vendorName: AWS
          name: AWSManagedRulesCommonRuleSet
    gcp:
      preconfiguredRules:
        - OWASP-CRS
    azure:
      managedRuleSets:
        - ruleSetType: OWASP
          ruleSetVersion: "3.2"
```

This keeps the CM envelope consistent and the shared config validated by CRD schema, while allowing each provider's native rule format without a false abstraction.

---

## Summary: Pattern → CM CRD family mapping

| Pattern | Source | Apply to |
|---------|--------|----------|
| Neutral intent resource | PVC, Gardener DNSEntry | `NfsVolume` capacity field (unify to k8s Quantity) |
| Normalize status, keep provider spec | Crossplane Managed Resources | Redis `engineVersion` and `replicasPerShard` naming (unify field names, map internally) |
| Provider-specific resource | ACK / ASO | `VpcPeering` — correct as-is, no change |
| Typed provider sub-struct | Crossplane Composition, Cluster API | Redis `parameters` field name (unify name, keep structural differences) |
| Portable container + unstructured payload | StorageClass `parameters` | Future `WafPolicy` rule configuration |
