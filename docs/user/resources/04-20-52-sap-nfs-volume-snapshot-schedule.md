# SapNfsVolumeSnapshotSchedule Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `sapnfsvolumesnapshotschedule.cloud-resources.kyma-project.io` namespaced custom resource (CR) automates [SapNfsVolumeSnapshot](./04-20-51-sap-nfs-volume-snapshot.md) creation for a given [SapNfsVolume](./04-20-50-sap-nfs-volume.md). It supports both recurring schedules using cron expressions and one-time snapshots.

The CR performs the following actions:

- Creates SapNfsVolumeSnapshot resources automatically at the specified interval, or once immediately for one-time mode.
- Stamps each created snapshot with a `deleteAfterDays` value derived from `maxRetentionDays`, enabling time-based automatic expiry.
- Enforces count-based retention: evicts the oldest `Ready` snapshot before creating a new one when `maxReadySnapshots` would be exceeded, and garbage-collects `Failed` snapshots beyond `maxFailedSnapshots`.
- Enables you to temporarily suspend or permanently stop snapshot creation.

Created snapshots are named using the pattern `{prefix}-{index}`, where `prefix` defaults to the schedule name and `index` is a monotonically incrementing counter.

> [!NOTE]
> Snapshot quota is shared across the entire OpenStack project. The default per-project limit is 50 snapshots across all volumes and schedules. The `maxReadySnapshots` default of `50` aligns with this limit but is applied per schedule — multiple schedules targeting different volumes in the same project compete for the same quota pool. Consider lowering `maxReadySnapshots` when running multiple schedules in the same project.

## How It Works <!-- {docsify-ignore} -->

An SapNfsVolumeSnapshotSchedule progresses through the following states:

| State | Description |
|-------|-------------|
| `Active` | The schedule is running normally and will create snapshots at the configured interval. |
| `Suspended` | The schedule has been paused via `spec.suspend: true`. No new snapshots are created, but existing ones are preserved. |
| `Done` | The schedule has completed: either a one-time snapshot finished, or `endTime` has passed. |
| `Error` | A transient error occurred. The reconciler retries automatically. |
| `Deleting` | The schedule is being deleted. Snapshot cascade deletion is performed if `deleteCascade` is `true`. |

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter | Type | Required | Immutable | Description |
|-----------|------|----------|-----------|-------------|
| **template** | object | Yes | No | Template for the SapNfsVolumeSnapshot resources created by this schedule. |
| **template.spec.sourceVolume** | object | Yes | No | Reference to the SapNfsVolume to snapshot. The volume must be in the `Ready` state when each snapshot is created. |
| **template.spec.sourceVolume.name** | string | Yes | No | Name of the source SapNfsVolume. |
| **template.spec.sourceVolume.namespace** | string | No | No | Namespace of the source `SapNfsVolume`. Defaults to the schedule's namespace if not provided. |
| **template.labels** | map\[string\]string | No | No | Labels applied to each created `SapNfsVolumeSnapshot`. Merged with schedule-managed labels. |
| **template.annotations** | map\[string\]string | No | No | Annotations applied to each created `SapNfsVolumeSnapshot`. |
| **schedule** | string | No | No | Cron expression for the recurring schedule. When empty or not specified, a single snapshot is created immediately (one-time mode) and the schedule transitions to `Done`. See [Cron syntax](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#schedule-syntax). |
| **prefix** | string | No | No | Prefix for the names of created `SapNfsVolumeSnapshot` resources. Defaults to the name of this schedule. |
| **startTime** | string | No | No | Time before which no snapshots are created, in RFC 3339 format (e.g., `2026-06-01T00:00:00Z`). When not set, the schedule becomes effective immediately. |
| **endTime** | string | No | No | Time after which no new snapshots are created, in RFC 3339 format. When reached, the schedule transitions to `Done`. When not set, the schedule runs indefinitely. |
| **maxRetentionDays** | int | No | No | Maximum number of days to retain each created snapshot. Stamped as `deleteAfterDays` on each snapshot at creation time. Defaults to `375`. Minimum: `1`. |
| **maxReadySnapshots** | int | No | No | Maximum number of `Ready` snapshots to retain. The oldest `Ready` snapshot is deleted before creating a new one when this limit would be exceeded. Defaults to `50`. Minimum: `1`. |
| **maxFailedSnapshots** | int | No | No | Maximum number of `Failed` snapshots to retain. Oldest snapshots beyond this count are garbage-collected. Defaults to `5`. Minimum: `1`. |
| **suspend** | boolean | No | No | When `true`, stops the schedule from creating new snapshots and sets state to `Suspended`. Existing snapshots are not affected. Defaults to `false`. |
| **deleteCascade** | boolean | No | No | When `true`, all `SapNfsVolumeSnapshot` resources created by this schedule are deleted when the schedule itself is deleted. When `false`, existing snapshots are preserved. Defaults to `false`. |

**Status:**

| Parameter | Type | Description |
|-----------|------|-------------|
| **state** | string | Current state of the schedule. Possible values: `Active`, `Suspended`, `Done`, `Error`, `Deleting`. |
| **conditions** | \[\]object | Represents the current state of the CR's conditions. |
| **conditions.lastTransitionTime** | string | Defines the date of the last condition status change. |
| **conditions.message** | string | Provides more details about the condition status change. |
| **conditions.reason** | string | Defines the reason for the condition status change. |
| **conditions.status** (required) | string | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.type** | string | Provides a short description of the condition. |
| **nextRunTimes** | \[\]string | Preview of up to 3 upcoming times when the next snapshots will be created. |
| **nextDeleteTimes** | map\[string\]string | Map of snapshot names to their expected time-based deletion time, calculated from `maxRetentionDays`. |
| **lastCreateRun** | string | Time when the last snapshot creation was triggered. |
| **lastCreatedBackup** | objectRef | Object reference of the last created snapshot. |
| **lastDeleteRun** | string | Time when the last retention-based deletion was performed. |
| **lastDeletedBackups** | \[\]objectRef | Object references of the snapshots deleted during the last retention run. |
| **schedule** | string | The cron expression of the currently active schedule. |
| **snapshotIndex** | int | Monotonically incrementing counter used for snapshot naming. |
| **backupCount** | int | Number of snapshots currently present in the system for this schedule. |

## Sample Custom Resources <!-- {docsify-ignore} -->

### Recurring Daily Snapshot

Create a daily snapshot at midnight, retaining the last 7 snapshots for up to 30 days:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotSchedule
metadata:
  name: daily-snapshot
  namespace: default
spec:
  template:
    spec:
      sourceVolume:
        name: my-sap-nfs-vol
  schedule: "0 0 * * *"
  prefix: daily
  maxRetentionDays: 30
  maxReadySnapshots: 7
```

### One-Time Snapshot

Create a single snapshot immediately, for example before a maintenance window:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotSchedule
metadata:
  name: pre-maintenance-snapshot
  namespace: default
spec:
  template:
    spec:
      sourceVolume:
        name: my-sap-nfs-vol
```

### Time-Bounded Weekly Schedule with Cascade Delete

Create weekly snapshots within a defined time window, and delete all managed snapshots when the schedule is removed:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotSchedule
metadata:
  name: weekly-snapshot
  namespace: default
spec:
  template:
    spec:
      sourceVolume:
        name: my-sap-nfs-vol
  schedule: "0 2 * * 0"
  prefix: weekly
  startTime: "2026-06-01T00:00:00Z"
  endTime: "2027-06-01T00:00:00Z"
  maxRetentionDays: 90
  maxReadySnapshots: 12
  deleteCascade: true
```

## Related Resources <!-- {docsify-ignore} -->

- [SapNfsVolume](./04-20-50-sap-nfs-volume.md) — The source volume to snapshot
- [SapNfsVolumeSnapshot](./04-20-51-sap-nfs-volume-snapshot.md) — The snapshot resources created by this schedule
- [SapNfsVolumeSnapshotRestore](./04-20-53-sap-nfs-volume-snapshot-restore.md) — Restore a volume from a snapshot
