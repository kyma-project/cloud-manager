# Progressive API Design — Worked Examples

> Companion to `progressive-api-design.md`. Each section shows the full lifecycle of a resource:
> what the user writes at each layer, what the controller generates internally, and what status looks like after provisioning.
> All examples use the same resource `name` so you can follow one object through all three layers.

---

## How to Read These Examples

Every example follows this structure:

```
┌─────────────────────────────────────────────────────────┐
│  What the user writes (one of three layers)             │
├─────────────────────────────────────────────────────────┤
│  What the controller detects / infers / translates      │
├─────────────────────────────────────────────────────────┤
│  What gets sent to the cloud API (internal — not YAML)  │
├─────────────────────────────────────────────────────────┤
│  What the user reads back in status                     │
└─────────────────────────────────────────────────────────┘
```

The user only ever sees the top and bottom boxes. The middle two are the controller's job.

---

## Resource 1: VpcPeering

### Context

VPC Peering connects two private networks so workloads can talk to each other privately.
The user's mental model: *"I want my application network to see my data network."*
They do not think in terms of `VPCPeeringConnection`, `CreateVpcPeeringConnection`, or ARM resource IDs.

---

### Layer 1 — Zero-config (same YAML on any cloud)

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: VpcPeering
metadata:
  name: app-to-data
  namespace: my-team
spec:
  remoteVpcId: "data-network"
```

**What the controller detects from Kyma Scope:**

```
cloud:           aws
localVpcId:      vpc-0aaa1111bbbb2222c   (from Scope)
localRegion:     us-east-1               (from Scope)
localAccountId:  111122223333            (from Scope)
```

**What gets sent to AWS EC2 API (controller-internal):**

```
CreateVpcPeeringConnection:
  VpcId:       vpc-0aaa1111bbbb2222c
  PeerVpcId:   data-network
  PeerRegion:  us-east-1              ← inferred: same region as local
  Tags:
    - Key: managed-by,  Value: kyma-cloud-manager
    - Key: kyma-scope,  Value: my-scope
```

**What the user reads back:**

```yaml
status:
  state: Ready
  peeringId: "pcx-0abc123def456789"
  conditions:
    - type: Ready
      status: "True"
      reason: PeeringActive
      message: "VPC peering connection is active."
```

---

### Layer 2 — Neutral tuning (user knows their architecture)

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: VpcPeering
metadata:
  name: app-to-data
  namespace: my-team
spec:
  remoteVpcId: "data-network"
  deleteRemotePeering: true    # "when I delete this, clean up the remote side too"
  allowDnsResolution: false    # "I manage DNS myself — don't touch it"
```

**What the controller translates (per cloud):**

| Neutral field | → AWS | → GCP | → Azure |
|---|---|---|---|
| `deleteRemotePeering: true` | Sets lifecycle annotation; controller calls `DeleteVpcPeeringConnection` on remote on teardown | Deletes both-direction `NetworkPeering` objects on teardown | Calls `DELETE` on both `VirtualNetworkPeering` objects on teardown |
| `allowDnsResolution: false` | `AccepterPeeringConnectionOptions.AllowDnsResolutionFromRemoteVpc: false` | No private DNS zone created | `allowVirtualNetworkAccess: false` on peering object |

**What gets sent to GCP Compute API (if this were a GCP cluster):**

```
networkPeerings.insert:
  name:                 app-to-data-local-to-peer
  network:              projects/my-proj/global/networks/kyma-network
  peerNetwork:          projects/my-proj/global/networks/data-network
  exportCustomRoutes:   false
  importCustomRoutes:   false
  # deleteRemotePeering handled by controller lifecycle logic
  # allowDnsResolution: false → no private zone created
```

**What the user reads back (same on all clouds):**

```yaml
status:
  state: Ready
  peeringId: "app-to-data-local-to-peer"
  conditions:
    - type: Ready
      status: "True"
      reason: PeeringActive
      message: "VPC peering connection is active."
```

---

### Layer 3 — Provider escape hatch (AWS cross-account, cross-region)

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: VpcPeering
metadata:
  name: app-to-data
  namespace: my-team
spec:
  remoteVpcId: "vpc-0fedcba9876543210"
  deleteRemotePeering: true
  # Layer 3: the user knows this is cross-account cross-region
  aws:
    remoteRegion: "us-west-2"
    remoteAccountId: "999988887777"
    remoteRouteTableUpdateStrategy: "AUTO"
```

**What the controller validates:**

```
cluster.cloud = "aws" ✓  (spec.aws matches — allowed)
spec.gcp not set      ✓
spec.azure not set    ✓
```

**What gets sent to AWS EC2 API:**

```
CreateVpcPeeringConnection:
  VpcId:        vpc-0aaa1111bbbb2222c   ← from Scope (Layer 0)
  PeerVpcId:    vpc-0fedcba9876543210   ← from spec.remoteVpcId (Layer 1)
  PeerRegion:   us-west-2              ← from spec.aws.remoteRegion (Layer 3)
  PeerOwnerId:  999988887777           ← from spec.aws.remoteAccountId (Layer 3)

# After acceptance:
ModifyVpcPeeringConnectionOptions:
  AccepterPeeringConnectionOptions:
    AllowDnsResolutionFromRemoteVpc: true   ← default (not set in spec)

# Route table updates (AUTO strategy):
CreateRoute (requester route table):
  RouteTableId:           rtb-0aaa111  ← discovered from Scope
  DestinationCidrBlock:   10.1.0.0/16  ← discovered from peer VPC
  VpcPeeringConnectionId: pcx-0abc123
```

**What the user reads back:**

```yaml
status:
  state: Ready
  peeringId: "pcx-0abc123def456789"
  conditions:
    - type: Ready
      status: "True"
      reason: PeeringActive
      message: "VPC peering connection is active (cross-account, us-east-1 ↔ us-west-2)."
```

---

### Error: Wrong provider sub-struct on GCP cluster

**What the user mistakenly writes on a GCP cluster:**

```yaml
spec:
  remoteVpcId: "data-network"
  aws:                          # ← this is a GCP cluster
    remoteAccountId: "999988887777"
```

**What the controller returns:**

```yaml
status:
  state: Error
  conditions:
    - type: Ready
      status: "False"
      reason: InvalidProviderExtension
      message: >
        spec.aws is not valid on this cluster. This cluster runs on GCP.
        Use spec.gcp for GCP-specific configuration, or remove the
        provider extension to use neutral defaults.
```

---

## Resource 2: RedisInstance

### Context

The user wants a managed Redis cache. Their mental model: *"I need a Redis 7 instance with 6 GB of memory."*
They do not think in terms of `ReplicationGroup`, `REDIS_7_0`, or `cacheNodeType`.

---

### Layer 1 — Zero-config

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: RedisInstance
metadata:
  name: session-cache
  namespace: my-team
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
```

**What the controller detects and infers:**

```
cloud:          gcp
tier:           STANDARD_HA     ← default for production Kyma clusters
replicaCount:   1               ← default
maintenanceWindow: none         ← default (GCP assigns)
```

**What the controller translates (Layer 1 fields, GCP):**

| User field | → GCP API field | Value |
|---|---|---|
| `memorySizeGb: 6` | `memorySizeGb` | `6` (passthrough — GCP uses integer GiB) |
| `engineVersion: "7.0"` | `redisVersion` | `"REDIS_7_0"` ← controller maps |

**What gets sent to GCP Memorystore API:**

```
redis.instances.create:
  name:          session-cache
  memorySizeGb:  6
  redisVersion:  REDIS_7_0         ← mapped from "7.0"
  tier:          STANDARD_HA       ← default
  replicaCount:  1                 ← default
  labels:
    managed-by: kyma-cloud-manager
```

**What the user reads back:**

```yaml
status:
  state: Ready
  primaryEndpoint: "10.0.0.5:6379"
  memorySizeGb: 6
  replicaCount: 1
  conditions:
    - type: Ready
      status: "True"
      reason: InstanceReady
      message: "Redis instance is ready."
```

---

### Layer 2 — Neutral tuning

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: RedisInstance
metadata:
  name: session-cache
  namespace: my-team
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  # Layer 2 — neutral tuning
  replicasPerShard: 2             # HA with 2 replicas
  parameters:                     # consistent name — was: redisConfigs (GCP), parameters (AWS)
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
```

**What the controller translates per cloud:**

| Neutral field | → AWS | → GCP | → Azure |
|---|---|---|---|
| `engineVersion: "7.0"` | `engineVersion: "7.0"` (passthrough) | `redisVersion: "REDIS_7_0"` | `redisVersion: "6.0"` (Azure maps differently) |
| `replicasPerShard: 2` | `ReplicasPerNodeGroup: 2` | `replicaCount: 2` | `replicasPerPrimary: 2` ← was the inconsistent name |
| `parameters: {k: v}` | `ParameterNameValues: [{Name: k, Value: v}]` | `redisConfigs: {k: v}` ← was the inconsistent name | validated against Azure's allowed key list, then passed through |

**What gets sent to AWS ElastiCache API (if this were an AWS cluster):**

```
CreateReplicationGroup:
  ReplicationGroupId:    session-cache
  CacheNodeType:         cache.r6g.large    ← inferred from memorySizeGb: 6
  EngineVersion:         7.0
  NumNodeGroups:         1
  ReplicasPerNodeGroup:  2                  ← mapped from replicasPerShard
  ParameterNameValues:
    - ParameterName: maxmemory-policy,  Value: volatile-lru
    - ParameterName: activedefrag,      Value: yes
```

**What the user reads back (same on all clouds):**

```yaml
status:
  state: Ready
  primaryEndpoint: "session-cache.abc123.ng.0001.use1.cache.amazonaws.com:6379"
  memorySizeGb: 6
  replicaCount: 2
  conditions:
    - type: Ready
      status: "True"
```

---

### Layer 3 — Provider escape hatch (GCP maintenance window)

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: RedisInstance
metadata:
  name: session-cache
  namespace: my-team
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  replicasPerShard: 2
  parameters:
    maxmemory-policy: volatile-lru
  # Layer 3: GCP-specific — no cross-cloud equivalent
  gcp:
    maintenancePolicy:
      dayOfWeek:
        day: "SATURDAY"
        startTime:
          hours: 15
    persistenceConfig:
      persistenceMode: "RDB"
      rdbSnapshotPeriod: "TWELVE_HOURS"
```

**What gets sent to GCP Memorystore API:**

```
redis.instances.create:
  memorySizeGb:   6
  redisVersion:   REDIS_7_0
  replicaCount:   2
  redisConfigs:                         ← mapped from parameters
    maxmemory-policy: volatile-lru
  maintenancePolicy:                    ← from spec.gcp (passed through)
    weeklyMaintenanceWindow:
      - day: SATURDAY
        startTime: { hours: 15 }
  persistenceConfig:                    ← from spec.gcp (passed through)
    persistenceMode: RDB
    rdbSnapshotPeriod: TWELVE_HOURS
```

---

### Layer 3 — Provider escape hatch (Azure TLS and network access)

```yaml
spec:
  memorySizeGb: 6
  engineVersion: "7.0"
  azure:
    minimumTlsVersion: "1.2"
    publicNetworkAccess: "Disabled"
    sku:
      family: "C"
      name: "Standard"
      capacity: 2
```

**What the controller validates and translates:**

```
cluster.cloud = "azure" ✓
minimumTlsVersion → maps directly to Azure Redis API field
publicNetworkAccess → maps directly
sku → passes through (Azure-native SKU vocabulary)
memorySizeGb: 6 → Azure SKU capacity inferred OR overridden by spec.azure.sku.capacity
```

---

## Resource 3: NfsVolume

### Context

The user wants shared file storage that multiple pods can read and write simultaneously.
Their mental model: *"I need 1 TB of shared storage."*
They do not think in terms of Filestore tiers, EFS performance modes, or Manila OpenStack API integers.

---

### Layer 1 — Zero-config (the unification win)

This is the clearest example of the pattern's value. Four different clouds had four different field names and formats for the same concept. After applying Pattern 2:

**What the user writes (identical on AWS, GCP, Azure, SAP):**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: NfsVolume
metadata:
  name: shared-data
  namespace: my-team
spec:
  capacity: "1Ti"
```

**What the controller translates:**

| Cloud | API call | Translated value |
|---|---|---|
| GCP (Filestore) | `capacityGb: 1024` | `1Ti` → `1024` GiB integer |
| SAP (OpenStack Manila) | `size: 1024` | `1Ti` → `1024` GiB integer |
| AWS (EFS) | Elastic sizing | `1Ti` used as PVC claim only; EFS scales automatically |
| Azure (Azure Files) | `diskSizeGB: 1024` | `1Ti` → `1024` GiB integer |

**Before this pattern was applied:**

```yaml
# GCP user had to write:
kind: GcpNfsVolume
spec:
  capacityGb: 1024        # integer — GCP Filestore API unit leaked into spec

# SAP user had to write:
kind: SapNfsVolume
spec:
  capacityGb: 100         # integer — OpenStack Manila unit leaked into spec

# AWS user could write:
kind: AwsNfsVolume
spec:
  capacity: "1Ti"         # this was already correct

# Azure user had to write:
kind: AzureNfsVolume
spec:
  capacity: "1Ti"         # this was already correct (Azure already used Quantity)
```

**After — one CRD, one field, four clouds:**

```yaml
kind: NfsVolume
spec:
  capacity: "1Ti"         # always — controller converts for each cloud
```

**What the user reads back:**

```yaml
status:
  state: Ready
  capacity: "1Ti"         # normalised — always Quantity regardless of cloud
  server: "10.0.0.10"
  path: "/vol1"
  conditions:
    - type: Ready
      status: "True"
      reason: NfsVolumeReady
      message: "NFS volume is ready."
```

---

### Layer 2 — No additional neutral fields needed

NfsVolume has no shared optional concepts across all clouds. Performance modes, tier names, and file share names are all provider-specific. They belong in Layer 3.

---

### Layer 3 — Provider escape hatch

```yaml
# GCP: Filestore tier and file share name
spec:
  capacity: "1Ti"
  gcp:
    tier: "REGIONAL"         # BASIC_HDD / BASIC_SSD / ZONAL / REGIONAL
    fileShareName: "vol1"    # Filestore file share name — GCP-specific

---
# AWS: EFS performance mode and throughput
spec:
  capacity: "1Ti"
  aws:
    performanceMode: "maxIO"              # generalPurpose / maxIO
    throughputMode: "provisioned"         # bursting / provisioned / elastic
    provisionedThroughputMibps: 500.0

---
# GCP: create from backup
spec:
  capacity: "1Ti"
  gcp:
    tier: "REGIONAL"
    sourceBackup: "projects/my-project/locations/us-east1/backups/my-backup"
```

---

## Resource 4: IpRange

### Context

IpRange is the **reference implementation** of the pattern — the ideal the other resources should achieve.
It already follows Layer 1 perfectly and needs no changes.

**What the user writes (same on all clouds, no translation needed):**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: IpRange
metadata:
  name: kyma-ip-range
  namespace: my-team
spec:
  cidr: "10.250.0.0/22"
```

**Why this is perfect:**
- `cidr` is a pure network concept — same meaning on AWS, GCP, Azure, Alicloud, OpenStack
- The controller routes to the correct provider at runtime (VPC, GCP VPC, Azure VNet, Alicloud VSwitch, OpenStack subnet) without any user input
- The value format (`10.250.0.0/22`) is IANA CIDR notation — not a provider-specific unit
- No Layer 2 or Layer 3 needed for the common case

**What the user reads back:**

```yaml
status:
  state: Ready
  cidr: "10.250.0.0/22"
  conditions:
    - type: Ready
      status: "True"
```

This is the test every other CM resource should pass: *can we write the same YAML on any cloud and have it work?*

---

## Resource 5: Backup and Schedule

### Context

The user wants automated backups of their NfsVolume. Their mental model: *"Back this up every night, keep 7 days of backups."*

---

### Layer 1 — Zero-config

**What the user writes:**

```yaml
apiVersion: cloud-manager.kyma-project.io/v1beta1
kind: NfsVolumeBackupSchedule
metadata:
  name: nightly-backup
  namespace: my-team
spec:
  source:
    volume:
      name: shared-data          # reference to the NfsVolume
  schedule: "0 2 * * *"         # cron: every night at 02:00
  maxRetentionDays: 7
```

**Before the pattern was applied — the inconsistency:**

```yaml
# GCP user wrote:
kind: GcpNfsVolumeBackup
spec:
  source:
    volume:
      name: shared-data         # nested under source.volume ✓

# SAP user wrote:
kind: SapNfsVolumeSnapshot
spec:
  sourceVolume:                 # top-level — different shape ✗
    name: shared-data
```

**After — one shape everywhere:**

```yaml
# Both write:
spec:
  source:
    volume:
      name: shared-data         # same shape on all providers
```

---

### Layer 2 — Neutral tuning

```yaml
spec:
  source:
    volume:
      name: shared-data
  schedule: "0 2 * * *"
  maxRetentionDays: 7
  maxReadyBackups: 5             # keep at most 5 ready backups at once
  suspend: false                 # pause scheduling without deleting the CR
```

All of these fields are consistent across providers — no translation needed. The controller passes them through to the provider's scheduling mechanism.

---

### Layer 3 — Provider escape hatch

```yaml
# GCP: cross-cluster backup access control — no AWS/Azure equivalent
spec:
  source:
    volume:
      name: shared-data
  schedule: "0 2 * * *"
  gcp:
    accessibleFrom:
      - "projects/peer-project/locations/us-east1/instances/peer-cluster"

---
# AWS: Backup Vault cold storage lifecycle
spec:
  source:
    volume:
      name: shared-data
  schedule: "0 2 * * *"
  aws:
    lifecycle:
      moveToColdStorageAfterDays: 30
      deleteAfterDays: 365
```

---

## The Full Layer Matrix

A summary of which fields exist at which layer for each resource family:

### VpcPeering

| Field | Layer | Cloud(s) |
|---|---|---|
| `remoteVpcId` | 1 | All |
| `deleteRemotePeering` | 2 | All |
| `allowDnsResolution` | 2 | All |
| `aws.remoteRegion` | 3 | AWS |
| `aws.remoteAccountId` | 3 | AWS |
| `aws.remoteRouteTableUpdateStrategy` | 3 | AWS |
| `gcp.remoteProject` | 3 | GCP |
| `gcp.importCustomRoutes` | 3 | GCP |
| `azure.remoteTenant` | 3 | Azure |
| `azure.remotePeeringName` | 3 | Azure |
| `azure.useRemoteGateway` | 3 | Azure |

### RedisInstance

| Field | Layer | Cloud(s) |
|---|---|---|
| `memorySizeGb` | 1 | All |
| `engineVersion` | 1 | All (controller maps `"7.0"` → `"REDIS_7_0"` for GCP) |
| `replicasPerShard` | 2 | All (controller maps → `replicasPerPrimary` for Azure) |
| `parameters` | 2 | All (controller maps → `redisConfigs` for GCP, validates for Azure) |
| `aws.autoMinorVersionUpgrade` | 3 | AWS |
| `aws.preferredMaintenanceWindow` | 3 | AWS |
| `gcp.maintenancePolicy` | 3 | GCP |
| `gcp.persistenceConfig` | 3 | GCP |
| `azure.minimumTlsVersion` | 3 | Azure |
| `azure.publicNetworkAccess` | 3 | Azure |
| `azure.sku` | 3 | Azure |

### NfsVolume

| Field | Layer | Cloud(s) |
|---|---|---|
| `capacity` | 1 | All (controller converts `"1Ti"` → `capacityGb: 1024` for GCP/SAP) |
| `gcp.tier` | 3 | GCP |
| `gcp.fileShareName` | 3 | GCP |
| `gcp.sourceBackup` | 3 | GCP |
| `aws.performanceMode` | 3 | AWS |
| `aws.throughputMode` | 3 | AWS |
| `aws.provisionedThroughputMibps` | 3 | AWS |

### NfsVolumeBackupSchedule

| Field | Layer | Cloud(s) |
|---|---|---|
| `source.volume.name` | 1 | All |
| `schedule` | 1 | All |
| `maxRetentionDays` | 2 | All |
| `maxReadyBackups` | 2 | All |
| `suspend` | 2 | All |
| `deleteCascade` | 2 | All |
| `gcp.accessibleFrom` | 3 | GCP |
| `aws.lifecycle` | 3 | AWS |

### IpRange

| Field | Layer | Cloud(s) |
|---|---|---|
| `cidr` | 1 | All — **no Layer 2 or Layer 3 needed** |

---

## Controller Translation Reference

The mappings the controller implements for Layer 2 fields:

### engineVersion

```go
// controller/internal/translator/redis.go

func TranslateEngineVersion(neutral string, cloud CloudProvider) string {
    switch cloud {
    case AWS:
        return neutral                      // "7.0" → "7.0" (passthrough)
    case GCP:
        return "REDIS_" + strings.ReplaceAll(neutral, ".", "_")
                                            // "7.0" → "REDIS_7_0"
    case Azure:
        return neutral                      // "7.0" → "7.0" (passthrough)
    case Alicloud:
        return neutral                      // "7.0" → "7.0" (passthrough)
    }
}
```

### capacity (NfsVolume)

```go
// controller/internal/translator/nfs.go

func TranslateCapacityToGib(quantity resource.Quantity) int64 {
    // "1Ti" → 1024
    // "100Gi" → 100
    // "500Gi" → 500
    return quantity.Value() / (1024 * 1024 * 1024)
}
```

### replicasPerShard

```go
// controller/internal/translator/redis.go

// Field is "replicasPerShard" in user spec on all clouds.
// Azure's API calls it "replicasPerPrimary".
// Controller writes Azure's name in the KCP object; user always writes neutral name.
func AzureReplicaFieldName() string {
    return "replicasPerPrimary"   // used only in KCP/Azure internal reconciler
}
```

### status.state normalisation

```go
// controller/internal/translator/status.go

func NormaliseState(providerState string, cloud CloudProvider) ResourceState {
    switch cloud {
    case AWS:
        switch providerState {
        case "available":   return StateReady
        case "creating":    return StateCreating
        case "deleting":    return StateDeleting
        case "failed":      return StateError
        }
    case GCP:
        switch providerState {
        case "READY":       return StateReady
        case "CREATING":    return StateCreating
        case "DELETING":    return StateDeleting
        case "MAINTENANCE": return StateReady   // still serving during maintenance
        case "FAILED":      return StateError
        }
    case Azure:
        switch providerState {
        case "Connected":   return StateReady
        case "Initiated":   return StateCreating
        case "Disconnected":return StateError
        }
    }
}
```

---

## End-to-End Lifecycle: VpcPeering on AWS

A complete trace from user `kubectl apply` to ready status:

```
1. User applies:
   kubectl apply -f vpc-peering.yaml

   kind: VpcPeering
   spec:
     remoteVpcId: "data-network"
     deleteRemotePeering: true

2. Controller reconciler triggered
   └── reads Kyma Scope
       └── cloud = "aws"
       └── localVpcId = "vpc-0aaa1111bbbb2222c"
       └── localRegion = "us-east-1"

3. Controller validates spec
   └── spec.aws not set ✓ (no provider sub-struct)
   └── spec.gcp not set ✓
   └── spec.azure not set ✓
   └── remoteVpcId present ✓

4. Controller applies Layer 0 defaults
   └── remoteRegion = "us-east-1" (same as local — inferred)

5. Controller calls AWS EC2 API
   └── CreateVpcPeeringConnection(
         VpcId:     "vpc-0aaa1111bbbb2222c",
         PeerVpcId: "data-network",
         PeerRegion: "us-east-1"
       )
   └── response: { VpcPeeringConnectionId: "pcx-0abc123" }

6. Controller watches for AWS state = "active"
   └── polls DescribeVpcPeeringConnections
   └── status: "pending-acceptance" → auto-accepts (same account)
   └── status: "active" ✓

7. Controller writes route table entries (deleteRemotePeering annotation stored)

8. Controller updates CR status
   └── status.state = "Ready"
   └── status.peeringId = "pcx-0abc123"
   └── status.conditions[Ready] = True

9. User reads:
   kubectl get vpcpeering app-to-data

   NAME          STATE   PEERING-ID           AGE
   app-to-data   Ready   pcx-0abc123def456     2m
```

---

## Progressive Disclosure in Documentation

The three-layer design also maps to documentation structure. Users read documentation progressively — they stop reading when they have what they need.

```
docs/
  vpcpeering/
    README.md          ← Layer 1 only. "Create a VPC peering in 3 lines."
    tuning.md          ← Layer 2. "Common tuning options."
    aws-advanced.md    ← Layer 3. "AWS-specific configuration."
    gcp-advanced.md    ← Layer 3. "GCP-specific configuration."
    azure-advanced.md  ← Layer 3. "Azure-specific configuration."
```

A new user reads `README.md` and stops. An AWS engineer doing cross-account peering reads `aws-advanced.md`. Nobody reads everything unless they need to.

This mirrors how Kubernetes itself documents `PersistentVolumeClaim`:
- The PVC page explains `capacity` and `accessModes` (Layer 1)
- A separate StorageClass page explains provisioner-specific `parameters` (Layer 3)
- Users who only need basic storage never visit the StorageClass page
