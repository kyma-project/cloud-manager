# NEW Reconciler Pattern (Provider-Specific CRDs)

**Authority**: Foundational architecture  
**Prerequisite For**: All new KCP reconciler work  
**Must Read Before**: Creating any new KCP reconciler (post-2024)

**Prerequisites**:
- MUST understand: [State Pattern](STATE_PATTERN.md)
- MUST understand: [Action Composition](ACTION_COMPOSITION.md)
- MUST have read: [AGENTS.md](../../../AGENTS.md)

**Skip This File If**:
- You are maintaining existing multi-provider CRDs (see [OLD Pattern](RECONCILER_OLD_PATTERN.md))
- You are working on SKR reconcilers only

## Pattern Status

**Status**: ✅ **REQUIRED FOR ALL NEW CODE**  
**Introduced**: 2024-06  
**Reference Implementation**: [pkg/kcp/provider/gcp/subnet/](../../../pkg/kcp/provider/gcp/subnet/)

## Rules: NEW Pattern

### MUST USE NEW Pattern

1. MUST use for ALL new KCP reconcilers created after 2024
2. MUST use provider-specific CRD names (`GcpSubnet`, `AzureVNetLink`)
3. MUST have state directly extend `focal.State` (no intermediate layer)
4. MUST place all logic in single provider package
5. MUST use `focal.New()` in action composition
6. MUST handle state factory errors explicitly

### MUST NOT

1. MUST NOT use for maintaining existing multi-provider CRDs
2. MUST NOT create multi-provider CRDs (use provider-specific instead)
3. MUST NOT add shared intermediate state layer
4. MUST NOT use provider switching (CRD is already provider-specific)
5. MUST NOT place logic across multiple packages

### ALWAYS

1. ALWAYS embed `focal.State` directly in provider state
2. ALWAYS inject cloud API clients via client providers
3. ALWAYS implement type-specific getter (`ObjAsGcpSubnet()`)
4. ALWAYS place one action per file

## Pattern Characteristics

| Aspect | NEW Pattern |
|--------|-------------|
| CRD Naming | Provider-specific (`GcpSubnet`) |
| State Layers | 2: focal.State → Provider State |
| State Extension | Direct (no intermediate) |
| Package | Single provider package |
| Provider Switch | None (CRD is provider-specific) |
| Examples | GcpSubnet, AzureVNetLink, AzureRedisEnterprise |

## CRD Structure

```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpSubnet  # Provider-specific name
metadata:
  name: my-subnet
spec:
  scope:
    name: my-scope
  remoteRef:
    namespace: kyma-namespace
    name: user-subnet
  cidr: 10.0.0.0/24
  purpose: PRIVATE
  # Only GCP-specific fields
status:
  id: projects/my-project/regions/us-central1/subnetworks/my-subnet
  state: Ready
  conditions: [...]
```

**Location**: [api/cloud-control/v1beta1/](../../../api/cloud-control/v1beta1/)

## Directory Structure

```
pkg/kcp/provider/gcp/subnet/
├── reconcile.go                    # Reconciler
├── state.go                        # State extends focal.State
├── client/                         # Cloud API client
│   └── client.go
├── loadSubnet.go                   # Actions
├── createSubnet.go
├── deleteSubnet.go
├── updateStatus.go
├── waitCreationOperationDone.go
└── ... other actions
```

## State Implementation

**state.go**:

```go
package subnet

import (
    "context"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    "cloud.google.com/go/compute/apiv1/computepb"
)

// State extends focal.State DIRECTLY (KEY characteristic)
type State struct {
    focal.State  // No intermediate layer
    
    // Cloud provider clients
    computeClient client.ComputeClient
    
    // Remote resources
    subnet *computepb.Subnetwork
    
    // Modification tracking
    updateMask []string
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    computeClientProvider gcpclient.GcpClientProvider[client.ComputeClient]
}

func NewStateFactory(
    computeClientProvider gcpclient.GcpClientProvider[client.ComputeClient],
) StateFactory {
    return &stateFactory{
        computeClientProvider: computeClientProvider,
    }
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:         focalState,  // Embed directly
        computeClient: f.computeClientProvider(),
    }, nil
}

// Type-specific getter (REQUIRED)
func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
    return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}
```

**Key Points**:
- State embeds `focal.State` directly
- Clients injected via providers
- Type-specific getter for convenience
- No shared intermediate state

## Reconciler Structure

**reconcile.go**:

```go
package subnet

import (
    "context"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    ctrl "sigs.k8s.io/controller-runtime"
)

type GcpSubnetReconciler interface {
    reconcile.Reconciler
}

type gcpSubnetReconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory  // Provider-specific
}

func NewGcpSubnetReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    stateFactory StateFactory,
) GcpSubnetReconciler {
    return &gcpSubnetReconciler{
        composedStateFactory: composedStateFactory,
        focalStateFactory:    focalStateFactory,
        stateFactory:         stateFactory,
    }
}

func (r *gcpSubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newFocalState(req.NamespacedName)
    action := r.newAction()
    
    return composed.Handling().
        WithMetrics("gcpsubnet", util.RequestObjToString(req)).
        Handle(action(ctx, state))
}

func (r *gcpSubnetReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{}),
        focal.New(),  // Loads scope, creates focal.State
        r.newFlow(),  // Provider-specific flow
    )
}

func (r *gcpSubnetReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Create provider-specific state from focal state
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            // MUST handle state factory errors
            composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap GcpSubnet state")
            subnet := st.Obj().(*cloudcontrolv1beta1.GcpSubnet)
            subnet.Status.State = cloudcontrolv1beta1.StateError
            return composed.UpdateStatus(subnet).
                SetExclusiveConditions(metav1.Condition{
                    Type:    cloudcontrolv1beta1.ConditionTypeError,
                    Status:  metav1.ConditionTrue,
                    Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
                    Message: "Failed to create GcpSubnet state",
                }).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }
        
        // Pass provider-specific state to actions
        return composed.ComposeActions(
            "gcpSubnet",
            actions.AddCommonFinalizer(),
            loadNetwork,
            loadSubnet,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    createSubnet,
                    waitCreationOperationDone,
                    updateStatusId,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteSubnet,
                    waitDeletionOperationDone,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)  // Pass provider-specific state
    }
}

func (r *gcpSubnetReconciler) newFocalState(name types.NamespacedName) focal.State {
    return r.focalStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.GcpSubnet{}),
    )
}
```

**Flow**:
1. `Reconcile()` creates focal state
2. `newAction()` composes: feature flags → `focal.New()` → `newFlow()`
3. `newFlow()` creates provider state, composes actions
4. Actions operate on provider-specific state

## Action Templates

### Load Action

**loadSubnet.go**:

```go
package subnet

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "google.golang.org/api/googleapi"
)

func loadSubnet(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    subnet := state.ObjAsGcpSubnet()
    
    // Skip if no ID (not created yet)
    if subnet.Status.Id == "" {
        return nil, ctx
    }
    
    gcpSubnet, err := state.computeClient.GetSubnet(ctx, 
        state.Scope().Spec.Gcp.Project,
        state.Scope().Spec.Gcp.Region,
        subnet.Spec.Name,
    )
    
    if err != nil {
        // 404 is OK - resource doesn't exist
        if googleapi.IsNotFound(err) {
            return nil, ctx
        }
        return err, ctx
    }
    
    state.subnet = gcpSubnet
    logger.Info("GCP Subnet loaded", "id", subnet.Status.Id)
    
    return nil, ctx
}
```

### Create Action

**createSubnet.go**:

```go
package subnet

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
)

func createSubnet(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Skip if already exists
    if state.subnet != nil {
        return nil, ctx
    }
    
    subnet := state.ObjAsGcpSubnet()
    
    logger.Info("Creating GCP subnet")
    
    operationId, err := state.computeClient.CreateSubnet(ctx, client.CreateSubnetRequest{
        ProjectId: state.Scope().Spec.Gcp.Project,
        Region:    state.Scope().Spec.Gcp.Region,
        Name:      subnet.Spec.Name,
        Network:   state.Scope().Spec.Gcp.VpcNetwork,
        Cidr:      subnet.Spec.Cidr,
        Purpose:   string(subnet.Spec.Purpose),
    })
    
    if err != nil {
        meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
            Message: err.Error(),
        })
        subnet.Status.State = cloudcontrolv1beta1.StateError
        state.UpdateObjStatus(ctx)
        return composed.StopWithRequeueDelay(time.Minute), nil
    }
    
    subnet.Status.OpIdentifier = operationId
    subnet.Status.State = cloudcontrolv1beta1.StateProcessing
    state.UpdateObjStatus(ctx)
    
    logger.Info("GCP subnet creation initiated", "operationId", operationId)
    
    return nil, ctx
}
```

### Wait Action

**waitCreationOperationDone.go**:

```go
package subnet

import (
    "context"
    "time"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "cloud.google.com/go/compute/apiv1/computepb"
)

func waitCreationOperationDone(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    subnet := state.ObjAsGcpSubnet()
    
    // Skip if no operation
    if subnet.Status.OpIdentifier == "" {
        return nil, ctx
    }
    
    operation, err := state.computeClient.GetRegionOperation(ctx,
        state.Scope().Spec.Gcp.Project,
        state.Scope().Spec.Gcp.Region,
        subnet.Status.OpIdentifier,
    )
    if err != nil {
        return err, ctx
    }
    
    // Still pending
    if operation.GetStatus() != computepb.Operation_DONE {
        logger.Info("Waiting for subnet creation", "status", operation.GetStatus())
        return composed.StopWithRequeueDelay(10 * time.Second), nil
    }
    
    // Check for errors
    if operation.Error != nil {
        logger.Error(fmt.Errorf("operation failed"), "Operation error", "error", operation.Error)
        meta.SetStatusCondition(subnet.Conditions(), metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
            Message: fmt.Sprintf("Operation failed: %v", operation.Error),
        })
        subnet.Status.State = cloudcontrolv1beta1.StateError
        state.UpdateObjStatus(ctx)
        return composed.StopWithRequeueDelay(time.Minute), nil
    }
    
    // Clear operation ID
    subnet.Status.OpIdentifier = ""
    state.UpdateObjStatus(ctx)
    
    logger.Info("Subnet creation complete")
    
    // Reload to get full details
    return loadSubnet(ctx, state)
}
```

### Status Update Action

**updateStatus.go**:

```go
package subnet

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    subnet := state.ObjAsGcpSubnet()
    
    // Skip if not loaded
    if state.subnet == nil {
        return nil, ctx
    }
    
    subnet.Status.Id = state.subnet.GetName()
    subnet.Status.State = cloudcontrolv1beta1.StateReady
    
    return composed.UpdateStatus(subnet).
        SetExclusiveConditions(metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeReady,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonReady,
            Message: "Subnet is ready",
        }).
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

## Controller Registration

**cmd/main.go**:

```go
// Create GCP clients
gcpClients, err := gcpclient.NewGcpClients(ctx, config.GcpConfig.CredentialsFile, config.GcpConfig.PeeringCredentialsFile, rootLogger)
if err != nil {
    setupLog.Error(err, "unable to create GCP clients")
    os.Exit(1)
}

// Create state factory
gcpSubnetStateFactory := gcpsubnet.NewStateFactory(
    gcpsubnetclient.NewComputeClientProvider(gcpClients),
)

// Register reconciler
if err = gcpsubnet.NewGcpSubnetReconciler(
    composedStateFactory,
    focalStateFactory,
    gcpSubnetStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "GcpSubnet")
    os.Exit(1)
}
```

## Common Pitfalls

### Pitfall 1: Not Handling 404 Errors

**Frequency**: Common  
**Impact**: Load actions fail when resource doesn't exist yet  
**Detection**: Error logs showing 404 during initial reconciliation

❌ **WRONG**:
```go
gcpSubnet, err := state.computeClient.GetSubnet(ctx, ...)
if err != nil {
    return err, ctx  // 404 treated as error
}
```

✅ **CORRECT**:
```go
gcpSubnet, err := state.computeClient.GetSubnet(ctx, ...)
if err != nil {
    if googleapi.IsNotFound(err) {
        return nil, ctx  // 404 is OK
    }
    return err, ctx
}
```

**Why It Fails**: 404 means resource not created yet (expected)  
**How to Fix**: Check for 404, return `nil, ctx` to continue  
**Prevention**: Always handle not-found as success case

### Pitfall 2: Missing State Factory Error Handling

**Frequency**: Occasional  
**Impact**: Reconciler panics on state creation failure  
**Detection**: Panic in newFlow() function

❌ **WRONG**:
```go
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, _ := r.stateFactory.NewState(ctx, st.(focal.State))  // Ignores error
        // ...
    }
}
```

✅ **CORRECT**:
```go
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap state")
            obj := st.Obj().(*cloudcontrolv1beta1.Type)
            obj.Status.State = cloudcontrolv1beta1.StateError
            return composed.UpdateStatus(obj).
                SetExclusiveConditions(metav1.Condition{...}).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }
        // ...
    }
}
```

**Why It Fails**: State factory can fail (client init errors)  
**How to Fix**: Check error, update status, stop reconciliation  
**Prevention**: Always handle state factory errors explicitly

### Pitfall 3: Adding Intermediate State Layer

**Frequency**: Rare  
**Impact**: Violates NEW pattern, adds unnecessary complexity  
**Detection**: State has intermediate layer between focal and provider

❌ **WRONG**:
```go
type SharedState struct {
    focal.State
    // Shared fields
}

type State struct {
    SharedState  // Intermediate layer
    // Provider fields
}
```

✅ **CORRECT**:
```go
type State struct {
    focal.State  // Direct extension
    // Provider fields
}
```

**Why It Fails**: NEW pattern requires direct extension  
**How to Fix**: Remove intermediate layer, embed focal.State directly  
**Prevention**: Follow NEW pattern state template exactly

## Summary Checklist

Before creating NEW pattern reconciler:
- [ ] Read [State Pattern](STATE_PATTERN.md)
- [ ] Read [Action Composition](ACTION_COMPOSITION.md)
- [ ] Verify using provider-specific CRD name
- [ ] Understand reference implementation: `pkg/kcp/provider/gcp/subnet/`

For state implementation:
- [ ] State embeds `focal.State` directly
- [ ] Clients injected via providers
- [ ] Type-specific getter implemented
- [ ] No intermediate state layer

For reconciler implementation:
- [ ] Uses `focal.New()` in action composition
- [ ] Handles state factory errors
- [ ] Passes provider state to actions
- [ ] Actions in separate files

For actions:
- [ ] Handle 404 as success case
- [ ] Update status on errors
- [ ] Add wait actions for async operations
- [ ] Use logger from context

## Related Documentation

**MUST READ NEXT**:
- [Add KCP Reconciler Guide](../guides/ADD_KCP_RECONCILER.md) - Step-by-step creation
- [Pattern Comparison](RECONCILER_PATTERN_COMPARISON.md) - NEW vs OLD

**REFERENCE**:
- [OLD Pattern](RECONCILER_OLD_PATTERN.md) - For maintaining legacy code
- [GCP Client NEW Pattern](GCP_CLIENT_NEW_PATTERN.md) - GCP-specific details
- [Quick Reference](../reference/QUICK_REFERENCE.md) - Code templates
