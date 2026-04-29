# Implementation Plan

- [x] 1. SapNfsVolumeSnapshot CRD and API types
- [x] 1.1 Define `SapNfsVolumeSnapshot` API types using kubebuilder
  - Create `api/cloud-resources/v1beta1/sapnfsvolumesnapshot_types.go` with `SapNfsVolumeSnapshotSpec`, `SapNfsVolumeSnapshotStatus`, kubebuilder markers, CEL immutability validation for `sourceVolume`, and all required interface methods (`State()`, `SetState()`, `Conditions()`, `GetObjectMeta()`, `SpecificToFeature()`, `SpecificToProviders()`, `CloneForPatchStatus()`)
  - Add `AnnotationSnapshotId` constant to a shared annotations file or `api/cloud-resources/v1beta1/`
  - Register types with `SchemeBuilder` in `init()`
  - Run `make manifests` to generate CRDs
  - _Requirements: 1.1, 1.3, 1.6, 1.8_

- [x] 1.2 Write API validation tests for `SapNfsVolumeSnapshot`
  - Create `internal/api-tests/skr_sapnfsvolumesnapshot_test.go` with builder struct and tests:
    - Valid create with sourceVolume (`canCreateSkr`)
    - Immutability of sourceVolume (`canNotChangeSkr` → "SourceVolume is immutable.")
  - _Requirements: 1.3_

- [x] 2. SapNfsVolumeSnapshotRestore CRD and API types
- [x] 2.1 Define `SapNfsVolumeSnapshotRestore` API types using kubebuilder
  - Create `api/cloud-resources/v1beta1/sapnfsvolumesnapshotrestore_types.go` with `SapNfsVolumeSnapshotRestoreSpec`, `SapNfsVolumeSnapshotRestoreDestination` (MinProperties/MaxProperties XOR), `SapNfsVolumeSnapshotNewVolume`, `SapNfsVolumeSnapshotRestoreStatus`, CEL immutability for `sourceSnapshot` and `destination`, and all interface methods
  - Register types with `SchemeBuilder`
  - Run `make manifests`
  - _Requirements: 3.9_

- [x] 2.2 Write API validation tests for `SapNfsVolumeSnapshotRestore`
  - Create `internal/api-tests/skr_sapnfsvolumesnapshotrestore_test.go` with builder and tests:
    - Valid create with existingVolume (`canCreateSkr`)
    - Valid create with newVolume (`canCreateSkr`)
    - Both set → MaxProperties violation (`canNotCreateSkr`)
    - Neither set → MinProperties violation (`canNotCreateSkr`)
    - Immutability of sourceSnapshot (`canNotChangeSkr`)
    - Immutability of destination (`canNotChangeSkr`)
  - _Requirements: 3.9_

- [x] 3. SapNfsVolumeSnapshotSchedule CRD and API types
- [x] 3.1 Define `SapNfsVolumeSnapshotSchedule` API types using kubebuilder
  - Create `api/cloud-resources/v1beta1/sapnfsvolumesnapshotschedule_types.go` with `SapNfsVolumeSnapshotScheduleSpec`, `SapNfsVolumeSnapshotTemplate`, `SapNfsVolumeSnapshotScheduleStatus`, kubebuilder markers with defaults, and all interface methods including `backupschedule.BackupSchedule` interface (`GetSourceRef`, `SetSourceRef`, `GetSchedule`, `SetSchedule`, `GetPrefix`, `SetPrefix`, `GetStartTime`, `SetStartTime`, `GetEndTime`, `SetEndTime`, `GetMaxRetentionDays`, `SetMaxRetentionDays`, `GetSuspend`, `SetSuspend`, `GetDeleteCascade`, `SetDeleteCascade`, `GetMaxReadyBackups`, `SetMaxReadyBackups`, `GetMaxFailedBackups`, `SetMaxFailedBackups`, and all status accessors)
  - Register types with `SchemeBuilder`
  - Run `make manifests`
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9_

- [x] 3.2 Write API validation tests for `SapNfsVolumeSnapshotSchedule`
  - Create `internal/api-tests/skr_sapnfsvolumesnapshotschedule_test.go` with builder and tests:
    - Valid create with template + schedule (`canCreateSkr`)
    - Valid create one-time (no schedule) (`canCreateSkr`)
  - _Requirements: 4.1, 4.2_

- [x] 4. KCP API change and CRD patch script
- [x] 4.1 Add `SnapshotId` to `NfsInstanceOpenStack` and update CRD version patch script
  - Add `SnapshotId string` field to `NfsInstanceOpenStack` in `api/cloud-control/v1beta1/nfsinstance_types.go`
  - Add version patch lines for all three new CRDs in `config/patchAfterMakeManifests.sh`
  - Run `make manifests` then `config/sync.sh`
  - _Requirements: 3.6 (new-volume restore needs SnapshotId propagation)_

- [ ] 5. OpenStack snapshot client
- [ ] 5.1 Implement `SnapshotClient` interface and client
  - Create `pkg/kcp/provider/sap/client/clientSnapshot.go` with `SnapshotClient` interface (`CreateSnapshot`, `GetSnapshot`, `DeleteSnapshot`, `ListSnapshots`, `RevertShareToSnapshot`) and `snapshotClient` struct implementation using gophercloud `snapshots` and `shares.Revert` packages; set microversion `2.27` for `RevertShareToSnapshot`
  - _Requirements: 1.1, 2.1, 3.1_

- [ ] 5.2 Register `SnapshotClient` in `ClientFactory`
  - Add `SnapshotClient(ctx) (SnapshotClient, error)` method to `ClientFactory` in `pkg/kcp/provider/sap/client/client.go`, reusing the existing `shareSvc`
  - _Requirements: 1.1_

- [ ] 6. SAP mock snapshot operations
- [ ] 6.1 Add snapshot operations to the SAP mock
  - Add `SnapshotClient` to `Clients` interface in `pkg/kcp/provider/sap/mock/types.go`
  - Implement `CreateSnapshot`, `GetSnapshot`, `DeleteSnapshot`, `ListSnapshots`, `RevertShareToSnapshot` on the mock project in `pkg/kcp/provider/sap/mock/` (store in mock state, toggle status from `creating` → `available` after first `GetSnapshot` call)
  - _Requirements: 1.1, 2.1, 3.1_

- [ ] 7. SapNfsVolumeSnapshot reconciler
- [ ] 7.1 Create state and reconciler skeleton
  - Create `pkg/skr/sapnfsvolumesnapshot/state.go` with `State` struct (embedding `composed.State`, fields for `Scope`, `SapNfsVolume`, `snapshot`, `sapClient`, `provider`), `StateFactory`, `NewStateFactory()`, helper `ObjAsSapNfsVolumeSnapshot()`
  - Create `pkg/skr/sapnfsvolumesnapshot/reconcile.go` with `Reconciler` struct, `Run()`, `newState()`, `newAction()`, `NewReconciler()`, and `composeActions()` wiring all actions (feature load, loadObj, loadScope, clientCreate, shortCircuit, markFailed, finalizer, snapshotLoad, sourceVolumeLoad, ttlExpiry, IfElse create/delete paths, StopAndForget)
  - _Requirements: 1.6, 1.8, 1.11, 2.4_

- [ ] 7.2 Implement `loadScope` and `clientCreate` actions
  - Create `pkg/skr/sapnfsvolumesnapshot/loadScope.go` — load `Scope` from KCP cluster by `KymaRef`
  - Create `pkg/skr/sapnfsvolumesnapshot/clientCreate.go` — construct SAP client using `SapClientProvider` with Scope credentials
  - _Requirements: 1.1 (OpenStack client needed for Manila calls)_

- [ ] 7.3 Implement `sourceVolumeLoad` action
  - Create `pkg/skr/sapnfsvolumesnapshot/sourceVolumeLoad.go` — load referenced `SapNfsVolume` from SKR cluster, validate it is in `Ready` state, resolve KCP `NfsInstance` to get Manila `shareId` from `status.stateData`
  - _Requirements: 1.2_

- [ ] 7.4 Implement `snapshotLoad` action
  - Create `pkg/skr/sapnfsvolumesnapshot/snapshotLoad.go` — if `status.openstackId` is set, fetch by ID; otherwise fallback to `ListSnapshots` by name+shareId, persist resolved `openstackId`
  - _Requirements: 1.1 (idempotent snapshot resolution)_

- [ ] 7.5 Implement `snapshotCreate` and `snapshotWaitAvailable` actions
  - Create `pkg/skr/sapnfsvolumesnapshot/snapshotCreate.go` — generate deterministic name, store in `status.id`, call `CreateSnapshot`, store `status.openstackId`, set state to `Creating`
  - Create `pkg/skr/sapnfsvolumesnapshot/snapshotWaitAvailable.go` — poll Manila snapshot status; if `available` → continue, if `error` → set Error state
  - _Requirements: 1.1, 1.4, 1.5, 1.6_

- [ ] 7.6 Implement `snapshotDelete` and `snapshotWaitDeleted` actions
  - Create `pkg/skr/sapnfsvolumesnapshot/snapshotDelete.go` — call `DeleteSnapshot`; if not found (404), proceed
  - Create `pkg/skr/sapnfsvolumesnapshot/snapshotWaitDeleted.go` — poll until 404 or `error_deleting`
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 7.7 Implement status actions, `shortCircuit`, `ttlExpiry`, and `markFailed`
  - Create `pkg/skr/sapnfsvolumesnapshot/statusReady.go` — set state `Ready`, update `sizeGb`
  - Create `pkg/skr/sapnfsvolumesnapshot/statusCreating.go` — set state `Creating`
  - Create `pkg/skr/sapnfsvolumesnapshot/statusDeleting.go` — set state `Deleting`
  - Create `pkg/skr/sapnfsvolumesnapshot/shortCircuit.go` — skip if already `Ready` and no changes
  - Create `pkg/skr/sapnfsvolumesnapshot/ttlExpiry.go` — check if `deleteAfterDays` TTL exceeded; if so trigger deletion
  - Create `pkg/skr/sapnfsvolumesnapshot/markFailed.go` — transition from `Error` to `Failed` for scheduled snapshots with newer successors
  - _Requirements: 1.4, 1.7, 1.8, 1.9, 1.10_

- [ ] 7.8 Wire controller: create `SapNfsVolumeSnapshot` controller setup
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshot_controller.go` with `SapNfsVolumeSnapshotReconciler`, `SapNfsVolumeSnapshotReconcilerFactory`, RBAC markers, and `SetupSapNfsVolumeSnapshotReconciler()` that registers with `SkrRegistry`
  - _Requirements: 1.1_

- [ ] 8. SapNfsVolumeSnapshotRestore reconciler
- [ ] 8.1 Create state and reconciler skeleton
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/state.go` with `State` struct (fields for `Scope`, `SourceSnapshot`, `DestinationVolume`, `CreatedVolume`, `sapClient`, `provider`), `StateFactory`, `NewStateFactory()`
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/reconcile.go` with `Reconciler`, `Run()`, `newAction()` wiring all actions (feature load, loadObj, loadScope, clientCreate, finalizer, sourceSnapshotLoad, IfElse non-delete/delete, within non-delete IfElse in-place/new-volume paths, StopAndForget)
  - _Requirements: 3.10, 3.12_

- [ ] 8.2 Implement `loadScope`, `clientCreate`, and `sourceSnapshotLoad` actions
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/loadScope.go` — load Scope from KCP cluster
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/clientCreate.go` — construct SAP client
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/sourceSnapshotLoad.go` — load `SapNfsVolumeSnapshot`, validate `Ready` state
  - _Requirements: 3.2, 3.11_

- [ ] 8.3 Implement in-place restore actions
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/destinationVolumeLoad.go` — load destination `SapNfsVolume`, resolve shareId from KCP `NfsInstance`
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/validateInPlace.go` — validate snapshot belongs to destination volume and is the most recent snapshot
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/restoreInPlace.go` — call `RevertShareToSnapshot()` with microversion 2.27, set state `InProgress`
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/restoreInPlaceWait.go` — poll share status until `available` or `reverting_error`
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.10_

- [ ] 8.4 Implement new-volume restore actions
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/restoreNewVolume.go` — create `SapNfsVolume` CR with snapshot-id annotation from template
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/restoreNewVolumeWait.go` — wait for new `SapNfsVolume` to reach `Ready`, record `createdVolume` in status
  - _Requirements: 3.5, 3.6, 3.7, 3.8_

- [ ] 8.5 Implement status actions
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/statusDone.go` — set state `Done`
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/statusFailed.go` — set state `Failed`
  - Create `pkg/skr/sapnfsvolumesnapshotrestore/statusInProgress.go` — set state `InProgress`
  - _Requirements: 3.3, 3.4, 3.8, 3.10, 3.11_

- [ ] 8.6 Wire controller: create `SapNfsVolumeSnapshotRestore` controller setup
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshotrestore_controller.go` with reconciler struct, factory, RBAC markers, and `SetupSapNfsVolumeSnapshotRestoreReconciler()`
  - _Requirements: 3.12_

- [ ] 9. Existing reconciler changes for snapshot-based volume creation
- [ ] 9.1 Propagate snapshot ID through SapNfsVolume → NfsInstance → Manila
  - Modify `pkg/skr/sapnfsvolume/kcpNfsInstanceCreate.go` to read `AnnotationSnapshotId` from the SapNfsVolume and set `SnapshotId` on `NfsInstanceOpenStack`
  - Modify `pkg/kcp/provider/sap/nfsinstance/shareCreate.go` to read `SnapshotId` from `NfsInstance.Spec.Instance.OpenStack` and pass it to `CreateShareOp()` instead of hardcoded `""`
  - _Requirements: 3.6 (new-volume restore depends on this plumbing)_

- [ ] 10. SapNfsVolumeSnapshotSchedule reconciler
- [ ] 10.1 Create state and reconciler skeleton
  - Create `pkg/skr/sapnfssnapshotschedule/state.go` with `State` struct implementing `backupschedule.ScheduleState` interface (all methods: `ObjAsBackupSchedule`, `GetScheduleCalculator`, `GetCronExpression`, `SetCronExpression`, `GetNextRunTime`, `SetNextRunTime`, `IsCreateRunCompleted`, `SetCreateRunCompleted`, `IsDeleteRunCompleted`, `SetDeleteRunCompleted`), plus provider-specific fields
  - Create `pkg/skr/sapnfssnapshotschedule/reconcile.go` with `Reconciler`, `Run()`, `NewReconciler()`, and `newAction()` composing shared `backupschedule.*` actions and SAP-specific actions (loadSnapshots, loadScope, loadSource, createSnapshot, deleteSnapshots, setStatusToActive, deleteCascade)
  - _Requirements: 4.1, 4.2, 4.10_

- [ ] 10.2 Implement schedule-specific actions
  - Create `pkg/skr/sapnfssnapshotschedule/loadSnapshots.go` — list `SapNfsVolumeSnapshot` CRs by schedule labels
  - Create `pkg/skr/sapnfssnapshotschedule/loadScope.go` — load Scope from KCP cluster
  - Create `pkg/skr/sapnfssnapshotschedule/loadSource.go` — load source `SapNfsVolume`, validate Ready
  - Create `pkg/skr/sapnfssnapshotschedule/createSnapshot.go` — create `SapNfsVolumeSnapshot` with deterministic name, schedule labels, template spec; override `deleteAfterDays` with `maxRetentionDays`
  - Create `pkg/skr/sapnfssnapshotschedule/deleteSnapshots.go` — evict oldest Ready snapshots exceeding `maxReadySnapshots`, garbage-collect Failed beyond `maxFailedSnapshots`
  - Create `pkg/skr/sapnfssnapshotschedule/deleteCascade.go` — delete all snapshots with matching schedule labels when `deleteCascade: true`
  - Create `pkg/skr/sapnfssnapshotschedule/setStatusToActive.go` — set schedule state to Active
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9_

- [ ] 10.3 Wire controller: create `SapNfsVolumeSnapshotSchedule` controller setup
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshotschedule_controller.go` with reconciler struct, factory, RBAC markers, and `SetupSapNfsVolumeSnapshotScheduleReconciler()`
  - _Requirements: 4.10_

- [ ] 11. Controller tests
- [ ] 11.1 Register all three new reconcilers in test suite
  - Add `SetupSapNfsVolumeSnapshotReconciler`, `SetupSapNfsVolumeSnapshotRestoreReconciler`, and `SetupSapNfsVolumeSnapshotScheduleReconciler` calls to `internal/controller/cloud-resources/suite_test.go`
  - _Requirements: all_

- [ ] 11.2 Write SapNfsVolumeSnapshot controller tests
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshot_test.go` using testinfra with `Eventually`/`LoadAndCheck`:
    - Create and delete: create snapshot → verify Manila snapshot created → mock `available` → verify Ready + sizeGb → delete → verify Manila snapshot deleted → finalizer removed
    - TTL expiry: create with `deleteAfterDays: 1` → advance fake clock → verify deletion triggered → verify removed
  - _Requirements: 1.1, 1.4, 1.7, 2.1, 2.2_

- [ ] 11.3 Write SapNfsVolumeSnapshotRestore controller tests
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshotrestore_test.go`:
    - In-place revert: create restore targeting existing volume → verify revert API called → mock `available` → verify Done
    - New-volume restore: create restore with newVolume → verify SapNfsVolume created with snapshot-id annotation → wait Ready → verify Done + createdVolume
  - _Requirements: 3.1, 3.3, 3.5, 3.7_

- [ ] 11.4 Write SapNfsVolumeSnapshotSchedule controller tests
  - Create `internal/controller/cloud-resources/sapnfsvolumesnapshotschedule_test.go`:
    - Recurring schedule with cascade delete: create schedule with cron → verify snapshots created → delete schedule with `deleteCascade: true` → verify all snapshots deleted
    - One-time schedule: create without cron → verify single snapshot created → verify Done
    - Retention count-based: create snapshots exceeding `maxReadySnapshots` → verify oldest evicted
  - _Requirements: 4.1, 4.2, 4.8, 4.9_

- [ ] 12. Build verification
- [ ] 12.1 Run `make manifests` and `make test` to verify everything compiles, CRDs generate, and all tests pass
  - _Requirements: all_
