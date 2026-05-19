# Status Mutation Patterns Reference

Two exclusive patterns for mutating object status. Use one per action — never mix them.

| Pattern | Location | Use For |
|---------|----------|---------|
| `StatusPatcherComposed` | `pkg/composed/statusPatcher.go` | **New code (recommended)** |
| `UpdateStatusBuilder` | `pkg/composed/updateStatus.go` | Legacy code and SKR condition sync |

---

## StatusPatcherComposed (Recommended for New Code)

Built-in change detection: only patches if status actually changed (DeepCopy comparison). Avoids unnecessary API calls.

### Constructor

```go
composed.NewStatusPatcherComposed[T ObjWithStatus](obj T) *StatusPatcherComposed[T]
```

### Methods

| Method | Purpose |
|--------|---------|
| `IsStale()` | True if `generation != observedGeneration` — use to skip no-op status updates |
| `MutateStatus(func(T))` | Modify status fields |
| `SetConditions(conditions...)` | Set conditions (auto-updates `observedGeneration`) |
| `RemoveConditions(types...)` | Remove conditions by type |
| `OnSuccess(handlers...)` | Called on successful patch |
| `OnFailure(handlers...)` | Called on failed patch |
| `OnStatusChanged(handlers...)` | Called only if resource version actually changed |
| `Run(ctx, client)` | Execute patch + handlers; returns `(error, context.Context)` |

### Pre-built Handlers

```go
composed.Continue               // continue to next action (default OnSuccess)
composed.Requeue                // return StopWithRequeue (default OnFailure)
composed.RequeueAfter(duration) // return StopWithRequeueDelay
composed.Forget                 // return StopAndForget
composed.Log(msg)               // log message
composed.LogError(err, msg)     // log error with message
composed.LogIf(condition, msg)  // conditionally log
```

### Pattern 1: Ready status with stale check

```go
func statusReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
        MutateStatus(func(obj *cloudcontrolv1beta1.VpcNetwork) {
            obj.Status.State = "Ready"
            obj.SetStatusProvisioned()
        }).
        OnStatusChanged(composed.Log("VpcNetwork is Ready")).
        Run(ctx, state.Cluster().K8sClient())
}
```

### Pattern 2: Initial status with stale check

```go
func statusInitial(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())
    if !sp.IsStale() {
        return nil, ctx  // generation already reflected in status
    }

    return sp.
        MutateStatus(func(obj *cloudcontrolv1beta1.Subscription) {
            obj.SetStatusProcessing()
        }).
        OnSuccess(composed.Continue).
        OnFailure(composed.Log("Error setting initial status")).
        Run(ctx, state.Cluster().K8sClient())
}
```

### Pattern 3: Error status with requeue

```go
func statusError(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
        MutateStatus(func(obj *cloudcontrolv1beta1.VpcNetwork) {
            obj.SetStatusProviderError(state.providerErr.Error())
        }).
        OnSuccess(
            composed.LogError(state.providerErr, "Provider error"),
            composed.Requeue,
        ).
        Run(ctx, state.Cluster().K8sClient())
}
```

### Pattern 4: Conditional logging on change

```go
return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
    MutateStatus(func(obj *cloudcontrolv1beta1.VpcNetwork) {
        obj.Status.Identifiers.Vpc = state.vpcId
    }).
    OnSuccess(
        composed.LogIf(state.wasCreated, "VpcNetwork created"),
        composed.LogIf(state.wasUpdated, "VpcNetwork updated"),
    ).
    Run(ctx, state.Cluster().K8sClient())
```

---

## UpdateStatusBuilder (Legacy)

Use when maintaining legacy reconcilers, or when you need `DeriveStateFromConditions()`, `KeepConditions()`, or `RemoveConditionIfReasonMatched()`.

### Entry Points

```go
composed.UpdateStatus(obj ObjWithConditions)  // full update (sends entire object)
composed.PatchStatus(obj ObjWithConditions)   // JSON patch (sends only changes)
```

### Builder Methods

| Method | Purpose |
|--------|---------|
| `SetCondition(cond)` | Add/update single condition |
| `SetExclusiveConditions(conds...)` | Replace all conditions with these |
| `RemoveConditions(types...)` | Remove conditions by type |
| `KeepConditions(types...)` | Keep only specified conditions |
| `RemoveConditionIfReasonMatched(type, reason)` | Conditional removal |
| `DeriveStateFromConditions(func)` | Derive `.Status.State` from conditions |
| `ErrorLogMessage(msg)` | Log message on error |
| `SuccessLogMsg(msg)` | Log message on success |
| `SuccessError(err)` | Error to return on success (default: `StopAndForget`) |
| `FailedError(err)` | Error to return on failure (default: `StopWithRequeue`) |
| `OnUpdateError(f)` | Custom error handler |
| `OnUpdateSuccess(f)` | Custom success handler |
| `Run(ctx, state)` | Execute; returns `(error, context.Context)` |

### Pattern 1: Set exclusive conditions

```go
return composed.UpdateStatus(state.ObjAsIpRange()).
    SetExclusiveConditions(metav1.Condition{
        Type:    cloudresourcesv1beta1.ConditionTypeReady,
        Status:  metav1.ConditionTrue,
        Reason:  "Ready",
        Message: "Resource is ready",
    }).
    DeriveStateFromConditions(state.MapConditionToState()).
    ErrorLogMessage("Error updating status").
    Run(ctx, state)
```

### Pattern 2: Multiple condition operations

```go
return composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
    SetCondition(metav1.Condition{
        Type:   cloudresourcesv1beta1.ConditionTypeReady,
        Status: metav1.ConditionTrue,
        Reason: cloudresourcesv1beta1.ConditionTypeReady,
    }).
    RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
    ErrorLogMessage("Error updating status with ready condition").
    SuccessError(composed.StopWithRequeue).
    Run(ctx, state)
```

### Pattern 3: Error condition with StopAndForget

```go
return composed.PatchStatus(state.ObjAsSchedule()).
    SetExclusiveConditions(metav1.Condition{
        Type:    cloudcontrolv1beta1.ConditionTypeError,
        Status:  metav1.ConditionTrue,
        Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
        Message: "Scope does not exist",
    }).
    SuccessError(composed.StopAndForget).
    Run(ctx, state)
```

---

## When to Use Which

| Situation | Use |
|-----------|-----|
| New reconciler | `StatusPatcherComposed` |
| Need change detection (avoid unnecessary API calls) | `StatusPatcherComposed` |
| Need `IsStale()` for generation-based reconciliation | `StatusPatcherComposed` |
| Need `OnStatusChanged()` for conditional logging | `StatusPatcherComposed` |
| Maintaining legacy reconciler | `UpdateStatusBuilder` |
| Need `DeriveStateFromConditions()` | `UpdateStatusBuilder` |
| Need `KeepConditions` / `RemoveConditionIfReasonMatched` | `UpdateStatusBuilder` |
| SKR reconciler syncing KCP conditions | `UpdateStatusBuilder` |

---

## Migration Guide

```go
// BEFORE (UpdateStatusBuilder)
return composed.UpdateStatus(obj).
    SetCondition(metav1.Condition{...}).
    ErrorLogMessage("Error updating status").
    SuccessError(composed.StopAndForget).
    Run(ctx, state)

// AFTER (StatusPatcherComposed)
return composed.NewStatusPatcherComposed(obj).
    SetConditions(metav1.Condition{...}).
    OnStatusChanged(composed.Log("Status updated")).
    OnSuccess(composed.Forget).
    Run(ctx, state.Cluster().K8sClient())
```

Key differences:
- Constructor takes just the object
- `Run()` takes `client` (not `state`)
- Use handler methods instead of `SuccessError`/`FailedError`
- Change detection is automatic

## See Also

- `references/primitives.md` — Full `pkg/composed` documentation, flow control errors, util.Timing
