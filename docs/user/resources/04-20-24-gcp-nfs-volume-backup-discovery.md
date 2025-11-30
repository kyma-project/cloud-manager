# GcpNfsVolumeBackupDiscovery Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `gcpnfsvolumebackupdiscovery.cloud-resources.kyma-project.io` cluster-scoped custom resource (CR) enables discovering GCP Filestore backups that are accessible to the current cluster. This resource is part of the GcpNfsVolumeBackup family and provides a way to find backups created in other clusters that have granted access through the `AccessibleFrom` field.

## How It Works

The backup discovery mechanism works through cross-cluster accessibility:

1. **Backup Creation**: A user in cluster A creates a GcpNfsVolumeBackup and includes the shoot ID of cluster B in the **AccessibleFrom** field.
2. **Discovery Request**: A user in cluster B applies a GcpNfsVolumeBackupDiscovery resource.
3. **Discovery Process**: The controller reconciles the discovery resource and searches for backups that have cluster B's shoot ID in their **AccessibleFrom** configuration.
4. **Results**: Discovered backups are listed in the **.status.availableBackups** field with detailed metadata.

> [!NOTE]
> The discovery resource is reconciled only once when created. The results represent a snapshot in time of available backups at that moment. If new backups are created later and granted access to this cluster, existing discovery resources will not automatically reflect these new backups. To discover newly available backups, you must create a new GcpNfsVolumeBackupDiscovery resource.

This enables secure cross-cluster backup sharing within the same global account and GCP project.

## Use Cases

- **Disaster Recovery**: Discover backups from a failed cluster to restore data in a new cluster
- **Data Migration**: Find backups created in one environment to restore in another
- **Cross-Environment Access**: Share specific backups between development, staging, and production clusters
- **Backup Inventory**: Get an overview of all accessible backups from other clusters

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

The GcpNfsVolumeBackupDiscovery resource currently has an empty specification. Discovery is automatically performed based on the cluster's shoot ID and available backups in the same GCP project.

**Status:**

| Parameter                                     | Type       | Description                                                                                                                        |
|-----------------------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state**                                     | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                                | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime**             | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**                        | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**                         | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)              | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**                           | string     | Provides a short description of the condition.                                                                                     |
| **discoverySnapshotTime**                     | string     | The timestamp when the discovery operation was performed. This represents a point-in-time snapshot of available backups.           |
| **availableBackupsCount**                     | int        | The total number of backups discovered that are accessible to this cluster.                                                        |
| **availableBackupUris**                       | \[\]string | A list of URIs for all discovered backups.                                                                                         |
| **availableBackups**                          | \[\]object | Detailed information about each discovered backup.                                                                                  |
| **availableBackups.uri**                      | string     | The backup URI in the format `{locationId}/{backupId}`.                                                                             |
| **availableBackups.location**                 | string     | The GCP region where the backup is stored.                                                                                         |
| **availableBackups.shootName**                | string     | The name of the shoot (cluster) where the backup was originally created.                                                           |
| **availableBackups.backupName**               | string     | The name of the original GcpNfsVolumeBackup resource that created this backup.                                                     |
| **availableBackups.backupNamespace**          | string     | The namespace of the original GcpNfsVolumeBackup resource.                                                                         |
| **availableBackups.volumeName**               | string     | The name of the GcpNfsVolume that was backed up.                                                                                   |
| **availableBackups.volumeNamespace**          | string     | The namespace of the GcpNfsVolume that was backed up.                                                                              |
| **availableBackups.creationTime**             | string     | The timestamp when the backup was created.                                                                                         |

## Sample Custom Resource

See an exemplary GcpNfsVolumeBackupDiscovery custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackupDiscovery
metadata:
  name: my-backup-discovery
spec: {}
---
# Example of discovered backups in status
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackupDiscovery
metadata:
  name: my-backup-discovery
spec: {}
status:
  state: Ready
  discoverySnapshotTime: "2024-01-15T10:30:00Z"
  availableBackupsCount: 2
  availableBackupUris:
    - "us-west1/cm-12345678-1234-5678-9abc-123456789012"
    - "us-east1/cm-87654321-4321-8765-cba9-210987654321"
  availableBackups:
    - uri: "us-west1/cm-12345678-1234-5678-9abc-123456789012"
      location: us-west1
      shootName: b2738ca
      backupName: my-volume-backup
      backupNamespace: default
      volumeName: my-nfs-volume
      volumeNamespace: default
      creationTime: "2024-01-14T14:20:00Z"
    - uri: "us-east1/cm-87654321-4321-8765-cba9-210987654321"
      location: us-east1
      shootName: c-65cc021
      backupName: critical-data-backup
      backupNamespace: production
      volumeName: critical-volume
      volumeNamespace: production
      creationTime: "2024-01-13T09:15:00Z"
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2024-01-15T10:30:00Z"
      reason: Ready
      message: Successfully discovered available Nfs Volume Backups from GCP
```

## Related Resources

- [GcpNfsVolumeBackup](./04-20-21-gcp-nfs-volume-backup.md) - Create backups of GCP Filestore volumes
- [GcpNfsVolumeRestore](./04-20-23-gcp-nfs-volume-restore.md) - Restore volumes from discovered backups
- [GcpNfsVolume](./04-20-20-gcp-nfs-volume.md) - The source volume type for backups