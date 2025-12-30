# Guide: Using Feature Flags

**Target Audience**: LLM coding agents  
**Prerequisites**: Understanding action composition, reconciler patterns  
**Purpose**: Control feature availability per landscape, provider, customer plan using feature flags  
**Context**: Cloud Manager uses YAML-based feature flag system with query language for conditional enabling

## Authority: Feature Flag Requirements

### MUST

- **MUST load feature context**: ALWAYS use `feature.LoadFeatureContextFromObj()` in reconciler composition
- **MUST check apiDisabled**: EVERY SKR reconciler checks `feature.ApiDisabled.Value(ctx)` first
- **MUST validate YAML**: Run `make test-ff` after editing feature flag YAML files
- **MUST start disabled**: New flags MUST default to `disabled` in defaultRule
- **MUST use query language**: Targeting rules use query expressions, not code

### MUST NOT

- **NEVER skip context loading**: Feature flags return wrong values without context
- **NEVER modify flags without validation**: Always run `make test-ff` before committing
- **NEVER hardcode landscape/provider**: Use feature flags, not if-statements in code
- **NEVER enable in production first**: Always enable in dev → stage → prod
- **NEVER leave stale flags**: Remove flags after full rollout

### ALWAYS

- **ALWAYS check flags at action entry**: Check at start of action function, not middle
- **ALWAYS return StopAndForget when disabled**: Use `return composed.StopAndForget, nil` pattern
- **ALWAYS document targeting rules**: Add descriptive `name` to each targeting rule
- **ALWAYS use descriptive flag names**: Flag name should indicate feature purpose

### NEVER

- **NEVER bypass feature flags**: Don't code around disabled features
- **NEVER use flags for business logic**: Flags control availability, not behavior
- **NEVER modify GA flags carelessly**: Changes affect production immediately
- **NEVER assume context keys**: Always populate context via LoadFeatureContextFromObj

## Feature Flag Files

### GA Features

**Location**: `pkg/feature/ff_ga.yaml`

Generally available features in production:

```yaml
apiDisabled:
  variations:
    enabled: false
    disabled: true
  targeting:
    - name: Disable NfsBackup
      query: feature == "nfsBackup"
      variation: disabled
    
    - name: Disable NFS On CCEE
      query: provider == "openstack" and feature == "nfs"
      variation: disabled
      
    - name: Disable all on trial
      query: brokerPlan == "trial"
      variation: disabled
  defaultRule:
    variation: enabled
```

### Edge/Experimental Features

**Location**: `pkg/feature/ff_edge.yaml`

Experimental features not yet GA:

```yaml
experimentalFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Enable in dev only
      query: landscape == "dev"
      variation: enabled
  defaultRule:
    variation: disabled
```

## Context Keys

Available context keys for query expressions:

| Key | Values | Source |
|-----|--------|--------|
| `landscape` | dev, stage, prod | From Scope |
| `feature` | nfs, redis, peering, etc | From resource type |
| `provider` | aws, gcp, azure, openstack | From Scope |
| `brokerPlan` | trial, standard, premium | From resource labels |
| `plane` | skr, kcp | From resource location |

Context is automatically populated from resource type, Scope, and labels.

## Using Feature Flags in Reconcilers

### Step 1: Load Feature Context

#### ❌ WRONG: Not Loading Context

```go
// NEVER: Checking flag without loading context
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        composed.LoadObj,
        someAction,  // WRONG: No feature context loaded!
    )
}

func someAction(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {  // WRONG: Returns wrong value!
        return composed.StopAndForget, nil
    }
    return nil, ctx
}
```

#### ✅ CORRECT: Load Context Before Checking

```go
// ALWAYS: Load feature context first
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),  // Load context
        composed.LoadObj,
        someAction,
    )
}

func someAction(ctx context.Context, state composed.State) (error, context.Context) {
    logger := composed.LoggerFromCtx(ctx)
    
    if feature.ApiDisabled.Value(ctx) {  // Now returns correct value
        logger.Info("API is disabled, skipping reconciliation")
        return composed.StopAndForget, nil
    }
    
    return nil, ctx
}
```

### Step 2: Check Feature Flags in Actions

#### ❌ WRONG: Continuing When Disabled

```go
// NEVER: Not stopping when feature disabled
func createResource(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        // WRONG: Should stop, not continue!
    }
    
    // Creates resource even when disabled!
    return nil, ctx
}
```

#### ✅ CORRECT: Stop When Disabled

```go
// ALWAYS: Return StopAndForget when disabled
func createResource(ctx context.Context, state composed.State) (error, context.Context) {
    logger := composed.LoggerFromCtx(ctx)
    
    if feature.ApiDisabled.Value(ctx) {
        logger.Info("API is disabled, skipping resource creation")
        return composed.StopAndForget, nil  // Stop reconciliation
    }
    
    // Only executes when enabled
    return nil, ctx
}
```

### Step 3: Conditional Logic Based on Flags

#### ❌ WRONG: Using Flags for Business Logic

```go
// NEVER: Using feature flags to control algorithm selection
func processData(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.UseNewAlgorithm.Value(ctx) {
        return newAlgorithm(data)  // WRONG: Business logic in flags!
    }
    return oldAlgorithm(data)
}
```

#### ✅ CORRECT: Use Flags for Availability Only

```go
// ALWAYS: Flags control availability, not behavior
func createResource(ctx context.Context, state composed.State) (error, context.Context) {
    // Check if feature is available
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil
    }
    
    // If available, use standard logic
    return standardCreation(ctx, state)
}

// Separate flag for experimental features
func enhancedCreateResource(ctx context.Context, state composed.State) (error, context.Context) {
    // Check if enhanced creation is available
    if !feature.EnhancedCreation.Value(ctx) {
        return composed.StopAndForget, nil
    }
    
    // Enhanced creation logic
    return enhancedCreation(ctx, state)
}
```

## Defining New Feature Flags

### Step 1: Choose File

- **GA features** → `pkg/feature/ff_ga.yaml` (production-ready)
- **Experimental features** → `pkg/feature/ff_edge.yaml` (testing/dev)

### Step 2: Add Flag Definition

#### ❌ WRONG: Enabled by Default

```yaml
# NEVER: New flag enabled by default
myNewFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Disable in prod
      query: landscape == "prod"
      variation: disabled  # WRONG: Should default to disabled!
  defaultRule:
    variation: enabled  # WRONG: New features should default disabled
```

#### ✅ CORRECT: Disabled by Default

```yaml
# ALWAYS: New flags start disabled, enable gradually
myNewFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    # Phase 1: Enable in dev only
    - name: Enable in dev
      query: landscape == "dev"
      variation: enabled
    
    # Phase 2: Enable in dev and stage (uncomment later)
    # - name: Enable in dev and stage
    #   query: landscape == "dev" or landscape == "stage"
    #   variation: enabled
  defaultRule:
    variation: disabled  # Start disabled
```

### Step 3: Add Constant in Code

**Location**: `pkg/feature/flags.go`

#### ❌ WRONG: No Code Constant

```go
// NEVER: Using string literal
if feature.Flag[bool]("myNewFeature").Value(ctx) {  // WRONG: No type safety!
    // ...
}
```

#### ✅ CORRECT: Typed Constant

```go
// ALWAYS: Define typed constant
package feature

// MyNewFeature enables batch operations for Redis clusters
// Currently enabled only in dev landscape
var MyNewFeature = Flag[bool]("myNewFeature")

// In reconciler:
func someAction(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.MyNewFeature.Value(ctx) {  // Type-safe!
        // Feature logic
    }
    return nil, ctx
}
```

### Step 4: Validate Configuration

#### ❌ WRONG: Not Validating

```yaml
# Invalid YAML (missing colon)
myNewFeature
  variations:
    enabled: true
```

**Result**: Silent failure, flag returns default value

#### ✅ CORRECT: Always Validate

```bash
# ALWAYS run after editing YAML
make test-ff
```

This validates:
- YAML syntax correct
- All flags have valid structure
- Query expressions syntactically valid
- No duplicate flag names

## Query Language

### Basic Comparisons

```yaml
# Exact match
query: landscape == "dev"

# Inequality
query: provider != "openstack"

# Multiple conditions (AND)
query: landscape == "prod" and provider == "gcp"

# Multiple conditions (OR)
query: provider == "aws" or provider == "azure"
```

### Complex Queries

#### ❌ WRONG: Invalid Syntax

```yaml
# NEVER: Invalid query syntax
- name: Invalid query
  query: landscape = "dev"  # WRONG: Single = instead of ==
  variation: enabled

- name: Missing quotes
  query: landscape == dev  # WRONG: Missing quotes around value
  variation: enabled
```

#### ✅ CORRECT: Valid Query Syntax

```yaml
# ALWAYS: Use correct operators and quotes
- name: Nested conditions
  query: (landscape == "prod" or landscape == "stage") and provider == "gcp"
  variation: enabled

- name: Multiple features
  query: feature == "nfs" or feature == "redis"
  variation: disabled

- name: Exclude specific combinations
  query: not (provider == "openstack" and feature == "nfs")
  variation: enabled
```

### Available Operators

| Operator | Usage | Example |
|----------|-------|---------|
| `==` | Equals | `provider == "gcp"` |
| `!=` | Not equals | `provider != "openstack"` |
| `and` | Logical AND | `landscape == "prod" and provider == "gcp"` |
| `or` | Logical OR | `provider == "aws" or provider == "azure"` |
| `not` | Logical NOT | `not (feature == "nfs")` |
| `()` | Grouping | `(a or b) and c` |

## Common Flag Patterns

### Pattern: Gradual Rollout

#### ❌ WRONG: Enabling Everywhere at Once

```yaml
# NEVER: Enabling in all landscapes simultaneously
newFeature:
  defaultRule:
    variation: enabled  # WRONG: Too risky!
```

#### ✅ CORRECT: Phase-Based Rollout

```yaml
# ALWAYS: Enable incrementally
newFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    # Phase 1: Dev only
    - name: Enable in dev
      query: landscape == "dev"
      variation: enabled
    
    # Phase 2: Dev and stage (uncomment after validation)
    # - name: Enable in dev and stage
    #   query: landscape == "dev" or landscape == "stage"
    #   variation: enabled
    
    # Phase 3: All landscapes (move to defaultRule after validation)
  defaultRule:
    variation: disabled
```

### Pattern: Provider-Specific Features

```yaml
gcpOnlyFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Enable for GCP only
      query: provider == "gcp"
      variation: enabled
  defaultRule:
    variation: disabled
```

### Pattern: Premium Feature

```yaml
premiumFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Enable for premium plan
      query: brokerPlan == "premium"
      variation: enabled
    
    - name: Enable for standard plan
      query: brokerPlan == "standard"
      variation: enabled
  defaultRule:
    variation: disabled  # Trial users don't get premium features
```

### Pattern: Disable on Trial

```yaml
apiDisabled:
  variations:
    enabled: false
    disabled: true
  targeting:
    - name: Disable on trial
      query: brokerPlan == "trial" and feature == "expensiveFeature"
      variation: disabled
  defaultRule:
    variation: enabled
```

### Pattern: Temporary Disable

```yaml
problematicFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Temporarily disable in prod due to GCP API issues
      query: landscape == "prod"
      variation: disabled
  defaultRule:
    variation: enabled
```

## Common Feature Flags

### apiDisabled

Controls whether entire resources/features are available:

```go
// MUST check in SKR reconcilers
if feature.ApiDisabled.Value(ctx) {
    logger.Info("API is disabled")
    return composed.StopAndForget, nil
}
```

**Use cases**:
- Disable feature in specific landscapes
- Disable for trial accounts
- Disable provider-specific features
- Emergency feature disable

## Debugging Feature Flags

### Check Current Context

#### ❌ WRONG: Not Logging Flag Values

```go
// NEVER: Silent flag checks without logging
func debugAction(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil  // Why was it disabled? Unknown!
    }
    return nil, ctx
}
```

#### ✅ CORRECT: Log Flag Values for Debugging

```go
// ALWAYS: Log flag values when debugging
func debugAction(ctx context.Context, state composed.State) (error, context.Context) {
    logger := composed.LoggerFromCtx(ctx)
    
    // Log all relevant flag values
    logger.Info("Feature flag status",
        "apiDisabled", feature.ApiDisabled.Value(ctx),
        "myFeature", feature.MyNewFeature.Value(ctx),
    )
    
    if feature.ApiDisabled.Value(ctx) {
        logger.Info("API is disabled, stopping reconciliation")
        return composed.StopAndForget, nil
    }
    
    return nil, ctx
}
```

### Test Specific Contexts

```go
// Create test context with specific values
ctx := context.Background()
ctx = context.WithValue(ctx, types.KeyLandscape, "dev")
ctx = context.WithValue(ctx, types.KeyProvider, "gcp")
ctx = context.WithValue(ctx, types.KeyBrokerPlan, "trial")

// Test flag evaluation
if feature.MyNewFeature.Value(ctx) {
    t.Log("Feature enabled in test context")
}
```

## Common Pitfalls

### Pitfall 1: Not Loading Context

**Symptom**: Flag always returns default value, ignoring targeting rules

**Cause**:
```go
// WRONG: No feature context loaded
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        composed.LoadObj,
        someAction,  // feature.ApiDisabled.Value(ctx) returns wrong value!
    )
}
```

**Solution**: ALWAYS load context first:
```go
// CORRECT: Load feature context
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),
        composed.LoadObj,
        someAction,
    )
}
```

### Pitfall 2: YAML Syntax Errors

**Symptom**: Flag returns unexpected value

**Cause**: YAML syntax error (e.g., wrong indentation, missing quotes)

**Solution**: ALWAYS run `make test-ff` after editing

### Pitfall 3: Wrong Query Syntax

**Symptom**: Targeting rule never matches

**Cause**:
```yaml
# WRONG: Single = instead of ==
query: landscape = "dev"

# WRONG: Missing quotes
query: landscape == dev
```

**Solution**: Use correct operators:
```yaml
# CORRECT: Double == and quotes
query: landscape == "dev"
```

### Pitfall 4: Targeting Rules Order

**Symptom**: Wrong variation returned

**Cause**: Targeting rules evaluated in order, first match wins

```yaml
targeting:
  - name: Disable everything
    query: true  # WRONG: Matches everything first!
    variation: disabled
  
  - name: Enable in dev
    query: landscape == "dev"  # Never reached!
    variation: enabled
```

**Solution**: Order rules from specific to general:
```yaml
targeting:
  - name: Enable in dev
    query: landscape == "dev"  # Specific first
    variation: enabled
  
  - name: Disable elsewhere
    query: true  # General last
    variation: disabled
```

## Validation Checklist

### Flag Definition
- [ ] Flag name is descriptive
- [ ] Variations defined (enabled/disabled)
- [ ] Targeting rules have descriptive names
- [ ] Query syntax is correct
- [ ] defaultRule is disabled for new flags
- [ ] Runs `make test-ff` successfully

### Code Integration
- [ ] Constant defined in `pkg/feature/flags.go`
- [ ] Feature context loaded in reconciler
- [ ] Flag checked at action entry point
- [ ] Returns StopAndForget when disabled
- [ ] Logged for debugging

### Rollout Plan
- [ ] Starts disabled in all landscapes
- [ ] Enables in dev first
- [ ] Validated in dev before stage
- [ ] Validated in stage before prod
- [ ] Documented why/when enabled

### Cleanup
- [ ] Remove flag after full rollout
- [ ] Remove flag checks from code
- [ ] Update documentation

## Summary: Key Rules

1. **Load Context First**: Use `feature.LoadFeatureContextFromObj()` before checking flags
2. **Check apiDisabled**: All SKR reconcilers check `feature.ApiDisabled.Value(ctx)`
3. **Start Disabled**: New flags default to `disabled`, enable gradually
4. **Validate YAML**: Always run `make test-ff` after editing
5. **Return StopAndForget**: When disabled, return `composed.StopAndForget, nil`
6. **Query Syntax**: Use `==` (not `=`), quotes around strings, proper operators
7. **Rule Order**: Specific rules before general rules
8. **Cleanup**: Remove flags after full rollout

## Next Steps

- [Add Controller Tests](CONTROLLER_TESTS.md)
- [Add KCP Reconciler](ADD_KCP_RECONCILER.md)
- [Add SKR Reconciler](ADD_SKR_RECONCILER.md)
