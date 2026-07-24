# SapNfsVolumeSnapshotRestore Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `sapnfsvolumesnapshotrestore.cloud-resources.kyma-project.io` namespaced custom resource (CR) triggers a one-shot restore of an [SapNfsVolumeSnapshot](./04-20-51-sap-nfs-volume-snapshot.md) to either an existing or a new [SapNfsVolume](./04-20-50-sap-nfs-volume.md).

## Overview <!-- {docsify-ignore} -->

SapNfsVolumeSnapshotRestore supports two restore paths, selected by the `destination` field:

- **Existing volume (in-place revert)**: reverts an existing SapNfsVolume to the snapshot's state using the Manila "Revert share to snapshot" API. The volume's current data is overwritten. Due to an OpenStack Manila constraint, only the **most recent** snapshot of a given volume can be used for in-place revert.

- **New volume (create from snapshot)**: creates a new SapNfsVolume pre-populated with the snapshot's data using the Manila "create share from snapshot" capability. This path works with any snapshot, not just the most recent, and produces an independent volume.

Exactly one of `destination.existingVolume` or `destination.newVolume` must be specified.

SapNfsVolumeSnapshotRestore is a one-shot operation. Once a restore reaches `Done` or `Failed`, it cannot be retried. Create a new resource to attempt the operation again.

## Prerequisites <!-- {docsify-ignore} -->

Before creating an SapNfsVolumeSnapshotRestore, ensure that:

- The source SapNfsVolumeSnapshot exists and is in `Ready` state.
- For existing-volume restores: the destination SapNfsVolume exists, is in `Ready` state, and the snapshot was taken from that volume.
- For existing-volume restores: the snapshot is the **most recent** snapshot of the destination volume.
- For new-volume restores: the `capacityGb` in the new volume spec is **greater than or equal to** the source snapshot's size in GiB.

## How It Works <!-- {docsify-ignore} -->

| State | Description |
|-------|-------------|
| `InProgress` | The restore operation has been accepted and is running (revert or new-volume creation in progress). |
| `Done` | The restore completed successfully. |
| `Failed` | A permanent error occurred (for example, the snapshot is not the most recent, or it does not belong to the destination volume). Check `.status.conditions` for details. No further retries are performed. |
| `Error` | A transient error occurred (for example, source or destination not yet ready). The reconciler retries automatically. |

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter | Type | Required | Immutable | Description |
|-----------|------|----------|-----------|-------------|
| **sourceSnapshot** | object | Yes | Yes | Reference to the SapNfsVolumeSnapshot to restore from. The snapshot must be in `Ready` state. |
| **sourceSnapshot.name** | string | Yes | Yes | Name of the source SapNfsVolumeSnapshot. |
| **sourceSnapshot.namespace** | string | No | Yes | Namespace of the source SapNfsVolumeSnapshot. Defaults to the namespace of this resource if not provided. |
| **destination** | object | Yes | Yes | Specifies where to restore the snapshot data. Exactly one of `existingVolume` or `newVolume` must be set. |
| **destination.existingVolume** | object | No* | Yes | Reference to an existing SapNfsVolume to revert in-place. *Required if `newVolume` is not set. |
| **destination.existingVolume.name** | string | Yes* | Yes | Name of the destination SapNfsVolume. *Required if `existingVolume` is set. |
| **destination.existingVolume.namespace** | string | No | Yes | Namespace of the destination SapNfsVolume. Defaults to the namespace of this resource if not provided. |
| **destination.newVolume** | object | No* | Yes | Defines a new SapNfsVolume to create from the snapshot. *Required if `existingVolume` is not set. |
| **destination.newVolume.metadata.name** | string | Yes* | Yes | Name of the new SapNfsVolume to create. *Required if `newVolume` is set. |
| **destination.newVolume.metadata.namespace** | string | No | Yes | Namespace of the new SapNfsVolume. Defaults to the namespace of this resource if not provided. |
| **destination.newVolume.metadata.labels** | map\[string\]string | No | Yes | Labels for the new SapNfsVolume. |
| **destination.newVolume.metadata.annotations** | map\[string\]string | No | Yes | Annotations for the new SapNfsVolume. |
| **destination.newVolume.spec.capacityGb** | int | Yes* | Yes | Capacity of the new volume in GiB. Must be greater than zero and greater than or equal to the snapshot's source share size. *Required if `newVolume` is set. |
| **destination.newVolume.spec.ipRange** | object | No | Yes | IpRange reference for the new volume. If omitted, the default IpRange is used. |
| **destination.newVolume.spec.ipRange.name** | string | No | Yes | Name of the existing IpRange to use. |
| **destination.newVolume.spec.volume** | object | No | Yes | PersistentVolume options for the new volume (name, labels, annotations). |
| **destination.newVolume.spec.volumeClaim** | object | No | Yes | PersistentVolumeClaim options for the new volume (name, labels, annotations). |

**Status:**

| Parameter | Type | Description |
|-----------|------|-------------|
| **state** | string | Current state of the restore operation. Possible values: `InProgress`, `Done`, `Failed`, `Error`. |
| **revertInitiated** | boolean | Indicates that the Manila revert API call was successfully submitted. Used internally for idempotency to prevent duplicate revert calls on reconciler requeue. |
| **createdVolume** | objectRef | Reference to the SapNfsVolume created during a new-volume restore. Populated only for `newVolume` destination restores. |
| **conditions** | \[\]object | Represents the current state of the CR's conditions. |
| **conditions.lastTransitionTime** | string | Defines the date of the last condition status change. |
| **conditions.message** | string | Provides more details about the condition status change. |
| **conditions.reason** | string | Defines the reason for the condition status change. |
| **conditions.status** (required) | string | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.type** | string | Provides a short description of the condition. |

## Limitations <!-- {docsify-ignore} -->

- **In-place revert (most recent snapshot only)**: the Manila "Revert share to snapshot" API only succeeds if the referenced snapshot is the most recent one for that share. If newer snapshots exist, the restore will set state to `Failed`. To revert to an older state, delete the newer snapshots first, then create a new SapNfsVolumeSnapshotRestore.
- **In-place revert (snapshot must belong to the destination volume)**: Manila only allows reverting a share to its own snapshot. Cross-volume revert is not supported.
- **New-volume restore (same availability zone)**: the new volume is created in the same OpenStack project and availability zone as the parent share of the snapshot.
- **Immutable after completion**: once a restore reaches `Done` or `Failed`, create a new SapNfsVolumeSnapshotRestore resource to retry.

## Sample Custom Resources <!-- {docsify-ignore} -->

### In-Place Revert to the Most Recent Snapshot

Revert an existing volume to the state captured in a snapshot:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotRestore
metadata:
  name: revert-to-snapshot
  namespace: default
spec:
  sourceSnapshot:
    name: my-snapshot
  destination:
    existingVolume:
      name: my-sap-nfs-vol
```

### Create a New Volume from a Snapshot

Restore snapshot data into a new, independent volume:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotRestore
metadata:
  name: restore-to-new-volume
  namespace: default
spec:
  sourceSnapshot:
    name: my-snapshot
  destination:
    newVolume:
      metadata:
        name: my-restored-vol
      spec:
        capacityGb: 1000
```

### Cross-Namespace Restore to a New Volume

Restore a snapshot from one namespace into a new volume in another namespace:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotRestore
metadata:
  name: cross-namespace-restore
  namespace: staging
spec:
  sourceSnapshot:
    name: production-snapshot
    namespace: production
  destination:
    newVolume:
      metadata:
        name: staging-restored-vol
        namespace: staging
        labels:
          env: staging
      spec:
        capacityGb: 2000
```

## Related Resources <!-- {docsify-ignore} -->

- [SapNfsVolume](./04-20-50-sap-nfs-volume.md): The source and destination volume resource
- [SapNfsVolumeSnapshot](./04-20-51-sap-nfs-volume-snapshot.md): The source snapshot to restore from
- [SapNfsVolumeSnapshotSchedule](./04-20-52-sap-nfs-volume-snapshot-schedule.md): Automate snapshot creation on a schedule
