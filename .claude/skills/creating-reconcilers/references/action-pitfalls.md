# Action Pitfalls

Common mistakes when implementing composed action reconcilers.

## 1. State type confusion

Assert to the wrong state level → runtime panic.

```go
// BAD: asserting composed.State directly to provider state
return &State{State: state.(*State)}  // PANIC if state is not *State

// GOOD: assert through the hierarchy
focalState := state.(focal.State)
return &State{State: focalState, client: client}
```

State hierarchy: `composed.State` → `kcpcommonaction.State` (or `focal.State`) → provider `*State`.

## 2. Missing StopAndForget at flow end

Without `StopAndForgetAction`, the controller requeues indefinitely even after successful work.

```go
// BAD: flow ends without termination
composed.ComposeActionsNoName(loadResource, updateStatus)

// GOOD:
composed.ComposeActionsNoName(loadResource, updateStatus, composed.StopAndForgetAction)
```

Every successful code path MUST end with `StopAndForget` or `StopAndForgetAction`.

## 3. Silent UpdateObjStatus errors

```go
// BAD: ignoring the error
state.UpdateObjStatus(ctx)

// GOOD on create/update path:
if err := state.UpdateObjStatus(ctx); err != nil {
    return err, ctx
}

// OK on error path (already returning an error):
_ = state.UpdateObjStatus(ctx)
```

## 4. No existence check before create

```go
// BAD: creates unconditionally
composed.ComposeActionsNoName(createResource, updateStatus)

// GOOD: load first, create only if absent
composed.ComposeActionsNoName(
    loadResource,
    composed.If(resourceNotExistsPredicate, createResource),
    updateStatus,
)
```

## 5. No wait action after async operation

Cloud API calls (create, delete, update) return an operation ID, not a result.

```go
// BAD: proceeds immediately
composed.ComposeActionsNoName(createResource, updateStatusReady)

// GOOD: poll until done
composed.ComposeActionsNoName(
    createResource,
    waitOperationDone,   // StopWithRequeueDelay until op.Done
    updateStatusReady,
)
```

## 6. Missing condition on error path

```go
// BAD: error returned but status shows nothing
if err != nil {
    return err, ctx
}

// GOOD: set condition before returning
if err != nil {
    composed.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
        Type:   cloudcontrolv1beta1.ConditionTypeError,
        Status: metav1.ConditionTrue,
        Reason: "CreateFailed",
        Message: err.Error(),
    })
    _ = state.UpdateObjStatus(ctx)
    return err, ctx
}
```

## 7. Missing KCP labels when creating from SKR

```go
// BAD: no labels
ObjectMeta: metav1.ObjectMeta{Name: id, Namespace: ns}

// GOOD: three required labels
Labels: map[string]string{
    cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
    cloudcontrolv1beta1.LabelRemoteName:      obj.GetName(),
    cloudcontrolv1beta1.LabelRemoteNamespace: obj.GetNamespace(),
},
```

Missing labels break status sync and make cross-cluster debugging impossible.

## 8. Removing finalizer before cloud deletion confirmed

```go
// BAD: finalizer removed immediately after delete call
composed.ComposeActionsNoName(deleteKcpResource, actions.RemoveCommonFinalizer())

// GOOD: wait for actual deletion
composed.ComposeActionsNoName(
    deleteKcpResource,
    waitKcpResourceDeleted,
    actions.RemoveCommonFinalizer(),
)
```

## 9. Mutating state in actions (state changes are local)

Actions receive a pointer to state but mutations to local-scope values are not shared between separate action pipelines.

Store shared data in context or in the State struct fields that are passed by pointer.

## 10. Using IfElse instead of separate If blocks

```go
// BAD:
composed.IfElse(composed.MarkedForDeletionPredicate, deleteFlow, createFlow)

// GOOD:
composed.If(composed.MarkedForDeletionPredicate, deleteFlow),
composed.If(composed.NotMarkedForDeletionPredicate, createFlow),
```

## 11. Calling time.Now() directly

```go
// BAD: not testable, clock not injected
nextRun := time.Now().Add(interval)

// GOOD: use injected clock
nextRun := r.clock.Now().Add(interval)
```

## 12. NEW pattern vs OLD pattern confusion

New resources (post-2024) MUST use provider-specific CRDs (`GcpRedisCluster`, not `RedisInstance`).
`RedisInstance`, `NfsInstance`, `IpRange` use the legacy multi-provider pattern — maintain as-is, do not replicate.

## 13. Using multi-provider state hierarchy for a single-provider resource

```go
// BAD: single-provider resource adds unnecessary types/ subpackage and interface layer
pkg/kcp/gcpsubnet/types/state.go   // not needed — only one provider ever uses this
pkg/kcp/gcpsubnet/state.go         // not needed — wraps nothing useful

// GOOD: single-provider resource puts everything in the provider package
pkg/kcp/provider/gcp/subnet/state.go      // State embeds focal.State directly
pkg/kcp/provider/gcp/subnet/reconcile.go  // No provider Switch, no types/ package
```

Only add the `types/` subpackage and shared state interface when **multiple providers** must be supported. The `types/` package exists solely to break the circular import between the shared reconciler (which references provider packages for the Switch) and the provider packages (which embed the shared state). See `references/kcp-single-provider.md` vs `references/kcp-multi-provider.md`.
