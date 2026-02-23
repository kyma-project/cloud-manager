# GcpNfsVolumeBackup Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `gcpnfsvolumebackup.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes the GCP Filestore
instance's backup.
While the GCP Filestore backup is created in the underlying cloud provider subscription, it needs its source GCP 
Filestore instance to be available. But upon its creation, it can be used independently of the source instance.

GCP Filestore backups are regional resources, and they are created in the same region as the source GCP Filestore 
instance unless specified otherwise.

For a given Gcp Filestore, backups are incremental, as long as they are created on the same region. 
This reduces latency on backup creation. However, if a backup is created in a different region from the latest backup, 
it will be a full backup.
To learn more, read [Filestore Backup Creation](https://cloud.google.com/filestore/docs/backups#backup-creation).

## Cross-Cluster Backup Sharing <!-- {docsify-ignore} -->

Backups can be shared across clusters within the same global account and GCP project using the **accessibleFrom** field. This enables scenarios such as:

- **Disaster Recovery**: Restore data from a backup created in another cluster
- **Data Migration**: Move data between development, staging, and production environments
- **Cross-Environment Testing**: Use production data snapshots in test environments

To discover backups shared from other clusters, use the [GcpNfsVolumeBackupDiscovery](04-20-24-gcp-nfs-volume-backup-discovery.md) resource.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type       | Description                                                                                                                                                                                                                                           |
|-----------------------------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **source**                  | object     | Required. Immutable. Specifies the source of the backup.                                                                                                                                                                                              |
| **source.volume**           | object     | Required. Reference of the existing source GcpNfsVolume that is backed up.                                                                                                                                                                            |
| **source.volume.name**      | string     | Required. Name of the source GcpNfsVolume.                                                                                                                                                                                                            |
| **source.volume.namespace** | string     | Optional. Namespace of the source GcpNfsVolume. Defaults to the namespace of the GcpNfsVolumeBackup resource if not provided.                                                                                                                         |
| **location**                | string     | Optional. Immutable. The GCP region where the backup is stored. If left empty, it defaults to the region of the cluster. Must be a valid [GCP region](https://cloud.google.com/filestore/docs/regions).                                              |
| **accessibleFrom**          | \[\]string | Optional. Array of shoot names or subaccount IDs that are granted access to restore from this backup. Use `"all"` to allow access from all shoots in the same global account and GCP project. `"all"` cannot be combined with other values. Max 10 items. |

**Status:**

| Parameter                         | Type              | Description                                                                                                                     |
|-----------------------------------|-------------------|---------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string            | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Creating`, `Error`, `Deleting`, `Deleted`, or `Failed`. |
| **id**                            | string            | The internal backup identifier (UUID).                                                                                          |
| **location**                      | string            | Signifies the location of the backup. This is particularly useful if location is not provided in the spec.                     |
| **capacity**                      | Quantity          | Provides the storage size of the backup.                                                                                        |
| **accessibleFrom**                | string            | Comma-separated list reflecting the **accessibleFrom** field in spec after the last successful reconciliation.                 |
| **conditions**                    | \[\]object        | Represents the current state of the CR's conditions.                                                                            |
| **conditions.lastTransitionTime** | string            | Defines the date of the last condition status change.                                                                           |
| **conditions.message**            | string            | Provides more details about the condition status change.                                                                        |
| **conditions.reason**             | string            | Defines the reason for the condition status change.                                                                             |
| **conditions.status** (required)  | string            | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                      |
| **conditions.type**               | string            | Provides a short description of the condition.                                                                                  |

## Sample Custom Resources <!-- {docsify-ignore} -->

### Basic Backup

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: my-backup
spec:
  source:
    volume:
      name: my-vol
  location: us-west1
```

### Backup with Cross-Cluster Access

Grant specific clusters access to restore from this backup:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: shared-backup
spec:
  source:
    volume:
      name: my-vol
      namespace: production
  accessibleFrom:
    - "a1b2c3d"   # shoot ID of target cluster
    - "e4f5g6h"   # another cluster's shoot ID
```

### Backup Accessible to All Clusters

Allow all clusters in the same global account and GCP project to restore from this backup:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: globally-accessible-backup
spec:
  source:
    volume:
      name: critical-data
  accessibleFrom:
    - "all"
```

## Related Resources <!-- {docsify-ignore} -->

- [GcpNfsVolume](04-20-20-gcp-nfs-volume.md) - The source volume for backups
- [GcpNfsVolumeBackupDiscovery](04-20-24-gcp-nfs-volume-backup-discovery.md) - Discover backups shared from other clusters
- [GcpNfsBackupSchedule](04-20-22-gcp-nfs-backup-schedule.md) - Automate backup creation on a schedule
