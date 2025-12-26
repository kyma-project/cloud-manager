# Guide: CRD Architecture Evolution

**Target Audience**: LLM coding agents  
**Prerequisites**: Understanding reconciler patterns, state hierarchies  
**Purpose**: Explain OLD vs NEW CRD patterns, why evolution happened, when to use each  
**Context**: Cloud Manager evolved from multi-provider CRDs to provider-specific CRDs for better maintainability

## Authority: CRD Pattern Requirements

### MUST

- **MUST use NEW Pattern**: ALL new code MUST use provider-specific CRDs (GcpRedisCluster style)
- **MUST understand both patterns**: Agents maintaining code must recognize OLD vs NEW
- **MUST never add to OLD**: NEVER add features to multi-provider CRDs (RedisInstance, NfsInstance, IpRange)
- **MUST create new CRDs**: For new capabilities, create provider-specific CRD, don't extend multi-provider
- **MUST use two-layer state**: NEW Pattern extends focal.State directly, no intermediate layer

### MUST NOT

- **NEVER use OLD Pattern for new code**: Multi-provider CRDs FORBIDDEN for development after 2024
- **NEVER add provider sections**: Don't create CRDs with `gcp:`, `aws:`, `azure:` sections
- **NEVER use provider switching**: NEW Pattern has direct reconciliation, no `BuildSwitchAction`
- **NEVER create three-layer state**: NEW Pattern uses composed → focal → provider (2 layers to focal)
- **NEVER modify multi-provider CRDs**: Only bug fixes, no new features

### ALWAYS

- **ALWAYS use provider-specific names**: `GcpRedisCluster`, `AzureRedisEnterprise` not `RedisCluster`
- **ALWAYS extend focal.State directly**: NEW Pattern state extends focal.State, not intermediate shared state
- **ALWAYS one reconciler per CRD**: Each provider-specific CRD has its own reconciler

### NEVER

- **NEVER mix patterns**: Don't add NEW Pattern logic to OLD Pattern resources
- **NEVER migrate existing CRDs**: Keep RedisInstance/NfsInstance/IpRange as-is for backward compatibility
- **NEVER assume multi-provider**: Default to provider-specific CRDs unless maintaining legacy

## Evolution Overview

**OLD Pattern** (Legacy): Single CRD for all providers, three-layer state, provider switching  
**NEW Pattern** (Current): Provider-specific CRDs, two-layer state, direct reconciliation

## OLD Pattern: Multi-Provider CRDs

### CRD Structure

#### ❌ WRONG (But Understanding Required for Maintenance)

```yaml
# One CRD for all providers
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance  # Single CRD
spec:
  scope:
    name: scope-name
  ipRange:
    name: iprange-name
  instance:
    gcp:              # GCP-specific section
      memorySize: 5
      tier: BASIC
      redisVersion: "7.0"
    aws:              # AWS-specific section
      cacheNodeType: cache.t3.micro
      engineVersion: "7.0"
    azure:            # Azure-specific section
      sku: Basic
      capacity: 1
```

### State Hierarchy (Three Layers)

```
┌─────────────────────────────┐
│   composed.State            │  ← Layer 1: Base K8s operations
│   (pkg/composed/state.go)   │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   focal.State               │  ← Layer 2: Scope management
│   (pkg/common/actions/      │
│    focal/state.go)          │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   Shared State (types.State)│  ← Layer 3: Shared domain logic
│   (pkg/kcp/redisinstance/   │     (e.g., IpRange loading)
│    types/state.go)          │
│   + ipRange                 │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   Provider State            │  ← Layer 4: Provider-specific
│   (pkg/kcp/provider/gcp/    │     resources and clients
│    redisinstance/state.go)  │
│   + gcpRedisInstance        │
│   + memorystoreClient       │
└─────────────────────────────┘
```

### Directory Structure

```
pkg/kcp/redisinstance/              # SHARED LAYER
├── reconciler.go                    # Main reconciler with provider switching
├── state.go                         # Shared state (extends focal.State)
├── types/
│   └── state.go                     # Shared state interface
├── loadIpRange.go                   # Shared actions
└── ...

pkg/kcp/provider/gcp/redisinstance/  # PROVIDER-SPECIFIC LAYER
├── new.go                           # Provider action composition
├── state.go                         # Provider state (extends shared state)
├── client/
│   └── client.go
├── loadRedis.go
├── createRedis.go
└── ...
```

### Provider Switching Reconciler

```go
// pkg/kcp/redisinstance/reconciler.go
func (r *redisInstanceReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{}),
        focal.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActions(
                "redisInstanceCommon",
                loadIpRange,  // Shared action
                composed.BuildSwitchAction(  // Provider switching
                    "providerSwitch",
                    nil,
                    composed.NewCase(
                        statewithscope.GcpProviderPredicate,
                        gcpredisinstance.New(r.gcpStateFactory),
                    ),
                    composed.NewCase(
                        statewithscope.AzureProviderPredicate,
                        azureredisinstance.New(r.azureStateFactory),
                    ),
                    composed.NewCase(
                        statewithscope.AwsProviderPredicate,
                        awsredisinstance.New(r.awsStateFactory),
                    ),
                ),
            )(ctx, newState(st.(focal.State)))
        },
    )
}
```

### Problems with OLD Pattern

#### Problem 1: Versioning Lockstep

Cannot version provider fields independently:

```yaml
# WRONG: Can't have different API versions per provider
kind: RedisInstance
spec:
  instance:
    gcp:
      # v1 fields
      tier: BASIC
    aws:
      # Want v2 fields, but requires CRD v2 for ALL providers!
      cacheNodeType: cache.t3.micro
```

#### Problem 2: Field Bloat

CRD contains ALL provider fields:

```go
type RedisInstanceInfo struct {
    Gcp   *RedisInstanceGcp   `json:"gcp,omitempty"`    // 20+ fields
    Aws   *RedisInstanceAws   `json:"aws,omitempty"`    // 15+ fields
    Azure *RedisInstanceAzure `json:"azure,omitempty"`  // 18+ fields
}
// Total: 50+ fields, user only uses ~15
```

#### Problem 3: Validation Complexity

```go
// Must validate exactly one provider set
func (r *RedisInstance) ValidateCreate() error {
    count := 0
    if r.Spec.Instance.Gcp != nil {
        count++
    }
    if r.Spec.Instance.Aws != nil {
        count++
    }
    if r.Spec.Instance.Azure != nil {
        count++
    }
    if count != 1 {
        return errors.New("exactly one provider must be specified")
    }
    return nil
}
```

#### Problem 4: Provider Coupling

Adding new provider feature requires:
- Modifying shared CRD
- Updating all provider reconcilers
- Testing all provider combinations
- Coordinating changes across teams

## NEW Pattern: Provider-Specific CRDs

### CRD Structure

#### ✅ CORRECT: Provider-Specific CRDs

```yaml
# GCP
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpRedisInstance  # Provider-specific
spec:
  scope:
    name: scope-name
  memorySize: 5
  tier: BASIC
  redisVersion: "7.0"
  # Only GCP fields

---
# AWS
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: AwsRedisInstance  # Provider-specific
spec:
  scope:
    name: scope-name
  cacheNodeType: cache.t3.micro
  engineVersion: "7.0"
  # Only AWS fields

---
# Azure
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: AzureRedisInstance  # Provider-specific
spec:
  scope:
    name: scope-name
  sku: Basic
  capacity: 1
  # Only Azure fields
```

### State Hierarchy (Two Layers to focal.State)

```
┌─────────────────────────────┐
│   composed.State            │  ← Base K8s operations
│   (pkg/composed/state.go)   │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   focal.State               │  ← Scope management
│   (pkg/common/actions/      │
│    focal/state.go)          │
└──────────────┬──────────────┘
               │ embeds (DIRECTLY)
               ↓
┌─────────────────────────────┐
│   Provider State            │  ← Provider-specific
│   (pkg/kcp/provider/gcp/    │     (No intermediate layer!)
│    redisinstance/state.go)  │
│   + gcpRedisInstance        │
│   + memorystoreClient       │
└─────────────────────────────┘
```

### Directory Structure

```
pkg/kcp/provider/gcp/redisinstance/  # SINGLE PROVIDER PACKAGE
├── reconcile.go                      # Reconciler with direct composition
├── state.go                          # State extends focal.State directly
├── client/
│   └── client.go                    # GCP-specific client
├── loadRedis.go
├── createRedis.go
├── updateRedis.go
├── deleteRedis.go
└── updateStatus.go
```

### Direct Reconciler (No Switching)

#### ❌ WRONG: Provider Switching in NEW Pattern

```go
// NEVER: Provider switching in NEW Pattern
func (r *reconciler) newFlow() composed.Action {
    return composed.BuildSwitchAction(  // WRONG: No switching needed!
        "providerSwitch",
        nil,
        composed.NewCase(GcpProviderPredicate, gcpAction),
    )
}
```

#### ✅ CORRECT: Direct Reconciliation

```go
// ALWAYS: Direct reconciliation in NEW Pattern
func (r *gcpRedisInstanceReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Create provider-specific state from focal state
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            return composed.LogErrorAndReturn(err, 
                "Error creating state", 
                composed.StopWithRequeue, ctx)
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
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteRedis,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)  // Pass provider-specific state directly
    }
}
```

### State Implementation Comparison

#### ❌ OLD Pattern: Three-Layer State

```go
// WRONG for new code: Intermediate shared state layer
// pkg/kcp/redisinstance/types/state.go
type State interface {
    focal.State
    
    // Shared fields across all providers
    IpRange() composed.PtrState[*cloudcontrolv1beta1.IpRange]
}

// pkg/kcp/provider/gcp/redisinstance/state.go
type State struct {
    types.State  // Extends SHARED state
    
    gcpRedisInstance  *redispb.Instance
    memorystoreClient client.MemorystoreClient
}
```

#### ✅ NEW Pattern: Direct Extension

```go
// CORRECT: State directly extends focal.State
// pkg/kcp/provider/gcp/redisinstance/state.go
type State struct {
    focal.State  // DIRECTLY extends focal.State
    
    gcpRedisInstance  *redispb.Instance
    memorystoreClient client.MemorystoreClient
}

// Simple state factory
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:             focalState,
        memorystoreClient: f.memorystoreClientProvider(),
    }, nil
}
```

### Benefits of NEW Pattern

| Benefit | Explanation |
|---------|-------------|
| **Independent Versioning** | GCP can be v1beta2, AWS stays v1beta1 |
| **Cleaner CRDs** | Only relevant fields per provider |
| **Simpler Validation** | Provider-specific rules only |
| **Easier Maintenance** | Changes affect one provider only |
| **Better Features** | Easy to add provider-unique capabilities |
| **Simpler State** | No intermediate shared layer |
| **Faster Development** | Teams work independently |

## Pattern Recognition

### How to Identify OLD Pattern

**Signs**:
```go
// 1. CRD has provider sections
type RedisInstanceInfo struct {
    Gcp   *RedisInstanceGcp   `json:"gcp,omitempty"`
    Aws   *RedisInstanceAws   `json:"aws,omitempty"`
    Azure *RedisInstanceAzure `json:"azure,omitempty"`
}

// 2. Shared state package
pkg/kcp/redisinstance/types/

// 3. Provider switching
composed.BuildSwitchAction(
    "providerSwitch",
    composed.NewCase(GcpProviderPredicate, gcpAction),
    composed.NewCase(AwsProviderPredicate, awsAction),
)

// 4. Three-layer state
type State struct {
    types.State  // Shared layer
    // provider fields
}
```

### How to Identify NEW Pattern

**Signs**:
```go
// 1. Provider-specific CRD name
type GcpRedisCluster struct {
    // Only GCP fields
}

// 2. State directly extends focal
type State struct {
    focal.State  // Direct extension
    // provider fields
}

// 3. No provider switching
func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, _ := r.stateFactory.NewState(ctx, st.(focal.State))
        return composed.ComposeActions(...)(ctx, state)
    }
}

// 4. Provider package only
pkg/kcp/provider/gcp/rediscluster/
```

## Migration Strategy

### For New Resources

#### ❌ WRONG: Creating Multi-Provider CRD

```go
// NEVER: Creating new multi-provider CRD
type RedisCluster struct {
    Spec RedisClusterSpec `json:"spec"`
}

type RedisClusterSpec struct {
    Cluster RedisClusterInfo `json:"cluster"`
}

type RedisClusterInfo struct {
    Gcp   *RedisClusterGcp   `json:"gcp,omitempty"`  // WRONG!
    Aws   *RedisClusterAws   `json:"aws,omitempty"`  // WRONG!
    Azure *RedisClusterAzure `json:"azure,omitempty"`  // WRONG!
}
```

#### ✅ CORRECT: Provider-Specific CRDs

```go
// ALWAYS: Create provider-specific CRDs
type GcpRedisCluster struct {
    Spec GcpRedisClusterSpec `json:"spec"`
}

type GcpRedisClusterSpec struct {
    ShardCount   int32  `json:"shardCount"`
    ReplicaCount int32  `json:"replicaCount"`
    // Only GCP fields
}

type AwsElastiCache struct {
    Spec AwsElastiCacheSpec `json:"spec"`
}

type AwsElastiCacheSpec struct {
    NumShards       int32  `json:"numShards"`
    ReplicasPerShard int32  `json:"replicasPerShard"`
    // Only AWS fields
}
```

### For Existing Resources

**Keep multi-provider CRDs for backward compatibility**:

| CRD | Pattern | Action |
|-----|---------|--------|
| RedisInstance | OLD | Keep as-is, bug fixes only |
| NfsInstance | OLD | Keep as-is, bug fixes only |
| IpRange | OLD | Keep as-is, bug fixes only |

**Do NOT add new features to OLD pattern CRDs**. Instead:
1. Create provider-specific CRD for new capability
2. Document migration path
3. Mark old CRD as deprecated (docs only)

### Example: Adding Redis Cluster Support

#### ❌ WRONG: Extending Old CRD

```yaml
# NEVER: Adding to existing RedisInstance
kind: RedisInstance
spec:
  instance:
    gcp:
      cluster:  # WRONG: Don't add new features to OLD pattern!
        shardCount: 3
```

#### ✅ CORRECT: New Provider-Specific CRD

```yaml
# ALWAYS: Create new provider-specific CRD
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpRedisCluster  # New CRD
spec:
  shardCount: 3
  replicaCount: 2
```

## SKR Impact

### OLD Pattern: SKR → Multi-Provider KCP

```yaml
# SKR: Provider-specific
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisInstance

# Creates KCP: Multi-provider
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance
spec:
  instance:
    gcp:  # Maps to GCP section
      memorySize: 5
```

**Complex mapping** required in SKR reconciler.

### NEW Pattern: SKR → Provider-Specific KCP

```yaml
# SKR: Provider-specific
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisCluster

# Creates KCP: Provider-specific
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpRedisCluster  # Same structure
spec:
  shardCount: 3
```

**Simple 1:1 mapping** in SKR reconciler.

## Pattern Comparison Table

| Aspect | OLD (Multi-Provider) | NEW (Provider-Specific) |
|--------|---------------------|------------------------|
| **CRD** | One CRD, provider sections | One CRD per provider |
| **State Layers** | 3 layers (composed → focal → shared → provider) | 2 layers to focal (composed → focal → provider) |
| **Versioning** | Coupled | Independent |
| **Field Count** | 50+ fields (all providers) | 15-20 fields (one provider) |
| **Validation** | Complex cross-provider | Simple provider-specific |
| **Provider Switching** | Yes (BuildSwitchAction) | No (direct reconciliation) |
| **Maintenance** | Harder (affects all) | Easier (affects one) |
| **Use For** | **Legacy only** | **All new features** |

## Decision Rules

### When to Use NEW Pattern

**ALWAYS** for:
- ✅ Any new resource type
- ✅ Provider-specific features
- ✅ Features unique to one cloud
- ✅ Experimental/beta features

**Examples**:
- `GcpRedisCluster`
- `AzureRedisEnterprise`
- `GcpSubnet`
- `AwsVpcPeering`

### When to Use OLD Pattern

**ONLY** for:
- ⚠️ Maintaining `RedisInstance`
- ⚠️ Maintaining `NfsInstance`
- ⚠️ Maintaining `IpRange`
- ⚠️ Bug fixes in legacy code

**NEVER** for new development.

## Common Pitfalls

### Pitfall 1: Adding Features to OLD Pattern

**Symptom**: Trying to add new capability to multi-provider CRD

**Cause**:
```go
// WRONG: Adding to RedisInstance
type RedisInstanceGcp struct {
    MemorySize int32 `json:"memorySize"`
    NewFeature string `json:"newFeature"`  // WRONG!
}
```

**Solution**: Create new provider-specific CRD:
```go
// CORRECT: New CRD for new capability
type GcpRedisCluster struct {
    ShardCount   int32  `json:"shardCount"`
    NewFeature   string `json:"newFeature"`
}
```

### Pitfall 2: Using Provider Switching in NEW Pattern

**Symptom**: Trying to add provider switching to provider-specific reconciler

**Cause**: Confusing patterns

**Solution**: NEW Pattern has ONE reconciler per CRD, no switching needed

### Pitfall 3: Creating Intermediate State Layer

**Symptom**: Creating shared state for provider-specific CRD

**Cause**:
```go
// WRONG: Intermediate layer in NEW Pattern
type State struct {
    sharedState types.State  // WRONG: No intermediate layer!
    gcpResource *Resource
}
```

**Solution**: Extend focal.State directly:
```go
// CORRECT: Direct extension
type State struct {
    focal.State  // Direct
    gcpResource *Resource
}
```

## Summary: Key Differences

### OLD Pattern (Legacy)
- **CRD**: Multi-provider with provider sections
- **State**: Three-layer hierarchy
- **Reconciler**: Provider switching with BuildSwitchAction
- **Use**: Maintain existing code only

### NEW Pattern (Current)
- **CRD**: Provider-specific (GcpRedisCluster)
- **State**: Extends focal.State directly (two layers to focal)
- **Reconciler**: Direct reconciliation, no switching
- **Use**: ALL new development

### Key Rule
**ALWAYS use NEW Pattern for new code**. OLD Pattern is legacy-only.

## Next Steps

- [Add KCP Reconciler](ADD_KCP_RECONCILER.md) - Use NEW pattern
- [Add SKR Reconciler](ADD_SKR_RECONCILER.md) - Use NEW pattern
- [Reconciler NEW Pattern](../architecture/RECONCILER_NEW_PATTERN.md)
- [Reconciler OLD Pattern](../architecture/RECONCILER_OLD_PATTERN.md)
