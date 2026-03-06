---
name: feature-flags
description: Work with Cloud Manager feature flags. Use when adding feature toggles, checking feature flag status, or implementing gradual rollout.
---

# Feature Flags

Manage feature flags for gradual rollout and feature gating.

## Quick Start

1. Define flag in `pkg/feature/`
2. Load flag context in reconciler
3. Check flag before execution
4. Configure flags in YAML files

## Flag Definition

**File**: `pkg/feature/feature.go`

```go
var (
    GcpNewResource = NewFeature("GcpNewResource", GcpNewResourceEnabled)
)

const (
    GcpNewResourceEnabled  = "enabled"
    GcpNewResourceDisabled = "disabled"
)
```

## Loading Flag Context

In reconciler's SetupWithManager:

```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpNewResource{}),
        // ... rest of actions
    )
}
```

## Checking Flags

```go
func checkFeatureDisabled(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.GcpNewResource.Value(ctx) == feature.GcpNewResourceDisabled {
        return composed.StopAndForget, nil
    }
    return nil, ctx
}
```

Or using the common pattern:

```go
func checkFeatureDisabled(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil
    }
    return nil, ctx
}
```

## Configuration Files

| File | Purpose | Environment |
|------|---------|-------------|
| `pkg/feature/ff_ga.yaml` | Generally Available | Production |
| `pkg/feature/ff_edge.yaml` | Edge/Experimental | Edge clusters |
| `config/featureToggles/featureToggles.local.yaml` | Local overrides | Development |

## YAML Format

```yaml
features:
  - name: GcpNewResource
    value: enabled
    # Optional targeting
    scope:
      - kyma-system/*    # All resources in namespace
      - */my-resource    # Specific resource name
```

## Targeting Rules

```yaml
# Enable for specific namespace
scope:
  - target-namespace/*

# Enable for specific resource
scope:
  - */specific-resource

# Enable for namespace/name combo
scope:
  - target-namespace/specific-resource

# Disable everywhere except listed
value: disabled
scope:
  - allowed-namespace/*
```

## Testing Flags

```go
It("Should respect disabled feature", func() {
    // Set flag to disabled in test context
    ctx := feature.ContextBuilderFromCtx(ctx).
        Feature(feature.GcpNewResource, feature.GcpNewResourceDisabled).
        Build()

    // Resource should be ignored
    err := reconciler.Reconcile(ctx, req)
    Expect(err).NotTo(HaveOccurred())

    // Verify no cloud resources created
    Expect(gcpMock.GetResource(id)).To(BeNil())
})
```

## Validate Flags

```bash
make test-ff
```

## Rules

### MUST
- Check feature flags in SKR reconcilers
- Load feature context before checking
- Use defined constants (not strings)

### MUST NOT
- Modify flag definitions without approval
- Skip flag checks in user-facing resources
- Hardcode flag values

## Checklist

- [ ] Flag defined in `pkg/feature/`
- [ ] Context loaded in reconciler
- [ ] Flag checked before execution
- [ ] YAML configuration added
- [ ] Tests cover enabled/disabled states
- [ ] `make test-ff` passes

## Related

- Full guide: [docs/agents/guides/FEATURE_FLAGS.md](../../../docs/agents/guides/FEATURE_FLAGS.md)
