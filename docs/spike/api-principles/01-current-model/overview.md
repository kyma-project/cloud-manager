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

### Unification 

No shape differences. Future demand for various types can be encapsulated within the class/type attribute that determines the reconciliation algorithm and APIs used.

**Portable intent**

> 💡Pattern taken from core Kubernetes resources like Service.type and Secret.type indicating a predefined enumerable set of possible values, without additional configuration.

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: IpRange
metadata:
  name: my-iprange
  # cluster-scoped, no namespace
spec:
  cidr: "10.250.0.0/22"
  type: <ip-range-type>
```

Types:
- aws/shared-subnet
- azure/shared-subnet
- azure/app-gateway-subnet
- gcp/psa
- gcp/psc-redis
- sap/shared-subnet
- alicloud/shared-subnet

### Technical note ⚠️

IpRange and GcpSubnet carry significant network reconfiguration pre-requisites that are hidden from the user, and their outcomes are the foundation for all other network bound resources (NFS, Redis, WAF, DNS...). Problem is lack of user exposed address space configuration. For some providers their reconciliation is extremely complex, constrained and cardinality sensitive. Deeper grooming required!

### Honestly what we get 🎁

Two kinds of same shape collapsed into one kind brings UI/API simplification, but on technical side it brings additional entanglement of unrelated functionality leveraged with all networking functional debt.

---

## VpcPeering

VpcPeering connects the Kyma cluster's VPC to a customer-owned VPC in the same cloud provider. All three resources are fully provider-specific; the underlying peering mechanism, identity model, and routing options differ enough that no field set is shared beyond `deleteRemotePeering`.

<table>
<tr>
<th>AwsVpcPeering</th>
<th>AzureVpcPeering</th>
<th>GcpVpcPeering</th>
</tr>
<tr style="vertical-align: top;">
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

### Unification

The remote VPC identifier can be collapsed into a single composite string representation specific to cloud provider:
- AWS - `arn:aws:ec2:us-east-1:123456789012:vpc/vpc-1234567890abcdef0`
- Azure - already in use - `/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet`
- GCP - `projects/my-project/global/networks/my-net`

**⚠️ Risk** providers not having a specific standard resource identifier where we must define **own custom format**.

**Otherwise**, the remote VPC identifier moves to the provider specific configuration and portable intent resource remains empty, which doesn't make much sense.

**Provider specific configuration** can be externalized into separate shapes.

<table>
<tr>
<th>Aws VpcPeeringConfig</th>
<th>Azure VpcPeeringConfig</th>
<th>Gcp VpcPeeringConfig</th>
</tr>
<tr>
<td>

```yaml
apiVersion: aws.cloud-resources.kyma-project.io/v1beta1
kind: VpcPeeringConfig
metadata:
  name: my-vpcpeering-config
  # cluster-scoped, no namespace
spec:
  # Immutable. Enum: AUTO, NONE, MATCHED, UNMATCHED
  remoteRouteTableUpdateStrategy: "AUTO"
```

</td>
<td>

```yaml
apiVersion: azure.cloud-resources.kyma-project.io/v1beta1
kind: VpcPeeringConfig
metadata:
  name: my-vpcpeering-config
  # cluster-scoped, no namespace
spec:
  # Required. Immutable. Max 80 chars.
  remotePeeringName: "my-peering"
  # Optional. Immutable. Default false.
  useRemoteGateway: false
```

</td>
<td>

```yaml
apiVersion: gcp.cloud-resources.kyma-project.io/v1beta1
kind: VpcPeeringConfig
metadata:
  name: my-vpcpeering-config
  # cluster-scoped, no namespace
spec:
  # Required. Immutable. 1-63 chars, lowercase alphanumeric+hyphens.
  remotePeeringName: "my-peering"
  # Immutable. Default false.
  importCustomRoutes: false
```

</td>
</tr>
</table>

**Portable intent** with inlined provider specific config.

> 💡 Pattern taken from Kubernetes ObjectReference like Event involvedObject, or Pod spec.volumes[*].configMap

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: VpcPeering
metadata:
  name: my-vpcpeering-config
  # cluster-scoped, no namespace
spec:
  # Required. Remote VPC Network identifier to peer with
  remoteId: "provider-specific-resource-id-containing-all-relevant-attributes"

  # Default: false. Whether to delete the remote side of the peering on delete.
  deleteRemotePeering: true

  # Required
  config:
    apiVersion: aws|azure|gcp.cloud-resources.kyma-project.io/v1beta1
    kind: VpcPeeringConfig
    name: my-vpcpeering-config
```

### Honestly what we get 🎁

N shapes replaces by N+1 shapes, where each CR still has all the fields it had before just in slightly different syntax. Must decide how it's driven by feature flags.

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
<tr style="vertical-align: top;">
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolume
metadata:
  name: my-volume
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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

### Unification

The common fields that can stay in the portable intent resource:
- ipRange
- volume
- volumeClaim

Some have restore related source field that can bound to the typed resource.

**Provider specific configuration** can be extracted into own types.

<table>
<tr>
<th>Aws NfsConfig</th>
<th>Azure NfsConfig</th>
<th>Alicloud NfsConfig</th>
<th>Sap NfsConfig</th>
</tr>
<tr>
<td>

```yaml
apiVersion: aws.cloud-resources.kyma-project.io/v1beta1
kind: NfsConfig
spec:
  # Default: generalPurpose. Enum: generalPurpose, maxIO.
  performanceMode: "generalPurpose"
  # Default: bursting. Enum: bursting, elastic.
  throughput: "bursting"
```

</td>
<td>

```yaml
apiVersion: azure.cloud-resources.kyma-project.io/v1beta1
kind: NfsConfig
spec:
  # Enum: BASIC_HDD, BASIC_SSD, ZONAL, REGIONAL.
  tier: "BASIC_HDD"
  # Capacity in GiB. Default: 2560. Valid ranges depend on tier.
  capacityGb: 1024

```

</td>
<td>

```yaml
apiVersion: alicloud.cloud-resources.kyma-project.io/v1beta1
kind: NfsConfig
spec:
  # Enum: Performance, Capacity, Premium.
  storageType: "Performance"
```

</td>
<td>

```yaml
apiVersion: sap.cloud-resources.kyma-project.io/v1beta1
kind: NfsConfig
spec:
  # Required. Integer GiB (not Kubernetes quantity). Must be > 0.
  capacityGb: 100
```

</td>
</tr>
</table>

**Portable intent** with inlined provider specific config.

> 💡 Pattern taken from Kubernetes ObjectReference like Event involvedObject, or Pod spec.volumes[*].configMap

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: NfsVolume
metadata:
  name: my-volume
  namespace: my-ns
spec:
  # Optional. Immutable.
  ipRange:
    name: my-iprange
  
  # Required
  config:
    apiVersion: aws|gcp|alicloud|sap.cloud-resources.kyma-project.io/v1beta1
    kind: NfsConfig
    name: my-nfs-config
    namespace: my-ns
    
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

  # Optional. Specified when created from backup/snapshot
  source:
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: Backup|Snapshot
    namespace: some-namespace
    name: my-backup
```

### Honestly what we get 🎁

N shapes replaces by N+1 shapes, where each CR still has all the fields it had before just in slightly different syntax. Must decide how it's driven by feature flags (ie Azure doesn't have NFS at all).

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
<tr style="vertical-align: top;">
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisInstance
metadata:
  name: my-redis
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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
  namespace: my-ns
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

### Unification

The common fields that can stay in the portable intent resource:
- ipRange
- authSecret

**Provider specific configuration** can be externalized into separate shapes.

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
<tr style="vertical-align: top;">
<td>

```yaml
apiVersion: aws.cloud-resources.kyma-project.io/v1beta1
kind: RedisInstanceConfig
spec:
  # Required. S1-S7 or P1-P6.
  # Service letter (S/P) immutable; only capacity number is mutable.
  redisTier: "S1"
  # Default: "7.0". Enum: "6.x", "7.0", "7.1".
  # Upgrade-only (cannot downgrade).
  engineVersion: "7.0"
  # Default: false.
  autoMinorVersionUpgrade: false
  # Optional. Format: ddd:hh24:mi-ddd:hh24:mi
  preferredMaintenanceWindow: "sun:23:00-mon:01:30"
  # Optional. Freeform Redis config map.
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
```

</td>
<td>

```yaml
apiVersion: gcp.cloud-resources.kyma-project.io/v1beta1
kind: RedisInstanceConfig
spec:
  # Required. S1-S8 or P1-P6. Service letter (S/P) immutable.
  redisTier: "P1"
  # Default: "REDIS_7_0". 
  # Enum: REDIS_6_X, REDIS_7_0, REDIS_7_2.
  # Upgrade-only (cannot downgrade).
  redisVersion: "REDIS_7_0"
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
```

</td>
<td>

```yaml
apiVersion: azure.cloud-resources.kyma-project.io/v1beta1
kind: RedisInstanceConfig
spec:
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
```

</td>
<td>

```yaml
apiVersion: alicloud.cloud-resources.kyma-project.io/v1beta1
kind: RedisInstanceConfig
spec:
  # Required. S1-S5 or P1-P5.
  # Service letter (S/P) immutable.
  redisTier: "S1"
  # Default: "7.0". Enum: "5.0", "6.0", "7.0".
  # Immutable after creation (unlike AWS/GCP).
  engineVersion: "7.0"
```

</td>
<td>

```yaml
apiVersion: aws.cloud-resources.kyma-project.io/v1beta1
kind: RedisClusterConfig
spec:
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
  # Optional. Format: ddd:hh24:mi-ddd:hh24:mi
  preferredMaintenanceWindow: "sun:23:00-mon:01:30"
  # Optional. Freeform Redis config map.
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
```

</td>
<td>

```yaml
apiVersion: gcp.cloud-resources.kyma-project.io/v1beta1
kind: RedisClusterConfig
spec:
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
```

</td>
<td>

```yaml
apiVersion: azure.cloud-resources.kyma-project.io/v1beta1
kind: RedisCluster
spec:
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
```

</td>
<td>

```yaml
apiVersion: alicloud.cloud-resources.kyma-project.io/v1beta1
kind: RedisClusterConfig
spec:
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
```

</td>
<td>

```yaml
apiVersion: azure.cloud-resources.kyma-project.io/v1beta1
kind: ManagedRedisConfig
spec:
  # Required. S1-S5, P1-P5, or C3-C7.
  # Family letter is immutable:
  #   S = Balanced non-HA EnterpriseCluster (dev)
  #   P = ComputeOptimized HA EnterpriseCluster
  #   C = ComputeOptimized HA OSSCluster (sharded)
  redisTier: "P1"
```

</td>
</tr>
</table>


**Portable intent** resource:

> 💡 Pattern taken from Kubernetes ObjectReference like Event involvedObject, or Pod spec.volumes[*].configMap

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: Redis
metadata:
  name: my-redis
  namespace: some-ns
spec:
  # Optional.
  ipRange:
    name: my-iprange
  
  # Optional. Controls the generated connection secret.
  authSecret:
    name: "my-redis-auth"
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
  
  # Required
  config:
    apiVersion: aws|azure|gcp|alicloud|sap.cloud-resources.kyma-project.io/v1beta1
    kind: RedisConfig|RedisClusterConfig|ManagedRedisConfig
    name: my-redis-config
    namespace: my-ns
```

### Honestly what we get 🎁

N shapes replaces by N+1 shapes, where each CR still has all the fields it had before just in slightly different syntax. Must decide how it's driven by feature flags.


---

## Other Resources

The following resources were found in `api/cloud-resources/v1beta1/` but do not fit into the groupings above.

### AzureVpcDnsLink

Azure-only. Links a customer DNS zone (Azure Private DNS Zone or DNS Resolver Ruleset) to the Kyma VNet. This complements `AzureVpcPeering`: after a VNet is peered, DNS names for private endpoints in the peer network still need to resolve inside the Kyma cluster's VNet. `AzureVpcDnsLink` establishes that DNS resolution bridge. It is not a VPC peering resource and has no equivalent on AWS or GCP.

**Lack of related resources from other providers dot not provide ground for the comparison**.

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

### CloudResources

Operator module CR — one per cluster, created by the Kyma module operator. This is not a cloud resource; it holds module-level status (state, served flag, conditions). End users do not create or manage this object.

**Required by KLM, must remain**.

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

[Other excluded resources](./excluded.md)
