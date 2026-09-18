
## Backup & Restore

Cloud Manager provides NFS volume backup, restore, and scheduling across providers. The three sub-families (one-time backup, restore, scheduled backup) are shown in separate tables. Note that Azure has no `NfsVolume` resource, so its backup resources operate on raw Kubernetes PVCs, and SAP/OpenStack uses "snapshot" terminology throughout.

### One-Time Backup

<table>
<tr>
<th>AwsNfsVolumeBackup</th>
<th>GcpNfsVolumeBackup</th>
<th>AzureRwxVolumeBackup</th>
<th>SapNfsVolumeSnapshot (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
spec:
  # Required. References a Cloud Manager AwsNfsVolume.
  source:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
  # Optional. AWS Backup Vault lifecycle policy.
  # Both fields are immutable after creation.
  lifecycle:
    # Days until backup is deleted.
    deleteAfterDays: 365
    # Days until backup moves to cold storage.
    # deleteAfterDays must be >= this + 90.
    moveToColdStorageAfterDays: 30
  # Optional. Immutable. AWS Region for backup storage.
  # Defaults to the source volume's region.
  location: "us-east-1"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
spec:
  # Required. Immutable. References a Cloud Manager GcpNfsVolume.
  source:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
  # Optional. Immutable. GCP region for backup storage.
  # Defaults to the source volume's region.
  location: "us-west1"
  # Optional. Cross-cluster access control.
  # List of shootNames or subaccountIds, or ["all"] for global access.
  # Max 10 entries. "all" cannot be combined with other values.
  accessibleFrom:
    - "all"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxVolumeBackup
metadata:
  name: my-backup
  namespace: kyma-system
  # Azure has no NfsVolume resource — backs a raw PVC directly.
spec:
  # Required. Immutable. References a Kubernetes PVC.
  source:
    pvc:
      name: "my-pvc"
      namespace: "kyma-system"  # optional
  # Optional. Azure region for backup storage.
  location: "westeurope"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshot
metadata:
  name: my-snapshot
  namespace: kyma-system
  # OpenStack/CCEE uses "snapshot" terminology
  # instead of "backup".
spec:
  # Required. Immutable. References a SapNfsVolume.
  sourceVolume:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. Days after which the snapshot is auto-deleted.
  # 0 = no automatic deletion.
  deleteAfterDays: 90
```

</td>
</tr>
</table>

### Restore

<table>
<tr>
<th>AwsNfsVolumeRestore</th>
<th>GcpNfsVolumeRestore</th>
<th>AzureRwxVolumeRestore</th>
<th>SapNfsVolumeSnapshotRestore (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. AWS restores to a subdirectory
  # (aws-backup-restore_*/...), not in-place.
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  # Use either backup (object ref) or backupUrl (raw GCP URI).
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"
    # Alternative to backup (mutually exclusive):
    # backupUrl: "us-west1/my-backup-id"
  # Required. Immutable. Target volume to restore into.
  destination:
    volume:
      name: "my-volume"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxVolumeRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  source:
    backup:
      name: "my-backup"
      namespace: "kyma-system"  # optional
  # Required. Immutable. Target PVC to restore into.
  destination:
    pvc:
      name: "my-pvc"
      namespace: "kyma-system"  # optional
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotRestore
metadata:
  name: my-restore
  namespace: kyma-system
spec:
  # Required. Immutable.
  sourceSnapshot:
    name: "my-snapshot"
    namespace: "kyma-system"  # optional
  # Required. Immutable. Exactly one of the two modes below.
  destination:
    # Mode 1: In-place revert of an existing volume.
    # The snapshot must be the most recent one for the volume.
    existingVolume:
      name: "my-volume"
      namespace: "kyma-system"
    # Mode 2: Create a new SapNfsVolume from the snapshot.
    # (mutually exclusive with existingVolume)
    # newVolume:
    #   metadata:
    #     name: "my-new-volume"
    #     namespace: "kyma-system"
    #   spec:
    #     capacityGb: 100
    #     ipRange:
    #       name: my-iprange
```

</td>
</tr>
</table>

### Scheduled Backup

<table>
<tr>
<th>AwsNfsBackupSchedule</th>
<th>GcpNfsBackupSchedule</th>
<th>AzureRwxBackupSchedule</th>
<th>SapNfsVolumeSnapshotSchedule (OpenStack)</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the AwsNfsVolume to back up.
  nfsVolumeRef:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. AWS region for backup storage.
  location: "us-east-1"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  # If true, deletes all child backups when schedule is deleted.
  deleteCascade: true       # default: false
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the GcpNfsVolume to back up.
  nfsVolumeRef:
    name: "my-volume"
    namespace: "kyma-system"  # optional
  # Optional. GCP region for backup storage.
  location: "us-west1"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
  # Optional. Cross-cluster access for scheduled backups.
  accessibleFrom:
    - "all"
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRwxBackupSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
spec:
  # Required. References the PVC to back up.
  # (Azure uses pvcRef, not nfsVolumeRef)
  pvcRef:
    name: "my-pvc"
    namespace: "kyma-system"  # optional
  # Optional. Azure region for backup storage.
  location: "westeurope"
  # Optional cron expression. If absent, one-shot backup.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated backup names.
  prefix: "my-backup"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 30      # default: 375
  maxReadyBackups: 10       # default: 100
  maxFailedBackups: 3       # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
```

</td>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolumeSnapshotSchedule
metadata:
  name: my-schedule
  namespace: kyma-system
  # OpenStack schedule — the source volume is
  # specified inside the snapshot template.
spec:
  # Optional cron expression. If absent, one-shot snapshot.
  schedule: "0 2 * * *"
  # Optional. Prefix for generated snapshot names.
  prefix: "my-snapshot"
  # Optional. Schedule start/end bounds.
  startTime: "2025-01-01T00:00:00Z"
  endTime: "2026-01-01T00:00:00Z"
  # Retention and cascade controls.
  maxRetentionDays: 90      # default: 375
  maxReadySnapshots: 10     # default: 50
  maxFailedSnapshots: 3     # default: 5
  suspend: false            # default: false
  deleteCascade: true       # default: false
  # Template for each generated SapNfsVolumeSnapshot.
  template:
    labels:
      app: my-app
    annotations:
      custom/annotation: "value"
    spec:
      sourceVolume:
        name: "my-volume"
        namespace: "kyma-system"
      deleteAfterDays: 90
```

</td>
</tr>
</table>

**Differences across providers:**
- **Source reference model:** AWS and GCP reference a Cloud Manager `NfsVolume` object (`source.volume.name`); Azure references a raw Kubernetes PVC (`source.pvc.name`) — there is no `AzureNfsVolume`; SAP/OpenStack uses `sourceVolume` (a Kubernetes `ObjectReference`).
- **Terminology:** SAP/OpenStack uses "snapshot" throughout instead of "backup", reflecting Manila's snapshot API.
- **Cross-cluster access:** GCP Backup uniquely adds `accessibleFrom` — a list of shoot names or subaccount IDs that may restore the backup, enabling cross-cluster backup sharing. No other provider has this field.
- **Discovery resource:** GCP uniquely has a companion `GcpNfsVolumeBackupDiscovery` cluster-scoped resource (see Other Resources) that the controller populates with available backup URIs, useful for cross-cluster restore workflows.
- **Lifecycle/TTL:** AWS Backup uniquely has a `lifecycle` field (cold storage transition + deletion rules), which maps to AWS Backup Vault lifecycle policies.
- **Create-from-backup at volume creation:** GCP NfsVolume supports `sourceBackup`/`sourceBackupUrl` directly in the volume spec, so a new volume can be pre-populated from a backup without a separate Restore object. SAP/OpenStack has the analogous `dataSource.snapshot`. AWS and Azure have no equivalent shortcut.
- **Restore destination modes:** SAP/OpenStack Restore uniquely supports two modes: `existingVolume` (in-place revert — requires the snapshot to be the most recent) and `newVolume` (creates a new `SapNfsVolume` from the snapshot).
- **Schedule source field:** AWS, GCP, and Azure schedule the source via a top-level reference field (`nfsVolumeRef` / `pvcRef`); SAP/OpenStack embeds the source inside a `template.spec.sourceVolume` field.

**Differences across features (Backup vs Restore vs Schedule):**
- Schedule adds cron scheduling (`schedule`), time bounds (`startTime`, `endTime`), and retention controls (`maxRetentionDays`, `maxReadyBackups`/`maxReadySnapshots`, `maxFailedBackups`/`maxFailedSnapshots`, `deleteCascade`, `suspend`).
- Schedule references the source volume/PVC directly; the controller creates individual Backup objects for each triggered run.
- Restore is a one-shot, fire-and-forget operation — it has no scheduling or retention fields.

---


### GcpNfsVolumeBackupDiscovery

GCP-only. Cluster-scoped, empty spec. The controller populates `status.availableBackups` with metadata for all GCP Filestore backups accessible from this shoot (across all namespaces and, if configured, cross-cluster backups). Consumers read this resource to discover backup URIs before constructing a `GcpNfsVolumeRestore`. This is a read-only discovery resource — users create it empty and the controller fills in the status.

<table>
<tr>
<th>GcpNfsVolumeBackupDiscovery</th>
</tr>
<tr>
<td>

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolumeBackupDiscovery
metadata:
  name: my-backup-discovery
  # cluster-scoped, no namespace
spec: {}
  # No spec fields. Controller populates status.availableBackups
  # with: uri, location, shootName, backupName, backupNamespace,
  # volumeName, volumeNamespace, creationTime.
```

</td>
</tr>
</table>
