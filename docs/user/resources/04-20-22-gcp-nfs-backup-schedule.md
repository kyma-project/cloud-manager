# GcpNfsBackupSchedule Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `gcpnfsbackupschedule.cloud-resources.kyma-project.io` namespaced custom resource (CR) automates
[GcpNfsVolumeBackup](04-20-21-gcp-nfs-volume-backup.md) creation for a given [GcpNfsVolume](04-20-20-gcp-nfs-volume.md).
It supports both recurring schedules using CRON expressions and one-time backups.

The CR performs the following actions:

- Creates `GcpNfsVolumeBackup` resources automatically at the specified interval or once at a given time.
- Automatically deletes backups that exceed the configured retention period (`maxRetentionDays`) or count limits (`maxReadyBackups`, `maxFailedBackups`).
- Enables you to temporarily suspend or resume backup creation and deletion.

Created backups are named using the pattern `{prefix}-{index}-{YYYYMMDD-HHMMSS}`, where `prefix` defaults to the
schedule name, `index` is an auto-incrementing counter, and the timestamp reflects the scheduled run time in UTC.

## Cross-Cluster Backup Sharing <!-- {docsify-ignore} -->

Backups created by this schedule can be shared across clusters within the same global account and GCP project using the **accessibleFrom** field. This enables scenarios such as:

- **Disaster Recovery**: Restore data from a backup created in another cluster
- **Data Migration**: Move data between development, staging, and production environments
- **Cross-Environment Testing**: Use production data snapshots in test environments

To discover backups shared from other clusters, use the [GcpNfsVolumeBackupDiscovery](04-20-24-gcp-nfs-volume-backup-discovery.md) resource.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type       | Required | Immutable | Description                                                                                                                                                                                                                     |
|-----------------------------|------------|----------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **nfsVolumeRef**            | object     | Yes      | No        | Reference to an existing GcpNfsVolume that must be in `Ready` state.                                                                                                                                                           |
| **nfsVolumeRef.name**       | string     | Yes      | No        | Name of the existing GcpNfsVolume.                                                                                                                                                                                             |
| **nfsVolumeRef.namespace**  | string     | No       | No        | Namespace of the existing GcpNfsVolume. Defaults to the namespace of the GcpNfsBackupSchedule resource if not provided.                                                                                                        |
| **location**                | string     | No       | No        | The GCP region where backups are stored. Defaults to the region of the source GcpNfsVolume. Must be a valid [GCP region](https://cloud.google.com/filestore/docs/regions).                                                     |
| **schedule**                | string     | No       | No        | CRON expression for the schedule. When empty or not specified, the schedule runs only once — at the specified `startTime`, or immediately if `startTime` is not set. See also [Schedule Syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax). |
| **prefix**                  | string     | No       | No        | Prefix for the name of the created `GcpNfsVolumeBackup` resources. Defaults to the name of this schedule.                                                                                                                      |
| **startTime**               | string     | No       | No        | Start time for the schedule in RFC 3339 format (e.g., `2026-06-01T00:00:00Z`). Value cannot be before the resource creation time. When not specified, the schedule becomes effective immediately.                               |
| **endTime**                 | string     | No       | No        | End time for the schedule in RFC 3339 format. Value cannot be before the `startTime` or, if `startTime` is not set, before the resource creation time. When not specified, the schedule runs indefinitely.                     |
| **maxRetentionDays**        | int        | No       | No        | Maximum number of days to retain backup resources. Backups older than this are automatically deleted. Default: 375. Minimum: 1.                                                                                                |
| **maxReadyBackups**         | int        | No       | No        | Maximum number of backups in `Ready` state to retain. Oldest backups exceeding this count are automatically deleted. Default: 100. Minimum: 1.                                                                                 |
| **maxFailedBackups**        | int        | No       | No        | Maximum number of backups in `Failed` state to retain. Oldest backups exceeding this count are automatically deleted. Default: 5. Minimum: 1.                                                                                  |
| **suspend**                 | boolean    | No       | No        | Specifies whether to suspend the schedule temporarily. While suspended, no backups are created or deleted. Defaults to `false`.                                                                                                |
| **deleteCascade**           | boolean    | No       | No        | Specifies whether to cascade delete all backup resources when this schedule is deleted. When `false`, backups are orphaned and must be deleted manually or via their own retention. Defaults to `false`.                        |
| **accessibleFrom**          | \[\]string | No       | No        | Array of shoot names or subaccount IDs that have access to the backups created by this schedule for restore. Use `"all"` to allow access from all shoots in the same global account and GCP project. `"all"` cannot be combined with other values. Max 10 items. |

**Status:**

| Parameter                         | Type              | Description                                                                                                                            |
|-----------------------------------|-------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string            | Signifies the current state of **CustomObject**. Its value can be either `Active`, `Suspended`, `Done`, `Error`, or `Deleting`.        |
| **conditions**                    | \[\]object        | Represents the current state of the CR's conditions.                                                                                   |
| **conditions.lastTransitionTime** | string            | Defines the date of the last condition status change.                                                                                  |
| **conditions.message**            | string            | Provides more details about the condition status change.                                                                               |
| **conditions.reason**             | string            | Defines the reason for the condition status change.                                                                                    |
| **conditions.status** (required)  | string            | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                             |
| **conditions.type**               | string            | Provides a short description of the condition.                                                                                         |
| **nextRunTimes**                  | \[\]string        | Provides a preview of up to 3 upcoming times when the next backups will be created.                                                    |
| **nextDeleteTimes**               | map\[string\]string | Provides the backup objects and their expected deletion time (calculated based on `maxRetentionDays`).                                |
| **lastCreateRun**                 | string            | Provides the time when the last backup was created.                                                                                    |
| **lastCreatedBackup**             | objectRef         | Provides the object reference of the last created backup.                                                                              |
| **lastDeleteRun**                 | string            | Provides the time when the last backup was deleted.                                                                                    |
| **lastDeletedBackups**            | \[\]objectRef     | Provides the object references of the last deleted backups.                                                                            |
| **schedule**                      | string            | Provides the cron expression of the current active schedule.                                                                           |
| **backupIndex**                   | int               | Provides the current index of the backup created by this schedule.                                                                     |
| **backupCount**                   | int               | Provides the number of backups currently present in the system.                                                                        |

## Sample Custom Resources <!-- {docsify-ignore} -->

### Recurring Daily Backup

Create a daily backup at midnight, retaining the last 7 backups for up to 30 days:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsBackupSchedule
metadata:
  name: daily-backup
spec:
  nfsVolumeRef:
    name: my-nfs-volume
  schedule: "0 0 * * *"
  prefix: daily
  maxRetentionDays: 30
  maxReadyBackups: 7
```

### One-Time Backup

Create a single backup immediately without a recurring schedule:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsBackupSchedule
metadata:
  name: one-time-backup
spec:
  nfsVolumeRef:
    name: my-nfs-volume
```

### Schedule with Cross-Cluster Access

Create weekly backups accessible from all clusters in the same global account:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsBackupSchedule
metadata:
  name: shared-weekly-backup
spec:
  nfsVolumeRef:
    name: my-nfs-volume
  schedule: "0 2 * * 0"
  prefix: weekly
  startTime: "2026-06-01T00:00:00Z"
  endTime: "2027-06-01T00:00:00Z"
  deleteCascade: true
  accessibleFrom:
    - "all"
```

## Related Resources <!-- {docsify-ignore} -->

- [GcpNfsVolume](04-20-20-gcp-nfs-volume.md) - The source volume for backups
- [GcpNfsVolumeBackup](04-20-21-gcp-nfs-volume-backup.md) - The backup resources created by this schedule
- [GcpNfsVolumeRestore](04-20-23-gcp-nfs-volume-restore.md) - Restore a volume from a backup
- [GcpNfsVolumeBackupDiscovery](04-20-24-gcp-nfs-volume-backup-discovery.md) - Discover backups shared from other clusters
