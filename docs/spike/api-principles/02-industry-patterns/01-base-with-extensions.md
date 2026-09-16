# Pattern: Common Base with Provider-Specific Extensions

This is the foundational pattern for Cloud Manager CRD API design. It appears across multiple production Kubernetes projects under different names, but the core idea is always the same: **shared user intent lives in a common base; provider-specific configuration lives in typed extensions.**

---

## Where This Pattern Comes From

### Cluster API (CAPI)
**Repo:** [kubernetes-sigs/cluster-api](https://github.com/kubernetes-sigs/cluster-api)
**Key file:** [`api/v1beta1/cluster_types.go`](https://github.com/kubernetes-sigs/cluster-api/blob/main/api/v1beta1/cluster_types.go)

Base `Cluster` holds provider-agnostic fields. Provider-specific resources (`AWSCluster`, `GCPCluster`, `AzureCluster`) live in separate repos and are referenced via `spec.infrastructureRef`.

```yaml
# What the user writes — base (provider-agnostic)
kind: Cluster
spec:
  clusterNetwork:
    pods:
      cidrBlocks: ["192.168.0.0/16"]
  infrastructureRef:          # typed pointer to provider extension
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta2
    kind: AWSCluster           # or GCPCluster, AzureCluster
    name: my-cluster

---
# Provider extension — user also creates this (CAPI's main cost)
kind: AWSCluster
spec:
  region: "us-east-1"
  network:
    vpc:
      cidrBlock: "10.0.0.0/16"
```

**Lesson for CM:** The base+ref split is correct. CM improves on CAPI by *generating* the provider object (KCP) from the user object (SKR) — the user only creates one resource, not two.

---

### Crossplane Composite Resources
**Repo:** [crossplane/crossplane](https://github.com/crossplane/crossplane)
**Key file:** [`apis/apiextensions/v1/composition_types.go`](https://github.com/crossplane/crossplane/blob/master/apis/apiextensions/v1/composition_types.go)

A `CompositeResourceDefinition` (XRD) defines the provider-agnostic schema (the Claim). A `Composition` maps Claim fields to provider-specific Managed Resource fields via patches and transforms — including value mapping (e.g. `tier: standard` → `cacheNodeType: cache.t3.small`).

```yaml
# User writes (Claim — neutral)
kind: RedisInstanceClaim
spec:
  tier: standard          # neutral concept
  engineVersion: "7.0"   # neutral concept
  memorySizeGb: 5

# Composition maps to AWS (internal, not user-facing)
# tier=standard → cacheNodeType=cache.t3.small
# engineVersion="7.0" → engineVersion="7.0" (passthrough for AWS)
# engineVersion="7.0" → redisVersion="REDIS_7_0" (transform for GCP)
```

**Lesson for CM:** The Composition principle — *user writes the concept, controller maps to provider vocabulary* — is exactly what CM should do for fields like `engineVersion` and `replicasPerShard`. The mapping is compiled Go in CM rather than YAML patches, which is strictly better (type-safe, testable).

---

### Kubernetes Gateway API
**Repo:** [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api)
**Key file:** [`apis/v1/gatewayclass_types.go`](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gatewayclass_types.go)

`GatewayClass` carries the common structure. `spec.parametersRef` is the typed escape hatch pointing to a provider-specific config object. Standard fields (listeners, routes) are on the base; anything provider-specific is in the referenced object.

```yaml
kind: GatewayClass
spec:
  controllerName: "example.com/nginx-controller"
  parametersRef:              # optional typed pointer to provider extension
    group: "gateway.example.com"
    kind: "NginxGatewayConfig"
    name: "nginx-params"
```

**Lesson for CM:** The `parametersRef` pattern is the right model for cases where most users don't need provider-specific tuning, but expert users must be able to reach it. The reference is typed and optional — not a free-form annotation.

---

## How Cloud Manager Uses This Pattern (KCP Layer)

CM already implements this pattern at the KCP layer. The KCP `RedisInstance` has a discriminated union (`spec.instance.gcp`, `spec.instance.aws`, etc.) where common lifecycle fields live at the top level and provider-specific fields live in the sub-struct.

**Key files:**
- [`api/cloud-control/v1beta1/redisinstance_types.go`](https://github.com/kyma-project/cloud-manager/blob/main/api/cloud-control/v1beta1/redisinstance_types.go)

```yaml
# KCP RedisInstance — base fields + provider extension (discriminated union)
# Current code: api/cloud-control/v1beta1/redisinstance_types.go
kind: RedisInstance                    # cloud-control API group
spec:
  scope:
    name: my-scope                     # base — same on all providers
  ipRange:
    name: my-iprange                   # base — same on all providers
  instance:
    gcp:                               # extension — only one of gcp/aws/azure/alicloud
      memorySizeGb: 6
      tier: "STANDARD_HA"
      redisVersion: "REDIS_7_0"        # KCP-internal: uses GCP API value directly

status:
  memorySizeGb: 6                      # normalized — same field regardless of provider
  replicaCount: 1                      # normalized
  primaryEndpoint: "10.0.0.5:6379"    # normalized
```

**The problem is at the SKR layer.** The KCP layer correctly uses provider vocabulary internally (`redisVersion: "REDIS_7_0"`). The SKR layer should expose the neutral concept (`engineVersion: "7.0"`) and let the controller translate. Currently the SKR layer leaks provider vocabulary into the user-facing spec on some resources.

---

## Applying the Pattern to Each CM CRD Family

The rule: fields where **the user decision is the same** across providers belong in the base (same field name, same type, same semantics). Fields where **the user decision is inherently provider-specific** belong in the extension.

---

### Redis Instance + Cluster

#### Fields that belong in the base (user decision is the same)

| Concept | Before (inconsistent) | After (base field) |
|---------|----------------------|-------------------|
| Engine version | `engineVersion: "7.0"` (AWS, Alicloud) · `redisVersion: "REDIS_7_0"` (GCP) · `redisVersion: "6.0"` (Azure) | `engineVersion: "7.0"` everywhere · controller maps to provider format |
| Config map | `parameters` (AWS) · `redisConfigs` (GCP) · `redisConfiguration` typed struct (Azure) | `parameters: map[string]string` everywhere · Azure struct shape preserved internally |
| Replicas per shard | `replicasPerShard` (AWS, GCP, Alicloud) · `replicasPerPrimary` (Azure) | `replicasPerShard` everywhere · controller maps to Azure's field name |
| Auth secret shape | `authSecret.name/labels/annotations` — consistent ✓ | no change needed |
| IpRange reference | `ipRange.name` — consistent ✓ | no change needed |

**Before:**
```yaml
kind: GcpRedisInstance
spec:
  redisTier: "S1"                  # GCP S1 — provider-native tier name
  redisVersion: "REDIS_7_0"        # GCP-specific field name and value format
  redisConfigs:                     # GCP-specific field name
    maxmemory-policy: volatile-lru
  maintenancePolicy:                # GCP-specific structure
    dayOfWeek:
      day: "SATURDAY"
      startTime:
        hours: 15

---
kind: AwsRedisInstance
spec:
  redisTier: "S1"                  # AWS S1 — different capacity mapping than GCP S1
  engineVersion: "7.0"             # different field name from GCP
  parameters:                       # different field name from GCP
    maxmemory-policy: volatile-lru
  preferredMaintenanceWindow: "sat:15:00-sat:16:00"  # different structure from GCP

---
kind: AzureRedisCluster
spec:
  redisTier: "C3"
  replicasPerPrimary: 1            # different field name from AWS/GCP
```

**After:**
```yaml
kind: GcpRedisInstance
spec:
  redisTier: "S1"                  # stays GCP-specific (capacity meaning differs per provider)
  engineVersion: "7.0"             # was: redisVersion: "REDIS_7_0"
  parameters:                       # was: redisConfigs
    maxmemory-policy: volatile-lru
  maintenancePolicy:                # stays GCP-specific (structure genuinely differs)
    dayOfWeek:
      day: "SATURDAY"
      startTime:
        hours: 15

---
kind: AwsRedisInstance
spec:
  redisTier: "S1"                  # stays AWS-specific (capacity meaning differs per provider)
  engineVersion: "7.0"             # unchanged — was already correct
  parameters:                       # unchanged — was already correct
    maxmemory-policy: volatile-lru
  preferredMaintenanceWindow: "sat:15:00-sat:16:00"  # stays AWS-specific

---
kind: AzureRedisCluster
spec:
  redisTier: "C3"
  replicasPerShard: 1              # was: replicasPerPrimary
```

#### Fields that stay provider-specific (user decision genuinely differs)

- `redisTier` — tier names encode provider-native sizing units (S1–S8 GCP, C1–C8 AWS clusters, P1–P5 Azure). The capacity per tier differs between providers even when the letter+number matches. Mapping to `size: small/medium/large` would lose fidelity.
- AWS `autoMinorVersionUpgrade`, `preferredMaintenanceWindow` — no GCP/Azure/Alicloud equivalent
- GCP `maintenancePolicy.dayOfWeek` — GCP-specific structured shape, no equivalent
- Azure `parameters` (renamed from `redisConfiguration`) — **proposed change**: Azure's API validates a defined set of config keys server-side; renaming the field makes it consistent with AWS/GCP while the controller still validates allowed keys against Azure's schema before calling the API
- Alicloud `engineVersion` immutability — Alicloud does not allow version upgrade after creation; this is a provider constraint reflected in the CRD validation rule, not a field name change

---

### NfsVolume

#### Fields that belong in the base

| Concept | Before (inconsistent) | After (base field) |
|---------|----------------------|-------------------|
| Capacity | `capacity: "10Gi"` k8s Quantity (AWS, Alicloud) · `capacityGb: 1024` integer (GCP, SAP) | `capacity` as k8s `resource.Quantity` everywhere · controller converts to integer for cloud API |
| PV/PVC control | `volume` + `volumeClaim` sub-specs — consistent ✓ | no change needed |
| IpRange reference | `ipRange.name` — consistent ✓ | no change needed |

**Before:**
```yaml
kind: GcpNfsVolume
spec:
  capacityGb: 1024            # integer, GCP API unit leaking into user spec

---
kind: SapNfsVolume
spec:
  capacityGb: 100             # integer, OpenStack API unit

---
kind: AwsNfsVolume
spec:
  capacity: "1Ti"             # k8s Quantity — correct pattern

---
kind: AlicloudNfsVolume
spec:
  capacity: "20Gi"            # k8s Quantity — already correct
```

**After:**
```yaml
kind: GcpNfsVolume
spec:
  capacity: "1Ti"             # was: capacityGb: 1024 · controller converts to integer

---
kind: SapNfsVolume
spec:
  capacity: "100Gi"           # was: capacityGb: 100

---
kind: AwsNfsVolume
spec:
  capacity: "1Ti"             # unchanged

---
kind: AlicloudNfsVolume
spec:
  capacity: "20Gi"            # unchanged
```

#### Fields that stay provider-specific

- GCP `tier` (BASIC_HDD/SSD/ZONAL/REGIONAL), `fileShareName` — Filestore-specific constructs
- GCP `sourceBackup` — GCP supports create-from-backup at volume creation; AWS/Azure do not
- AWS `performanceMode`, `throughput` — EFS-specific I/O knobs
- Alicloud elastic sizing — `capacity` controls PV/PVC claim size only, not cloud quota; this is a provider constraint, not an inconsistency in the field shape

---

### Backup / Restore / Schedule

#### Fields that belong in the base

| Concept | Before (inconsistent) | After (base field) |
|---------|----------------------|-------------------|
| Source reference field | `source.volume.name` (AWS/GCP Backup) · `sourceVolume.name` (SAP) | `source.volume.name` everywhere (where source is an NfsVolume) |
| Schedule fields | `schedule`, `maxRetentionDays`, `maxReadyBackups`, `suspend`, `deleteCascade` — consistent ✓ | no change needed |

**Before:**
```yaml
kind: GcpNfsVolumeBackup
spec:
  source:
    volume:
      name: "my-volume"       # nested under source.volume

---
kind: SapNfsVolumeSnapshot
spec:
  sourceVolume:               # top-level, different shape
    name: "my-volume"
```

**After:**
```yaml
kind: GcpNfsVolumeBackup
spec:
  source:
    volume:
      name: "my-volume"       # unchanged

---
kind: SapNfsVolumeSnapshot
spec:
  source:                     # was: sourceVolume (top-level)
    volume:
      name: "my-volume"
```

#### Fields that stay provider-specific

- Azure uses `source.pvc.name` — Azure has no NfsVolume resource; backup operates on raw PVCs. This is structurally correct, not an inconsistency.
- SAP "snapshot" terminology throughout — reflects Manila's own API
- GCP `accessibleFrom` — cross-cluster backup access control, no equivalent on other providers
- AWS `lifecycle` — maps to AWS Backup Vault cold storage + TTL policies

---

### VpcPeering

The three resources keep separate kinds (Pattern 4 — provider-specific resource is correct). But within each resource the base fields currently sit flat alongside provider-specific fields with no structural separation. Applying the base+extension principle makes the shared intent visible.

**Shared fields across all three (base):**
- `deleteRemotePeering` — identical semantics on AWS, GCP, Azure
- `remotePeeringName` — same concept on GCP and Azure; AWS auto-assigns (not present)

**Provider-specific fields (extension):**
- AWS: `remoteVpcId`, `remoteRegion`, `remoteAccountId`, `remoteRouteTableUpdateStrategy`
- GCP: `remoteVpc`, `remoteProject`, `importCustomRoutes`
- Azure: `remoteVnet`, `remoteTenant`, `useRemoteGateway`

Azure shows the split most clearly — two base-concept fields alongside three Azure-specific fields:

**Before (flat — no structural separation):**
```yaml
kind: AzureVpcPeering
spec:
  remotePeeringName: "my-peering"
  remoteVnet: "/subscriptions/.../virtualNetworks/my-vnet"
  deleteRemotePeering: true
  remoteTenant: "00000000-..."
  useRemoteGateway: false
```

**After (base fields at top level, provider extension under named key):**
```yaml
kind: AzureVpcPeering
spec:
  # base — shared concept with GCP
  remotePeeringName: "my-peering"
  deleteRemotePeering: true
  # Azure-specific extension
  azure:
    remoteVnet: "/subscriptions/.../virtualNetworks/my-vnet"
    remoteTenant: "00000000-..."
    useRemoteGateway: false

---
kind: GcpVpcPeering
spec:
  # base — shared concept with Azure
  remotePeeringName: "my-peering"
  deleteRemotePeering: true
  # GCP-specific extension
  gcp:
    remoteVpc: "my-remote-vpc"
    remoteProject: "my-gcp-project"
    importCustomRoutes: false

---
kind: AwsVpcPeering
spec:
  # base — deleteRemotePeering only (AWS auto-assigns remotePeeringName)
  deleteRemotePeering: true
  # AWS-specific extension
  aws:
    remoteVpcId: "vpc-0a1b2c3d4e5f"
    remoteRegion: "us-east-1"
    remoteAccountId: "123456789012"
    remoteRouteTableUpdateStrategy: "AUTO"
```

The three separate kinds remain. This is not a unification — it is structural clarity within each resource so users immediately see which fields are conceptually shared and which are provider-specific.

---

### IpRange

**No changes.** `IpRange` is already the ideal expression of this pattern: one neutral field (`cidr`), provider routing handled at runtime by the controller.

`GcpSubnet` is a documented exception — GCP Redis Cluster requires Private Service Connect, a different network construct from the Private Service Access IP range used by all other features.

---

## Summary

| CRD family | Changes | Stays provider-specific |
|------------|---------|------------------------|
| Redis Instance + Cluster | Unify `engineVersion`, `parameters`, `replicasPerShard` field names; rename Azure `redisConfiguration` → `parameters` (proposed) | `redisTier`, maintenance windows |
| NfsVolume | Unify `capacity` to k8s `resource.Quantity` | Tier/performance/fileShareName fields, sourceBackup |
| Backup / Restore / Schedule | Unify `source.volume` reference shape | Azure PVC source, SAP snapshot terminology, GCP `accessibleFrom` |
| VpcPeering | Move provider fields under named extension key (`azure:`, `gcp:`, `aws:`) | Remote VPC identity, routing, cross-account/tenant fields |
| IpRange | No changes | Already the ideal neutral resource |
