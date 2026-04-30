# Status Mutation Patterns Reference

This document covers the two exclusive patterns for mutating object status in cloud-manager reconcilers.

## Overview

Cloud-manager provides two status mutation patterns:

| Pattern | Location | Recommended For |
|---------|----------|-----------------|
| UpdateStatusBuilder | `pkg/composed/updateStatus.go` | Legacy code, complex condition filtering |
| StatusPatcherComposed | `pkg/composed/statusPatcher.go` | **New code (recommended)** |

**Rule**: Use one pattern per action, never mix them.

---

## StatusPatcherComposed (Recommended)

The modern pattern with built-in change detection and handler chains.

### Basic Usage

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

### Key Features

1. **Change Detection**: Only patches if status actually changed (compares DeepCopy)
2. **IsStale() Check**: Built-in generation vs observedGeneration check
3. **Handler Chains**: Separate handlers for success, failure, and status-changed events

### Constructor

```go
composed.NewStatusPatcherComposed[T ObjWithStatus](obj T) *StatusPatcherComposed[T]
```

### Methods

| Method | Purpose | Returns |
|--------|---------|---------|
| `IsStale()` | Check if generation differs from observedGeneration | `bool` |
| `MutateStatus(func(T))` | Modify status fields | `*StatusPatcherComposed[T]` |
| `SetConditions(conditions...)` | Set conditions (auto-updates observedGeneration) | `*StatusPatcherComposed[T]` |
| `RemoveConditions(types...)` | Remove conditions by type | `*StatusPatcherComposed[T]` |
| `OnSuccess(handlers...)` | Handlers called on successful patch | `*StatusPatcherComposed[T]` |
| `OnFailure(handlers...)` | Handlers called on failed patch | `*StatusPatcherComposed[T]` |
| `OnStatusChanged(handlers...)` | Handlers called only if resource version changed | `*StatusPatcherComposed[T]` |
| `Run(ctx, client)` | Execute patch and handlers | `(error, context.Context)` |

### Pre-built Handlers

```go
composed.Continue     // Continue to next action (default success)
composed.Requeue      // Return StopWithRequeue (default failure)
composed.RequeueAfter(duration)  // Return StopWithRequeueDelay
composed.Forget       // Return StopAndForget
composed.Log(msg)     // Log message
composed.LogError(err, msg)  // Log error with message
composed.LogIf(condition, msg)  // Conditionally log
```

### Default Behavior

- **OnSuccess**: `Continue()` (continue to next action)
- **OnFailure**: `Requeue()` + `Log("failed to patch status...")`
- **OnStatusChanged**: none

### Common Patterns

#### Pattern 1: Initial Status with Stale Check

```go
func statusInitial(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    sp := composed.NewStatusPatcherComposed(state.ObjAsSubscription())
    if !sp.IsStale() {
        return nil, ctx  // Already up to date
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

#### Pattern 2: Conditional Logging

```go
func statusUpdate(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    return composed.NewStatusPatcherComposed(state.ObjAsVpcNetwork()).
        MutateStatus(func(obj *cloudcontrolv1beta1.VpcNetwork) {
            obj.Status.Identifiers.Vpc = state.vpcId
        }).
        OnSuccess(
            composed.LogIf(state.wasCreated, "VpcNetwork created"),
            composed.LogIf(state.wasUpdated, "VpcNetwork updated"),
        ).
        Run(ctx, state.Cluster().K8sClient())
}
```

#### Pattern 3: Error Status with Requeue

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

---

## UpdateStatusBuilder (Legacy)

The original pattern using a fluent builder API.

### Entry Points

```go
// Full update (sends entire object)
composed.UpdateStatus(obj ObjWithConditions)

// JSON patch (sends only changes)
composed.PatchStatus(obj ObjWithConditions)
```

### Builder Methods

| Method | Purpose |
|--------|---------|
| `SetCondition(cond)` | Add/update single condition |
| `SetExclusiveConditions(conds...)` | Replace all conditions |
| `RemoveConditions(types...)` | Remove conditions by type |
| `KeepConditions(types...)` | Keep only specified conditions |
| `RemoveConditionIfReasonMatched(type, reason)` | Conditional removal |
| `ErrorLogMessage(msg)` | Log message on error |
| `SuccessLogMsg(msg)` | Log message on success |
| `SuccessError(err)` | Error to return on success |
| `FailedError(err)` | Error to return on failure |
| `SuccessErrorNil()` | Force nil return on success |
| `DeriveStateFromConditions(func)` | Derive state from conditions |
| `Run(ctx, state)` | Execute and return `(error, context.Context)` |

### Default Behavior

- **SuccessError**: `StopAndForget`
- **FailedError**: `StopWithRequeue`

### Common Patterns

#### Pattern 1: Set Exclusive Conditions

```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
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
}
```

#### Pattern 2: Multiple Condition Operations

```go
func updateStatusReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    return composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
        SetCondition(metav1.Condition{
            Type:    cloudresourcesv1beta1.ConditionTypeReady,
            Status:  metav1.ConditionTrue,
            Reason:  cloudresourcesv1beta1.ConditionTypeReady,
            Message: "Ready",
        }).
        RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
        ErrorLogMessage("Error updating status with ready condition").
        SuccessError(composed.StopWithRequeue).
        Run(ctx, state)
}
```

#### Pattern 3: Error Condition with StopAndForget

```go
func updateStatusError(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    return composed.PatchStatus(state.ObjAsSchedule()).
        SetExclusiveConditions(metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonScopeNotFound,
            Message: "Scope does not exist",
        }).
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

---

## When to Use Which

### Use StatusPatcherComposed When:

- Writing new reconcilers
- You need change detection (avoid unnecessary API calls)
- You want `IsStale()` for generation-based reconciliation
- You need handler chains with flow control
- You want `OnStatusChanged()` for conditional logging

### Use UpdateStatusBuilder When:

- Maintaining legacy reconcilers
- You need `DeriveStateFromConditions()`
- You need complex condition filtering (`KeepConditions`, `RemoveConditionIfReasonMatched`)
- Working with SKR reconcilers that sync KCP conditions

---

## Migration Guide

### Before (UpdateStatusBuilder)

```go
return composed.UpdateStatus(obj).
    SetCondition(metav1.Condition{...}).
    ErrorLogMessage("Error updating status").
    SuccessError(composed.StopAndForget).
    Run(ctx, state)
```

### After (StatusPatcherComposed)

```go
return composed.NewStatusPatcherComposed(obj).
    SetConditions(metav1.Condition{...}).
    OnStatusChanged(composed.Log("Status updated")).
    OnSuccess(composed.Forget).
    Run(ctx, state.Cluster().K8sClient())
```

Key differences:
- Constructor takes just the object
- `Run()` takes `client` not `state`
- Use handlers instead of `SuccessError`/`FailedError`
- Change detection is automatic

---

## See Also

- `references/flow-control.md` - Action return patterns
- `references/primitives.md` - Full pkg/composed documentation
- `pkg/composed/statusPatcher.go` - Source implementation
- `pkg/composed/updateStatus.go` - Legacy implementation
