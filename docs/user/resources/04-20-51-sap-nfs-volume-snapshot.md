# SapNfsVolumeSnapshot Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `sapnfsvolumesnapshot.cloud-resources.kyma-project.io` namespaced custom resource (CR) describes a point-in-time, read-only snapshot of an [SapNfsVolume](./04-20-50-sap-nfs-volume.md). Snapshots are created using the OpenStack Manila Share Snapshots API and let you capture the state of an NFS share at a specific moment for data recovery.

Unlike backups, snapshots are bound to their parent share. Snapshot data is not copied to separate storage; it resides on the same backend as the source volume and cannot exist independently. You cannot delete an SapNfsVolume while it has associated snapshots. You must remove all snapshots first.

> [!NOTE]
> Manila snapshots are subject to per-project quotas. The default limits are 50 snapshots and 1000 GiB of total snapshot storage per OpenStack project. All SapNfsVolumeSnapshot resources across all volumes in the same project share this quota. Contact your OpenStack administrator to adjust these limits if needed.

## How It Works <!-- {docsify-ignore} -->

An SapNfsVolumeSnapshot progresses through the following states:

| State | Description |
|-------|-------------|
| `Creating` | The Manila snapshot creation request has been submitted and is in progress. |
| `Ready` | The snapshot is available and can be used for restore operations. |
| `Error` | A transient error occurred. The reconciler retries automatically. |
| `Deleting` | The Manila snapshot is being deleted. |
| `Failed` | A terminal error occurred that will not be retried. Check `.status.conditions` for details. |

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter | Type | Required | Immutable | Description |
|-----------|------|----------|-----------|-------------|
| **sourceVolume** | object | Yes | Yes | Reference to the SapNfsVolume to snapshot. The volume must exist and be in `Ready` state at the time of creation. |
| **sourceVolume.name** | string | Yes | Yes | Name of the source SapNfsVolume. |
| **sourceVolume.namespace** | string | No | Yes | Namespace of the source SapNfsVolume. Defaults to the namespace of this resource if not provided. |
| **deleteAfterDays** | int | No | No | Number of days after which the snapshot is automatically deleted. `0` disables automatic deletion. Defaults to `0`. |

**Status:**

| Parameter | Type | Description |
|-----------|------|-------------|
| **state** | string | Current lifecycle state. Possible values: `Creating`, `Ready`, `Error`, `Deleting`, `Failed`. |
| **id** | string | Internal identifier of the snapshot resource. |
| **openstackId** | string | UUID of the Manila snapshot, as assigned by OpenStack. |
| **sizeGb** | int | Snapshot size in GiB, as reported by Manila. Populated when the snapshot reaches `Ready` state. |
| **shareId** | string | UUID of the Manila share the snapshot belongs to. |
| **conditions** | \[\]object | Represents the current state of the CR's conditions. |
| **conditions.lastTransitionTime** | string | Defines the date of the last condition status change. |
| **conditions.message** | string | Provides more details about the condition status change. |
| **conditions.reason** | string | Defines the reason for the condition status change. |
| **conditions.status** (required) | string | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.type** | string | Provides a short description of the condition. |

## Limitations <!-- {docsify-ignore} -->

- **Snapshots are bound to their parent share**: a snapshot cannot be moved to a different volume or OpenStack project. It exists in the same project and availability zone as the source SapNfsVolume.
- **An SapNfsVolume cannot be deleted while it has active snapshots**: all snapshots referencing a volume must be deleted before the volume can be removed.
- **In-place revert supports the most recent snapshot only**: when restoring to an existing volume using SapNfsVolumeSnapshotRestore, the snapshot must be the most recent one taken from that volume. This is an OpenStack Manila constraint.
- **Project-wide quota**: the 50-snapshot default limit is shared across all volumes and schedules in the same OpenStack project.

## Sample Custom Resources <!-- {docsify-ignore} -->

### Basic Snapshot

Create a one-off snapshot of an NFS volume:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshot
metadata:
  name: my-snapshot
  namespace: default
spec:
  sourceVolume:
    name: my-sap-nfs-vol
```

### Snapshot with Automatic Expiry

Create a snapshot that is automatically deleted after 30 days:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshot
metadata:
  name: my-expiring-snapshot
  namespace: default
spec:
  sourceVolume:
    name: my-sap-nfs-vol
  deleteAfterDays: 30
```

## Related Resources <!-- {docsify-ignore} -->

- [SapNfsVolume](./04-20-50-sap-nfs-volume.md): The source volume to snapshot
- [SapNfsVolumeSnapshotSchedule](./04-20-52-sap-nfs-volume-snapshot-schedule.md): Automate snapshot creation on a schedule
- [SapNfsVolumeSnapshotRestore](./04-20-53-sap-nfs-volume-snapshot-restore.md): Restore a volume from a snapshot
