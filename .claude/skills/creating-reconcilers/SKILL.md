---
name: creating-reconcilers
description: Use when creating or modifying reconcilers in Cloud Manager — adding a new KCP resource (cloud-control API), a new SKR resource (cloud-resources API), or adding a new provider (AWS, GCP, Azure, OpenStack) to an existing resource.
---

# Creating Reconcilers

Cloud Manager has three reconciler types. Read the correct reference BEFORE writing any code.
Announce which type you are implementing in your first response.

## Reconciler Type Decision

| Scenario | Type | Reference |
|---|---|---|
| KCP resource reconciled across multiple providers (AWS/GCP/Azure/OpenStack) | KCP Multi-Provider | `references/kcp-multi-provider.md` |
| KCP resource for one specific provider only (e.g. GcpSubnet) | KCP Single-Provider | `references/kcp-single-provider.md` |
| SKR resource that creates and manages a corresponding KCP resource | SKR | `references/skr-reconciler.md` |
| SKR resource with no backing KCP resource (schedules, backups) | SKR-Only | `references/skr-only-pattern.md` |

> **Modifying existing legacy code** — If the resource already uses `focal.State` (`NfsInstance`, `RedisInstance`, `IpRange`), read the existing implementation and continue it. Do not introduce kcpcommonaction.State patterns unless explicitly asked to migrate.

## Architecture Overview

### KCP (cloud-control) — Control Plane

Runs in Kyma Control Plane. Reconciles cloud provider resources (VPC, NFS, Redis, etc.). Branches by provider (AWS, GCP, Azure, OpenStack) using a Switch predicate. Provider-specific StateFactory initializes cloud clients.

### SKR (cloud-resources) — Data Plane

Runs in each SAP BTP Kyma Runtime (user's cluster). No provider branching — linear pipelines. Syncs SKR spec to KCP objects, KCP status back to SKR. May also create local K8s resources (PV, PVC, Secrets).

## Quick Reference

### Core Types (pkg/composed)

| Type | Signature | Purpose |
|------|-----------|---------|
| `Action` | `func(ctx context.Context, state State) (error, context.Context)` | Executable unit of work |
| `State` | Interface | Holds reconciliation context, K8s object, cluster access |
| `StateFactory` | `NewState(name, obj) State` | Creates State instances |
| `Predicate` | `func(ctx context.Context, state State) bool` | Conditional branching |

### Composition Functions

| Function | Usage |
|----------|-------|
| `ComposeActionsNoName(actions...)` | Sequential action pipeline (PREFERRED) |
| `ComposeActions(name, actions...)` | Named sequential pipeline (avoid) |
| `If(predicate, action)` | Execute action if predicate is true |
| `Switch(default, cases...)` | Multi-way branching |
| `NewCase(predicate, action)` | Case for Switch |

### Flow Control Errors

| Error | Behavior |
|-------|----------|
| `StopAndForget` | End reconciliation, no requeue |
| `StopWithRequeue` | End and requeue immediately |
| `StopWithRequeueDelay(d)` | End and requeue after duration |
| `Break` | Exit current composition |

### Built-in Predicates

- `MarkedForDeletionPredicate` — Object has deletion timestamp
- `NotMarkedForDeletionPredicate` — Object not marked for deletion (use this instead of `Not(MarkedForDeletionPredicate)`)

## Coding Conventions

ALWAYS follow these rules:

1. **Use `ComposeActionsNoName`** — avoid `ComposeActions` with names
2. **One action per line** when composing
3. **Separate delete and create/update** with comment markers:
   ```go
   // delete ================================================================================
   // create/update =========================================================================
   ```
4. **Use separate `If` blocks** instead of `IfElse`:
   ```go
   // CORRECT:
   composed.If(composed.MarkedForDeletionPredicate, deleteFlow),
   composed.If(composed.NotMarkedForDeletionPredicate, createUpdateFlow),

   // WRONG:
   composed.IfElse(composed.MarkedForDeletionPredicate, deleteFlow, createUpdateFlow)
   ```
5. **Use `NotMarkedForDeletionPredicate`** — not `Not(MarkedForDeletionPredicate)`
6. **Don't generate speculative actions** — use comment placeholders unless explicitly specified
7. **Sequential only** — actions execute one at a time. NEVER compose actions to run in parallel.
8. **Finalizer law** — every resource requiring deletion cleanup MUST have a finalizer. Add it on the create path. Remove it ONLY after deletion is fully confirmed (see pitfall #8).
9. **Feature flag gate** — every SKR reconciler MUST include `composed.If(feature.ApiDisabledPredicate, composed.StopAndForgetAction)` early in the pipeline, after loading feature context (see pitfall #9).
10. **Injectable clock** — if a reconciler performs any time-based logic (scheduling, expiry, rate limiting), the StateFactory MUST accept `clock.Clock`. Use `clock.RealClock{}` in production and `clock.NewFakeClock()` in tests (see pitfall #12).

## Critical Pitfalls (Summary)

See `references/action-pitfalls.md`. Most frequent:
- **#1** state type assertion panic — ALWAYS assert through the full hierarchy, never skip levels
- **#2** missing `StopAndForget` at flow end — every successful path MUST terminate
- **#7** missing KCP labels when creating from SKR — breaks status sync and cross-cluster debugging
- **#14** adding `types/` subpackage to single-provider resources — only needed when multiple providers share a state interface

## References

**Read exactly ONE** (from the decision table above):
- `references/kcp-multi-provider.md` — KCP multi-provider pattern (AWS/GCP/Azure/OpenStack)
- `references/kcp-single-provider.md` — KCP single-provider pattern (GcpSubnet)
- `references/skr-reconciler.md` — SKR pattern (with backing KCP resource)
- `references/skr-only-pattern.md` — SKR-Only pattern (no KCP resource)

**Read alongside your flow reference:**
- `references/action-pitfalls.md` — Full GOOD/BAD examples for pitfalls listed above
- `references/feature-flags.md` — Required for all SKR flows; see pitfall #9
- `references/primitives.md` — Deep-dive on pkg/composed API; read only if you need detail beyond the Quick Reference above
