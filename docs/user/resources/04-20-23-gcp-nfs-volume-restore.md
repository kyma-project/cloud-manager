# GcpNfsVolumeRestore Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `gcpnfsvolumerestore.cloud-resources.kyma-project.io` namespaced custom resource (CR) triggers a one-shot restore of a GCP Filestore backup onto an existing BASIC-tier Filestore instance.

## Overview <!-- {docsify-ignore} -->

A GcpNfsVolumeRestore restores the contents of a GCP Filestore backup onto a target Filestore instance. The restore can target either the same instance the backup was created from, or a different existing instance.

You can specify the source backup in two ways:
- **By reference** — point to an existing [GcpNfsVolumeBackup](./04-20-21-gcp-nfs-volume-backup.md) resource using `source.backup`.
- **By URL** — provide a direct backup path using `source.backupUrl` in the format `{location_id}/{backup_id}`.

Exactly one of `source.backup` or `source.backupUrl` must be set.

> [!NOTE]
> This resource performs an **in-place restore** on an existing Filestore instance and is only supported for **BASIC_HDD** and **BASIC_SSD** tiers.
> To restore a ZONAL or REGIONAL backup, use the `sourceBackup` field on a new [GcpNfsVolume](./04-20-20-gcp-nfs-volume.md) instead.

To learn more, read [Restoring from backups](https://cloud.google.com/filestore/docs/backup-restore).

## Prerequisites <!-- {docsify-ignore} -->

Before creating a GcpNfsVolumeRestore, ensure that:

- A destination [GcpNfsVolume](./04-20-20-gcp-nfs-volume.md) exists and is in the `Ready` state.
- A source [GcpNfsVolumeBackup](./04-20-21-gcp-nfs-volume-backup.md) exists and is in the `Ready` state — or you have a valid `backupUrl`.
- The destination Filestore capacity is **equal to or greater than** the capacity of the source backup's Filestore.
- The destination Filestore uses a **BASIC** tier (`BASIC_HDD` or `BASIC_SSD`).

## Limitations <!-- {docsify-ignore} -->

- **BASIC tiers only** — ZONAL and REGIONAL tiers do not support in-place restore. See [Supported tiers](https://cloud.google.com/filestore/docs/backup-restore).
- **Capacity constraint** — the destination instance capacity must be >= the source backup's instance capacity. See [Backup/Restore limitations](https://cloud.google.com/filestore/docs/backups#limitations-storage).
- **No concurrent restores** — only one restore can run against a given GcpNfsVolume at a time. Additional restore requests targeting the same volume will wait until the active restore completes.
- **Immutable after completion** — once a restore reaches `Done` or `Failed`, it cannot be retried. Create a new GcpNfsVolumeRestore resource instead.

## How It Works <!-- {docsify-ignore} -->

GcpNfsVolumeRestore is a one-shot operation. The resource progresses through the following states and cannot be reused:

| State | Description |
|-------|-------------|
| `Processing` | The restore request has been accepted and is being validated. |
| `InProgress` | The GCP Filestore restore operation is running. The source backup and destination instance must remain available. |
| `Done` | The restore completed successfully. The destination Filestore instance now reflects the contents of the source backup. |
| `Failed` | The restore failed. Check `.status.conditions` for details. |
| `Error` | An unexpected error occurred during reconciliation. |

> [!NOTE]
> Both the source GCP Filestore backup and the destination GCP Filestore instance must remain available while the restore operation is `InProgress`.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                        | Type   | Required | Immutable | Description                                                                                                                                                                                                |
|----------------------------------|--------|----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **source**                       | object | Yes      | Yes       | Specifies the source backup for the restore operation. Exactly one of `backup` or `backupUrl` must be set.                                                                                                 |
| **source.backup**                | object | No*      | Yes       | Reference to an existing GcpNfsVolumeBackup resource. *Required if `backupUrl` is not set.                                                                                                                 |
| **source.backup.name**           | string | Yes*     | Yes       | Name of the source GcpNfsVolumeBackup. *Required if `source.backup` is specified.                                                                                                                          |
| **source.backup.namespace**      | string | Yes*     | Yes       | Namespace of the source GcpNfsVolumeBackup. Defaults to the namespace of the GcpNfsVolumeRestore resource if not provided. *Required if `source.backup` is specified.                                      |
| **source.backupUrl**             | string | No*      | Yes       | Direct backup identifier in the format `{location_id}/{backup_id}`. *Required if `source.backup` is not set. Must match pattern `^(.+)/(.+)$`.                                                            |
| **destination**                  | object | Yes      | Yes       | Specifies the destination of the restore operation.                                                                                                                                                        |
| **destination.volume**           | object | Yes      | Yes       | Reference to an existing GcpNfsVolume that the backup is restored onto.                                                                                                                                    |
| **destination.volume.name**      | string | Yes      | Yes       | Name of the destination GcpNfsVolume.                                                                                                                                                                      |
| **destination.volume.namespace** | string | No       | Yes       | Namespace of the destination GcpNfsVolume. Defaults to the namespace of the GcpNfsVolumeRestore resource if not provided.                                                                                  |

**Status:**

| Parameter                         | Type       | Description                                                                                                                             |
|-----------------------------------|------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| **state**                         | string     | Current state of the GcpNfsVolumeRestore. Possible values: `Processing`, `InProgress`, `Done`, `Failed`, or `Error`.                    |
| **opIdentifier**                  | string     | Operation identifier used to track the underlying GCP Filestore restore operation.                                                      |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                                    |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                                   |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                                |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                     |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                              |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                          |

## Sample Custom Resources <!-- {docsify-ignore} -->

### Restore from a GcpNfsVolumeBackup Reference

Restore a backup onto the same volume it was created from:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeRestore
metadata:
  name: my-restore
  namespace: my-namespace
spec:
  source:
    backup:
      name: my-backup
      namespace: my-namespace
  destination:
    volume:
      name: my-vol
      namespace: my-namespace
```

### Restore from a Backup URL

Restore using a direct backup URL (useful for cross-project or discovered backups):

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeRestore
metadata:
  name: my-restore-from-url
  namespace: my-namespace
spec:
  source:
    backupUrl: "us-central1/my-backup-id"
  destination:
    volume:
      name: my-vol
      namespace: my-namespace
```

### Cross-Instance Restore

Restore a backup onto a different Filestore instance than the one it was created from:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeRestore
metadata:
  name: cross-instance-restore
  namespace: production
spec:
  source:
    backup:
      name: staging-backup
      namespace: staging
  destination:
    volume:
      name: production-vol
      namespace: production
```

## Related Resources <!-- {docsify-ignore} -->

- [GcpNfsVolume](./04-20-20-gcp-nfs-volume.md) — The destination volume for restores
- [GcpNfsVolumeBackup](./04-20-21-gcp-nfs-volume-backup.md) — The source backup resource
- [GcpNfsVolumeBackupDiscovery](./04-20-24-gcp-nfs-volume-backup-discovery.md) — Discover backups shared from other clusters
- [GcpNfsBackupSchedule](./04-20-22-gcp-nfs-backup-schedule.md) — Automate backup creation on a schedule
