---
name: composed-action
description: |
  Create KCP or SKR reconcilers using the composed action pattern. Use this skill when:
  - Creating a new reconciler for cloud-control (KCP) or cloud-resources (SKR) API
  - Adding provider support (AWS, GCP, Azure, OpenStack) to existing reconcilers
  - Understanding composed action flow, state hierarchies, or provider branching
  - Implementing cloud provider client facades in reconciler state
  - Creating or adding new actions to existing reconcilers
  - Modifying action flow or changing how actions interact
  - Understanding action return values and flow control
  - Updating object status in reconcilers
  Trigger phrases: "create reconciler", "add provider flow", "composed action",
  "KCP reconciler", "SKR reconciler", "state factory", "provider branching",
  "provider skeleton", "reconciler skeleton", "composed action flow",
  "create action", "add action", "modify action", "change action",
  "action flow", "action return", "StopWithRequeue", "StopWithRequeueDelay",
  "status update", "StatusPatcher", "UpdateStatus"
---

# Composed Reconciler Pattern

This skill helps you create Kubernetes reconcilers using the cloud-manager composed action framework.

## Quick Reference

### Core Types (pkg/composed)

| Type | Signature | Purpose |
|------|-----------|---------|
| `Action` | `func(ctx, state) (error, context.Context)` | Executable unit of work |
| `State` | Interface | Holds reconciliation context, K8s object |
| `Predicate` | `func(ctx, state) bool` | Conditional branching |

### Composition Functions

| Function | Usage |
|----------|-------|
| `ComposeActionsNoName(actions...)` | Sequential pipeline (ALWAYS use this) |
| `If(predicate, action)` | Execute if predicate true |
| `Switch(default, cases...)` | Multi-way branching |

### Flow Control Returns

| Return | Effect |
|--------|--------|
| `nil, ctx` | Continue to next action |
| `StopWithRequeue, ctx` | Stop, requeue immediately |
| `StopWithRequeueDelay(d), ctx` | Stop, requeue after delay |
| `StopAndForget, ctx` | Stop, no requeue |

**CRITICAL**: Never return `nil, nil`. Always return the context.

## Architecture Overview

### KCP (cloud-control) - Control Plane

- Runs in Kyma Control Plane cluster
- Reconciles cloud provider resources (VPC, NFS, Redis)
- **Branches by provider** (AWS, GCP, Azure, OpenStack)
- Uses state hierarchy with embedding
- Provider StateFactory initializes cloud clients

### SKR (cloud-resources) - Data Plane

- Runs in each SAP BTP Kyma Runtime
- **No provider branching** - linear flows
- Syncs SKR spec to KCP objects
- May create local K8s resources (PV, PVC, Secrets)

## Interactive Workflow

When using this skill, provide or be asked about:

1. **Reconciler type**: KCP or SKR?
2. **Resource name**: e.g., "NfsInstance", "RedisCluster"
3. **For KCP**: Which providers? (AWS, GCP, Azure, OpenStack)
4. **Cloud client needed?**: Will the flow call cloud provider APIs?

## What Do You Need?

| Task | Reference |
|------|-----------|
| Create KCP reconciler | `references/kcp-reconciler.md` |
| Create SKR reconciler | `references/skr-reconciler.md` |
| Add/modify actions | `references/flow-control.md` |
| Update object status | `references/status-mutation.md` |
| Provider client setup | `references/provider-clients.md` |
| Coding conventions | `references/conventions.md` |
| pkg/composed primitives | `references/primitives.md` |

## Directory Layouts

### KCP Layout

```
internal/controller/cloud-control/
└── {resource}_controller.go

pkg/kcp/{resource}/
├── types/state.go          # State interface
├── reconciler.go           # Main reconciler
├── state.go                # State struct
└── {action}.go             # Actions

pkg/kcp/provider/{provider}/{resource}/
├── new.go                  # Provider entry point
├── state.go                # Provider state + client
└── {action}.go             # Provider actions
```

### SKR Layout

```
internal/controller/cloud-resources/
└── {resource}_controller.go

pkg/skr/{resource}/
├── reconciler.go
├── state.go
├── loadKcp{Resource}.go
├── createKcp{Resource}.go
└── updateStatus.go
```

## Coding Conventions (Critical)

1. **Use `ComposeActionsNoName`** - never `ComposeActions`
2. **One action per line** when composing
3. **Separate flows** with comment markers:
   ```go
   // delete ================================================================================
   // create/update =========================================================================
   ```
4. **Use separate `If` blocks**:
   ```go
   composed.If(composed.MarkedForDeletionPredicate, deleteFlow),
   composed.If(composed.NotMarkedForDeletionPredicate, createFlow),
   ```
5. **Use `NotMarkedForDeletionPredicate`** - not `Not(MarkedForDeletionPredicate)`
6. **Never return `nil, nil`** - always return context

## KCP Reconciler Template (Minimal)

```go
func (r *reconciler) newAction() composed.Action {
    providerFlow := composed.Switch(
        nil,
        composed.NewCase(kcpcommonaction.AwsProviderPredicate, aws{resource}.New(r.awsStateFactory)),
        composed.NewCase(kcpcommonaction.GcpProviderPredicate, gcp{resource}.New(r.gcpStateFactory)),
        composed.NewCase(kcpcommonaction.AzureProviderPredicate, azure{resource}.New(r.azureStateFactory)),
        composed.NewCase(kcpcommonaction.OpenStackProviderPredicate, sap{resource}.New(r.sapStateFactory)),
    )

    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.{Resource}{}),
        kcpcommonaction.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActionsNoName(
                // delete ================================================================================
                composed.If(
                    composed.MarkedForDeletionPredicate,
                    composed.ComposeActionsNoName(
                        providerFlow,
                        actions.PatchRemoveCommonFinalizer(),
                    ),
                ),
                // create/update =========================================================================
                composed.If(
                    composed.NotMarkedForDeletionPredicate,
                    composed.ComposeActionsNoName(
                        actions.PatchAddCommonFinalizer(),
                        providerFlow,
                        statusReady,
                    ),
                ),
            )(ctx, newState(st.(kcpcommonaction.State)))
        },
    )
}
```

## SKR Reconciler Template (Minimal)

```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.{Resource}{}),
        composed.LoadObj,
        loadKcp{Resource},

        // delete ================================================================================
        composed.If(
            composed.MarkedForDeletionPredicate,
            composed.ComposeActionsNoName(
                deleteKcp{Resource},
                actions.RemoveCommonFinalizer(),
            ),
        ),

        // create/update =========================================================================
        composed.If(
            composed.NotMarkedForDeletionPredicate,
            composed.ComposeActionsNoName(
                actions.AddCommonFinalizer(),
                createKcp{Resource},
                waitKcpReady,
                updateStatus,
            ),
        ),

        composed.StopAndForgetAction,
    )
}
```

## Example Files in Codebase

| Pattern | Example |
|---------|---------|
| KCP Controller | `internal/controller/cloud-control/vpcnetwork_controller.go` |
| KCP Reconciler | `pkg/kcp/vpcnetwork/reconciler.go` |
| Provider State | `pkg/kcp/provider/aws/vpcnetwork/state.go` |
| Provider New() | `pkg/kcp/provider/aws/vpcnetwork/new.go` |
| SKR Controller | `internal/controller/cloud-resources/gcpnfsvolume_controller.go` |
| SKR Reconciler | `pkg/skr/gcpnfsvolume/reconciler.go` |
