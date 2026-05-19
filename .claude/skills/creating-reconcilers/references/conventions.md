# Coding Conventions Reference

Deep-dive conventions for cloud-manager reconcilers. Rules that affect code structure and safety — not covered by the composition rules in SKILL.md.

## Action Function Signature

```go
func actionName(ctx context.Context, st composed.State) (error, context.Context)
```

**CRITICAL**: Always return the context. Never return `nil, nil`.

```go
// CORRECT
return nil, ctx           // success
return err, ctx           // error
return composed.StopWithRequeue, ctx   // flow control

// WRONG — breaks the context chain for the next action
return nil, nil
```

## File Naming

| File type | Convention | Example |
|-----------|-----------|---------|
| Action | camelCase describing the operation | `shootLoad.go`, `createVolume.go` |
| State | `state.go` | — |
| Reconciler | `reconciler.go` | — |
| Type interface | `types/state.go` | — |

**One action per file.** Improves code navigation, git history readability, and merge conflict reduction.

## State Casting and Type-Safe Accessors

Cast the generic `composed.State` to your specific state type at the start of every action:

```go
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    // use state.YourFields
}
```

Define type-safe accessor methods to avoid repeated casts across actions:

```go
// In state.go
func (s *State) ObjAsVpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
    return s.Obj().(*cloudcontrolv1beta1.VpcNetwork)
}

// Usage in actions
obj := state.ObjAsVpcNetwork()
obj.Spec.SomeField  // type-safe
```

## Error Handling

Always use `LogErrorAndReturn` when returning an error with flow control — never drop the error silently:

```go
// CORRECT
if err != nil {
    return composed.LogErrorAndReturn(err, "Error loading resource", composed.StopWithRequeue, ctx)
}

// WRONG — loses error context
if err != nil {
    return composed.StopWithRequeue, ctx
}
```

Check errors immediately after the call — never defer:

```go
// CORRECT
obj, err := client.Get(ctx, name)
if err != nil {
    return composed.LogErrorAndReturn(err, "Error getting object", composed.StopWithRequeue, ctx)
}

// WRONG — deferred check may test the wrong error
obj, err := client.Get(ctx, name)
// ... other code ...
if err != nil { ... }
```

## Logger Usage

```go
logger := composed.LoggerFromCtx(ctx)
```

Add context values when entering a significant section, loading a resource used in subsequent actions, or when the values will help debug an action chain:

```go
ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
    "vpcId", state.vpcId,
    "region", state.region,
))
```

Log with structured fields:

```go
logger.Info("Resource created", "resourceId", id)
logger.Error(err, "Failed to create resource")
```

## Import Organization

Group imports in this order; goimports enforces blank-line separation:

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    // Project APIs
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

    // Project packages
    "github.com/kyma-project/cloud-manager/pkg/composed"
)
```

## Interface Compliance

Verify interface implementation at compile time by placing this line immediately after the type definition:

```go
type state struct {
    kcpcommonaction.State
    // fields...
}

var _ types.State = (*state)(nil)  // panics at compile time if state doesn't implement types.State
```

Same pattern for clients and factories:

```go
var _ Client = (*client)(nil)
var _ StateFactory = (*stateFactory)(nil)
```

## See Also

- `references/primitives.md` — Flow control errors, composition functions, util.Timing
- `references/action-pitfalls.md` — Common mistakes with GOOD/BAD examples
