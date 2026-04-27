---
name: composed-reconciler
description: |
  Create KCP or SKR reconcilers using the composed action pattern. Use this skill when:
  - Creating a new reconciler for cloud-control (KCP) or cloud-resources (SKR) API
  - Adding provider support (AWS, GCP, Azure, OpenStack) to existing reconcilers
  - Understanding composed action flow, state hierarchies, or provider branching
  - Implementing cloud provider client facades in reconciler state
  Trigger phrases: "create reconciler", "add provider flow", "composed action",
  "KCP reconciler", "SKR reconciler", "state factory", "provider branching"
---

# Composed Reconciler Pattern

This skill helps you create Kubernetes reconcilers using the cloud-manager composed action framework.

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

- `MarkedForDeletionPredicate` - Object has deletion timestamp
- `NotMarkedForDeletionPredicate` - Object not marked for deletion (use this instead of `Not(MarkedForDeletionPredicate)`)

## Architecture Overview

Cloud Manager has two API groups with different reconciliation patterns:

### KCP (cloud-control) - Control Plane

- Runs in Kyma Control Plane cluster
- Reconciles cloud provider resources (VPC, NFS, Redis, etc.)
- **Branches by provider** (AWS, GCP, Azure, OpenStack)
- Uses state hierarchy with embedding
- Provider-specific StateFactory initializes cloud clients

### SKR (cloud-resources) - Data Plane

- Runs in each SAP BTP Kyma Runtime (user's cluster)
- **No provider branching** - simpler linear flows
- Syncs SKR spec to KCP objects
- Syncs KCP status back to SKR
- May create local K8s resources (PV, PVC, Secrets)

## Interactive Workflow

When using this skill, provide or be asked about:

1. **Reconciler type**: KCP or SKR?
2. **Resource name**: e.g., "NfsInstance", "RedisCluster"
3. **For KCP**: Which providers? (AWS, GCP, Azure, OpenStack)
4. **KCP base state** (if new resource):
   - `kcpcommonaction.State` (RECOMMENDED) - new pattern
   - `focal.State` (LEGACY) - for backward compatibility
5. **Cloud client needed?**: Will the flow call cloud provider APIs?

## Coding Conventions

### ALWAYS follow these rules:

1. **Use `ComposeActionsNoName`** - avoid `ComposeActions` with names
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
5. **Use `NotMarkedForDeletionPredicate`** - not `Not(MarkedForDeletionPredicate)`
6. **Don't generate speculative actions** - use comment placeholders unless explicitly specified

## KCP Reconciler Structure

### Directory Layout

```
internal/controller/cloud-control/
└── {resource}_controller.go        # Thin wrapper, delegates to pkg/kcp

pkg/kcp/{resource}/
├── types/
│   └── state.go                    # State interface (breaks circular deps)
├── reconciler.go                   # Main reconciler
├── state.go                        # State struct
└── {action}.go                     # Individual actions

pkg/kcp/provider/{provider}/{resource}/
├── new.go                          # Provider New() entry point
├── state.go                        # Provider state with cloud client
└── {action}.go                     # Provider-specific actions
```

### Circular Dependency Resolution

Reconciler references providers (for Switch), but providers embed reconciler state.
**Solution**: Define state interface in `types/` subpackage.

### Controller Template

```go
// internal/controller/cloud-control/{resource}_controller.go
func Setup{Resource}Reconciler(
    kcpManager manager.Manager,
    awsStateFactory aws{resource}.StateFactory,
    azureStateFactory azure{resource}.StateFactory,
    gcpStateFactory gcp{resource}.StateFactory,
    sapStateFactory sap{resource}.StateFactory,
) error {
    return New{Resource}Reconciler(
        kcp{resource}.New(
            composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
            kcpcommonaction.NewStateFactory(),
            awsStateFactory,
            azureStateFactory,
            gcpStateFactory,
            sapStateFactory,
        ),
    ).SetupWithManager(kcpManager)
}
```

### Reconciler Template (kcpcommonaction.State - RECOMMENDED)

```go
// pkg/kcp/{resource}/reconciler.go
func New(
    composedStateFactory composed.StateFactory,
    kcpCommonStateFactory kcpcommonaction.StateFactory,
    awsStateFactory aws{resource}.StateFactory,
    azureStateFactory azure{resource}.StateFactory,
    gcpStateFactory gcp{resource}.StateFactory,
    sapStateFactory sap{resource}.StateFactory,
) {Resource}Reconciler {
    return &reconciler{
        composedStateFactory:  composedStateFactory,
        kcpCommonStateFactory: kcpCommonStateFactory,
        awsStateFactory:       awsStateFactory,
        azureStateFactory:     azureStateFactory,
        gcpStateFactory:       gcpStateFactory,
        sapStateFactory:       sapStateFactory,
    }
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newKcpCommonState(req.NamespacedName)
    action := r.newAction()

    return composed.Handling().
        WithMetrics("{resource}", util.RequestObjToString(req)).
        Handle(action(ctx, state))
}

func (r *reconciler) newKcpCommonState(name types.NamespacedName) kcpcommonaction.State {
    return r.kcpCommonStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.{Resource}{}),
    )
}

func (r *reconciler) newAction() composed.Action {
    providerFlow := composed.Switch(
        nil,
        composed.NewCase(kcpcommonaction.AwsProviderPredicate, aws{resource}.New(r.awsStateFactory)),
        composed.NewCase(kcpcommonaction.AzureProviderPredicate, azure{resource}.New(r.azureStateFactory)),
        composed.NewCase(kcpcommonaction.GcpProviderPredicate, gcp{resource}.New(r.gcpStateFactory)),
        composed.NewCase(kcpcommonaction.OpenStackProviderPredicate, sap{resource}.New(r.sapStateFactory)),
    )

    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.{Resource}{}),
        kcpcommonaction.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActionsNoName(
                // common actions placeholder

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
                        // validation actions placeholder
                        providerFlow,
                        statusReady,
                    ),
                ),
            )(ctx, newState(st.(kcpcommonaction.State)))
        },
    )
}
```

### Provider State Template

```go
// pkg/kcp/provider/{provider}/{resource}/state.go
type State struct {
    types.State           // Embed state interface from types package
    client      Client    // Provider-specific client
}

type StateFactory interface {
    NewState(ctx context.Context, baseState types.State) (context.Context, *State, error)
}

func NewStateFactory(clientProvider ClientProvider) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, baseState types.State) (context.Context, *State, error) {
    client, err := f.clientProvider(
        ctx,
        baseState.Subscription().Status.SubscriptionInfo.{Provider}.{Info},
    )
    if err != nil {
        return ctx, nil, err
    }
    return ctx, &State{State: baseState, client: client}, nil
}
```

### Provider New() Template

```go
// pkg/kcp/provider/{provider}/{resource}/new.go
func New(sf StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        baseState := st.(types.State)
        cctx, state, err := sf.NewState(ctx, baseState)
        if err != nil {
            return err, ctx
        }

        return composed.ComposeActionsNoName(
            // delete ================================================================================
            composed.If(
                composed.MarkedForDeletionPredicate,
                composed.ComposeActionsNoName(
                    infraDelete,
                ),
            ),

            // create/update =========================================================================
            composed.If(
                composed.NotMarkedForDeletionPredicate,
                composed.ComposeActionsNoName(
                    infraCreateUpdate,
                ),
            ),
        )(cctx, state)
    }
}
```

## SKR Reconciler Structure

### Directory Layout

```
internal/controller/cloud-resources/
└── {resource}_controller.go        # Factory pattern, registers with SKR runtime

pkg/skr/{resource}/
├── reconciler.go                   # Main reconciler
├── state.go                        # State with KcpCluster + SkrCluster
├── createKcp{Resource}.go          # Sync spec to KCP
├── updateStatus.go                 # Sync KCP status back to SKR
└── {action}.go                     # Other actions (local resources, etc.)
```

### Controller Template

```go
// internal/controller/cloud-resources/{resource}_controller.go
type {Resource}ReconcilerFactory struct{}

func (f *{Resource}ReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
    return &{Resource}Reconciler{
        reconciler: {resource}.NewReconcilerFactory().New(args),
    }
}

func Setup{Resource}Reconciler(reg skrruntime.SkrRegistry) error {
    return reg.Register().
        WithFactory(&{Resource}ReconcilerFactory{}).
        For(&cloudresourcesv1beta1.{Resource}{}).
        Complete()
}
```

### Reconciler Template

```go
// pkg/skr/{resource}/reconciler.go
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
                // local actions placeholder - add only when specified
            ),
        ),

        composed.StopAndForgetAction,
    )
}
```

### SKR State Template

```go
// pkg/skr/{resource}/state.go
type State struct {
    composed.State
    KymaRef      klog.ObjectRef
    KcpCluster   composed.StateCluster
    Kcp{Resource} *cloudcontrolv1beta1.{Resource}
}

type stateFactory struct {
    baseStateFactory composed.StateFactory
    scopeProvider    scopeprovider.ScopeProvider
    kcpCluster       composed.StateCluster
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
    kymaRef, err := f.scopeProvider.GetScope(ctx, req.NamespacedName)
    if err != nil {
        return nil, err
    }
    return &State{
        State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.{Resource}{}),
        KymaRef:    kymaRef,
        KcpCluster: f.kcpCluster,
    }, nil
}
```

## Reference Files

For deeper understanding, see:
- `references/primitives.md` - Detailed pkg/composed documentation
- `references/kcp-reconciler.md` - KCP patterns with examples
- `references/skr-reconciler.md` - SKR patterns with examples

## Example Files in Codebase

| Pattern | Example Files |
|---------|---------------|
| KCP Controller | `internal/controller/cloud-control/vpcnetwork_controller.go` |
| KCP Reconciler (new) | `pkg/kcp/vpcnetwork/reconciler.go` |
| KCP Reconciler (legacy) | `pkg/kcp/nfsinstance/reconciler.go` |
| Provider State | `pkg/kcp/provider/aws/vpcnetwork/state.go` |
| Provider New() | `pkg/kcp/provider/aws/vpcnetwork/new.go` |
| SKR Controller | `internal/controller/cloud-resources/gcpnfsvolume_controller.go` |
| SKR Reconciler | `pkg/skr/gcpnfsvolume/reconciler.go` |
| SKR State | `pkg/skr/awsredisinstance/state.go` |
