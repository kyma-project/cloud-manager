# GcpNfsVolume Custom Resource

The `gcpnfsvolume.cloud-resources.kyma-project.io` custom resource (CR) describes the GCP Filestore
instance that can be used as a ReadWriteMany (RWX) volume in the cluster. Once the GCP Filestore instance is provisioned
in the underlying cloud provider subscription, also the corresponding PersistentVolume (PV) and
PersistentVolumeClaim (PVC) are created in the RWX mode, so they can be used from multiple cluster workloads.
To use the GcpNfsVolume CR as a volume in the cluster workload, specify the workload volume of the `persistentVolumeClaim` type.
A created GcpNfsVolume can be deleted only where there are no workloads that
are using it, and when PV and PVC are unbound.

## IP Address Requirements <!-- {docsify-ignore} -->

The zonal GCP Filestore requires 8 and the regional GCP Filestore requires 64 IP addresses. To learn
more, read [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range). 
Those IP addresses are
allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the GcpNfsVolume,
then the default IpRange is used. If a default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology
when you want to be in control of the network segments in order to avoid range conflicts with other networks.

## Tier Selection and Capacity Planning <!-- {docsify-ignore} -->

You must specify the GcpNfsVolume capacity according to your tier selection. The `tier` field defaults to `BASIC_HDD` 
and can be one of `BASIC_HDD`, `BASIC_SSD`, `ZONAL`, or `REGIONAL`. Legacy tiers are not supported.

Each tier has specific capacity requirements and constraints:

| Tier | Min Capacity | Max Capacity | Scalability Rules | Capacity Changes |
|------|--------------|--------------|-------------------|------------------|
| **BASIC_HDD** | 1024 GiB | 65,433 GiB | Any value in range | Can be increased only (immutable downwards) |
| **BASIC_SSD** | 2560 GiB | 65,433 GiB | Any value in range | Can be increased only (immutable downwards) |
| **ZONAL** | 1024 GiB | 102,400 GiB | 1024-9984: divisible by 256<br>10240-102400: divisible by 2560 | Can be increased or decreased |
| **REGIONAL** | 1024 GiB | 102,400 GiB | 1024-9984: divisible by 256<br>10240-102400: divisible by 2560 | Can be increased or decreased |

**Examples of valid capacities:**
- BASIC_HDD: 1024, 1500, 2560, 65433
- BASIC_SSD: 2560, 3000, 10000
- ZONAL/REGIONAL: 1024, 1280 (1024+256), 1536 (1024+512), 10240, 12800 (10240+2560)

For more information, see [GCP Filestore service tiers](https://cloud.google.com/filestore/docs/service-tiers).

## PersistentVolume and PersistentVolumeClaim Management <!-- {docsify-ignore} -->

By default, the created PV and PVC have the same name as the GcpNfsVolume resource, but you can optionally
specify their names, labels and annotations if needed. If PV or PVC already exist with the same name as the one
being created, the provisioned GCP Filestore remains and the GcpNfsVolume is put into the `Error` state.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Required | Immutable | Description                                                                                                                                                                                                                                                                                                |
|-----------------------------|---------------------|----------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                 | object              | No       | Yes       | IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created automatically.                                                                                                                                                                      |
| **ipRange.name**            | string              | Yes*     | Yes       | Name of the existing IpRange to use. *Required if `ipRange` is specified.                                                                                                                                                                                                                                 |
| **capacityGb**              | int                 | Yes      | Partial   | Capacity of the volume in GiB. Must meet tier-specific requirements (see capacity table above). Note: BASIC_HDD and BASIC_SSD can only be increased. ZONAL and REGIONAL can be increased or decreased within tier constraints.                                                                            |
| **location**                | string              | No       | Yes       | **Deprecated.** This field is no longer used and will be removed in a future version. The location is always determined automatically based on the cluster region or one of its zones depending on the tier.                                                                                              |
| **tier**                    | string              | No       | Yes       | The GCP Filestore tier configuration option. Defaults to `BASIC_HDD`. Must be one of: `BASIC_HDD`, `BASIC_SSD`, `ZONAL`, or `REGIONAL`. Each tier has different capacity requirements and file share name length limits. For more information, see [Service tiers](https://cloud.google.com/filestore/docs/service-tiers). |
| **fileShareName**           | string              | No       | Yes       | The name of the file share. Defaults to `vol1`. Must match pattern `^[a-z][a-z0-9_]*[a-z0-9]$`. Maximum length: 16 characters for BASIC tiers, 63 characters for ZONAL/REGIONAL tiers.                                                                                                                   |
| **sourceBackup**            | object              | No       | Yes       | Source backup for restoring an existing backup while creating a new Filestore instance. The capacity must be equal to or greater than the source Filestore's capacity. Tier limitations also apply. Mutually exclusive with `sourceBackupUrl`. See [GcpNfsVolumeRestore](./04-20-23-gcp-nfs-volume-restore.md). |
| **sourceBackup.name**       | string              | Yes*     | Yes       | Name of the source GcpNfsVolumeBackup. *Required if `sourceBackup` is specified.                                                                                                                                                                                                                          |
| **sourceBackup.namespace**  | string              | Yes*     | Yes       | Namespace of the source GcpNfsVolumeBackup. *Required if `sourceBackup` is specified.                                                                                                                                                                                                                     |
| **sourceBackupUrl**         | string              | No       | Yes       | Direct URL to a backup in the format `<project>/<location>/<instance>/<backup>`. Mutually exclusive with `sourceBackup`. Must match pattern `^(.+)/(.+)$`.                                                                                                                                               |
| **volume**                  | object              | No       | No        | PersistentVolume options. Optional.                                                                                                                                                                                                                                                                       |
| **volume.name**             | string              | No       | No        | PersistentVolume name. Defaults to the name of the GcpNfsVolume resource.                                                                                                                                                                                                                                |
| **volume.labels**           | map\[string\]string | No       | No        | PersistentVolume labels. Defaults to nil.                                                                                                                                                                                                                                                                 |
| **volume.annotations**      | map\[string\]string | No       | No        | PersistentVolume annotations. Defaults to nil.                                                                                                                                                                                                                                                            |
| **volumeClaim**             | object              | No       | No        | PersistentVolumeClaim options. Optional.                                                                                                                                                                                                                                                                  |
| **volumeClaim.name**        | string              | No       | Yes       | PersistentVolumeClaim name. Defaults to the name of the GcpNfsVolume resource. Cannot be changed after creation.                                                                                                                                                                                          |
| **volumeClaim.labels**      | map\[string\]string | No       | No        | PersistentVolumeClaim labels. Defaults to nil.                                                                                                                                                                                                                                                            |
| **volumeClaim.annotations** | map\[string\]string | No       | No        | PersistentVolumeClaim annotations. Defaults to nil.                                                                                                                                                                                                                                                       |

**Status:**

| Parameter                         | Type       | Description                                                                                                                                         |
|-----------------------------------|------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Current state of the GcpNfsVolume. Possible values: `Ready`, `Processing`, `Creating`, `Updating`, `Deleting`, or `Error`.                          |
| **id**                            | string     | The GCP Filestore instance ID.                                                                                                                      |
| **hosts**                         | \[\]string | List of NFS hosts (DNS names or IP addresses) that clients can use to connect to the volume.                                                        |
| **location**                      | string     | The location where the volume is provisioned (automatically determined based on cluster region/zones and tier).                                     |
| **capacity**                      | quantity   | The provisioned capacity of the volume. Matches `spec.capacityGb` when provisioning or capacity changes are complete.                               |
| **protocol**                      | string     | The NFS protocol used (typically `NFSv3`).                                                                                                          |
| **conditions**                    | \[\]object | Current conditions of the resource. Used for detailed status tracking.                                                                               |
| **conditions.lastTransitionTime** | string     | Timestamp of the last condition status change.                                                                                                      |
| **conditions.message**            | string     | Human-readable message providing details about the condition status change.                                                                         |
| **conditions.reason**             | string     | Machine-readable reason for the condition status change (e.g. `IpRangeNotReady`).                             |
| **conditions.status** (required)  | string     | Status of the condition. Possible values: `True`, `False`, or `Unknown`.                                                                            |
| **conditions.type**               | string     | Type of condition (e.g., `Ready`, `Error`).                                                                                                         |

## Sample Custom Resources <!-- {docsify-ignore} -->

### Basic Example with Default Settings

This example creates a basic HDD volume with the minimum required configuration:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-basic-volume
spec:
  capacityGb: 1024  # Minimum for BASIC_HDD tier
---
apiVersion: v1
kind: Pod
metadata:
  name: basic-workload
spec:
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: my-basic-volume
  containers:
    - name: workload
      image: nginx
      volumeMounts:
        - mountPath: "/mnt/data"
          name: data
```

### High-Performance SSD Volume

This example creates a high-performance SSD volume with custom PV/PVC names:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-ssd-volume
spec:
  tier: BASIC_SSD
  capacityGb: 2560  # Minimum for BASIC_SSD tier
  fileShareName: ssdshare
  volume:
    name: custom-pv-name
    labels:
      environment: production
      tier: ssd
  volumeClaim:
    name: custom-pvc-name
    labels:
      environment: production
```

### Regional Volume for High Availability

This example creates a regional volume with proper capacity alignment:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-regional-volume
spec:
  tier: REGIONAL
  capacityGb: 1280  # 1024 + 256 (divisible by 256)
  fileShareName: regional_share_001
  ipRange:
    name: my-custom-range
```

### Zonal Volume with Large Capacity

This example creates a zonal volume with capacity in the upper range:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-large-volume
spec:
  tier: ZONAL
  capacityGb: 12800  # 10240 + 2560 (divisible by 2560 in upper range)
  fileShareName: large_zonal_share
```

## Related Resources <!-- {docsify-ignore} -->

- [IpRange CR](./04-10-iprange.md) - IP address range management
- [GcpNfsVolumeBackup CR](./04-20-21-gcp-nfs-volume-backup.md) - Backup management
- [GcpNfsVolumeRestore CR](./04-20-23-gcp-nfs-volume-restore.md) - Restore procedures
- [GCP Filestore Documentation](https://cloud.google.com/filestore/docs) - Official Google Cloud documentation
