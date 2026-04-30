# Coding Conventions Reference

This document covers coding conventions for cloud-manager reconcilers.

## Action Structure

### Function Signature

Every action MUST use this exact signature:

```go
func actionName(ctx context.Context, st composed.State) (error, context.Context)
```

**IMPORTANT**: Always return the context. Never return `nil, nil`.

### File Naming

- **Action files**: Use camelCase describing the operation
  - `shootLoad.go`, `createVolume.go`, `allocateIpRange.go`
- **State files**: `state.go`
- **Reconciler files**: `reconciler.go`
- **Type interface files**: `types/state.go`

### One Action Per File

Each action should be in its own file. This improves:
- Code navigation
- Git history readability
- Merge conflict reduction

## State Casting

Cast the generic `composed.State` to your specific state type at the start of actions:

```go
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    // Now use state.YourFields
}
```

### Type-Safe Accessors

Define accessor methods for type-safe object access:

```go
// In state.go
func (s *State) ObjAsVpcNetwork() *cloudcontrolv1beta1.VpcNetwork {
    return s.Obj().(*cloudcontrolv1beta1.VpcNetwork)
}
```

Usage:
```go
obj := state.ObjAsVpcNetwork()
obj.Spec.SomeField  // Type-safe access
```

## Logger Usage

### Getting the Logger

```go
logger := composed.LoggerFromCtx(ctx)
```

### Adding Context Values

```go
ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
    "vpcId", state.vpcId,
    "region", state.region,
))
```

Add context values when:
- Entering a significant section of code
- Loading a resource that will be used in subsequent actions
- Values help debug issues in this action chain

### Logging Messages

```go
logger.Info("Resource created", "resourceId", id)
logger.Error(err, "Failed to create resource")
```

## Import Organization

### Standard Grouping

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
    "github.com/kyma-project/cloud-manager/pkg/util"
)
```

### KCP-Specific Imports

```go
import (
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
)
```

### SKR-Specific Imports

```go
import (
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
)
```

## Interface Compliance

Verify interface implementation at compile time:

```go
var _ types.State = (*state)(nil)
var _ Client = (*client)(nil)
var _ StateFactory = (*stateFactory)(nil)
```

Place these after the type definition:

```go
type state struct {
    kcpcommonaction.State
    // fields...
}

var _ types.State = (*state)(nil)  // Compile-time check
```

## Comment Markers

### Delete/Create Separation

Use these exact comment markers to separate flows in reconcilers:

```go
// delete ================================================================================
composed.If(
    composed.MarkedForDeletionPredicate,
    composed.ComposeActionsNoName(
        // deletion actions
    ),
),

// create/update =========================================================================
composed.If(
    composed.NotMarkedForDeletionPredicate,
    composed.ComposeActionsNoName(
        // creation/update actions
    ),
),
```

## Composition Rules

### Use ComposeActionsNoName

**ALWAYS** use `ComposeActionsNoName`, never `ComposeActions`:

```go
// CORRECT
composed.ComposeActionsNoName(
    action1,
    action2,
    action3,
)

// WRONG - name parameter is ignored
composed.ComposeActions("myActions",
    action1,
    action2,
    action3,
)
```

### One Action Per Line

```go
// CORRECT
composed.ComposeActionsNoName(
    actions.PatchAddCommonFinalizer(),
    validateSpec,
    providerFlow,
    statusReady,
)

// WRONG - hard to read
composed.ComposeActionsNoName(action1, action2, action3)
```

### Separate If Blocks

Use separate `If` blocks instead of `IfElse`:

```go
// CORRECT
composed.If(composed.MarkedForDeletionPredicate, deleteFlow),
composed.If(composed.NotMarkedForDeletionPredicate, createUpdateFlow),

// WRONG
composed.IfElse(composed.MarkedForDeletionPredicate, deleteFlow, createUpdateFlow)
```

### Use NotMarkedForDeletionPredicate

```go
// CORRECT
composed.If(composed.NotMarkedForDeletionPredicate, flow)

// WRONG
composed.If(composed.Not(composed.MarkedForDeletionPredicate), flow)
```

## Error Handling

### Always Use LogErrorAndReturn

```go
// CORRECT
if err != nil {
    return composed.LogErrorAndReturn(err, "Error loading resource", composed.StopWithRequeue, ctx)
}

// WRONG - loses error context
if err != nil {
    return composed.StopWithRequeue, ctx
}
```

### Check Errors Immediately

```go
// CORRECT
obj, err := client.Get(ctx, name)
if err != nil {
    return composed.LogErrorAndReturn(err, "Error getting object", composed.StopWithRequeue, ctx)
}

// WRONG - deferred error check
obj, err := client.Get(ctx, name)
// ... other code ...
if err != nil {  // Might be checking wrong error
    return ...
}
```

## Placeholder Comments

When generating reconciler skeletons, use placeholder comments for unspecified actions:

```go
composed.ComposeActionsNoName(
    actions.PatchAddCommonFinalizer(),
    // validation actions placeholder
    providerFlow,
    // status update placeholder
)
```

Don't generate speculative actions - let the user specify what they need.

## See Also

- `references/flow-control.md` - Action return patterns
- `references/kcp-reconciler.md` - KCP reconciler patterns
- `references/skr-reconciler.md` - SKR reconciler patterns
