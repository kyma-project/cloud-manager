# Composed Package Primitives Reference

This document provides detailed documentation of the `pkg/composed` package primitives.

## Action Type

**File**: `pkg/composed/action.go`

```go
type Action func(ctx context.Context, state State) (error, context.Context)
```

The fundamental unit of work in the composed framework. Actions:
- Receive context and state
- Return error (may be nil or a flow control error) and optionally modified context
- Can be composed sequentially or conditionally

### Built-in Actions

| Action | Purpose |
|--------|---------|
| `Noop` | Does nothing, returns nil |
| `StopAndForgetAction` | Returns `StopAndForget` error |
| `StopWithRequeueAction` | Returns `StopWithRequeue` error |
| `StopWithRequeueDelayAction(d)` | Returns `StopWithRequeueDelay(d)` |
| `LoadObj` | Loads the K8s object from cluster |

## State Interface

**File**: `pkg/composed/state.go`

```go
type State interface {
    Cluster() StateCluster
    Name() types.NamespacedName
    Obj() client.Object
    SetObj(client.Object)

    LoadObj(ctx context.Context, opts ...client.GetOption) error
    UpdateObj(ctx context.Context, opts ...client.UpdateOption) error
    UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error
    PatchObjStatus(ctx context.Context) error

    PatchObjAddFinalizer(ctx context.Context, f string) (bool, error)
    PatchObjRemoveFinalizer(ctx context.Context, f string) (bool, error)
}
```

State provides:
- Access to the Kubernetes object being reconciled
- Cluster-level resources (client, reader, event recorder)
- Methods for object lifecycle operations

### StateCluster Interface

```go
type StateCluster interface {
    K8sClient() client.Client
    ApiReader() client.Reader
    EventRecorder() events.EventRecorder
    Scheme() *runtime.Scheme
}
```

### Creating State

```go
// Create StateCluster from controller-runtime cluster
cluster := composed.NewStateClusterFromCluster(mgr)

// Create StateFactory
factory := composed.NewStateFactory(cluster)

// Create State for an object
state := factory.NewState(
    types.NamespacedName{Name: "my-obj", Namespace: "default"},
    &cloudcontrolv1beta1.MyResource{},
)
```

## StateFactory Interface

**File**: `pkg/composed/state.go`

```go
type StateFactory interface {
    NewState(name types.NamespacedName, obj client.Object) State
}
```

Factory pattern for creating State instances with a shared cluster reference.

## Composition Functions

**File**: `pkg/composed/action.go`

### IMPORTANT: Prefer Nameless Primitives

Many composition functions have two variants:
- **Nameless (PREFERRED)**: `Switch`, `ComposeActionsNoName`
- **Named (DEPRECATED)**: `BuildSwitchAction`, `ComposeActions`

The named variants accept a `name string` as the first parameter, but this parameter is **ignored** - it exists only for backward compatibility with legacy code.

**Always use the nameless variant in new code:**

| Instead of | Use |
|------------|-----|
| `ComposeActions(name, ...)` | `ComposeActionsNoName(...)` |
| `BuildSwitchAction(name, ...)` | `Switch(...)` |

### ComposeActionsNoName (PREFERRED)

```go
func ComposeActionsNoName(actions ...Action) Action
```

Composes actions into a sequential pipeline. **Always use this over `ComposeActions`**.

Execution semantics:
- Actions execute in order
- Stops on first error (unless it's a non-returning flow control error)
- Context flows through: each action receives the context from the previous
- Respects context cancellation

### ComposeActions

```go
func ComposeActions(_ string, actions ...Action) Action
```

Same as `ComposeActionsNoName` but with a name parameter. The name is currently unused but was intended for debugging/tracing. **Avoid using this in new code.**

### Execution Example

```go
composed.ComposeActionsNoName(
    action1,  // executes first
    action2,  // receives context from action1
    action3,  // receives context from action2
    // if any returns error, pipeline stops
)
```

## Predicate Type

**File**: `pkg/composed/predicate.go`

```go
type Predicate func(ctx context.Context, state State) bool
```

Boolean function for conditional branching.

### Built-in Predicates

**File**: `pkg/composed/markedForDeletionPredicate.go`

| Predicate | Returns true when |
|-----------|-------------------|
| `MarkedForDeletionPredicate` | Object has deletion timestamp |
| `NotMarkedForDeletionPredicate` | Object NOT marked for deletion |
| `ObjIsNil` | State's object is nil |

**Note**: Always use `NotMarkedForDeletionPredicate` instead of `Not(MarkedForDeletionPredicate)`.

### Predicate Combinators

```go
func Not(p Predicate) Predicate      // Logical NOT
func All(predicates ...Predicate) Predicate  // Logical AND
func Any(predicates ...Predicate) Predicate  // Logical OR
```

## Conditional Branching

**File**: `pkg/composed/predicate.go`

### If

```go
func If(condition Predicate, actions ...Action) Action
```

Executes actions only if condition is true. Returns nil error if condition is false.

**PREFERRED pattern for delete/create separation:**
```go
composed.If(composed.MarkedForDeletionPredicate, deleteFlow),
composed.If(composed.NotMarkedForDeletionPredicate, createUpdateFlow),
```

### IfElse

```go
func IfElse(condition Predicate, trueAction Action, falseAction Action) Action
```

Two-way branching. **Avoid in favor of separate `If` blocks.**

### Switch

```go
func Switch(defaultAction Action, cases ...Case) Action
```

Multi-way branching. Executes the first case whose predicate returns true. Falls back to defaultAction if none match.

### Case and NewCase

```go
type Case interface {
    Predicate(ctx context.Context, state State) bool
    Action(ctx context.Context, state State) (error, context.Context)
}

func NewCase(p Predicate, actions ...Action) Case
```

Creates a case for use with Switch.

**Example: Provider branching**
```go
providerFlow := composed.Switch(
    nil,  // no default
    composed.NewCase(kcpcommonaction.AwsProviderPredicate, awsFlow),
    composed.NewCase(kcpcommonaction.GcpProviderPredicate, gcpFlow),
    composed.NewCase(kcpcommonaction.AzureProviderPredicate, azureFlow),
    composed.NewCase(kcpcommonaction.OpenStackProviderPredicate, sapFlow),
)
```

### BuildSwitchAction (DEPRECATED)

```go
func BuildSwitchAction(name string, defaultAction Action, cases ...Case) Action
```

**DEPRECATED**: The `name` parameter is ignored. Always use `Switch()` instead.

### BreakIf

```go
func BreakIf(predicate Predicate) Action
```

Returns `Break` error if predicate is true, stopping the current composition.

## Flow Control Errors

**File**: `pkg/composed/errors.go`

### FlowControlError Interface

```go
type FlowControlError interface {
    error
    ShouldReturnError() bool
}
```

Special errors that control reconciliation flow.

### Built-in Flow Control Errors

| Error | ShouldReturnError | Behavior |
|-------|-------------------|----------|
| `StopAndForget` | true | End reconciliation, no requeue |
| `StopWithRequeue` | true | End and requeue immediately |
| `StopWithRequeueDelay(d)` | true | End and requeue after duration |
| `Break` | false | Exit current composition without error |

### Check Functions

```go
func IsStopAndForget(err error) bool
func IsStopWithRequeue(err error) bool
func IsStopWithRequeueDelay(err error) bool
func IsBreak(err error) bool
func IsTerminal(err error) bool  // Any stop error
```

### Usage in Actions

```go
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    // Check precondition
    if !ready {
        return composed.StopWithRequeueDelay(30 * time.Second), nil
    }

    // Success, continue to next action
    return nil, ctx
}
```

## Result Handling Bridge

**File**: `pkg/composed/bridge.go`

Converts composed action results to controller-runtime reconciliation results.

```go
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newState(req.NamespacedName)
    action := r.newAction()

    return composed.Handling().
        WithMetrics("myresource", util.RequestObjToString(req)).
        Handle(action(ctx, state))
}
```

### Handler Methods

| Method | Purpose |
|--------|---------|
| `Handling()` | Create new Handler |
| `WithMetrics(controller, name)` | Add metrics labels |
| `WithNoLog()` | Disable error logging |
| `Handle(err, ctx)` | Convert to ctrl.Result |

## Status Update Builder

**File**: `pkg/composed/updateStatus.go`

Fluent builder for status updates.

```go
return composed.UpdateStatus(obj).
    SetCondition(metav1.Condition{
        Type:    "Ready",
        Status:  metav1.ConditionTrue,
        Reason:  "Reconciled",
        Message: "Resource is ready",
    }).
    RemoveConditions("Error", "Updating").
    ErrorLogMessage("Error updating status").
    SuccessError(composed.StopWithRequeue).
    Run(ctx, state)
```

### Builder Methods

| Method | Purpose |
|--------|---------|
| `UpdateStatus(obj)` | Create builder for Update |
| `PatchStatus(obj)` | Create builder for Patch |
| `SetCondition(cond)` | Add/update a condition |
| `SetExclusiveConditions(conds...)` | Set conditions, remove all others |
| `RemoveConditions(types...)` | Remove conditions by type |
| `KeepConditions(types...)` | Keep only these condition types |
| `ErrorLogMessage(msg)` | Log message on error |
| `OnUpdateError(f)` | Custom error handler |
| `OnUpdateSuccess(f)` | Custom success handler |
| `SuccessError(err)` | Return this error on success |
| `Run(ctx, state)` | Execute the update |

## Logger Utilities

**File**: `pkg/composed/logger.go`

```go
// Get logger from context
logger := composed.LoggerFromCtx(ctx)

// Add logger to context
ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("key", "value"))

// Log error and return
return composed.LogErrorAndReturn(err, "Error message", composed.StopWithRequeue, ctx)
```
