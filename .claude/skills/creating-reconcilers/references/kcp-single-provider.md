# KCP Single-Provider Reconciler Pattern

Use this reference when:
- The KCP resource belongs to one specific provider only (e.g. GcpSubnet)
- There is NO `pkg/kcp/{resource}/` shared package — the reconciler lives entirely in `pkg/kcp/provider/{provider}/{resource}/`
- No `types/` subpackage is needed (no cross-provider references, no circular import problem)

Do NOT use for resources that must be reconciled across multiple providers — see `kcp-multi-provider.md`

## Overview

Single-provider KCP reconcilers skip the shared abstraction layer. The reconciler, state, and all actions live in the provider package. `focal.State` is embedded directly — no `types.State` interface layer.

## Directory Structure

```
internal/controller/cloud-control/
└── {resource}_controller.go          # Setup function, single stateFactory

pkg/kcp/provider/gcp/{resource}/
├── reconcile.go                      # Reconciler struct + newAction() + newFlow()
├── state.go                          # State embeds focal.State directly
├── ignorant.go                       # Optional: ignore key logic
└── {action}.go                       # Actions
```

Note: NO `pkg/kcp/{resource}/` shared package, NO `types/` subpackage.

## State Pattern

State embeds `focal.State` directly — no interface layer between composed and provider state.

**File**: `pkg/kcp/provider/gcp/subnet/state.go`

```go
package subnet

type State struct {
    focal.State                               // Direct embed, no interface
    computeClient             client.ComputeClient
    networkConnectivityClient client.NetworkConnectivityClient
    regionOperationsClient    client.RegionOperationsClient

    subnet                  *computepb.Subnetwork
    serviceConnectionPolicy *networkconnectivitypb.ServiceConnectionPolicy
    network                 *cloudcontrolv1beta1.Network
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}
```

StateFactory accepts `focal.State` (not a `types.State` interface) and returns `*State` directly.

```go
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    computeClient := f.computeClientProvider(focalState.Scope().Spec.Scope.Gcp.Project)
    // ... initialize other clients
    return &State{State: focalState, computeClient: computeClient, ...}, nil
}
```

## Reconciler Pattern

The reconciler holds `composedStateFactory`, `focalStateFactory`, and ONE `stateFactory`. No provider Switch.

**File**: `pkg/kcp/provider/gcp/subnet/reconcile.go`

```go
package subnet

type gcpSubnetReconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory
}

func NewGcpSubnetReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    stateFactory StateFactory,
) GcpSubnetReconciler {
    return &gcpSubnetReconciler{
        composedStateFactory: composedStateFactory,
        focalStateFactory:    focalStateFactory,
        stateFactory:         stateFactory,
    }
}

func (r *gcpSubnetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    if Ignore.ShouldIgnoreKey(req) {
        return ctrl.Result{}, nil
    }

    state := r.newFocalState(req.NamespacedName)
    action := r.newAction()

    return composed.Handling().
        WithMetrics("kcpgcpsubnet", util.RequestObjToString(req)).
        Handle(action(ctx, state))
}

func (r *gcpSubnetReconciler) newFocalState(name types.NamespacedName) focal.State {
    return r.focalStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.GcpSubnet{}),
    )
}

func (r *gcpSubnetReconciler) newAction() composed.Action {
    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{}),
        focal.New(),       // Loads Scope + feature context (not kcpcommonaction.New())
        r.newFlow(),
    )
}

func (r *gcpSubnetReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            // Set error condition and stop
            subnet := st.Obj().(*cloudcontrolv1beta1.GcpSubnet)
            subnet.Status.State = cloudcontrolv1beta1.StateError
            return composed.UpdateStatus(subnet).
                SetExclusiveConditions(metav1.Condition{
                    Type:    cloudcontrolv1beta1.ConditionTypeError,
                    Status:  metav1.ConditionTrue,
                    Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
                    Message: "Failed to create state",
                }).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }

        return composed.ComposeActionsNoName(
            actions.AddCommonFinalizer(),

            // delete ================================================================================
            composed.If(
                composed.MarkedForDeletionPredicate,
                composed.ComposeActionsNoName(
                    // provider delete actions
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),

            // create/update =========================================================================
            composed.If(
                composed.NotMarkedForDeletionPredicate,
                composed.ComposeActionsNoName(
                    // provider create/update actions
                    updateStatus,
                ),
            ),

            composed.StopAndForgetAction,
        )(ctx, state)
    }
}
```

## Controller Setup Pattern

The controller accepts only one `stateFactory` — no AWS/Azure/OpenStack variants.

**File**: `internal/controller/cloud-control/gcpsubnet_controller.go`

```go
func SetupGcpSubnetReconciler(
    ctx context.Context,
    kcpManager manager.Manager,
    computeClientProvider gcpclient.GcpClientProvider[gcpsubnetclient.ComputeClient],
    networkConnectivityClientProvider gcpclient.GcpClientProvider[gcpsubnetclient.NetworkConnectivityClient],
    regionOperationsClientProvider gcpclient.GcpClientProvider[gcpsubnetclient.RegionOperationsClient],
    env abstractions.Environment,
) error {
    return NewGcpSubnetReconciler(
        subnet.NewGcpSubnetReconciler(
            composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
            focal.NewStateFactory(),                          // loads Scope
            subnet.NewStateFactory(                           // single provider factory
                computeClientProvider,
                networkConnectivityClientProvider,
                regionOperationsClientProvider,
                env,
            ),
        ),
    ).SetupWithManager(ctx, kcpManager)
}
```

Contrast with multi-provider: no `kcpcommonaction.NewStateFactory()`, no `aws/azure/gcp/sap` factory arguments.

## Key Differences from Multi-Provider Pattern

| Aspect | Single-Provider | Multi-Provider |
|--------|----------------|----------------|
| State base | `focal.State` direct embed | `kcpcommonaction.State` via `types.State` interface |
| `types/` subpackage | NONE | Required (breaks circular import) |
| Provider Switch | NONE | `composed.Switch` with 4 predicates |
| Loader | `focal.New()` | `kcpcommonaction.New()` |
| Controller args | 1 stateFactory + client providers | 4 stateFactories (aws/azure/gcp/sap) |
| Package location | `pkg/kcp/provider/gcp/{resource}/` only | `pkg/kcp/{resource}/` + per-provider packages |

## Example Files in Codebase

| Component | File |
|-----------|------|
| Controller setup | `internal/controller/cloud-control/gcpsubnet_controller.go` |
| Reconciler + flow | `pkg/kcp/provider/gcp/subnet/reconcile.go` |
| State (focal.State embed) | `pkg/kcp/provider/gcp/subnet/state.go` |
