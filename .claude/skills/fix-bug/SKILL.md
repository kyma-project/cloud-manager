---
name: fix-bug
description: Debug and fix issues in Cloud Manager reconcilers. Use when troubleshooting errors, investigating test failures, debugging reconciliation issues, or when something is not working.
---

# Fix Bug / Troubleshoot

Debug and resolve issues in Cloud Manager reconcilers.

## Quick Start

1. Identify error type (test failure, reconciler stuck, status not updating)
2. Check this guide for common pitfalls
3. Add logging to narrow down issue
4. Verify action flow and return values
5. Check mock state if testing

## Common Pitfalls

### 1. Reconciler Never Completes

**Symptom**: Resource stays in Processing forever

**Cause**: Missing `StopAndForget` at end of flow

**Fix**:
```go
// ❌ WRONG
composed.ComposeActions("main",
    loadResource,
    createResource,
    updateStatus,
)

// ✅ CORRECT
composed.ComposeActions("main",
    loadResource,
    createResource,
    updateStatus,
    composed.StopAndForgetAction,  // MUST have this
)
```

### 2. Status Not Updating

**Symptom**: Status fields remain empty or stale

**Cause**: Missing `UpdateObjStatus()` call

**Fix**:
```go
// ❌ WRONG
obj.Status.Id = remoteResource.Id
return nil, ctx

// ✅ CORRECT
obj.Status.Id = remoteResource.Id
err := state.UpdateObjStatus(ctx)
if err != nil {
    return err, ctx
}
return nil, ctx
```

### 3. Test Timeout

**Symptom**: `Eventually()` times out

**Cause**: Mock state not transitioning

**Fix**:
```go
// Missing state transition - reconciler waits forever
By("Then resource is ready", func() {
    Eventually(LoadAndCheck)...
})

// ✅ ADD mock state transition
By("When cloud marks resource ready", func() {
    infra.GcpMock().SetResourceState(id, "READY")
})

By("Then resource is ready", func() {
    Eventually(LoadAndCheck)...
})
```

### 4. Wrong Action Return Value

**Symptom**: Unexpected requeue or flow stops early

**Cause**: Incorrect return values

**Reference**:
| Return | Effect |
|--------|--------|
| `nil, ctx` | Continue to next action |
| `error, ctx` | Stop + requeue with backoff |
| `composed.StopAndForget, nil` | Stop successfully, no requeue |
| `composed.StopWithRequeue, nil` | Stop + immediate requeue |
| `composed.StopWithRequeueDelay(d), nil` | Stop + requeue after delay |

### 5. 404 Causes Failure

**Symptom**: Error when resource doesn't exist in cloud

**Cause**: Not handling NotFound as success

**Fix**:
```go
// ❌ WRONG
resource, err := client.Get(ctx, id)
if err != nil {
    return err, ctx  // 404 causes error!
}

// ✅ CORRECT
resource, err := client.Get(ctx, id)
if err != nil {
    if IsNotFound(err) {
        return nil, ctx  // 404 is OK - resource just doesn't exist
    }
    return err, ctx
}
```

### 6. Wrong State Type Assertion

**Symptom**: Panic or nil pointer

**Cause**: Casting to wrong state type

**Fix**:
```go
// ❌ WRONG
state := st.(*focal.State)  // Wrong type

// ✅ CORRECT
state := st.(*State)  // Your provider state type
```

### 7. Finalizer Not Removed

**Symptom**: Resource deletion stuck

**Cause**: Missing finalizer removal in delete flow

**Fix**:
```go
composed.ComposeActions("delete",
    deleteResource,
    waitDeleted,
    actions.RemoveCommonFinalizer(),  // MUST have this
    composed.StopAndForgetAction,
)
```

### 8. State Factory Error Ignored

**Symptom**: Nil pointer panic in actions

**Cause**: Not checking state factory error

**Fix**:
```go
// ❌ WRONG
state, _ := factory.NewState(ctx, focalState)

// ✅ CORRECT
state, err := factory.NewState(ctx, focalState)
if err != nil {
    return err, ctx
}
```

### 9. Synchronous Test Assertion

**Symptom**: Test fails intermittently

**Cause**: Not using Eventually() for reconciled state

**Fix**:
```go
// ❌ WRONG - reconciler may not have run yet
Expect(obj.Status.State).To(Equal("Ready"))

// ✅ CORRECT
Eventually(LoadAndCheck).
    WithArguments(ctx, client, obj,
        NewObjActions(),
        HavingState("Ready"),
    ).Should(Succeed())
```

### 10. Condition Not Set Correctly

**Symptom**: Conditions not appearing in status

**Cause**: Not using SetCondition helper

**Fix**:
```go
// ❌ WRONG - manual condition management
obj.Status.Conditions = append(obj.Status.Conditions, condition)

// ✅ CORRECT
state.SetCondition(
    cloudcontrolv1beta1.ConditionTypeReady,
    metav1.ConditionTrue,
    "Ready",
    "Resource is ready",
)
```

### 11. CRD Changes Not Reflected

**Symptom**: New fields not available

**Cause**: Forgot to regenerate

**Fix**:
```bash
make manifests
make generate
# For SKR resources also:
./config/patchAfterMakeManifests.sh
./config/sync.sh
```

### 12. Client Provider Not Wired

**Symptom**: Client is nil

**Cause**: Missing registration in main.go

**Fix**: Check `cmd/main.go` for proper client provider setup

### 13. Race Condition in Mock

**Symptom**: Test fails randomly

**Cause**: Missing mutex in mock

**Fix**:
```go
func (m *mock) SetState(id, state string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.resources[id].State = state
}
```

## Debugging Techniques

### Add Logging

```go
logger := composed.LoggerFromCtx(ctx)
logger.Info("Action executing",
    "resource", state.Name(),
    "status", obj.Status.State,
    "remoteExists", state.remoteResource != nil,
)
```

### Check Resource Status

```bash
kubectl get <resource> -o yaml
kubectl describe <resource>
```

### View Controller Logs

```bash
# Local
make run

# Deployed
kubectl logs -n kyma-system deployment/cloud-manager -f | grep <resource>
```

### Run Single Test

```bash
go test ./internal/controller/cloud-control -v \
    -ginkgo.focus="<TestName>" \
    -ginkgo.v
```

## Action Return Value Reference

| Scenario | Return |
|----------|--------|
| Continue to next action | `nil, ctx` |
| Unexpected error | `error, ctx` |
| Success, done | `composed.StopAndForget, nil` |
| Wait for async operation | `composed.StopWithRequeueDelay(10*time.Second), nil` |
| Check again immediately | `composed.StopWithRequeue, nil` |

## Related

- Full pitfalls: [docs/agents/reference/COMMON_PITFALLS.md](../../../docs/agents/reference/COMMON_PITFALLS.md)
- Quick reference: [docs/agents/reference/QUICK_REFERENCE.md](../../../docs/agents/reference/QUICK_REFERENCE.md)
- Testing: `/write-tests`
