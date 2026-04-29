# Requirements Document

## Introduction

Customers have requested backup and restore for SAP Converge Cloud NFS volumes, as already available for GCP and AWS. Since Manila's Backup API is not viable on our deployment (see below), this feature uses the stable Snapshots API as the closest alternative.

This feature adds snapshot support for SAP Converge Cloud (OpenStack) NFS volumes within the Cloud Manager ecosystem. Snapshots allow users to create point-in-time, read-only copies of their SAP NFS volumes using the OpenStack Manila Share Snapshots API (available since Manila v2 base, non-experimental).

The OpenStack version targeted is **Antelope** (antelope-20260219211138). The Manila Share Snapshots API is a **core, non-experimental API** that has been part of Manila since the initial v2.0 microversioned release. Snapshots have never been tagged as experimental and do not require the `X-OpenStack-Manila-API-Experimental` header ([Experimental APIs](https://docs.openstack.org/manila/2023.1/contributor/experimental_apis.html)). The v2.27 "Revert share to snapshot" addition is also non-experimental ([Microversion History](https://docs.openstack.org/manila/2023.1/contributor/api_microversion_history.html)).

### Key Characteristics of OpenStack Manila Snapshots

- **Snapshot** = a point-in-time, read-only copy of a share's data
- **Status lifecycle**: `creating` → `available` (success) or `error` (failure); `deleting` → deleted or `error_deleting`
- **Snapshots are tied to a share** and cannot exist independently
- **A share can be created from a snapshot** (restore use-case)
- **A share can be reverted to its most recent snapshot** (API v2.27+)

### Snapshot Limitations

The Snapshots API is stable and available on Antelope, but it comes with significant constraints that **differ from GCP Filestore backups** (which are independent, relocatable resources that persist after source instance deletion):

1. **Snapshots are bound to their parent share and cannot exist without it.** They live on the same backend storage, referenced by `share_id`. ([Share Snapshots API](https://docs.openstack.org/api-ref/shared-file-system/#share-snapshots))

2. **A share cannot be deleted while it has snapshots.** Manila rejects the request; all snapshots must be deleted first. ([Admin Guide](https://docs.openstack.org/manila/2023.1/admin/shared-file-systems-snapshots.html))

3. **In-place revert only works with the most recent snapshot.** The "Revert share to snapshot" API (v2.27+) fails if the referenced snapshot is not the latest. Deleting newer snapshots makes an older snapshot eligible for revert — verified against Antelope. ([Revert API](https://docs.openstack.org/api-ref/shared-file-system/#revert-share-to-snapshot-since-api-v2-27))

4. **A new share can be created from any snapshot** (not restricted to the latest), producing an independent share with the snapshot's data. The new share must be in the same availability zone as the parent share of the snapshot. ([Create Share](https://docs.openstack.org/api-ref/shared-file-system/#create-share))

5. **Snapshots are not relocatable.** They exist within the same OpenStack project and availability zone as the parent share.

6. **Snapshots are subject to per-project quotas.** Manila enforces a project-level quota on the total number of snapshots (`maxTotalShareSnapshots`, default: **50**) and total snapshot storage (`maxTotalSnapshotGigabytes`, default: **1000 GiB**). These quotas are shared across all shares in the project. Administrators can adjust quotas via `manila quota-update`. ([Quotas and Limits](https://docs.openstack.org/manila/2023.1/admin/shared-file-systems-quotas.html))

### Why Snapshots Instead of Backups

OpenStack Manila does offer a Share Backup API ([API ref](https://docs.openstack.org/api-ref/shared-file-system/#share-backups-since-api-v2-80), [Admin Guide](https://docs.openstack.org/manila/latest/admin/shared-file-systems-share-backup-management.html) — note: this page does not exist in the 2023.1/Antelope docs since backups were added later) that provides true off-site backups — data copied to separate backup storage, independent of the source share lifecycle. This would be the ideal solution, matching the semantics of GCP Filestore backups. However, it is **not viable** for this feature for two reasons:

1. **The Backup API is not available on our Antelope deployment.** The Backup API was introduced at microversion **v2.80**, but our target Antelope release has a maximum microversion of **v2.78** ([Microversion History](https://docs.openstack.org/manila/2023.1/contributor/api_microversion_history.html) — "2.78 (Maximum in 2023.1/Antelope)"). The backup endpoints (`/v2/share-backups`) simply do not exist on our infrastructure. Even on newer Manila releases, the Backup API remains **experimental** — every request requires the `X-OpenStack-Manila-API-Experimental: True` header and the API may change or be removed ([Experimental APIs](https://docs.openstack.org/manila/2023.1/contributor/experimental_apis.html)). It also requires the cloud operator to configure a backup driver and a dedicated `manila-data` service.

We therefore use the stable **Share Snapshots API** as a temporary solution to give customers point-in-time recovery capabilities now. Should SAP Converge Cloud upgrade to a Manila version ≥ Bobcat (v2.82+) and enable the backup driver, a future `SapNfsVolumeBackup` CRD could wrap the Backup API for true off-site, lifecycle-independent backups.

---

## Open Questions

### OQ-1: CRD naming — "Backup" vs. "Snapshot"

Options:
- ~~**(A) Backup naming**: Use `SapNfsVolumeBackup`, `SapNfsVolumeBackupSchedule`, `SapNfsVolumeRestore` — consistent with `GcpNfsVolumeBackup` naming convention.~~
- **(B) Snapshot naming**: Use `SapNfsVolumeSnapshot`, `SapNfsVolumeSnapshotSchedule`, `SapNfsVolumeSnapshotRestore` — reflects the actual underlying OpenStack resource.

**Decision: `SapNfsVolumeSnapshot`** (option B — Snapshot naming)

**Rationale:** Manila has a separate Backup API (since v2.80) that may become available if we upgrade. Using `SapNfsVolumeSnapshot` reserves the "Backup" name for a future CRD wrapping that API and avoids a naming conflict.

_All CRD names in this document use the decided `Snapshot` naming: `SapNfsVolumeSnapshot`, `SapNfsVolumeSnapshotSchedule`, `SapNfsVolumeSnapshotRestore`._

### OQ-2: Handling parent SapNfsVolume deletion while snapshots exist

Since OpenStack Manila does not allow deleting a share that has dependent snapshots (see Limitation 2), we need to decide what happens when a user deletes a `SapNfsVolume` while `SapNfsVolumeSnapshot` resources referencing it still exist.

Options:
- **(A) Block deletion and set warning/error**: The `SapNfsVolume` reconciler checks for dependent `SapNfsVolumeSnapshot` resources, sets a warning or error condition, and refuses to proceed with deletion until all snapshots are removed by the user.
- **(B) Cascade-delete snapshots first**: The `SapNfsVolume` reconciler automatically deletes all dependent `SapNfsVolumeSnapshot` resources first, then proceeds with volume deletion.

**Decision: TBD**

_Once decided, a corresponding requirement must be added to implement the chosen behavior in the `SapNfsVolume` reconciler's deletion path._

### OQ-3: Restore mechanism — revert vs. create-from-snapshot

Options:
- ~~**(A) In-place revert only**: Only support reverting an existing volume to its most recent snapshot.~~
- ~~**(B) Create-from-snapshot only**: Only support creating a new volume from any snapshot.~~
- **(C) Both mechanisms**: Support both in-place revert and create-from-snapshot via a discriminated destination.

**Decision: (C) Both mechanisms** via a discriminated destination.

**Rationale:** Customers requested restore to behave like `GcpNfsVolumeBackup`, which supports both in-place revert and creating a new volume. Offering both paths gives users maximum flexibility — in-place revert for quick recovery of the latest state, and create-from-snapshot for restoring older snapshots or creating populated copies.

**Provider-specific constraints:**
- In-place revert is subject to Limitation 3 — only the most recent snapshot can be used. The controller validates this and returns `Error` if the snapshot is not the latest.
- New-resource restore works with any snapshot, producing an independent volume pre-populated with the snapshot's data.

---

## Requirements

### Requirement 1: SAP NFS Volume Snapshot Lifecycle

**User Story:** As a Kyma cluster user, I want to create point-in-time snapshots of my SAP NFS volumes, so that I have copies of my data for recovery purposes.

#### Acceptance Criteria — Creation

1. WHEN a user creates a `SapNfsVolumeSnapshot` resource referencing a `SapNfsVolume` THEN the system SHALL create a snapshot of that volume's underlying Manila share.
2. WHEN the snapshot is created THEN the system SHALL validate that the referenced `SapNfsVolume` exists and is in `Ready` state. IF the source volume does not exist or is not `Ready` THEN the system SHALL set the state to `Error` with a descriptive message.
3. WHEN the snapshot is created THEN the source volume reference SHALL be immutable after creation.
4. WHEN the OpenStack snapshot is successfully created THEN the system SHALL set the state to `Ready` and report the snapshot size in the status.
5. WHEN the OpenStack snapshot creation fails THEN the system SHALL set the state to `Error` and record the failure.
6. WHEN a snapshot is being created THEN the system SHALL set the state to `Creating`.

#### Acceptance Criteria — Automatic Expiry

7. WHEN a snapshot has a configured time-to-live (delete-after-days) and its age exceeds that value THEN the system SHALL automatically delete the snapshot, following the same deletion flow as Requirement 2.

#### Acceptance Criteria — Status Reporting

8. The snapshot SHALL report its lifecycle state with values: `Creating`, `Ready`, `Error`, `Deleting`, `Failed`.
9. WHEN a terminal error occurs that will not be retried THEN the system SHALL set the state to `Failed`.
10. WHEN a transient error occurs THEN the system SHALL set the state to `Error` and retry.

#### Acceptance Criteria — Feature Gating

11. WHEN the `nfsBackup` feature flag is disabled THEN the snapshot reconciler SHALL be inactive.

**Note:** Manila snapshots are not relocatable or shareable across projects (see Limitation 5), so the snapshot CRD omits location and access-control fields.

### Requirement 2: Delete an SAP NFS Volume Snapshot

**User Story:** As a Kyma cluster user, I want to delete a snapshot of my SAP NFS volume, so that I can free up storage when the snapshot is no longer needed.

#### Acceptance Criteria

1. WHEN a user deletes a `SapNfsVolumeSnapshot` resource THEN the system SHALL delete the corresponding OpenStack Manila snapshot before allowing the Kubernetes resource to be removed.
2. WHEN the OpenStack snapshot is not found during deletion (already deleted externally) THEN the system SHALL proceed with resource cleanup without error.
3. WHEN the OpenStack snapshot deletion fails THEN the system SHALL set the state to `Error` and retry.
4. WHEN a snapshot is being deleted THEN the system SHALL set the state to `Deleting`.

### Requirement 3: Restore SAP NFS Volume from Snapshot

**User Story:** As a Kyma cluster user, I want to restore my SAP NFS volume from a previously taken snapshot, so that I can recover data to a known good state.

The `SapNfsVolumeSnapshotRestore` resource supports two restore paths via a discriminated destination: restoring into an **existing** volume (in-place revert) or creating a **new** volume from the snapshot.

#### Acceptance Criteria — In-place restore (existing destination)

1. WHEN a user creates a `SapNfsVolumeSnapshotRestore` targeting an existing `SapNfsVolume` as destination THEN the system SHALL revert the volume's underlying Manila share to the referenced snapshot using the Manila "Revert share to snapshot" API (v2.27+).
2. WHEN the restore is created THEN the system SHALL validate that:
   - The referenced `SapNfsVolumeSnapshot` exists and is in `Ready` state.
   - The destination `SapNfsVolume` exists and is in `Ready` state.
   - The snapshot belongs to the destination volume (Manila only allows reverting a share to its own snapshot).
   - The snapshot is the **most recent** snapshot of the volume (Manila constraint — see Limitation 3). IF the snapshot is not the latest THEN the system SHALL set the state to `Error` with a descriptive message.
3. WHEN the revert operation succeeds THEN the system SHALL set the restore state to `Done`.
4. WHEN the revert operation fails THEN the system SHALL set the restore state to `Failed` and record the error.

#### Acceptance Criteria — New-resource restore (new destination)

5. WHEN a user creates a `SapNfsVolumeSnapshotRestore` targeting a new volume as destination THEN the system SHALL create a new `SapNfsVolume` pre-populated with the snapshot's data. The user provides a volume template (name, capacity, etc.) as part of the restore resource.
6. WHEN the new volume is created THEN it SHALL use the Manila "create share from snapshot" capability to populate the new share with the snapshot's data.
7. WHEN the new `SapNfsVolume` reaches `Ready` state THEN the system SHALL set the restore state to `Done` and record a reference to the created volume in the restore's status.
8. WHEN provisioning or restore fails THEN the system SHALL set the restore state to `Failed` and record the error.

#### Acceptance Criteria — Common

9. WHEN the `SapNfsVolumeSnapshotRestore` resource is created THEN the source snapshot and destination choice SHALL be immutable after creation.
10. WHEN a restore is in progress THEN the system SHALL set the state to `InProgress`.
11. WHEN validation fails (references not found, wrong state, snapshot not latest for in-place) THEN the system SHALL set the state to `Error` with a descriptive message.
12. WHEN the `nfsBackup` feature flag is disabled THEN the restore reconciler SHALL be inactive.

**References:**
- [Manila Revert Share to Snapshot (API v2.27+)](https://docs.openstack.org/api-ref/shared-file-system/#revert-share-to-snapshot-since-api-v2-27)
- [Manila Create Share with snapshot_id](https://docs.openstack.org/api-ref/shared-file-system/#create-share)

### Requirement 4: Scheduled Snapshot Creation

**User Story:** As a Kyma cluster user, I want to schedule automatic periodic snapshots of my SAP NFS volume, so that I have regular point-in-time copies without manual intervention.

**Note:** OpenStack Manila has no native snapshot scheduling capability. All scheduling logic is implemented by Cloud Manager.

#### Acceptance Criteria — Scheduling

1. WHEN a user creates a `SapNfsVolumeSnapshotSchedule` with a cron expression THEN the system SHALL automatically create `SapNfsVolumeSnapshot` resources at the times defined by the cron expression (recurring schedule).
2. WHEN a user creates a `SapNfsVolumeSnapshotSchedule` without a cron expression THEN the system SHALL create a single snapshot immediately (one-time schedule) and transition to `Done` state after completion.
3. WHEN the schedule creates a snapshot THEN the snapshot name SHALL be deterministic, incorporating the schedule name and a monotonically incrementing index.
4. WHEN the schedule is suspended THEN the system SHALL stop creating new snapshots and transition to `Suspended` state, but SHALL NOT delete existing snapshots.
5. WHEN the schedule has a configured start time THEN the system SHALL not create snapshots before that time.
6. WHEN the schedule has a configured end time THEN the system SHALL stop creating snapshots after that time and transition to `Done` state.

#### Acceptance Criteria — Retention

7. WHEN the schedule creates a snapshot THEN it SHALL apply a time-based retention (maximum retention days) on each created snapshot. The snapshot controller enforces the TTL via the snapshot's lifecycle settings.
8. WHEN the schedule is active THEN the system SHALL enforce count-based retention:
   - Before creating a new snapshot: evict the oldest `Ready` snapshot if the total count would exceed the configured maximum ready snapshots (default: 50). Note: the default aligns with the Manila per-project snapshot quota of 50 (see Limitation 6). **Caution:** this quota is project-wide across all shares, so multiple schedules targeting different volumes in the same project will compete for the same quota pool.
   - Garbage-collect `Failed` snapshots beyond the configured maximum failed snapshots (default: 5).

Time-based and count-based retention operate independently. If time-based retention expires snapshots faster than the count limit is reached, the count-based limit becomes a no-op — this is expected, not an error.

#### Acceptance Criteria — Deletion

9. WHEN cascade deletion is enabled and the schedule is deleted THEN the system SHALL delete all `SapNfsVolumeSnapshot` resources created by the schedule. WHEN cascade deletion is disabled (default) THEN existing snapshots SHALL be preserved.

#### Acceptance Criteria — Feature Gating

10. WHEN the `nfsBackup` feature flag is disabled THEN the schedule reconciler SHALL be inactive.
