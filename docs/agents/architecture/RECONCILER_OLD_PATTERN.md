# OLD Reconciler Pattern (Multi-Provider CRDs)

**Authority**: Reference only (legacy pattern)  
**Prerequisite For**: Maintaining existing multi-provider CRDs  
**Must Read Before**: Modifying RedisInstance, NfsInstance, or IpRange reconcilers

**Prerequisites**:
- MUST understand: [State Pattern](STATE_PATTERN.md)
- MUST understand: [Action Composition](ACTION_COMPOSITION.md)
- MUST have read: [NEW Pattern](RECONCILER_NEW_PATTERN.md)

**Skip This File If**:
- You are creating new reconcilers (use [NEW Pattern](RECONCILER_NEW_PATTERN.md))
- You are not working on RedisInstance, NfsInstance, or IpRange

## Pattern Status

**Status**: ⚠️ **LEGACY - DO NOT USE FOR NEW CODE**  
**Used By**: RedisInstance, NfsInstance, IpRange (existing only)  
**Reference**: [pkg/kcp/redisinstance/](../../../pkg/kcp/redisinstance/)

## Rules: OLD Pattern

### ONLY FOR

1. ONLY maintain existing RedisInstance, NfsInstance, IpRange
2. ONLY understand this pattern when fixing bugs in legacy code
3. ONLY reference when comparing with NEW pattern

### NEVER

1. NEVER create new multi-provider CRDs
2. NEVER use three-layer state hierarchy for new code
3. NEVER add new providers to existing multi-provider CRDs
4. NEVER replicate this pattern in new reconcilers
5. NEVER use BuildSwitchAction in NEW pattern reconcilers

### IF MAINTAINING

1. IF fixing bug in RedisInstance/NfsInstance/IpRange, MUST use existing pattern
2. IF adding feature, MUST follow three-layer hierarchy
3. IF modifying shared logic, MUST update all provider implementations
4. IF changing state, MUST verify impact on all providers

## Pattern Characteristics

| Aspect | OLD Pattern |
|--------|-------------|
| CRD Naming | Multi-provider (`RedisInstance`) |
| State Layers | 3: focal → shared → provider |
| Package | Split (shared + provider) |
| Provider Switch | BuildSwitchAction required |
| Complexity | High |
| Status | Legacy only |

## CRD Structure

```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance  # Multi-provider
spec:
  scope:
    name: my-scope
  ipRange:
    name: iprange-name
  instance:
    gcp:              # GCP section
      memorySize: 5
      tier: BASIC
    aws:              # AWS section
      cacheNodeType: cache.t3.micro
    azure:            # Azure section
      sku: Basic
```

**Key Characteristic**: Single CRD with provider-specific sections

## Directory Structure

```
pkg/kcp/redisinstance/              # Shared layer
├── reconciler.go                    # Provider switching
├── state.go                         # Shared state
├── types/state.go                   # Shared interface
├── loadIpRange.go                   # Shared actions
└── ...

pkg/kcp/provider/gcp/redisinstance/  # Provider layer
├── new.go                           # Provider actions
├── state.go                         # Provider state
├── client/
├── loadRedis.go
├── createRedis.go
└── ...
```

## Three-Layer State Hierarchy

```
Layer 1: focal.State           → Scope management
           ↓
Layer 2: Shared State          → Shared domain logic (IpRange)
         types.State
           ↓
Layer 3: Provider State        → Provider-specific (GCP Redis)
```

**Location**:
- Layer 1: [pkg/common/actions/focal/](../../../pkg/common/actions/focal/)
- Layer 2: [pkg/kcp/redisinstance/types/](../../../pkg/kcp/redisinstance/types/)
- Layer 3: [pkg/kcp/provider/gcp/redisinstance/](../../../pkg/kcp/provider/gcp/redisinstance/)

## Shared State Interface

**pkg/kcp/redisinstance/types/state.go**:

```go
type State interface {
    focal.State  // Extends focal
    ObjAsRedisInstance() *v1beta1.RedisInstance
    IpRange() *v1beta1.IpRange
    SetIpRange(r *v1beta1.IpRange)
}
```

## Shared State Implementation

**pkg/kcp/redisinstance/state.go**:

```go
type state struct {
    focal.State  // Embeds focal
    ipRange *cloudcontrolv1beta1.IpRange  // Shared across providers
}

func newState(focalState focal.State) types.State {
    return &state{State: focalState}
}

func (s *state) ObjAsRedisInstance() *cloudcontrolv1beta1.RedisInstance {
    return s.Obj().(*cloudcontrolv1beta1.RedisInstance)
}

func (s *state) IpRange() *cloudcontrolv1beta1.IpRange {
    return s.ipRange
}

func (s *state) SetIpRange(r *cloudcontrolv1beta1.IpRange) {
    s.ipRange = r
}
```

## Provider State

**pkg/kcp/provider/gcp/redisinstance/state.go**:

```go
type State struct {
    types.State  // Extends shared state
    
    gcpRedisInstance  *redispb.Instance     // GCP-specific
    memorystoreClient client.MemorystoreClient
    updateMask        []string
}

type StateFactory interface {
    NewState(ctx context.Context, redisInstanceState types.State) (*State, error)
}

func (f *stateFactory) NewState(ctx context.Context, redisInstanceState types.State) (*State, error) {
    return &State{
        State:             redisInstanceState,
        memorystoreClient: f.memorystoreClientProvider(),
    }, nil
}
```

## Reconciler with Provider Switching

**pkg/kcp/redisinstance/reconciler.go**:

```go
func (r *redisInstanceReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{}),
        focal.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActions(
                "redisInstanceCommon",
                loadIpRange,  // Shared action
                composed.BuildSwitchAction(  // Switch to provider
                    "providerSwitch",
                    nil,
                    composed.NewCase(GcpProviderPredicate, 
                        gcpredisinstance.New(r.gcpStateFactory)),
                    composed.NewCase(AzureProviderPredicate, 
                        azureredisinstance.New(r.azureStateFactory)),
                    composed.NewCase(AwsProviderPredicate, 
                        awsredisinstance.New(r.awsStateFactory)),
                ),
            )(ctx, newState(st.(focal.State)))  // Create shared state
        },
    )
}
```

**Flow**:
1. Load feature flags
2. `focal.New()` creates focal state
3. Create shared state from focal state
4. Execute shared actions (e.g., loadIpRange)
5. Provider switching routes to provider-specific logic
6. Provider creates provider-specific state from shared state

## Provider Action Composition

**pkg/kcp/provider/gcp/redisinstance/new.go**:

```go
func New(stateFactory StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Convert shared state to provider state
        state, err := stateFactory.NewState(ctx, st.(types.State))
        if err != nil {
            return err, ctx
        }

        return composed.ComposeActions(
            "gcpRedisInstance",
            actions.AddCommonFinalizer(),
            loadRedis,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    createRedis,
                    waitRedisAvailable,
                    modifyMemorySizeGb,
                    updateRedis,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteRedis,
                    waitRedisDeleted,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)  // Pass provider state
    }
}
```

## Provider Predicate

**pkg/kcp/redisinstance/predicates.go**:

```go
func GcpProviderPredicate(ctx context.Context, st composed.State) bool {
    redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
    return redisInstance.Spec.Instance.Gcp != nil
}

func AzureProviderPredicate(ctx context.Context, st composed.State) bool {
    redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
    return redisInstance.Spec.Instance.Azure != nil
}

func AwsProviderPredicate(ctx context.Context, st composed.State) bool {
    redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
    return redisInstance.Spec.Instance.Aws != nil
}
```

## Common Pitfalls

### Pitfall 1: Modifying Shared State Without Updating Providers

**Frequency**: Occasional  
**Impact**: Some providers break when shared logic changes  
**Detection**: Tests fail for specific providers after shared change

❌ **WRONG**:
```go
// Modify shared state interface
type State interface {
    focal.State
    ObjAsRedisInstance() *v1beta1.RedisInstance
    NewSharedMethod() string  // Added but not implemented in providers
}
```

✅ **CORRECT**:
```go
// 1. Update shared interface
type State interface {
    focal.State
    ObjAsRedisInstance() *v1beta1.RedisInstance
    NewSharedMethod() string
}

// 2. Update shared implementation
func (s *state) NewSharedMethod() string {
    return "value"
}

// 3. Verify all provider states still compile
```

**Why It Fails**: Provider states embed shared interface, must implement all methods  
**How to Fix**: Update shared implementation and verify all providers compile  
**Prevention**: Run tests for all providers after changing shared state

### Pitfall 2: Using OLD Pattern for New Code

**Frequency**: Rare  
**Impact**: Code review rejection, technical debt  
**Detection**: Multi-provider CRD created after 2024

❌ **WRONG**:
```go
// Creating new multi-provider CRD
kind: NewResource
spec:
  instance:
    gcp: {...}
    azure: {...}
```

✅ **CORRECT**:
```go
// Create provider-specific CRDs
kind: GcpNewResource
spec:
  # Only GCP fields
  
kind: AzureNewResource
spec:
  # Only Azure fields
```

**Why It Fails**: OLD pattern is legacy, NEW pattern is required  
**How to Fix**: Use NEW pattern with provider-specific CRDs  
**Prevention**: Read [NEW Pattern](RECONCILER_NEW_PATTERN.md) before creating reconcilers

### Pitfall 3: Missing Provider Case in Switch

**Frequency**: Rare  
**Impact**: Reconciler does nothing for that provider  
**Detection**: No reconciliation happens for specific provider

❌ **WRONG**:
```go
composed.BuildSwitchAction(
    "providerSwitch",
    nil,
    composed.NewCase(GcpProviderPredicate, gcpAction),
    // Missing Azure and AWS cases
)
```

✅ **CORRECT**:
```go
composed.BuildSwitchAction(
    "providerSwitch",
    nil,
    composed.NewCase(GcpProviderPredicate, gcpAction),
    composed.NewCase(AzureProviderPredicate, azureAction),
    composed.NewCase(AwsProviderPredicate, awsAction),
)
```

**Why It Fails**: No action executes for missing provider cases  
**How to Fix**: Add case for each provider in CRD spec  
**Prevention**: Verify all provider sections have corresponding cases

## Pattern Comparison

| Feature | OLD Pattern | NEW Pattern |
|---------|-------------|-------------|
| CRD | Multi-provider | Provider-specific |
| State Layers | 3 (focal → shared → provider) | 2 (focal → provider) |
| Packages | Split (shared + provider) | Single provider |
| Switching | BuildSwitchAction | None needed |
| Complexity | High | Low |
| Versioning | Coupled | Independent |
| Use For | Legacy only | All new code |

## Summary Checklist

When maintaining OLD pattern code:
- [ ] Verify you are working on RedisInstance, NfsInstance, or IpRange
- [ ] Understand three-layer state hierarchy
- [ ] Check impact on all providers when changing shared code
- [ ] Run tests for all providers
- [ ] Update all provider predicates if adding new provider section

When creating new code:
- [ ] STOP - do not use OLD pattern
- [ ] Use [NEW Pattern](RECONCILER_NEW_PATTERN.md) instead
- [ ] Create provider-specific CRD
- [ ] Use two-layer state hierarchy

## Related Documentation

**MUST READ NEXT**:
- [NEW Pattern](RECONCILER_NEW_PATTERN.md) - Required for all new code
- [Pattern Comparison](RECONCILER_PATTERN_COMPARISON.md) - Side-by-side comparison

**REFERENCE**:
- [State Pattern](STATE_PATTERN.md) - State hierarchy details
- [Action Composition](ACTION_COMPOSITION.md) - BuildSwitchAction details
