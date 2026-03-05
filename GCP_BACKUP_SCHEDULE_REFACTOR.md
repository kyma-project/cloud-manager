# GCP Backup Schedule Refactor Plan

**Date**: 2026-03-05  
**Branch**: `refactor-gcp-nfs-backup-scheduler`  
**Scope**: Refactor `pkg/skr/backupschedule` — extract common scheduling logic, create provider-specific reconcilers, gate with `backupScheduleV2` feature flag.

---

## Current State Analysis

### Architecture

The `pkg/skr/backupschedule/` package is a **shared multi-provider SKR reconciler** that handles three backup schedule types:

- `GcpNfsBackupSchedule` (GCP)
- `AwsNfsBackupSchedule` (AWS)
- `AzureRwxBackupSchedule` (Azure)

It uses a strategy pattern (`backupImpl` interface) to abstract provider differences, instantiated via `ScheduleType` enum. This is **pure SKR** — no backing KCP component.

### Problems with Current Implementation

1. **Multi-provider monolith**: All three providers share one reconciler with runtime type switching, making it hard to evolve independently.
2. **`backupImpl` interface is a leaky abstraction**: Actions like `createBackup`, `deleteBackups`, `loadBackups` directly call `state.backupImpl.*()` methods, coupling shared actions to provider-specific types.
3. **Mutable state sharing**: `State` holds `backupImpl`, `Backups`, `SourceRef`, `Scope`, `cronExpression`, `nextRunTime`, `createRunCompleted`, `deleteRunCompleted` — all mutated across action boundaries.
4. **Mixed concerns in actions**: `createBackup` both creates the backup object AND updates schedule status. `deleteBackups` both deletes AND updates status.
5. **Hardcoded error reasons**: Uses `cloudcontrolv1beta1.ReasonGcpError` even for non-GCP providers.
6. **No IfElse for create/delete path**: The reconciler uses a single linear chain with `composed.MarkedForDeletionPredicate` checks inside each action, instead of using `composed.IfElse`.
7. **Package-level mutable global**: `ToleranceInterval` is mutated in tests via `backupschedule.ToleranceInterval = 120 * time.Second`.
8. **No clock abstraction**: Actions call `time.Now()` directly, making controller tests slow (real-time waits of ~2+ minutes) and flaky.
9. **Useless unit tests**: Most unit tests in the package test trivial behavior with heavy mocking that doesn't validate real interactions.
10. **`removeFinalizer` uses deprecated pattern**: Calls `controllerutil.RemoveFinalizer` + `state.UpdateObj` instead of `actions.RemoveCommonFinalizer()`.

### Existing Precedent

The `pkg/skr/gcpnfsvolumebackup/` package already follows the v1/v2 split pattern:
- `pkg/skr/gcpnfsvolumebackup/v1/` — old implementation
- `pkg/skr/gcpnfsvolumebackup/v2/` — refactored implementation
- Controller factory in `gcpnfsvolumebackup_controller.go` checks `feature.GcpBackupV2.Value()` to pick v1 or v2.
- Tests: `gcpnfsvolumebackup_test.go` (v1, skips when flag on) and `gcpnfsvolumebackup_v2_test.go` (v2, skips when flag off).

---

## Design Decisions

### D1: Common Scheduling Package + Provider-Specific Reconcilers

The cron/scheduling logic is truly identical across providers (cron is cron). Rather than each provider copying the scheduling code, the common package exports **reusable composed actions** that operate on the `BackupSchedule` interface. Provider-specific actions (load source, create backup, delete backups, load backups) live entirely in the provider package.

**Common** (`pkg/skr/backupschedule/`):
- `BackupSchedule` interface (already exists — CRD types implement it)
- `ScheduleCalculator` struct with `clock.Clock` field
- Cron expression parsing + validation
- Next run time calculation (one-time + recurring)
- Time comparison with tolerance
- Reusable composed actions: `checkCompleted`, `checkSuspension`, `validateSchedule`, `validateTimes`, `calculateOnetimeSchedule`, `calculateRecurringSchedule`, `evaluateNextRun`
- Constants (`MaxSchedules`)

**Provider-specific** (each in its own reconciler package):
- `pkg/skr/gcpnfsbackupschedule/` — full reconciler with GCP-concrete types
- (Later) `pkg/skr/awsnfsbackupschedule/`, `pkg/skr/azurerwxbackupschedule/`

### D2: Clock Injection via `k8s.io/utils/clock`

All time-dependent logic uses an injected `clock.Clock` instead of `time.Now()`. This is already a project dependency used in `e2e/`.

**Common scheduling package** — `ScheduleCalculator` struct:
```go
type ScheduleCalculator struct {
    Clock clock.Clock
}

func (c *ScheduleCalculator) GetRemainingTime(target time.Time) time.Duration {
    now := c.Clock.Now().UTC()
    // ... calculation with tolerance
}

func (c *ScheduleCalculator) NextRunTimes(expr *cronexpr.Expression, start *time.Time, count int) []time.Time {
    now := c.Clock.Now().UTC()
    if start != nil && start.After(now) {
        return expr.NextN(start.UTC(), count)
    }
    return expr.NextN(now.UTC(), count)
}
```

**Provider state** embeds it:
```go
type State struct {
    composed.State
    // ...
    Scheduler *backupschedule.ScheduleCalculator
}
```

**Production**: `clock.RealClock{}`
**Controller tests**: `clock.NewFakeClock(fixedTime)` — tests call `fakeClock.Step(duration)` to advance time, making them fully deterministic and fast.

**Requeue delays** use `util.Timing.T*()` as usual. With `SetSpeedyTimingForTests()` (divider=100), delays are orders of magnitude shorter in tests:

| Production | Test |
|---|---|
| 100ms | 1ms |
| 1000ms | 10ms |
| 10000ms | 100ms |
| 60000ms | 600ms |

So after `fakeClock.Step()`, the reconciler requeues in ~100ms, sees the clock has advanced past the target time, and proceeds immediately. No 2-minute real-time waits.

### D3: Single Lifecycle Flow (No Separate Create/Delete Tests)

Delete is part of create — a single test scenario covers the full lifecycle:

1. Create schedule → verify backups get created → delete schedule → verify cascade cleanup → verify finalizer removed → verify schedule gone

This matches the established project pattern (e.g., `gcpnfsvolumebackup_v2_test.go` does create → verify Ready → delete → verify gone in one `It` block). Separate delete scenarios are only needed when **preconditions differ** (e.g., "delete while backup creation is in progress").

### D4: Feature Flag at Factory Level

The `backupScheduleV2` flag is checked in the controller factory's `New()` method (same pattern as `gcpnfsvolumebackup_controller.go`), not inside reconciler actions. Only GCP gets v2; AWS and Azure stay on v1.

### D5: Tolerance as Configuration, Not Global Mutable

The current `ToleranceInterval` global is replaced by a field on `ScheduleCalculator`. Production uses 1 second, tests configure whatever they need via the calculator instance — no package-level mutation.

---

## Refactor Steps

### Step 1: Create Feature Flag `backupScheduleV2`

**Files to create/modify:**

1. **Create** `pkg/feature/ffBackupScheduleV2.go`:
   ```go
   package feature

   import "context"

   const BackupScheduleV2FlagName = "backupScheduleV2"

   var BackupScheduleV2 = &backupScheduleV2Info{}

   type backupScheduleV2Info struct{}

   func (k *backupScheduleV2Info) Value(ctx context.Context) bool {
       return provider.BoolVariation(ctx, BackupScheduleV2FlagName, false)
   }
   ```

2. **Modify** `pkg/feature/ff_ga.yaml` — add entry (default: disabled):
   ```yaml
   backupScheduleV2:
     variations:
       enabled: true
       disabled: false
     defaultRule:
       variation: disabled
   ```

3. **Modify** `pkg/feature/ff_edge.yaml` — add entry (default: disabled):
   ```yaml
   backupScheduleV2:
     variations:
       enabled: true
       disabled: false
     defaultRule:
       variation: disabled
   ```

4. **Modify** `.github/workflows/pr-checks.yml` — add matrix entry + env var:
   - Matrix: `backupScheduleV2: "false"` and `backupScheduleV2: "true"`
   - Env: `FF_BACKUP_SCHEDULE_V2: ${{ matrix.backupScheduleV2 }}`

**Verification**: `make build` passes, flag defaults to `false`.

---

### Step 2: Move Current Implementation to `v1/`

**Action**: Move all files from `pkg/skr/backupschedule/` into `pkg/skr/backupschedule/v1/`.

Files to move (change `package backupschedule` → `package v1`):
- `reconciler.go`
- `state.go`
- `backupschedule.go` (BackupSchedule interface)
- `backupImpl.go`
- `backupImplGcpNfs.go`
- `backupImplAwsNfs.go`
- `backupImplAzureRwx.go`
- `compare.go`
- `addFinalizer.go`
- `checkCompleted.go`
- `checkSuspension.go`
- `validateSchedule.go`
- `validateTimes.go`
- `calculateOnetimeSchedule.go`
- `calculateRecurringSchedule.go`
- `evaluateNextRun.go`
- `loadScope.go`
- `loadSource.go`
- `loadBackups.go`
- `createBackup.go`
- `deleteBackups.go`
- `deleteCascade.go`
- `removeFinalizer.go`
- All `*_test.go` files

**Shared types in root `pkg/skr/backupschedule/`**: The `BackupSchedule` interface and `ScheduleType` constants stay in the root package (or are re-exported) so external references compile. The `ScheduleCalculator` and common scheduling actions are created here.

**Update imports** in controllers:
- `gcpnfsbackupschedule_controller.go`: `backupschedulev1 "github.com/kyma-project/cloud-manager/pkg/skr/backupschedule/v1"`
- `awsnfsbackupschedule_controller.go`: same
- `azurerwxbackupschedule_controller.go`: same
- `gcpnfsbackupschedule_test.go`: `backupschedulev1 "github.com/kyma-project/cloud-manager/pkg/skr/backupschedule/v1"` (for `ToleranceInterval`)

**Verification**: `make build` and `make test` pass with no behavior change.

---

### Step 3: Create Common Scheduling Package

Create reusable scheduling logic in `pkg/skr/backupschedule/` (the root package):

```
pkg/skr/backupschedule/
├── backupschedule.go          # BackupSchedule interface (kept from v1)
├── calculator.go              # ScheduleCalculator with clock.Clock
├── checkCompleted.go          # Reusable action (operates on BackupSchedule)
├── checkSuspension.go         # Reusable action
├── validateSchedule.go        # Reusable action
├── validateTimes.go           # Reusable action
├── calculateOnetimeSchedule.go  # Reusable action
├── calculateRecurringSchedule.go # Reusable action
├── evaluateNextRun.go         # Reusable action
├── calculator_test.go         # Unit tests for cron/time logic
```

#### `calculator.go`

```go
package backupschedule

import (
    "time"
    "github.com/gorhill/cronexpr"
    "k8s.io/utils/clock"
)

const MaxSchedules = 3

type ScheduleCalculator struct {
    Clock     clock.Clock
    Tolerance time.Duration
}

func NewScheduleCalculator(clk clock.Clock, tolerance time.Duration) *ScheduleCalculator {
    return &ScheduleCalculator{Clock: clk, Tolerance: tolerance}
}

func (c *ScheduleCalculator) Now() time.Time {
    return c.Clock.Now().UTC()
}

func (c *ScheduleCalculator) GetRemainingTime(target time.Time) time.Duration {
    return c.GetRemainingTimeWithTolerance(target, c.Tolerance)
}

func (c *ScheduleCalculator) GetRemainingTimeWithTolerance(target time.Time, tolerance time.Duration) time.Duration {
    now := c.Now()
    timeLeft := target.Unix() - now.Unix()
    if math.Abs(float64(timeLeft)) <= tolerance.Seconds() {
        return 0
    }
    return time.Duration(timeLeft) * time.Second
}

func (c *ScheduleCalculator) NextRunTimes(expr *cronexpr.Expression, start *time.Time, count int) []time.Time {
    now := c.Now()
    if start != nil && !start.IsZero() && start.After(now) {
        return expr.NextN(start.UTC(), count)
    }
    return expr.NextN(now.UTC(), count)
}
```

#### Common actions as reusable composed actions

The common actions (checkCompleted, checkSuspension, validateSchedule, validateTimes, calculateOnetimeSchedule, calculateRecurringSchedule, evaluateNextRun) operate on a lightweight **`ScheduleState` interface**:

```go
// ScheduleState is the minimal interface the common scheduling actions need.
// Provider states implement this by embedding their own state and exposing these methods.
type ScheduleState interface {
    composed.State
    ObjAsBackupSchedule() BackupSchedule
    GetScheduleCalculator() *ScheduleCalculator
    GetCronExpression() *cronexpr.Expression
    SetCronExpression(expr *cronexpr.Expression)
    GetNextRunTime() time.Time
    SetNextRunTime(t time.Time)
    IsCreateRunCompleted() bool
    SetCreateRunCompleted(v bool)
    IsDeleteRunCompleted() bool
    SetDeleteRunCompleted(v bool)
}
```

Each common action asserts `st.(ScheduleState)` and operates through the interface. Provider states implement `ScheduleState` with their concrete fields.

#### Unit tests (`calculator_test.go`)

These test **concrete business logic** — warranted per AGENTS.md:

- **Cron parsing**: valid expressions → correct next N times; invalid → error
- **Edge cases**: midnight rollover, month boundaries, DST transitions
- **Start time in future**: schedules from start time, not `Now()`
- **GetRemainingTime**: within tolerance → 0; outside → proper duration
- **GetRemainingTimeWithTolerance(0)**: exact comparison
- **Fake clock**: all tests use `clock.NewFakeClock(fixedTime)` for determinism

---

### Step 4: Create v2 GCP-specific Reconciler

Create `pkg/skr/gcpnfsbackupschedule/` for the GCP NFS backup schedule v2 implementation.

**Rationale**: Each provider gets its own reconciler package at `pkg/skr/<resource>/`, matching the project convention (e.g., `pkg/skr/gcpnfsvolume/`, `pkg/skr/gcpnfsvolumebackup/`). v2 eliminates the `backupImpl` indirection entirely.

**Directory structure**:
```
pkg/skr/gcpnfsbackupschedule/
├── state.go
├── reconciler.go
├── loadScope.go
├── loadSource.go
├── loadBackups.go
├── createBackup.go
├── deleteBackups.go
├── deleteCascade.go
├── updateStatus.go
```

#### `state.go`

```go
package gcpnfsbackupschedule

import (
    "context"

    "github.com/gorhill/cronexpr"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
    "k8s.io/klog/v2"
    "k8s.io/utils/clock"
)

type State struct {
    composed.State
    KymaRef    klog.ObjectRef
    KcpCluster composed.StateCluster
    SkrCluster composed.StateCluster

    // Common scheduling
    Scheduler      *backupschedule.ScheduleCalculator
    cronExpression *cronexpr.Expression
    nextRunTime    time.Time
    createRunDone  bool
    deleteRunDone  bool

    // GCP-specific (concrete types, no interfaces)
    Scope     *cloudcontrolv1beta1.Scope
    SourceRef *cloudresourcesv1beta1.GcpNfsVolume
    Backups   []*cloudresourcesv1beta1.GcpNfsVolumeBackup
}

// Implement backupschedule.ScheduleState interface
func (s *State) ObjAsBackupSchedule() backupschedule.BackupSchedule {
    return s.Obj().(backupschedule.BackupSchedule)
}
func (s *State) GetScheduleCalculator() *backupschedule.ScheduleCalculator {
    return s.Scheduler
}
func (s *State) GetCronExpression() *cronexpr.Expression { return s.cronExpression }
func (s *State) SetCronExpression(e *cronexpr.Expression) { s.cronExpression = e }
func (s *State) GetNextRunTime() time.Time { return s.nextRunTime }
func (s *State) SetNextRunTime(t time.Time) { s.nextRunTime = t }
func (s *State) IsCreateRunCompleted() bool { return s.createRunDone }
func (s *State) SetCreateRunCompleted(v bool) { s.createRunDone = v }
func (s *State) IsDeleteRunCompleted() bool { return s.deleteRunDone }
func (s *State) SetDeleteRunCompleted(v bool) { s.deleteRunDone = v }

// GCP-specific getter
func (s *State) ObjAsGcpNfsBackupSchedule() *cloudresourcesv1beta1.GcpNfsBackupSchedule {
    return s.Obj().(*cloudresourcesv1beta1.GcpNfsBackupSchedule)
}

type StateFactory interface {
    NewState(ctx context.Context, baseState composed.State) (*State, error)
}
```

Key improvements:
- **No `backupImpl` interface** — all types are concrete GCP types.
- **Typed `SourceRef`** — `*cloudresourcesv1beta1.GcpNfsVolume` instead of `composed.ObjWithConditions`.
- **Typed `Backups`** — `[]*cloudresourcesv1beta1.GcpNfsVolumeBackup` instead of `[]client.Object`.
- **Implements `ScheduleState`** — common scheduling actions work on this state.
- **`ScheduleCalculator` injected** — no `time.Now()` calls anywhere.

#### `reconciler.go`

```go
package gcpnfsbackupschedule

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster,
    skrCluster cluster.Cluster, env abstractions.Environment, clk clock.Clock) Reconciler {
    // ... factory wiring with clk ...
}

func (r *Reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpNfsBackupScheduleV2",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpNfsBackupSchedule{}),
        composed.LoadObj,
        actions.AddCommonFinalizer(),
        composed.IfElse(
            composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "gcpNfsBackupScheduleV2-main",
                // Common scheduling actions (from pkg/skr/backupschedule/)
                backupschedule.CheckCompleted,
                backupschedule.CheckSuspension,
                backupschedule.ValidateSchedule,
                backupschedule.ValidateTimes,
                backupschedule.CalculateOnetimeSchedule,
                backupschedule.CalculateRecurringSchedule,
                backupschedule.EvaluateNextRun,
                // GCP-specific actions
                loadScope,
                loadSource,
                loadBackups,
                createBackup,
                deleteBackups,
                updateStatus,
            ),
            composed.ComposeActions(
                "gcpNfsBackupScheduleV2-delete",
                loadBackups,
                deleteCascade,
                actions.RemoveCommonFinalizer(),
            ),
        ),
        composed.StopAndForgetAction,
    )
}
```

Key improvements:
- **Proper `IfElse`** for create vs delete path — no more `MarkedForDeletion` checks inside every action.
- **`actions.AddCommonFinalizer()` / `actions.RemoveCommonFinalizer()`** — standard helpers instead of custom.
- **Common scheduling actions imported** from `backupschedule` package.
- **GCP-specific actions** directly use concrete types.
- **Clock injected via `NewReconciler`** — passed through to `ScheduleCalculator`.

#### Action Improvements

Each GCP-specific action is simpler than v1:

**`createBackup.go`**: Directly constructs `*cloudresourcesv1beta1.GcpNfsVolumeBackup` with GcpNfsVolumeBackupSpec — no `backupImpl.getBackupObject()` indirection.

**`loadBackups.go`**: Directly lists `cloudresourcesv1beta1.GcpNfsVolumeBackupList`, assigns `[]*cloudresourcesv1beta1.GcpNfsVolumeBackup` to state — no `toObjectSlice()`.

**`loadSource.go`**: Directly loads `*cloudresourcesv1beta1.GcpNfsVolume`, checks ready condition directly — no `sourceToObjWithConditionAndState()`.

**`deleteCascade.go`**: Only runs in delete branch (IfElse), no deletion-predicate check needed.

---

### Step 5: Wire v2 into the GCP Controller via Feature Flag

**Modify** `internal/controller/cloud-resources/gcpnfsbackupschedule_controller.go`:

Follow the same pattern as `gcpnfsvolumebackup_controller.go`:

```go
package cloudresources

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    backupschedulev1 "github.com/kyma-project/cloud-manager/pkg/skr/backupschedule/v1"
    "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsbackupschedule"
    "k8s.io/utils/clock"
    // ...
)

type gcpNfsBackupScheduleRunner interface {
    Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

type GcpNfsBackupScheduleReconciler struct {
    reconciler gcpNfsBackupScheduleRunner
}

func (r *GcpNfsBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    return r.reconciler.Run(ctx, req)
}

type GcpNfsBackupScheduleReconcilerFactory struct {
    env abstractions.Environment
    clk clock.Clock
}

func (f *GcpNfsBackupScheduleReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
    if feature.BackupScheduleV2.Value(context.Background()) {
        reconciler := gcpnfsbackupschedule.NewReconciler(
            args.KymaRef, args.KcpCluster, args.SkrCluster, f.env, f.clk,
        )
        return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
    }

    reconciler := backupschedulev1.NewReconciler(
        args.KymaRef, args.KcpCluster, args.SkrCluster, f.env,
        backupschedulev1.GcpNfsBackupSchedule,
    )
    return &GcpNfsBackupScheduleReconciler{reconciler: &reconciler}
}

func SetupGcpNfsBackupScheduleReconciler(
    reg skrruntime.SkrRegistry,
    env abstractions.Environment,
    clk clock.Clock,
) error {
    return reg.Register().
        WithFactory(&GcpNfsBackupScheduleReconcilerFactory{env: env, clk: clk}).
        For(&cloudresourcesv1beta1.GcpNfsBackupSchedule{}).
        Complete()
}
```

**Keep** AWS and Azure controllers pointing to `v1` only (not part of this refactor).

**Modification to `suite_test.go`**:
```go
// Create a fake clock for v2 controller tests
var testFakeClock *clock.FakeClock

// In BeforeSuite:
testFakeClock = clock.NewFakeClock(time.Now())
Expect(SetupGcpNfsBackupScheduleReconciler(infra.Registry(), env, testFakeClock)).NotTo(HaveOccurred())
```

The `testFakeClock` is available to all test files in the package. v1 tests don't use it (v1 calls `time.Now()` internally). v2 tests call `testFakeClock.Step()` to advance time.

**Verification**: `make build` passes. Feature flag defaults to v1.

---

### Step 6: Controller Tests for v2

**Create** `internal/controller/cloud-resources/gcpnfsbackupschedule_v2_test.go`:

Uses `testFakeClock` to make tests fast and deterministic.

```go
package cloudresources

import (
    "context"
    "fmt"
    "time"

    "github.com/kyma-project/cloud-manager/pkg/feature"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    . "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR GcpNfsBackupSchedule V2", func() {
    BeforeEach(func() {
        if !feature.BackupScheduleV2.Value(context.Background()) {
            Skip("Skipping v2 tests because backupScheduleV2 feature flag is disabled")
        }
    })

    It("Scenario: Recurring schedule full lifecycle", func() {
        // Given: Scope, GcpNfsVolume in Ready state
        // When: Create GcpNfsBackupSchedule with cron "* * * * *"
        // Then: Schedule gets Active state and NextRunTimes populated

        // When: Advance fake clock past first scheduled run
        testFakeClock.Step(2 * time.Minute)

        // Then: GcpNfsVolumeBackup is created with correct labels and GCP fields
        // And: Schedule status updated (LastCreateRun, BackupIndex, BackupCount)

        // When: Delete GcpNfsBackupSchedule
        // Then: Cascade deletes backups
        // And: Finalizer removed
        // And: Schedule object gone
    })

    It("Scenario: One-time schedule", func() {
        // Given: Scope, GcpNfsVolume in Ready state
        // When: Create GcpNfsBackupSchedule with no cron, start time = now + 1min
        // And: Advance fake clock past start time
        testFakeClock.Step(2 * time.Minute)

        // Then: Single backup created
        // And: Schedule transitions to Done after backup is created
    })

    It("Scenario: Schedule with AccessibleFrom", func() {
        // When: Create schedule with AccessibleFrom = ["shoot-1", "shoot-2"]
        // And: Advance clock
        // Then: Created backup has .spec.accessibleFrom matching
    })

    It("Scenario: Suspension", func() {
        // Given: Active schedule
        // When: Set Suspend = true
        // Then: State → Suspended, NextRunTimes cleared
    })

    // Additional scenarios: invalid cron, source not ready, scope not found
})
```

**Key testing pattern with fake clock:**
1. Test creates schedule with start time relative to `testFakeClock.Now()`
2. Reconciler runs, sees time hasn't arrived → returns `StopWithRequeueDelay(util.Timing.T10000ms())` → requeues in **100ms** (test timing, divider=100)
3. Test calls `testFakeClock.Step(2 * time.Minute)`
4. Reconciler requeues, `clock.Now()` shows past target → creates backup
5. `Eventually(LoadAndCheck)` picks up the backup

No real-time waits. Tests complete in milliseconds per scenario.

**Modify** existing `gcpnfsbackupschedule_test.go`:
- Add skip when v2 flag is enabled:
  ```go
  BeforeEach(func() {
      if feature.BackupScheduleV2.Value(context.Background()) {
          Skip("Skipping v1 tests because backupScheduleV2 feature flag is enabled")
      }
  })
  ```

**Verification**: Tests pass with flag=false (v1 tests run, v2 skip) and flag=true (v2 tests run, v1 skip).

---

### Step 7: Update CI Pipeline

**Modify** `.github/workflows/pr-checks.yml`:
- Add to matrix:
  ```yaml
  - name: Build & Test
    backupScheduleV2: "false"
    # ...
  - name: Build & Test (... flags ...)
    backupScheduleV2: "true"
    # ...
  ```
- Add env mapping:
  ```yaml
  FF_BACKUP_SCHEDULE_V2: ${{ matrix.backupScheduleV2 }}
  ```

---

### Step 8: Cleanup and Documentation

1. Remove useless unit tests from v1 (tests that only mock everything and verify mock calls).
2. Update `AGENTS.md` if any new patterns emerge.
3. Update `docs/agents/reference/COMMON_PITFALLS.md` with any issues found during refactoring.

---

## Testing Strategy

### Unit Tests: Common Scheduling Logic (`pkg/skr/backupschedule/calculator_test.go`)

These test **concrete business logic** — warranted per AGENTS.md rules:

| Category | Test Cases |
|----------|-----------|
| Cron parsing | Valid expressions → correct next N times; Invalid → error |
| Edge cases | Midnight rollover, month boundaries, DST transitions |
| Start time | Future start → schedules from start; Past start → schedules from now |
| `GetRemainingTime` | Within tolerance → 0; Outside → proper duration |
| `GetRemainingTimeWithTolerance(0)` | Exact comparison, no tolerance |
| Fake clock | All tests use `clock.NewFakeClock(fixedTime)` for determinism |

### Controller Tests: GCP Provider (`internal/controller/cloud-resources/gcpnfsbackupschedule_v2_test.go`)

Full integration tests using `testinfra` framework with fake clock:

| Scenario | Validates |
|----------|-----------|
| Recurring schedule full lifecycle | Create → backup created → delete → cascade cleanup → finalizer removed |
| One-time schedule | Create → single backup → state transitions to Done |
| Schedule with AccessibleFrom | Backup inherits AccessibleFrom from schedule |
| Suspension | Suspend=true → state Suspended, NextRunTimes cleared |
| Invalid cron expression | Error state + condition |
| Source not ready | Error state + appropriate condition |
| Scope not found | Error state + StopAndForget |

### Test Configuration

- **v1 tests**: Skip when `backupScheduleV2` flag is enabled
- **v2 tests**: Skip when `backupScheduleV2` flag is disabled
- **CI matrix**: Runs both configurations (flag=false and flag=true)
- **Fake clock**: Shared `testFakeClock` variable in `suite_test.go`, stepped in v2 tests
- **Timing**: `util.SetSpeedyTimingForTests()` (divider=100) ensures requeue delays are ~1-100ms

---

## File Inventory

### Files to Create

| File | Purpose |
|------|---------|
| `pkg/feature/ffBackupScheduleV2.go` | Feature flag definition |
| `pkg/skr/backupschedule/calculator.go` | `ScheduleCalculator` with `clock.Clock` |
| `pkg/skr/backupschedule/schedule_state.go` | `ScheduleState` interface for common actions |
| `pkg/skr/backupschedule/checkCompleted.go` | Reusable common action |
| `pkg/skr/backupschedule/checkSuspension.go` | Reusable common action |
| `pkg/skr/backupschedule/validateSchedule.go` | Reusable common action |
| `pkg/skr/backupschedule/validateTimes.go` | Reusable common action |
| `pkg/skr/backupschedule/calculateOnetimeSchedule.go` | Reusable common action |
| `pkg/skr/backupschedule/calculateRecurringSchedule.go` | Reusable common action |
| `pkg/skr/backupschedule/evaluateNextRun.go` | Reusable common action |
| `pkg/skr/backupschedule/calculator_test.go` | Unit tests for scheduling logic |
| `pkg/skr/gcpnfsbackupschedule/state.go` | GCP-specific state |
| `pkg/skr/gcpnfsbackupschedule/reconciler.go` | GCP-specific reconciler |
| `pkg/skr/gcpnfsbackupschedule/loadScope.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/loadSource.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/loadBackups.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/createBackup.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/deleteBackups.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/deleteCascade.go` | GCP action |
| `pkg/skr/gcpnfsbackupschedule/updateStatus.go` | GCP action |
| `internal/controller/cloud-resources/gcpnfsbackupschedule_v2_test.go` | v2 controller tests |

### Files to Move (to `v1/` subfolder)

All files currently in `pkg/skr/backupschedule/` (they become `pkg/skr/backupschedule/v1/`).

### Files to Modify

| File | Change |
|------|--------|
| `pkg/feature/ff_ga.yaml` | Add `backupScheduleV2` entry |
| `pkg/feature/ff_edge.yaml` | Add `backupScheduleV2` entry |
| `.github/workflows/pr-checks.yml` | Add matrix + env var |
| `internal/controller/cloud-resources/gcpnfsbackupschedule_controller.go` | v1/v2 factory switching, accept `clock.Clock` |
| `internal/controller/cloud-resources/awsnfsbackupschedule_controller.go` | Update import to v1 |
| `internal/controller/cloud-resources/azurerwxbackupschedule_controller.go` | Update import to v1 |
| `internal/controller/cloud-resources/suite_test.go` | Add `testFakeClock`, update setup call |
| `internal/controller/cloud-resources/gcpnfsbackupschedule_test.go` | Add v2-skip guard |

---

## Key Design Decisions Summary

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | Common scheduling package + provider-specific reconcilers | Cron logic is identical across providers; provider actions differ |
| D2 | Clock injection via `k8s.io/utils/clock` | Deterministic tests, no real-time waits |
| D3 | Single lifecycle flow in tests | Matches project pattern; delete is part of create scenario |
| D4 | Feature flag at factory level | Same pattern as `gcpnfsvolumebackup_controller.go` |
| D5 | Tolerance as `ScheduleCalculator` config | Replaces global mutable `ToleranceInterval` |
| D6 | GCP-only for v2 first | AWS/Azure stay on v1, avoids scope creep |
| D7 | No `backupImpl` in v2 | Concrete types, no strategy indirection |
| D8 | `IfElse` for create/delete | No per-action `MarkedForDeletion` checks |
| D9 | Standard finalizer helpers | `actions.AddCommonFinalizer()` / `RemoveCommonFinalizer()` |
| D10 | Common actions via `ScheduleState` interface | Thin interface, reusable actions, no `backupImpl` |

---

## Execution Order

| # | Step | Depends On | Risk |
|---|------|------------|------|
| 1 | Create feature flag | None | Low |
| 2 | Move current code to v1/ | Step 1 | Medium (import updates) |
| 3 | Create common scheduling package | Step 2 | Medium |
| 4 | Create v2 GCP reconciler | Steps 2-3 | Medium |
| 5 | Wire v2 into controller | Steps 1-4 | Low |
| 6 | Write v2 controller tests | Steps 1-5 | Low |
| 7 | Update CI pipeline | Step 1 | Low |
| 8 | Cleanup and docs | Steps 1-7 | Low |

Steps 1 and 2 should be done first and verified independently (`make build && make test`). Step 3 creates the common package with unit tests. Steps 4-6 form the core v2 implementation. Steps 7-8 are finalization.

---

## Progress Tracker

| # | Step | Complete |
|---|------|----------|
| 1 | Create feature flag `backupScheduleV2` | Yes |
| 2 | Move current implementation to `v1/` | Yes |
| 3 | Create common scheduling package | Yes |
| 4 | Create v2 GCP-specific reconciler | Yes |
| 5 | Wire v2 into GCP controller via feature flag | Yes |
| 6 | Controller tests for v2 | |
| 7 | Update CI pipeline | Yes |
| 8 | Cleanup and documentation | |
