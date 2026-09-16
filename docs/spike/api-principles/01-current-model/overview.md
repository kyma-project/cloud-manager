# Current Cloud Manager SKR Resource Model — Overview

Cloud Manager's SKR API grew organically, feature by feature, with no shared resource-modeling approach. The result is a mix of provider-neutral resources (e.g. `IpRange`) and fully provider-specific ones (e.g. `AwsRedisInstance`, `GcpVpcPeering`), where some provider differences are real and justified by the underlying cloud service, while others are accidental divergence in naming, field shapes, or immutability rules. This document inventories all current SKR resources, shows their full schemas side by side, and summarises where and why they differ.

---

## IpRange and GcpSubnet

`IpRange` allocates a private IP range for Cloud Manager features (NFS, Redis) that require private service access into the Kyma cluster's VPC. It is provider-neutral — the same single-field spec works on all providers. `GcpSubnet` is a GCP-only counterpart that allocates a dedicated subnet for GCP Redis Cluster's Private Service Connect, a different network construct from the IP ranges used by Filestore and Memorystore.

<table>
<tr>
<th>IpRange</th>
<th>GcpSubnet</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: IpRange
metadata:
  name: my-iprange
  # cluster-scoped, no namespace
spec:
  # Optional. Immutable once set.
  # If omitted, a CIDR is auto-allocated from the cluster's VPC.
  cidr: "10.250.0.0/22"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpSubnet
metadata:
  name: my-gcpsubnet
  # cluster-scoped, no namespace
spec:
  # Required. CIDR for the GCP subnet.
  cidr: "10.250.4.0/22"
```

</td>
</tr>
</table>

**Differences across providers:**
- `IpRange` is provider-neutral — the same `cidr` field works on AWS, GCP, Azure, Alicloud, and OpenStack.
- `GcpSubnet` is GCP-only. It exists because GCP Redis Cluster requires a dedicated subnet for Private Service Connect, which is a different network construct from the Private Service Access IP range used by Filestore and Memorystore.

**Differences across features:**
- `IpRange` is referenced by NFS volumes and Redis instances/clusters across all providers (except `GcpRedisCluster`, which uses `GcpSubnet` instead).
- `GcpSubnet` is referenced only by `GcpRedisCluster`. Once GcpRedisCluster reaches GA, GcpSubnet is planned to move to an "undefined" feature gate.

---

## VpcPeering

VpcPeering connects the Kyma cluster's VPC to a customer-owned VPC in the same cloud provider. All three resources are fully provider-specific; the underlying peering mechanism, identity model, and routing options differ enough that no field set is shared beyond `deleteRemotePeering`.

<table>
<tr>
<th>AwsVpcPeering</th>
<th>AzureVpcPeering</th>
<th>GcpVpcPeering</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsVpcPeering
metadata:
  name: my-peering
  # cluster-scoped, no namespace
spec:
  # Required. Immutable.
  remoteVpcId: "vpc-0a1b2c3d4e5f"
  # Required. Immutable.
  remoteRegion: "us-east-1"
  # Required. Immutable.
  remoteAccountId: "123456789012"
  # Whether to delete the remote side of the peering on delete.
  deleteRemotePeering: true
  # Immutable. Enum: AUTO, NONE, MATCHED, UNMATCHED
  remoteRouteTableUpdateStrategy: "AUTO"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureVpcPeering
metadata:
  name: my-peering
  # cluster-scoped, no namespace
spec:
  # Required. Immutable. Max 80 chars.
  remotePeeringName: "my-peering"
  # Required. Immutable. Full Azure resource ID of remote VNet.
  remoteVnet: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet"
  # Whether to delete the remote side of the peering on delete.
  deleteRemotePeering: true
  # Optional. Immutable. AAD tenant ID for cross-tenant peering.
  remoteTenant: "00000000-0000-0000-0000-000000000001"
  # Optional. Immutable. Default false.
  useRemoteGateway: false
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpVpcPeering
metadata:
  name: my-peering
  # cluster-scoped, no namespace
spec:
  # Required. Immutable. 1-63 chars, lowercase alphanumeric+hyphens.
  remotePeeringName: "my-peering"
  # Required. Immutable. VPC name (not ID) in the remote project.
  remoteVpc: "my-remote-vpc"
  # Required. Immutable. GCP project ID, 6-30 chars.
  remoteProject: "my-gcp-project"
  # Whether to delete the remote side of the peering on delete.
  deleteRemotePeering: true
  # Immutable. Default false.
  importCustomRoutes: false
```

</td>
</tr>
</table>

**Differences across providers:**
- All three are fully provider-specific with no shared fields beyond `deleteRemotePeering`.
- AWS uniquely requires `remoteAccountId` (cross-account identity) and adds `remoteRouteTableUpdateStrategy` to control route-table propagation.
- GCP uniquely requires `remoteProject` and adds `importCustomRoutes` to control whether the Kyma side imports custom routes from the peer.
- Azure uniquely adds `remoteTenant` (cross-tenant / cross-subscription peering) and `useRemoteGateway`.
- The remote VPC is identified differently: AWS by opaque `remoteVpcId`; GCP by human-readable `remoteVpc` name plus `remoteProject`; Azure by a full ARM resource ID string in `remoteVnet`.

---

## NfsVolume

NfsVolume provisions a managed NFS file system and exposes it as a Kubernetes PersistentVolume and PersistentVolumeClaim. Each provider maps to a different cloud NFS service (EFS on AWS, Filestore on GCP, NAS on Alicloud, Manila on OpenStack/CCEE). Azure has no NfsVolume resource — it backs customer-managed PVCs backed by Azure Files directly via the Backup & Restore resources.

<table>
<tr>
<th>AwsNfsVolume</th>
<th>GcpNfsVolume</th>
<th>AlicloudNfsVolume</th>
<th>SapNfsVolume (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolume
metadata:
  name: my-volume
  namespace: kyma-system
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  # Required. Kubernetes resource quantity.
  capacity: "10Gi"
  # Default: generalPurpose. Enum: generalPurpose, maxIO.
  performanceMode: "generalPurpose"
  # Default: bursting. Enum: bursting, elastic.
  throughput: "bursting"
  # Optional. Controls the PersistentVolume name and labels.
  volume:
    name: "my-pv"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  # Optional. Controls the PersistentVolumeClaim name and labels.
  volumeClaim:
    name: "my-pvc"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-volume
  namespace: kyma-system
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  # Deprecated. Immutable. Omit — zone is auto-selected.
  # location: ""
  # Immutable. Default: BASIC_HDD.
  # Enum: BASIC_HDD, BASIC_SSD, ZONAL, REGIONAL.
  tier: "BASIC_HDD"
  # Immutable. Default: vol1.
  fileShareName: "vol1"
  # Capacity in GiB. Default: 2560.
  # Valid ranges depend on tier.
  capacityGb: 1024
  # Optional. Immutable. Create volume from an existing backup.
  sourceBackup:
    name: "my-backup"
    namespace: "kyma-system"
  # Optional. Controls the PersistentVolume name and labels.
  volume:
    name: "my-pv"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  # Optional. Name is immutable. Controls PVC.
  volumeClaim:
    name: "my-pvc"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudNfsVolume
metadata:
  name: my-volume
  namespace: kyma-system
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  # Required. Kubernetes resource quantity.
  # AliCloud NAS is elastic — this sizes the PV/PVC only,
  # not the provisioned cloud quota.
  capacity: "20Gi"
  # Immutable. Default: Performance.
  # Enum: Performance, Capacity, Premium.
  storageType: "Performance"
  # Optional. Controls the PersistentVolume name and labels.
  volume:
    name: "my-pv"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  # Optional. Controls PVC name and labels.
  volumeClaim:
    name: "my-pvc"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolume
metadata:
  name: my-volume
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. Integer GiB (not Kubernetes quantity). Must be > 0.
  capacityGb: 100
  # Optional. Controls PersistentVolume name and labels.
  volume:
    name: "my-pv"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  # Optional. Controls PVC name and labels.
  volumeClaim:
    name: "my-pvc"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  # Optional. Immutable. Create volume from an existing snapshot.
  # Exactly one property allowed.
  dataSource:
    snapshot:
      name: "my-snapshot"
      namespace: "kyma-system"
```

</td>
</tr>
</table>

**Differences across providers:**
- **Capacity type:** AWS and Alicloud use Kubernetes `resource.Quantity` (e.g. `"10Gi"`); GCP and SAP/OpenStack use an integer `capacityGb` field.
- **AWS** has `performanceMode` and `throughput` — EFS-specific I/O performance knobs not present on any other provider.
- **GCP** has `tier` (BASIC_HDD/BASIC_SSD/ZONAL/REGIONAL), `fileShareName` (the export path within the instance), and `sourceBackup` — allowing creation of a volume directly from an existing backup without a separate Restore object.
- **Alicloud** NAS is elastic: the `capacity` value controls only the PV/PVC claim size and is not propagated to the cloud provider (no provisioned quota).
- **SAP/OpenStack** has `dataSource.snapshot` for creating a volume from a Manila snapshot at creation time, analogous to GCP's `sourceBackup`.
- **Azure** has no NfsVolume resource — Backup & Restore works directly with raw Kubernetes PVCs.

---

## Redis

Cloud Manager provides managed Redis for both single-node HA instances (`*RedisInstance`) and sharded clusters (`*RedisCluster`). Both sub-families are shown in the same table to make Instance vs. Cluster differences immediately visible. `AzureManagedRedis` is a separate resource mapping to a different Azure product (Azure Managed Redis / Enterprise) from `AzureRedisInstance` (Azure Cache for Redis / OSS).

<table>
<tr>
<th>AwsRedisInstance</th>
<th>GcpRedisInstance</th>
<th>AzureRedisInstance</th>
<th>AlicloudRedisInstance</th>
<th>AwsRedisCluster</th>
<th>GcpRedisCluster</th>
<th>AzureRedisCluster</th>
<th>AlicloudRedisCluster</th>
<th>AzureManagedRedis</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisInstance
metadata:
  name: my-redis
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. S1-S7 or P1-P6.
  # Service letter (S/P) immutable;
  # only capacity number is mutable.
  redisTier: "S1"
  # Default: "7.0". Enum: "6.x", "7.0", "7.1".
  # Upgrade-only (cannot downgrade).
  engineVersion: "7.0"
  # Default: false.
  autoMinorVersionUpgrade: false
  # Default: false.
  authEnabled: true
  # Optional. Format: ddd:hh24:mi-ddd:hh24:mi
  preferredMaintenanceWindow: "sun:23:00-mon:01:30"
  # Optional. Freeform Redis config map.
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisInstance
metadata:
  name: my-redis
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. S1-S8 or P1-P6.
  # Service letter (S/P) immutable.
  redisTier: "P1"
  # Default: "REDIS_7_0".
  # Enum: REDIS_6_X, REDIS_7_0, REDIS_7_2.
  # Upgrade-only (cannot downgrade).
  redisVersion: "REDIS_7_0"
  # Default: false.
  authEnabled: true
  # Optional. Freeform Redis config map.
  redisConfigs:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  # Optional. Structured maintenance window.
  maintenancePolicy:
    dayOfWeek:
      day: "SATURDAY"
      startTime:
        hours: 15
        minutes: 45
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisInstance
metadata:
  name: my-redis
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. P1-P5 or S1-S5.
  redisTier: "P1"
  # Default: "6.0". Immutable after creation.
  redisVersion: "6.0"
  # Optional. Immutable after creation. Typed struct (not freeform map).
  redisConfiguration:
    maxclients: "8"
    maxfragmentationmemory-reserved: "50"
    maxmemory-delta: "50"
    maxmemory-policy: "volatile-lru"
    maxmemory-reserved: "50"
    notify-keyspace-events: ""
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudRedisInstance
metadata:
  name: my-redis
  namespace: kyma-system
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  # Required. S1-S5 or P1-P5.
  # Service letter (S/P) immutable.
  redisTier: "S1"
  # Default: "7.0". Enum: "5.0", "6.0", "7.0".
  # Immutable after creation (unlike AWS/GCP).
  engineVersion: "7.0"
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisCluster
metadata:
  name: my-redis-cluster
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. C1-C8.
  redisTier: "C1"
  # Required. 1-500 shards.
  shardCount: 3
  # Default: 1. Range: 0-5.
  replicasPerShard: 1
  # Default: "7.0". Enum: "6.x", "7.0", "7.1".
  # Upgrade-only.
  engineVersion: "7.0"
  # Default: false.
  autoMinorVersionUpgrade: false
  # Default: false.
  authEnabled: true
  # Optional. Format: ddd:hh24:mi-ddd:hh24:mi
  preferredMaintenanceWindow: "sun:23:00-mon:01:30"
  # Optional. Freeform Redis config map.
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisCluster
metadata:
  name: my-redis-cluster
  namespace: kyma-system
spec:
  # Optional. References a GcpSubnet object,
  # NOT an IpRange — the only SKR resource that
  # uses GcpSubnetRef instead of IpRangeRef.
  subnet:
    name: my-gcpsubnet
  # Required. Enum: C1, C3, C4, C6.
  redisTier: "C3"
  # Required. Min: 1.
  # Max depends on replicasPerShard: 250/125/83.
  shardCount: 5
  # Default: 1. Range: 0-2.
  replicasPerShard: 1
  # Optional. Freeform Redis config map.
  redisConfigs:
    maxmemory-policy: volatile-lru
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisCluster
metadata:
  name: my-redis-cluster
  namespace: kyma-system
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. Enum: C3, C4, C5, C6, C7.
  redisTier: "C3"
  # Optional. Range: 0-10.
  shardCount: 3
  # Optional. Immutable. Range: 0-10.
  replicasPerPrimary: 1
  # Optional. (Azure uses "replicasPerPrimary",
  # not "replicasPerShard" like other providers.)
  redisVersion: "6"
  # Optional. Immutable after creation. Typed struct.
  redisConfiguration:
    maxclients: "8"
    maxfragmentationmemory-reserved: "50"
    maxmemory-delta: "50"
    maxmemory-policy: "volatile-lru"
    maxmemory-reserved: "50"
    notify-keyspace-events: ""
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudRedisCluster
metadata:
  name: my-redis-cluster
  namespace: kyma-system
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  # Required. Enum: C3, C4, C5, C6, C7.
  # Each tier sets per-shard memory.
  redisTier: "C3"
  # Required. Range: 1-32.
  shardCount: 4
  # Must be 0. Read-replica config not yet
  # enabled for current tier SKUs.
  replicasPerShard: 0
  # Default: "7.0". Enum: "5.0", "6.0", "7.0".
  # Immutable after creation.
  engineVersion: "7.0"
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureManagedRedis
metadata:
  name: my-redis
  namespace: kyma-system
  # Different Azure product: Azure Managed Redis
  # (Enterprise tier), NOT Azure Cache for Redis.
spec:
  # Optional.
  ipRange:
    name: my-iprange
  # Required. S1-S5, P1-P5, or C3-C7.
  # Family letter is immutable:
  #   S = Balanced non-HA EnterpriseCluster (dev)
  #   P = ComputeOptimized HA EnterpriseCluster
  #   C = ComputeOptimized HA OSSCluster (sharded)
  redisTier: "P1"
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
```

</td>
</tr>
</table>

**Differences across providers:**
- **Version field naming is inconsistent:** AWS and Alicloud use `engineVersion` with dotted numeric values (e.g. `"7.0"`, `"6.x"`); GCP uses `redisVersion` with underscore-prefixed enum values (e.g. `"REDIS_7_0"`); Azure uses `redisVersion` with short numeric strings (e.g. `"6.0"`, `"6"`).
- **Configuration field shape differs:** AWS uses `parameters` (freeform `map[string]string`); GCP uses `redisConfigs` (freeform `map[string]string`); Azure uses `redisConfiguration` — a typed struct with named fields, making it impossible to pass an unrecognised config key.
- **Immutability rules differ:** Azure `redisConfiguration` and `redisVersion` are immutable after creation; all other providers allow in-place upgrades (version upgrade-only, not downgrade).
- **Maintenance window:** AWS uses a freeform string (`"sun:23:00-mon:01:30"`); GCP uses a structured object (`maintenancePolicy.dayOfWeek`); Azure and Alicloud have no maintenance window field.
- **Auth toggle:** AWS and GCP expose an explicit `authEnabled` boolean. Azure and Alicloud always require authentication — there is no toggle.

**Differences across features (Instance vs Cluster):**
- Cluster adds `shardCount` and either `replicasPerShard` (AWS, GCP, Alicloud) or `replicasPerPrimary` (Azure) — an inconsistency in naming for the same concept.
- `GcpRedisCluster` uniquely uses `subnet` (a `GcpSubnetRef`) instead of `ipRange` (an `IpRangeRef`) for private networking — the only SKR resource in the entire API that does not use `IpRange`.
- `AzureManagedRedis` maps to a different Azure product (Azure Managed Redis / Enterprise) from `AzureRedisInstance` (Azure Cache for Redis / OSS). The tier letter encodes both the product and the clustering mode: S = single-node non-HA, P = HA, C = HA sharded OSSCluster. This product distinction has no parallel in the other providers.
- Alicloud Cluster's `replicasPerShard` must be `0` — read-replica configuration is not yet enabled for the current cloud-native CE cluster SKUs, so the field is effectively reserved.

---

## Backup & Restore

Cloud Manager provides NFS volume backup, restore, and scheduling across providers. The three sub-families (one-time backup, restore, scheduled backup) are shown in separate tables. Note that Azure has no `NfsVolume` resource, so its backup resources operate on raw Kubernetes PVCs, and SAP/OpenStack uses "snapshot" terminology throughout.

### One-Time Backup

<table>
<tr>
<th>AwsNfsVolumeBackup</th>
<th>GcpNfsVolumeBackup</th>
<th>AzureRwxVolumeBackup</th>
<th>SapNfsVolumeSnapshot (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
spec:
  # Required. References a Cloud Manager AwsNfsVolume.
  source:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
  # Optional. AWS Backup Vault lifecycle policy.
  # Both fields are immutable after creation.
  lifecycle:
    # Days until backup is deleted.
    deleteAfterDays: 365
    # Days until backup moves to cold storage.
    # deleteAfterDays must be >= this + 90.
    moveToColdStorageAfterDays: 30
  # Optional. Immutable. AWS Region for backup storage.
  # Defaults to the source volume's region.
  location: "us-east-1"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
spec:
  # Required. Immutable. References a Cloud Manager GcpNfsVolume.
  source:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
  # Optional. Immutable. GCP region for backup storage.
  # Defaults to the source volume's region.
  location: "us-west1"
  # Optional. Cross-cluster access control.
  # List of shootNames or subaccountIds, or ["all"] for global access.
  # Max 10 entries. "all" cannot be combined with other values.
  accessibleFrom:
    - "all"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
  # Azure has no NfsVolume resource — backs a raw PVC directly.
spec:
  # Required. Immutable. References a Kubernetes PVC.
  source:
    pvc:
      name: "my-pvc"
      namespace: "kyma-system"  # optional
  # Optional. Azure region for backup storage.
  location: "westeurope"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshot
metadata:
  name: my-snapshot
  namespace: kyma-system
  # OpenStack/CCEE uses "snapshot" terminology
  # instead of "backup".
spec:
  # Required. Immutable. References a SapNfsVolume.
  sourceVolume:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. Days after which the snapshot is auto-deleted.
  # 0 = no automatic deletion.
  deleteAfterDays: 90
```

</td>
</tr>
</table>

### Restore

<table>
<tr>
<th>AwsNfsVolumeRestore</th>
<th>GcpNfsVolumeRestore</th>
<th>AzureRwxVolumeRestore</th>
<th>SapNfsVolumeSnapshotRestore (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. AWS restores to a subdirectory
  # (aws-backup-restore_*/...), not in-place.
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  # Use either backup (object ref) or backupUrl (raw GCP URI).
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"
    # Alternative to backup (mutually exclusive):
    # backupUrl: "us-west1/my-backup-id"
  # Required. Immutable. Target volume to restore into.
  destination:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"  # optional
  # Required. Immutable. Target PVC to restore into.
  destination:
    pvc:
      name: "my-pvc"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  sourceSnapshot:
    name: "my-snapshot"
    namespace: "kyma-system"  # optional
  # Required. Immutable. Exactly one of the two modes below.
  destination:
    # Mode 1: In-place revert of an existing volume.
    # The snapshot must be the most recent one for the volume.
    existingVolume:
      name: "my-volume"
      namespace: "kyma-system"
    # Mode 2: Create a new SapNfsVolume from the snapshot.
    # (mutually exclusive with existingVolume)
    # newVolume:
    #   metadata:
    #     name: "my-new-volume"
    #     namespace: "kyma-system"
    #   spec:
    #     capacityGb: 100
    #     ipRange:
    #       name: my-iprange
```

</td>
</tr>
</table>

### Scheduled Backup

<table>
<tr>
<th>AwsNfsBackupSchedule</th>
<th>GcpNfsBackupSchedule</th>
<th>AzureRwxBackupSchedule</th>
<th>SapNfsVolumeSnapshotSchedule (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the AwsNfsVolume to back up.
  nfsVolumeRef:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. AWS region for backup storage.
  location: "us-east-1"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  # If true, deletes all child backups when schedule is deleted.
  deleteCascade: true       # default: false
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the GcpNfsVolume to back up.
  nfsVolumeRef:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. GCP region for backup storage.
  location: "us-west1"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
  # Optional. Cross-cluster access for scheduled backups.
  accessibleFrom:
    - "all"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the PVC to back up.
  # (Azure uses pvcRef, not nfsVolumeRef)
  pvcRef:
    name: "my-pvc"
    namespace: "kyma-system"  # optional
  # Optional. Azure region for backup storage.
  location: "westeurope"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
  # OpenStack schedule — the source volume is
  # specified inside the snapshot template.
spec:
  # Optional cron expression. If absent, one-shot snapshot.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated snapshot names.
  prefix: "my-snapshot"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 90      # default: 375
  maxReadySnapshots: 10     # default: 50
  maxFailedSnapshots: 3     # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
  # Template for each generated SapNfsVolumeSnapshot.
  template:
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
    spec:
      sourceVolume:
        name: "my-volume"
        namespace: "kyma-system"
      deleteAfterDays: 90
```

</td>
</tr>
</table>

**Differences across providers:**
- **Source reference model:** AWS and GCP reference a Cloud Manager `NfsVolume` object (`source.volume.name`); Azure references a raw Kubernetes PVC (`source.pvc.name`) — there is no `AzureNfsVolume`; SAP/OpenStack uses `sourceVolume` (a Kubernetes `ObjectReference`).
- **Terminology:** SAP/OpenStack uses "snapshot" throughout instead of "backup", reflecting Manila's snapshot API.
- **Cross-cluster access:** GCP Backup uniquely adds `accessibleFrom` — a list of shoot names or subaccount IDs that may restore the backup, enabling cross-cluster backup sharing. No other provider has this field.
- **Discovery resource:** GCP uniquely has a companion `GcpNfsVolumeBackupDiscovery` cluster-scoped resource (see Other Resources) that the controller populates with available backup URIs, useful for cross-cluster restore workflows.
- **Lifecycle/TTL:** AWS Backup uniquely has a `lifecycle` field (cold storage transition + deletion rules), which maps to AWS Backup Vault lifecycle policies.
- **Create-from-backup at volume creation:** GCP NfsVolume supports `sourceBackup`/`sourceBackupUrl` directly in the volume spec, so a new volume can be pre-populated from a backup without a separate Restore object. SAP/OpenStack has the analogous `dataSource.snapshot`. AWS and Azure have no equivalent shortcut.
- **Restore destination modes:** SAP/OpenStack Restore uniquely supports two modes: `existingVolume` (in-place revert — requires the snapshot to be the most recent) and `newVolume` (creates a new `SapNfsVolume` from the snapshot).
- **Schedule source field:** AWS, GCP, and Azure schedule the source via a top-level reference field (`nfsVolumeRef` / `pvcRef`); SAP/OpenStack embeds the source inside a `template.spec.sourceVolume` field.

**Differences across features (Backup vs Restore vs Schedule):**
- Schedule adds cron scheduling (`schedule`), time bounds (`startTime`, `endTime`), and retention controls (`maxRetentionDays`, `maxReadyBackups`/`maxReadySnapshots`, `maxFailedBackups`/`maxFailedSnapshots`, `deleteCascade`, `suspend`).
- Schedule references the source volume/PVC directly; the controller creates individual Backup objects for each triggered run.
- Restore is a one-shot, fire-and-forget operation — it has no scheduling or retention fields.

---

## Other Resources

The following resources were found in `api/cloud-resources/v1beta1/` but do not fit into the groupings above.

### AzureVpcDnsLink

Azure-only. Links a customer DNS zone (Azure Private DNS Zone or DNS Resolver Ruleset) to the Kyma VNet. This complements `AzureVpcPeering`: after a VNet is peered, DNS names for private endpoints in the peer network still need to resolve inside the Kyma cluster's VNet. `AzureVpcDnsLink` establishes that DNS resolution bridge. It is not a VPC peering resource and has no equivalent on AWS or GCP.

<table>
<tr>
<th>AzureVpcDnsLink</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureVpcDnsLink
metadata:
  name: my-dns-link
  # cluster-scoped, no namespace
spec:
  # Required. Immutable. Max 80 chars.
  remoteLinkName: "my-dns-link"
  # Exactly one of remotePrivateDnsZone or remoteDnsResolverRuleset
  # must be specified (XValidation enforced).

  # Option A: link to a private DNS zone.
  remotePrivateDnsZone: "my-zone.privatelink.database.windows.net"

  # Option B: link to a DNS resolver ruleset (mutually exclusive with A).
  # remoteDnsResolverRuleset: "/subscriptions/.../dnsResolverRulesets/my-ruleset"

  # Optional. Immutable. AAD tenant ID for cross-tenant DNS links.
  remoteTenant: "00000000-0000-0000-0000-000000000001"
```

</td>
</tr>
</table>

### GcpNfsVolumeBackupDiscovery

GCP-only. Cluster-scoped, empty spec. The controller populates `status.availableBackups` with metadata for all GCP Filestore backups accessible from this shoot (across all namespaces and, if configured, cross-cluster backups). Consumers read this resource to discover backup URIs before constructing a `GcpNfsVolumeRestore`. This is a read-only discovery resource — users create it empty and the controller fills in the status.

<table>
<tr>
<th>GcpNfsVolumeBackupDiscovery</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackupDiscovery
metadata:
  name: my-backup-discovery
  # cluster-scoped, no namespace
spec: {}
  # No spec fields. Controller populates status.availableBackups
  # with: uri, location, shootName, backupName, backupNamespace,
  # volumeName, volumeNamespace, creationTime.
```

</td>
</tr>
</table>

### CloudResources

Operator module CR — one per cluster, created by the Kyma module operator. This is not a cloud resource; it holds module-level status (state, served flag, conditions). End users do not create or manage this object.

<table>
<tr>
<th>CloudResources</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: CloudResources
metadata:
  name: cloud-resources
  # cluster-scoped, no namespace
spec: {}
  # No user-facing spec fields.
  # Managed by the Kyma module operator.
  # status.state: Ready | Processing | Error | Deleting | Warning
  # status.served: "True" | "False"
```

</td>
</tr>
</table>
