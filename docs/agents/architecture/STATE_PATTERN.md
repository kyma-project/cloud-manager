# State Pattern - Three-Layer Hierarchy

**Authority**: Architecture foundation  
**Prerequisite For**: All reconciler work  
**Must Read Before**: Writing any reconciler code

## MUST UNDERSTAND: State Hierarchy

```
Layer 1: composed.State  → Base K8s operations
         ↓ embeds
Layer 2: focal.State     → Scope management (cloud provider context)
         ↓ embeds
Layer 3: Provider State  → Cloud API clients + remote resources
```

## Layer 1: composed.State

**Location**: `pkg/composed/state.go`

**Provides**:
- `Obj()` - Kubernetes object being reconciled
- `K8sClient()` - Kubernetes client for API operations
- `Name()` - NamespacedName of object
- `UpdateObj(ctx)` - Update spec/metadata
- `UpdateObjStatus(ctx)` - Update status (ALWAYS use for status)
- `Cluster()` - Cluster info (client, scheme, event recorder)

**Use Directly**: NEVER. Always use focal.State or provider-specific state.

## Layer 2: focal.State

**Location**: `pkg/common/actions/focal/state.go`

**Extends**: composed.State

**Adds**:
- `Scope()` - Cloud provider scope (project/subscription/account)
- `SetScope(scope)` - Set scope resource
- `ObjAsCommonObj()` - Access common object interface

**Scope Resource Definition**:
```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: Scope
spec:
  provider: gcp  # or aws, azure
  gcp:
    project: my-project-id
    vpcNetwork: my-vpc
```

**Created By**: `focal.New()` action

**Use When**:
- All KCP reconcilers
- Need cloud provider project/account context
- Provider-agnostic scope operations

## Layer 3: Provider-Specific State

**Location**: `pkg/kcp/provider/<provider>/<resource>/state.go`

**Extends**: focal.State (NEW pattern) OR shared.State (OLD pattern)

### NEW Pattern (REQUIRED for new code)

```go
type State struct {
    focal.State  // Direct extension - NO intermediate layer
    
    // Cloud API clients
    computeClient client.ComputeClient
    
    // Remote cloud resources
    subnet *computepb.Subnetwork
    
    // Operation tracking
    updateMask []string
}

func (s *State) ObjAsGcpSubnet() *v1beta1.GcpSubnet {
    return s.Obj().(*v1beta1.GcpSubnet)
}
```

**State Factory**:
```go
type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    computeClient := f.computeClientProvider()
    
    return &State{
        State:         focalState,  // Embed focal.State
        computeClient: computeClient,
    }, nil
}
```

**Examples**: `pkg/kcp/provider/gcp/subnet/`, `pkg/kcp/provider/azure/redisenterprise/`

### OLD Pattern (FORBIDDEN for new code)

```go
// Shared state interface (pkg/kcp/redisinstance/types/state.go)
type State interface {
    focal.State
    ObjAsRedisInstance() *v1beta1.RedisInstance
    IpRange() *v1beta1.IpRange
    SetIpRange(*v1beta1.IpRange)
}

// Provider state (pkg/kcp/provider/gcp/redisinstance/state.go)
type State struct {
    types.State  // Extends shared state (adds extra layer)
    
    gcpRedisInstance  *redispb.Instance
    memorystoreClient client.MemorystoreClient
}
```

**DO NOT USE**: Three-layer hierarchy (composed → focal → shared → provider). Only maintain existing code.

## State Creation Flow

```go
// Reconciler entry point
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        focal.New(),  // Step 1: Create focal.State from composed.State
        r.newFlow(),  // Step 2: Create provider-specific state
    )
}

// Provider flow
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // MUST check error
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            logger := composed.LoggerFromCtx(ctx)
            logger.Error(err, "Failed to create state")
            return composed.UpdateStatus(st.Obj()).
                SetConditionError(cloudcontrolv1beta1.ConditionTypeError, "StateInitFailed", err.Error()).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }
        
        // Pass provider state to actions
        return composed.ComposeActions(
            "flow",
            loadResource,
            createResource,
            updateStatus,
        )(ctx, state)  // provider-specific state
    }
}
```

## State Usage in Actions

**Type Assertion**:
```go
func createSubnet(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)  // Assert to provider-specific type
    
    obj := state.ObjAsGcpSubnet()  // Type-safe getter
    
    // Use state fields
    err := state.computeClient.CreateSubnet(ctx, ...)
    
    return nil, ctx
}
```

**REQUIRED Patterns**:
- ALWAYS assert to correct state type
- ALWAYS use type-safe getters (`ObjAsGcpSubnet()` not raw casting)
- NEVER assume state type without checking

## State Factory Pattern

**Purpose**: Dependency injection for testing and client management

**NEW Pattern Factory**:
```go
type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    clientProvider gcpclient.GcpClientProvider[client.ComputeClient]
}

func NewStateFactory(clientProvider gcpclient.GcpClientProvider[client.ComputeClient]) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    client := f.clientProvider()  // Get cloud API client
    
    if client == nil {
        return nil, fmt.Errorf("client provider returned nil")
    }
    
    return &State{
        State:  focalState,
        client: client,
    }, nil
}
```

**Factory Registration** (in `internal/controller/setup.go`):
```go
stateFactory := subnet.NewStateFactory(
    gcpclient.NewCachedGcpClientProvider(
        gcpclient.NewComputeClientProvider(),
    ),
)

reconciler := subnet.NewReconciler(mgr.GetClient(), stateFactory)
reconciler.SetupWithManager(mgr)
```

## State Method Patterns

**Type-Safe Getters** (REQUIRED):
```go
func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
    return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}
```

**Resource Accessors**:
```go
func (s *State) IpRange() *cloudcontrolv1beta1.IpRange {
    return s.ipRange
}

func (s *State) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
    s.ipRange = r
}
```

**Update Mask Tracking** (GCP pattern):
```go
func (s *State) SetMemorySizeGb(size int32) {
    s.updateMask = append(s.updateMask, "memory_size_gb")
    s.gcpRedisInstance.MemorySizeGb = size
}
```

## Rules: State Management

### MUST

1. MUST embed focal.State in provider state (NEW pattern)
2. MUST handle state factory errors
3. MUST use type-safe getters for CRD access
4. MUST assert to correct state type in actions
5. MUST create new state per reconciliation

### MUST NOT

1. MUST NOT use shared intermediate state layer (OLD pattern)
2. MUST NOT ignore state factory errors
3. MUST NOT use raw type assertions without type-safe getters
4. MUST NOT cache state across reconciliations
5. MUST NOT mutate state concurrently

## Common Pitfalls

### Pitfall 1: Wrong State Type Assertion

❌ **WRONG**:
```go
func action(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*focal.State)  // Wrong type
    subnet := state.Obj().(*v1beta1.GcpSubnet)  // Fails
}
```

✅ **CORRECT**:
```go
func action(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)  // Provider-specific state
    subnet := state.ObjAsGcpSubnet()  // Type-safe getter
}
```

### Pitfall 2: Not Handling State Factory Error

❌ **WRONG**:
```go
state, _ := r.stateFactory.NewState(ctx, st.(focal.State))  // Ignoring error
return composed.ComposeActions(...)(ctx, state)  // May panic
```

✅ **CORRECT**:
```go
state, err := r.stateFactory.NewState(ctx, st.(focal.State))
if err != nil {
    logger := composed.LoggerFromCtx(ctx)
    logger.Error(err, "State creation failed")
    return composed.UpdateStatus(st.Obj()).
        SetConditionError(v1beta1.ConditionTypeError, "StateInitFailed", err.Error()).
        SuccessError(composed.StopAndForget).
        Run(ctx, st)
}
return composed.ComposeActions(...)(ctx, state)
```

### Pitfall 3: Not Embedding State

❌ **WRONG**:
```go
type State struct {
    FocalState focal.State  // Field, not embedding
}
```

✅ **CORRECT**:
```go
type State struct {
    focal.State  // Embedding (anonymous field)
    // ... other fields
}
```

### Pitfall 4: Passing Wrong State Type

❌ **WRONG**:
```go
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        return composed.ComposeActions(
            "flow",
            providerAction,  // Expects provider state
        )(ctx, st)  // Passing focal.State - WRONG
    }
}
```

✅ **CORRECT**:
```go
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            return err, ctx
        }
        return composed.ComposeActions(
            "flow",
            providerAction,  // Expects provider state
        )(ctx, state)  // Passing provider state - CORRECT
    }
}
```

## State Lifecycle

**Per-Reconciliation**:
1. composed.State created with K8s object
2. focal.New() wraps with scope
3. StateFactory creates provider state with clients
4. Actions execute with provider state
5. State discarded after reconciliation

**NO caching**: Each reconciliation creates fresh state

**Thread-safe**: No shared mutable state

## Summary Checklist

Before writing reconciler:
- [ ] Understand three-layer hierarchy
- [ ] Know which pattern to use (NEW for new code)
- [ ] Understand state factory pattern
- [ ] Know how to handle state creation errors
- [ ] Know how to assert state types in actions

For NEW pattern:
- [ ] Provider state embeds focal.State directly
- [ ] State factory takes focal.State parameter
- [ ] Type-safe getters implemented
- [ ] No shared intermediate state layer

For maintaining OLD pattern:
- [ ] Understand shared state layer
- [ ] State factory takes shared state parameter
- [ ] Provider state embeds shared state
- [ ] DO NOT replicate this pattern

## Related Documentation

**MUST READ NEXT**:
- [ACTION_COMPOSITION.md](ACTION_COMPOSITION.md) - How actions use state
- [RECONCILER_NEW_PATTERN.md](RECONCILER_NEW_PATTERN.md) - Complete reconciler examples
- [RECONCILER_PATTERN_COMPARISON.md](RECONCILER_PATTERN_COMPARISON.md) - NEW vs OLD patterns

**REFERENCE**:
- [QUICK_REFERENCE.md](../reference/QUICK_REFERENCE.md) - State method cheat sheet
- [COMMON_PITFALLS.md](../reference/COMMON_PITFALLS.md) - State-related errors

---

**Last Updated**: 2025-01-24  
**Pattern Status**: Fundamental architecture  
**Agent Compliance**: MANDATORY
