# Features, the CRD→feature mapping, and how `apiDisabled` controls exposed APIs

## Why `feature` is derived, not typed

Targeting rules like `query: feature == "nfsBackup"` compare against the `feature` context key. That key is
**not** free text the user supplies — it is set from the resource being reconciled. Each CRD type
implements `SpecificToFeature()` (see `api/cloud-resources/v1beta1/*_types.go`) returning one of the feature
constants. To evaluate correctly, first map the resource the user is asking about to its feature name.

If `SpecificToFeature()` returns `""` (empty), the object is **not specific to one feature** — no
`feature == "…"` rule matches it, so it falls through to `defaultRule` (unless another key like
`brokerPlan` or `globalAccount` matches). `CloudResources` and `IpRange` return empty.

## CRD kind → feature name

| CRD kind | `feature` |
|---|---|
| AwsNfsVolume, GcpNfsVolume, SapNfsVolume | `nfs` |
| AwsNfsVolumeBackup, AwsNfsVolumeRestore, AwsNfsBackupSchedule | `nfsBackup` |
| GcpNfsVolumeBackup, GcpNfsVolumeRestore, GcpNfsBackupSchedule, GcpNfsVolumeBackupDiscovery | `nfsBackup` |
| AzureRwxVolumeBackup, AzureRwxVolumeRestore, AzureRwxBackupSchedule | `nfsBackup` |
| SapNfsVolumeSnapshot, SapNfsVolumeSnapshotRestore, SapNfsVolumeSnapshotSchedule | `nfsBackup` |
| AwsVpcPeering, AzureVpcPeering, GcpVpcPeering | `peering` |
| AwsRedisInstance, AzureRedisInstance, GcpRedisInstance | `redis` |
| AwsRedisCluster, AzureRedisCluster, GcpRedisCluster | `rediscluster` |
| GcpSubnet | `rediscluster` (temporary; moves to undefined after GcpRedisCluster GA) |
| AzureManagedRedis | `azureManagedRedis` |
| AzureVpcDnsLink | `vpcdnslink` |
| CloudResources, IpRange | `""` (not feature-specific) |

This is the source of truth as of the mapping extracted from the code. If a new CRD is added, grep
`api/cloud-resources/v1beta1/*_types.go` for `SpecificToFeature` to confirm. Some feature values seen in
configs (e.g. `waf`, `certificate`) are for CRDs not yet in this list or handled elsewhere — treat a
config `feature == "X"` as authoritative for what the rule targets even if X isn't in the table.

## How `apiDisabled` controls which APIs are exposed

`apiDisabled` is the master switch for **which CRDs are installed to an SKR and which controllers start**.
Its variations are inverted, which trips people up:

```yaml
variations:
  enabled: false     # "enabled" => apiDisabled=false => the API IS available
  disabled: true     # "disabled" => apiDisabled=true  => the API is NOT installed
```

So `variation: enabled` means the API is **exposed**; `variation: disabled` means it's **hidden**. Resolve
the label through `variations`, never assume.

The development/rollout flow this implements:
1. New CRD + reconciler added; not GA yet.
2. **dev** config enables everything → all APIs available for testing on dev SKRs.
3. **stage/prod** enable everything for the **Phoenix team global account**
   (masked here as `00000000-0000-0000-0000-000000000000`) via a top rule, then **disable** all non-GA APIs below it.
4. **Front-runner clients** on stage/prod get specific features enabled above the disable rules, targeted by
   `globalAccount`, `subAccount`, or `shoot`.
5. GA features have no disable rule, so they fall through to `defaultRule: enabled`.

This is why rule **order** is everything: the per-client `Enable …` rules sit *above* the blanket
`Disable …` rules, and the Phoenix `Enable all` rule sits at the very top.

## The `CloudResources` code-level override

`ffApiDisabled.go` short-circuits before consulting the flag: if `allKindGroups` contains
`cloudresources.cloud-resources.kyma-project.io`, `Value()` returns `false` (API enabled) regardless of any
targeting rule. The `CloudResources` CRD is the entrypoint resource and is **always** exposed. When a
question is about the `CloudResources` kind specifically, the answer is "always enabled, by code, not by
config."