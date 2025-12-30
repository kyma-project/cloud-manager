# Reconciler Pattern Comparison

**Authority**: Reference  
**Prerequisite For**: Pattern selection decisions  
**Must Read Before**: Choosing between patterns or understanding legacy code

**Prerequisites**:
- MUST have read: [NEW Pattern](RECONCILER_NEW_PATTERN.md)
- MUST have read: [OLD Pattern](RECONCILER_OLD_PATTERN.md)
- MUST understand: [State Pattern](STATE_PATTERN.md)

**Skip This File If**:
- You already know you are creating new code (use NEW Pattern)
- You are only reading this as reference after using other docs

## Pattern Selection Decision Tree

```
Are you creating NEW code (after 2024)?
├─ YES → MUST use NEW Pattern
│  └─ Provider-specific CRD (GcpSubnet)
│
└─ NO → Are you maintaining existing code?
   ├─ RedisInstance, NfsInstance, IpRange
   │  └─ Use OLD Pattern (legacy only)
   │
   └─ Other resources
      └─ MUST use NEW Pattern
```

## Side-by-Side Comparison

| Aspect | NEW Pattern | OLD Pattern |
|--------|-------------|-------------|
| CRD Naming | Provider-specific (`GcpSubnet`) | Multi-provider (`RedisInstance`) |
| State Layers | 2 (focal → provider) | 3 (focal → shared → provider) |
| State Extension | Direct from focal | Via shared intermediate |
| Package | Single provider | Shared + provider split |
| Reconciler | In provider package | In shared package |
| Provider Switch | None (CRD is specific) | BuildSwitchAction required |
| Complexity | Low | High |
| Versioning | Independent per provider | Coupled across providers |
| Status | **Required for new code** | **Legacy only** |
| Examples | GcpSubnet, AzureVNetLink | RedisInstance, NfsInstance, IpRange |

## Directory Structure Comparison

### NEW Pattern

```
pkg/kcp/provider/gcp/subnet/
├── reconcile.go          # Reconciler
├── state.go              # Extends focal.State
├── client/
├── loadSubnet.go
├── createSubnet.go
├── deleteSubnet.go
└── updateStatus.go
```

**State**: `composed.State → focal.State → GcpSubnet.State`

### OLD Pattern

```
pkg/kcp/redisinstance/              # Shared
├── reconciler.go                    # Provider switching
├── state.go                         # Shared state
├── types/state.go                   # Interface
├── loadIpRange.go
└── ...

pkg/kcp/provider/gcp/redisinstance/  # Provider
├── new.go                           # Actions
├── state.go                         # Extends shared
├── client/
├── loadRedis.go
└── ...
```

**State**: `composed.State → focal.State → redisinstance.State → GcpRedisInstance.State`

## CRD Comparison

### NEW Pattern CRD

```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpSubnet  # Provider in name
spec:
  scope: {...}
  cidr: 10.0.0.0/24
  purpose: PRIVATE
  # Only GCP fields
```

**Characteristics**:
- Provider in name
- Only provider-specific fields
- Independent versioning

### OLD Pattern CRD

```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance  # Generic name
spec:
  scope: {...}
  ipRange: {...}
  instance:
    gcp:          # GCP section
      memorySize: 5
    aws:          # AWS section
      cacheNodeType: cache.t3.micro
    azure:        # Azure section
      sku: Basic
```

**Characteristics**:
- Generic name
- Multiple provider sections
- Coupled versioning

## Reconciler Flow Comparison

### NEW Pattern Flow

```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{}),
        focal.New(),  // focal.State
        r.newFlow(),  // Direct to provider
    )
}

func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // focal.State → Provider.State (direct)
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        
        return composed.ComposeActions(
            "gcpSubnet",
            loadSubnet,
            createSubnet,
            updateStatus,
        )(ctx, state)
    }
}
```

**Flow**: focal.State → Provider.State (1 step)

### OLD Pattern Flow

```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{}),
        focal.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActions(
                "common",
                loadIpRange,  // Shared action
                composed.BuildSwitchAction(  // Switch to provider
                    "switch",
                    nil,
                    composed.NewCase(GcpPredicate, gcpredisinstance.New(factory)),
                    composed.NewCase(AzurePredicate, azureredisinstance.New(factory)),
                ),
            )(ctx, newState(st.(focal.State)))  // focal → shared
        },
    )
}

// Provider package
func New(factory StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // shared.State → Provider.State
        state, err := factory.NewState(ctx, st.(types.State))
        
        return composed.ComposeActions(
            "gcpRedis",
            loadRedis,
            createRedis,
        )(ctx, state)
    }
}
```

**Flow**: focal.State → shared.State → Provider.State (2 steps)

## Pattern Usage Rules

| Situation | Pattern | Rule |
|-----------|---------|------|
| Creating new resource | NEW | REQUIRED |
| Created after 2024 | NEW | REQUIRED |
| Provider-specific features | NEW | REQUIRED |
| Maintaining RedisInstance | OLD | Understand only |
| Maintaining NfsInstance | OLD | Understand only |
| Maintaining IpRange | OLD | Understand only |
| Adding new provider to existing multi-provider CRD | OLD | Ask for approval |
| Migrating OLD to NEW | - | Not recommended |

## Benefits Comparison

| Benefit | NEW Pattern | OLD Pattern |
|---------|-------------|-------------|
| Simple state hierarchy | ✅ (2 layers) | ❌ (3 layers) |
| Easy to understand | ✅ | ❌ |
| Independent versioning | ✅ | ❌ |
| Clear package structure | ✅ | ❌ |
| No provider switching | ✅ | ❌ |
| Easy testing | ✅ | ❌ |
| Maintainable | ✅ | ❌ |

## Code Examples Comparison

### State Definition

**NEW Pattern**:
```go
type State struct {
    focal.State  // Direct
    computeClient client.ComputeClient
    subnet *computepb.Subnetwork
}
```

**OLD Pattern**:
```go
// Shared state
type State interface {
    focal.State
    IpRange() *v1beta1.IpRange
}

// Provider state
type State struct {
    types.State  // Via shared
    memorystoreClient client.MemorystoreClient
    gcpRedisInstance *redispb.Instance
}
```

### State Factory

**NEW Pattern**:
```go
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:         focalState,  // Direct embed
        computeClient: f.clientProvider(),
    }, nil
}
```

**OLD Pattern**:
```go
// Shared factory
func newState(focalState focal.State) types.State {
    return &state{State: focalState}
}

// Provider factory
func (f *stateFactory) NewState(ctx context.Context, sharedState types.State) (*State, error) {
    return &State{
        State:             sharedState,  // Via shared
        memorystoreClient: f.clientProvider(),
    }, nil
}
```

## Migration Considerations

**IF** migrating OLD to NEW (rare):

1. Create provider-specific CRDs
2. Remove shared state layer
3. Remove provider switching
4. Move actions to provider package
5. Update tests

**HOWEVER**: Migration not recommended for stable resources. Use NEW for new resources.

## Common Mistakes

### Mistake 1: Using OLD Pattern for New Code

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
spec: {...}

kind: AzureNewResource
spec: {...}
```

### Mistake 2: Adding Shared Layer to NEW Pattern

❌ **WRONG**:
```go
type SharedState struct {
    focal.State
}

type State struct {
    SharedState  // Violates NEW pattern
}
```

✅ **CORRECT**:
```go
type State struct {
    focal.State  // Direct extension
}
```

### Mistake 3: Using NEW Pattern for OLD Resources

❌ **WRONG**:
```go
// Refactoring RedisInstance to NEW pattern
// High risk, not recommended
```

✅ **CORRECT**:
```go
// Maintain RedisInstance with OLD pattern
// Use NEW pattern for new resources only
```

## Summary Checklist

When choosing pattern:
- [ ] Creating new code? → NEW Pattern
- [ ] After 2024? → NEW Pattern
- [ ] Maintaining RedisInstance/NfsInstance/IpRange? → OLD Pattern (understand)
- [ ] In doubt? → NEW Pattern

For NEW Pattern:
- [ ] Provider-specific CRD name
- [ ] 2-layer state (focal → provider)
- [ ] Single provider package
- [ ] No provider switching

For OLD Pattern:
- [ ] Multi-provider CRD
- [ ] 3-layer state (focal → shared → provider)
- [ ] Shared + provider packages
- [ ] BuildSwitchAction required

## Related Documentation

**MUST READ NEXT**:
- [NEW Pattern](RECONCILER_NEW_PATTERN.md) - For creating new code
- [OLD Pattern](RECONCILER_OLD_PATTERN.md) - For maintaining legacy code

**REFERENCE**:
- [State Pattern](STATE_PATTERN.md) - State hierarchy details
- [Add KCP Reconciler Guide](../guides/ADD_KCP_RECONCILER.md) - Step-by-step
