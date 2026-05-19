# Feature Flags

Feature flags control feature availability per landscape, provider, and broker plan. They do NOT control business logic.

## Files

| File | Purpose |
|------|---------|
| `pkg/feature/ff_ga.yaml` | Generally available features |
| `pkg/feature/ff_edge.yaml` | Experimental/dev features |
| `pkg/feature/flags.go` | Typed Go constants for each flag |

## YAML Structure

```yaml
myFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Enable in dev
      query: landscape == "dev"
      variation: enabled
  defaultRule:
    variation: disabled   # New flags MUST default to disabled
```

## Context Keys

| Key | Values |
|-----|--------|
| `landscape` | dev, stage, prod |
| `feature` | nfs, redis, peering, etc. |
| `provider` | aws, gcp, azure, openstack |
| `brokerPlan` | trial, standard, premium |
| `plane` | skr, kcp |

Context is auto-populated from the resource type, Scope, and labels via `feature.LoadFeatureContextFromObj`.

## Using Flags in Reconcilers

```go
// 1. Load context — MUST be first in action pipeline
feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),

// 2. Check apiDisabled — MUST be early in every SKR reconciler
composed.If(feature.ApiDisabledPredicate, composed.StopAndForgetAction),

// 3. Check other flags
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.MyFeature.Value(ctx) {
        // feature-specific path
    }
    return nil, ctx
}
```

Without step 1, all flags return their default value regardless of targeting rules.

## Defining a New Flag

1. Add YAML entry (disabled by default) to `ff_ga.yaml` or `ff_edge.yaml`
2. Add typed constant to `pkg/feature/flags.go`:
   ```go
   var MyFeature = Flag[bool]("myFeature")
   ```
3. Run `make test-ff` to validate YAML syntax and query expressions

## Query Language

```yaml
# Exact match
query: landscape == "dev"

# AND / OR / NOT
query: landscape == "prod" and provider == "gcp"
query: provider == "aws" or provider == "azure"
query: not (provider == "openstack" and feature == "nfs")

# Grouping
query: (landscape == "dev" or landscape == "stage") and provider == "gcp"
```

Use `==` (not `=`). Quote string values. Order rules specific → general; first match wins.

## Gradual Rollout Pattern

```yaml
myFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Phase 1 — dev only
      query: landscape == "dev"
      variation: enabled
    # Phase 2 — uncomment after dev validation:
    # - name: Phase 2 — dev and stage
    #   query: landscape == "dev" or landscape == "stage"
    #   variation: enabled
  defaultRule:
    variation: disabled
```

## Rules

- New flags MUST default to `disabled`
- Run `make test-ff` after every YAML edit
- Never hardcode landscape/provider in code — use flags
- FFs may gate business logic (algorithm selection, new behavior) during gradual rollout — but the old code path and the flag MUST be removed after full rollout. Flags are not permanent configuration switches.
- Remove flags after full rollout
