# Action Composition Pattern

**Authority**: Foundational architecture  
**Prerequisite For**: All reconciler work  
**Must Read Before**: Writing any action or reconciler code

**Prerequisites**:
- MUST understand: [State Pattern](STATE_PATTERN.md)
- MUST have read: [AGENTS.md](../../../AGENTS.md)

**Skip This File If**:
- You are only reading/modifying CRD definitions (not reconciliation logic)

## Action Function Signature

```go
type Action func(ctx context.Context, state State) (error, context.Context)
```

**Location**: [pkg/composed/action.go](../../../pkg/composed/action.go)

## Rules: Action Composition

### MUST

1. MUST execute actions sequentially (never parallel)
2. MUST end successful flows with `composed.StopAndForgetAction` OR `return composed.StopAndForget, nil`
3. MUST return `nil, ctx` to continue to next action
4. MUST return `error, ctx` for failures (triggers requeue)
5. MUST place one action per file
6. MUST name actions descriptively (`loadSubnet`, not `action1`)
7. MUST check preconditions early in action
8. MUST update status on errors before returning
9. MUST use logger from context: `composed.LoggerFromCtx(ctx)`

### MUST NOT

1. MUST NOT execute actions in parallel
2. MUST NOT swallow errors (return `nil, ctx` when error occurred)
3. MUST NOT continue reconciliation after deletion complete
4. MUST NOT skip wait actions for async cloud operations
5. MUST NOT put multiple actions in single file

### ALWAYS

1. ALWAYS pass context forward (each action can modify)
2. ALWAYS handle state factory errors
3. ALWAYS verify resource state before operations

## Action Return Values

| Return | Effect | Requeue | Use When |
|--------|--------|---------|----------|
| `nil, ctx` | Continue | N/A | Normal flow |
| `error, ctx` | Stop, log | Yes (backoff) | Unexpected failure |
| `composed.StopAndForget, nil` | Stop | No | Complete success |
| `composed.StopWithRequeue, nil` | Stop | Yes (immediate) | Check again soon |
| `composed.StopWithRequeueDelay(d), nil` | Stop | Yes (after d) | Wait for async op |
| `composed.TerminalError(err), nil` | Stop | No | Unrecoverable |

**Location**: [pkg/composed/errors.go](../../../pkg/composed/errors.go)

## Action Composition Syntax

```go
composed.ComposeActions(
    "actionGroupName",
    action1,
    action2,
    action3,
)
```

**Execution**:
- Sequential order (1 → 2 → 3)
- Stops on first error
- Context passed forward

## Action Patterns

### Pattern 1: Conditional Execution

```go
composed.IfElse(
    predicate,           // func(ctx, state) bool
    actionWhenTrue,
    actionWhenFalse,
)
```

❌ **WRONG** [missing predicate]:
```go
composed.IfElse(
    nil,  // No predicate
    createResource,
    updateResource,
)
```

✅ **CORRECT**:
```go
composed.IfElse(
    composed.Not(composed.MarkedForDeletionPredicate),
    composed.ComposeActions(
        "create-update",
        createResource,
        updateStatus,
    ),
    composed.ComposeActions(
        "delete",
        deleteResource,
        removeFinalizer,
        composed.StopAndForgetAction,
    ),
)
```

### Pattern 2: Provider Switching (OLD Pattern Only)

**ONLY FOR**: Multi-provider CRDs (`RedisInstance`, `NfsInstance`, `IpRange`)  
**NEVER**: Use in NEW Pattern reconcilers

```go
composed.BuildSwitchAction(
    "providerSwitch",
    nil,  // Default action
    composed.NewCase(GcpProviderPredicate, gcpAction),
    composed.NewCase(AzureProviderPredicate, azureAction),
    composed.NewCase(AwsProviderPredicate, awsAction),
)
```

### Pattern 3: Wait-for-Ready

```go
composed.ComposeActions(
    "provision",
    createResource,           // Starts async operation
    waitResourceAvailable,    // Polls until ready
    updateStatus,             // Updates status when ready
)
```

❌ **WRONG** [missing wait action]:
```go
composed.ComposeActions(
    "provision",
    createResource,  // Returns immediately
    updateStatus,    // Status shows CREATING, not READY
)
```

✅ **CORRECT**:
```go
composed.ComposeActions(
    "provision",
    createResource,
    waitResourceReady,  // Wait for provisioning
    updateStatus,       // Now shows READY
)
```

**Wait action template**:
```go
func waitResourceReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    if state.remoteResource.State != "READY" {
        logger := composed.LoggerFromCtx(ctx)
        logger.Info("Waiting for resource to become ready")
        
        return composed.StopWithRequeueDelay(10 * time.Second), nil
    }
    
    return nil, ctx  // Ready, continue
}
```

### Pattern 4: Declarative Status Update

**Location**: [pkg/composed/updateStatus.go](../../../pkg/composed/updateStatus.go)

```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    obj := state.ObjAsTyped()
    
    return composed.UpdateStatus(obj).
        SetExclusiveConditions(metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeReady,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonReady,
            Message: "Resource is ready",
        }).
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

## Action Templates

### Create Action Template

```go
func createResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Check if already exists
    if state.remoteResource != nil {
        return nil, ctx
    }
    
    // Check prerequisites
    if state.dependency == nil {
        return fmt.Errorf("dependency must be loaded first"), ctx
    }
    
    logger.Info("Creating resource")
    
    err := state.client.Create(ctx, params)
    if err != nil {
        logger.Error(err, "Failed to create resource")
        return err, ctx
    }
    
    logger.Info("Resource created successfully")
    
    return nil, ctx
}
```

### Load Action Template

```go
func loadResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    resource, err := state.client.Get(ctx, resourceName)
    if err != nil {
        if isNotFound(err) {
            logger.Info("Resource not found")
            return nil, ctx  // Not found is OK
        }
        logger.Error(err, "Failed to load resource")
        return err, ctx
    }
    
    state.remoteResource = resource
    
    return nil, ctx
}
```

### Delete Action Template

```go
func deleteResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Check if already deleted
    if state.remoteResource == nil {
        return nil, ctx
    }
    
    logger.Info("Deleting resource")
    
    err := state.client.Delete(ctx, resourceName)
    if err != nil && !isNotFound(err) {
        logger.Error(err, "Failed to delete resource")
        return err, ctx
    }
    
    logger.Info("Resource deleted successfully")
    
    return nil, ctx
}
```

## Complete Flow Examples

### Simple Create-Update-Delete Flow

```go
composed.ComposeActions(
    "main",
    actions.AddCommonFinalizer(),
    loadRemoteResource,
    composed.IfElse(
        composed.Not(composed.MarkedForDeletionPredicate),
        composed.ComposeActions(
            "create-update",
            createResource,
            updateResource,
            updateStatus,
        ),
        composed.ComposeActions(
            "delete",
            deleteResource,
            actions.RemoveCommonFinalizer(),
            composed.StopAndForgetAction,
        ),
    ),
    composed.StopAndForgetAction,
)
```

### Complex Multi-Step Flow

```go
composed.ComposeActions(
    "gcpRedisInstance",
    actions.AddCommonFinalizer(),
    loadRedis,
    composed.IfElse(
        composed.Not(composed.MarkedForDeletionPredicate),
        composed.ComposeActions(
            "create-update",
            createRedis,
            updateStatusId,
            addUpdatingCondition,
            waitRedisAvailable,
            modifyMemorySizeGb,
            modifyReplicaCount,
            updateRedis,
            updateStatus,
        ),
        composed.ComposeActions(
            "delete",
            removeReadyCondition,
            deleteRedis,
            waitRedisDeleted,
            actions.RemoveCommonFinalizer(),
            composed.StopAndForgetAction,
        ),
    ),
    composed.StopAndForgetAction,
)
```

## Context Usage

### Feature Flags

```go
composed.ComposeActions(
    "main",
    feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.ResourceType{}),
    checkFeatureFlags,
    // ...
)

func checkFeatureFlags(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil
    }
    return nil, ctx
}
```

### Carrying Values

```go
func action1(ctx context.Context, st composed.State) (error, context.Context) {
    ctx = context.WithValue(ctx, "key", "value")
    return nil, ctx
}

func action2(ctx context.Context, st composed.State) (error, context.Context) {
    value := ctx.Value("key").(string)
    // Use value...
    return nil, ctx
}
```

## Common Pitfalls

### Pitfall 1: Missing StopAndForget

**Frequency**: Common  
**Impact**: Infinite requeue loop  
**Detection**: Reconciler keeps running after successful completion

❌ **WRONG**:
```go
composed.ComposeActions(
    "delete",
    deleteResource,
    removeFinalizer,
    // MISSING: StopAndForgetAction
)
```

✅ **CORRECT**:
```go
composed.ComposeActions(
    "delete",
    deleteResource,
    removeFinalizer,
    composed.StopAndForgetAction,
)
```

**Why It Fails**: Without explicit stop, reconciler requeues indefinitely  
**How to Fix**: Add `composed.StopAndForgetAction` at end of delete flow  
**Prevention**: Always end delete flows with `StopAndForgetAction`

### Pitfall 2: Swallowing Errors

**Frequency**: Occasional  
**Impact**: Silent failures, incorrect status  
**Detection**: Operations fail but no error in logs

❌ **WRONG**:
```go
func action(ctx context.Context, st composed.State) (error, context.Context) {
    err := doSomething()
    if err != nil {
        return nil, ctx  // Swallows error
    }
}
```

✅ **CORRECT**:
```go
func action(ctx context.Context, st composed.State) (error, context.Context) {
    err := doSomething()
    if err != nil {
        return err, ctx  // Propagates error
    }
}
```

**Why It Fails**: Returning `nil, ctx` when error occurred hides failures  
**How to Fix**: Return `error, ctx` to propagate error  
**Prevention**: Always return errors, never return `nil` when operation failed

### Pitfall 3: No Status Update on Error

**Frequency**: Common  
**Impact**: User sees stale status  
**Detection**: CR status doesn't reflect actual error state

❌ **WRONG**:
```go
func createResource(ctx context.Context, st composed.State) (error, context.Context) {
    err := state.client.Create(ctx, params)
    if err != nil {
        return err, ctx  // No status update
    }
}
```

✅ **CORRECT**:
```go
func createResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    obj := state.ObjAsTyped()
    
    err := state.client.Create(ctx, params)
    if err != nil {
        meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
            Message: err.Error(),
        })
        obj.Status.State = cloudcontrolv1beta1.StateError
        state.UpdateObjStatus(ctx)
        
        return composed.StopWithRequeueDelay(time.Minute), nil
    }
    
    return nil, ctx
}
```

**Why It Fails**: Status remains stale, user doesn't see error  
**How to Fix**: Update status before returning error  
**Prevention**: Always update status conditions on errors

### Pitfall 4: Missing Wait for Async Operations

**Frequency**: Common  
**Impact**: Status shows CREATING when operation complete  
**Detection**: Status never progresses to READY

❌ **WRONG**:
```go
composed.ComposeActions(
    "provision",
    createResource,  // Returns immediately
    updateStatus,    // Sets status to READY prematurely
)
```

✅ **CORRECT**:
```go
composed.ComposeActions(
    "provision",
    createResource,
    waitResourceReady,  // Wait for completion
    updateStatus,       // Now accurate
)
```

**Why It Fails**: Cloud operations are async, immediate status update incorrect  
**How to Fix**: Add wait action between create and status update  
**Prevention**: Always add wait actions for cloud provider operations

## Summary Checklist

Before writing actions:
- [ ] Read [State Pattern](STATE_PATTERN.md)
- [ ] Understand action return values
- [ ] Know which pattern (NEW vs OLD)

For each action:
- [ ] One action per file
- [ ] Descriptive name
- [ ] Check preconditions early
- [ ] Log operations
- [ ] Update status on errors
- [ ] Return correct error/context

For action composition:
- [ ] Sequential execution (no parallel)
- [ ] End with StopAndForget
- [ ] Add wait actions for async ops
- [ ] Handle deletion separately

## Related Documentation

**MUST READ NEXT**:
- [NEW Reconciler Pattern](RECONCILER_NEW_PATTERN.md) - Complete reconciler examples
- [State Pattern](STATE_PATTERN.md) - State hierarchy and management

**REFERENCE**:
- [Quick Reference](../reference/QUICK_REFERENCE.md) - Action return value cheat sheet
- [Common Pitfalls](../reference/COMMON_PITFALLS.md) - Troubleshooting guide
