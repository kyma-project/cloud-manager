# Action Flow Control Reference

This document explains action return values and their impact on reconciliation flow control.

## Action Signature

Every action has this signature:

```go
func actionName(ctx context.Context, st composed.State) (error, context.Context)
```

**CRITICAL**: Actions MUST NEVER return `nil, nil`. Always return the context:
- Success: `return nil, ctx`
- Error: `return err, ctx`

The composition framework uses the returned context for subsequent actions. Returning `nil` for context breaks the chain.

## Return Pattern Summary

| Pattern | Effect | Use When |
|---------|--------|----------|
| `return nil, ctx` | Continue to next action | Action succeeded |
| `return composed.StopWithRequeue, ctx` | Stop and requeue immediately | Transient failure |
| `return composed.StopWithRequeueDelay(d), ctx` | Stop and requeue after delay | Async polling |
| `return composed.StopAndForget, ctx` | Stop, no requeue | Terminal state reached |
| `return composed.LogErrorAndReturn(...)` | Log error + flow control | Error with logging |

## Continue Flow

```go
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    // Do work...
    
    // Success - continue to next action
    return nil, ctx
}
```

Use when:
- Action completed successfully
- Condition check passed
- No immediate work needed, let next action handle it

## Stop and Requeue Immediately

```go
return composed.StopWithRequeue, ctx
```

Use when:
- API call failed (transient error)
- Required resource not found but expected to appear
- Conflict during update (optimistic locking)

**Example:**
```go
err := state.Cluster().K8sClient().Get(ctx, name, obj)
if err != nil {
    return composed.LogErrorAndReturn(err, "Error loading resource", composed.StopWithRequeue, ctx)
}
```

## Stop and Requeue with Delay

```go
return composed.StopWithRequeueDelay(duration), ctx
```

### Timing Guidance

Use the `util.Timing` package for consistent delays across the codebase:

| Timing | Duration | Use For |
|--------|----------|---------|
| `util.Timing.T100ms()` | 100ms | Fast polling, immediate re-check |
| `util.Timing.T1000ms()` | 1s | Quick checks, short waits |
| `util.Timing.T10000ms()` | 10s | **Standard async polling (DEFAULT)** |
| `util.Timing.T60000ms()` | 60s | Long operations (SAP shares, backups) |
| `util.Timing.T300000ms()` | 5min | Very long operations (network deletion) |

### When to Use Each

**T100ms - T1000ms**: Resource just created/updated, checking immediate availability
```go
// Just created resource, check quickly
return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
```

**T10000ms (DEFAULT)**: Waiting for cloud provider operations
```go
// Waiting for AWS/GCP/Azure resource provisioning
if status != "available" {
    return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
```

**T60000ms**: Long-running operations or rate limiting
```go
// SAP share creation, backup operations
return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
```

**T300000ms**: Major infrastructure changes
```go
// Network deletion, large resource cleanup
return composed.StopWithRequeueDelay(util.Timing.T300000ms()), ctx
```

## Stop and Forget

```go
return composed.StopAndForget, ctx
```

Use when:
- Reconciliation reached desired terminal state
- Permanent error that shouldn't be retried
- Object is in steady state, no more work needed

**Example:**
```go
if meta.IsStatusConditionTrue(*obj.Status.Conditions, ConditionTypeReady) {
    return composed.StopAndForget, ctx
}
```

## Error Logging Pattern

Always use `LogErrorAndReturn` when logging errors with flow control:

```go
composed.LogErrorAndReturn(err, "Error message", composed.StopWithRequeue, ctx)
```

Parameters:
1. `err` - The actual error to log
2. `msg` - Human-readable error description
3. `result` - Flow control error (`StopWithRequeue`, `StopWithRequeueDelay(d)`, `StopAndForget`)
4. `ctx` - The context to return

**Example:**
```go
vpc, err := state.client.DescribeVpc(ctx, vpcId)
if err != nil {
    return composed.LogErrorAndReturn(err, "Error describing VPC", composed.StopWithRequeue, ctx)
}
```

## Decision Flowchart

```
Action completed?
├── No, error occurred
│   ├── Transient/retriable? → StopWithRequeue
│   ├── Needs time to resolve? → StopWithRequeueDelay(duration)
│   └── Permanent/terminal? → StopAndForget
│
└── Yes, success
    ├── More work to do? → return nil, ctx (continue)
    ├── Waiting for async operation? → StopWithRequeueDelay(duration)
    └── Terminal state reached? → StopAndForget
```

## Flow Control Error Interface

All flow control errors implement `FlowControlError`:

```go
type FlowControlError interface {
    error
    ShouldReturnError() bool
}
```

The `ShouldReturnError()` method determines if the error propagates to the controller-runtime result.

## Check Functions

```go
composed.IsStopAndForget(err) bool
composed.IsStopWithRequeue(err) bool
composed.IsStopWithRequeueDelay(err) bool
composed.IsBreak(err) bool
composed.IsTerminal(err) bool  // Any terminal error
```

## Example: Complete Action

```go
func waitForResourceReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Load resource
    resource, err := state.client.GetResource(ctx, state.resourceId)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error getting resource", composed.StopWithRequeue, ctx)
    }
    
    // Check status
    switch resource.Status {
    case "creating":
        logger.Info("Resource still creating, waiting...")
        return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
        
    case "available":
        logger.Info("Resource is ready")
        return nil, ctx  // Continue to next action
        
    case "failed":
        logger.Info("Resource creation failed permanently")
        return composed.StopAndForget, ctx
        
    default:
        logger.Info("Unknown status, retrying", "status", resource.Status)
        return composed.StopWithRequeue, ctx
    }
}
```

## See Also

- `references/status-mutation.md` - Status update patterns
- `references/primitives.md` - Full pkg/composed documentation
