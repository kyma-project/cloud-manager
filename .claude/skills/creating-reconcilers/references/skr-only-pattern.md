# SKR-Only Pattern (No Backing KCP Resource)

Use this reference when:
- The SKR resource manages other SKR-level objects (backup schedules, snapshot schedules)
- The resource does NOT create a backing KCP resource — it calls KCP clients directly
- Business logic runs in SKR, not delegated to KCP

Do NOT use if the resource creates a KCP resource — see `skr-reconciler.md`

Canonical example: `pkg/skr/gcpnfsvolumebackup/v2/` (NFS backup) and `pkg/skr/gcpnfsbackupschedule/` (backup schedule).

---

## Architecture: Shared Package + Provider Package

**Shared package** (`pkg/skr/backupschedule/`):
- `ScheduleState` interface — minimal contract for common actions
- `ScheduleCalculator` — cron/time logic with injected `clock.Clock`
- Reusable composed actions — operate on `ScheduleState`, no provider awareness

**Provider package** (`pkg/skr/gcpnfsbackupschedule/`):
- `State` struct implementing `ScheduleState` with concrete GCP types
- Provider-specific actions (loadScope, loadSource, createBackup, etc.)

## State Definition

```go
package gcpnfsbackupschedule

type State struct {
    composed.State                          // NOT kcpcommonaction.State
    KymaRef    klog.ObjectRef
    KcpCluster composed.StateCluster
    SkrCluster composed.StateCluster

    // Common scheduling (implements backupschedule.ScheduleState)
    Scheduler      *backupschedule.ScheduleCalculator
    cronExpression *cronexpr.Expression
    nextRunTime    time.Time
    createRunDone  bool
    deleteRunDone  bool

    // Provider-specific — concrete types, no interfaces
    Scope     *cloudcontrolv1beta1.Scope
    SourceRef *cloudresourcesv1beta1.GcpNfsVolume
    Backups   []*cloudresourcesv1beta1.GcpNfsVolumeBackup
}

// Implement backupschedule.ScheduleState interface
func (s *State) ObjAsBackupSchedule() backupschedule.BackupSchedule { ... }
func (s *State) GetScheduleCalculator() *backupschedule.ScheduleCalculator { ... }
func (s *State) GetCronExpression() *cronexpr.Expression { ... }
func (s *State) SetCronExpression(e *cronexpr.Expression) { ... }
```

## Action Composition

```go
func (r *Reconciler) newAction() composed.Action {
    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsBackupSchedule{}),
        composed.If(feature.ApiDisabledPredicate, composed.StopAndForgetAction),
        composed.LoadObj,
        actions.AddCommonFinalizer(),
        composed.If(
            composed.NotMarkedForDeletionPredicate,
            composed.ComposeActionsNoName(
                // Common scheduling actions from shared package
                backupschedule.CheckCompleted,
                backupschedule.CheckSuspension,
                backupschedule.ValidateSchedule,
                backupschedule.CalculateNextRun,
                backupschedule.EvaluateNextRun,
                // Provider-specific actions
                loadScope,
                loadSource,
                loadBackups,
                createBackup,
                deleteBackups,
                updateStatus,
            ),
        ),
        composed.If(
            composed.MarkedForDeletionPredicate,
            composed.ComposeActionsNoName(
                loadBackups,
                deleteCascade,
                actions.RemoveCommonFinalizer(),
            ),
        ),
        composed.StopAndForgetAction,
    )
}
```

## v1/v2 Feature Flag Factory Pattern

The controller factory checks a feature flag to select the reconciler version:

```go
func (f *GcpNfsBackupScheduleReconcilerFactory) New(args ReconcilerArguments) reconcile.Reconciler {
    if feature.BackupScheduleV2.Value(context.Background()) {
        reconciler := gcpnfsbackupschedule.NewReconciler(
            args.KymaRef, args.KcpCluster, args.SkrCluster, f.env, f.clk,
        )
        return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
    }
    // v1 fallback
    reconciler := backupschedulev1.NewReconciler(...)
    return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
}
```

## Clock Injection

```go
// State factory accepts clock
func NewReconciler(kymaRef klog.ObjectRef, ..., clk clock.Clock) Reconciler { ... }

// Production: clock.RealClock{} passed by controller setup
// Tests: clock.NewFakeClock(time.Now()) — advance with fakeClock.Step(duration)
```

```go
// In suite_test.go
var testFakeClock *clock.FakeClock

// In BeforeSuite
testFakeClock = clock.NewFakeClock(time.Now())
Expect(SetupReconciler(infra.Registry(), env, testFakeClock)).NotTo(HaveOccurred())

// In test — advance time past scheduled run
testFakeClock.Step(2 * time.Minute)
```

## Example Files

| Component | File |
|-----------|------|
| Shared state interface | `pkg/skr/backupschedule/schedule_state.go` |
| Shared actions | `pkg/skr/backupschedule/backupschedule.go` |
| Provider state | `pkg/skr/gcpnfsbackupschedule/state.go` |
| Provider reconciler | `pkg/skr/gcpnfsbackupschedule/reconciler.go` |
| v1/v2 controller factory | `internal/controller/cloud-resources/gcpnfsvolumebackup_controller.go` |
