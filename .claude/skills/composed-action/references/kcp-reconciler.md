# KCP Reconciler Patterns Reference

This document provides detailed patterns for KCP (cloud-control) reconcilers.

## Overview

KCP reconcilers run in the Kyma Control Plane and manage cloud provider resources. They feature:
- Provider-specific branching (AWS, GCP, Azure, OpenStack)
- Multi-layer state hierarchy with embedding
- Cloud client initialization via StateFactory

## State Hierarchy Patterns

### Pattern A: kcpcommonaction.State (RECOMMENDED)

**Status**: New pattern, all reconcilers should migrate to this.

```
composed.State
    └── kcpcommonaction.State
            └── {resource}.State (types.State interface)
                    └── provider/{provider}/{resource}.State
```

**What kcpcommonaction.New() loads**:
- Subscription
- IpRange
- VpcNetwork
- GcpSubnet

**Example**: `pkg/kcp/vpcnetwork/`

### Pattern B: focal.State (LEGACY)

**Status**: Legacy pattern, used by NfsInstance, RedisInstance, IpRange. Migration in progress.

```
composed.State
    └── focal.State
            └── {resource}.State
                    └── provider/{provider}/{resource}.State
```

**What focal.New() loads**:
- Scope (from ScopeRef on the object)
- Feature context

**Example**: `pkg/kcp/nfsinstance/`

## Directory Structure

```
internal/controller/cloud-control/
└── {resource}_controller.go          # Controller setup, delegates to reconciler

pkg/kcp/{resource}/
├── types/
│   └── state.go                      # State interface (breaks circular deps)
├── reconciler.go                     # Reconciler struct and newAction()
├── state.go                          # State struct implementation
├── ignore.go                         # Optional: ignore key logic
└── {action}.go                       # Individual action files

pkg/kcp/provider/{provider}/{resource}/
├── new.go                            # Provider entry point: New(StateFactory)
├── state.go                          # Provider state with cloud client
└── {action}.go                       # Provider-specific actions
```

## Circular Dependency Resolution

The reconciler must reference providers (for Switch), but providers must embed reconciler state.

**Solution**: Define state interface in `types/` subpackage.

```go
// pkg/kcp/{resource}/types/state.go
package types

type State interface {
    kcpcommonaction.State  // or focal.State for legacy
    
    // Resource-specific methods
    ObjAs{Resource}() *cloudcontrolv1beta1.{Resource}
    // ... other accessors
}
```

```go
// pkg/kcp/{resource}/state.go
package {resource}

type state struct {
    kcpcommonaction.State  // Embed
    // resource-specific fields
}

// Verify interface compliance
var _ types.State = (*state)(nil)

func (s *state) ObjAs{Resource}() *cloudcontrolv1beta1.{Resource} {
    return s.Obj().(*cloudcontrolv1beta1.{Resource})
}
```

```go
// pkg/kcp/provider/aws/{resource}/state.go
package aws{resource}

type State struct {
    types.State        // Embed interface, not struct
    client  Client     // AWS client
}
```

## Controller Setup Example

**File**: `internal/controller/cloud-control/vpcnetwork_controller.go`

```go
package cloudcontrol

func SetupVpcNetworkReconciler(
    kcpManager manager.Manager,
    awsStateFactory awsvpcnetwork.StateFactory,
    azureStateFactory azurevpcnetwork.StateFactory,
    gcpStateFactory gcpvpcnetwork.StateFactory,
    sapStateFactory sapvpcnetwork.StateFactory,
) error {
    return NewVpcNetworkReconciler(
        vpcnetwork.New(
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

## Reconciler Implementation (kcpcommonaction.State)

**File**: `pkg/kcp/vpcnetwork/reconciler.go`

```go
package vpcnetwork

type VpcNetworkReconciler interface {
    reconcile.Reconciler
}

type reconciler struct {
    composedStateFactory  composed.StateFactory
    kcpCommonStateFactory kcpcommonaction.StateFactory
    awsStateFactory       awsvpcnetwork.StateFactory
    azureStateFactory     azurevpcnetwork.StateFactory
    gcpStateFactory       gcpvpcnetwork.StateFactory
    sapStateFactory       sapvpcnetwork.StateFactory
}

func New(
    composedStateFactory composed.StateFactory,
    kcpCommonStateFactory kcpcommonaction.StateFactory,
    awsStateFactory awsvpcnetwork.StateFactory,
    azureStateFactory azurevpcnetwork.StateFactory,
    gcpStateFactory gcpvpcnetwork.StateFactory,
    sapStateFactory sapvpcnetwork.StateFactory,
) VpcNetworkReconciler {
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
    if Ignore != nil && Ignore.ShouldIgnoreKey(req) {
        return ctrl.Result{}, nil
    }

    kcpCommonState := r.newKcpCommonState(req.NamespacedName)
    action := r.newAction()

    return composed.Handling().
        WithMetrics("vpcnetwork", util.RequestObjToString(req)).
        Handle(action(ctx, kcpCommonState))
}

func (r *reconciler) newKcpCommonState(name types.NamespacedName) kcpcommonaction.State {
    return r.kcpCommonStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.VpcNetwork{}),
    )
}

func (r *reconciler) newAction() composed.Action {
    providerFlow := composed.Switch(
        nil,
        composed.NewCase(kcpcommonaction.AwsProviderPredicate, awsvpcnetwork.New(r.awsStateFactory)),
        composed.NewCase(kcpcommonaction.AzureProviderPredicate, azurevpcnetwork.New(r.azureStateFactory)),
        composed.NewCase(kcpcommonaction.GcpProviderPredicate, gcpvpcnetwork.New(r.gcpStateFactory)),
        composed.NewCase(kcpcommonaction.OpenStackProviderPredicate, sapvpcnetwork.New(r.sapStateFactory)),
    )

    return composed.ComposeActionsNoName(
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.VpcNetwork{}),
        kcpcommonaction.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActionsNoName(
                nameDetermine,

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
                        specCidrBlocksValidate,
                        providerFlow,
                        statusReady,
                    ),
                ),
            )(ctx, newState(st.(kcpcommonaction.State)))
        },
    )
}
```

## Provider Predicates

**kcpcommonaction predicates** (for new pattern):
- `kcpcommonaction.AwsProviderPredicate`
- `kcpcommonaction.GcpProviderPredicate`
- `kcpcommonaction.AzureProviderPredicate`
- `kcpcommonaction.OpenStackProviderPredicate`

**statewithscope predicates** (for legacy focal pattern):
- `statewithscope.AwsProviderPredicate`
- `statewithscope.GcpProviderPredicate`
- `statewithscope.AzureProviderPredicate`
- `statewithscope.OpenStackProviderPredicate`

## Provider State Pattern

**File**: `pkg/kcp/provider/aws/vpcnetwork/state.go`

```go
package awsvpcnetwork

type State struct {
    types.State                        // Embed interface
    client awsvpcnetworkclient.Client  // AWS client
}

type StateFactory interface {
    NewState(ctx context.Context, baseState types.State) (context.Context, *State, error)
}

type stateFactory struct {
    clientProvider awsclient.SkrClientProvider[awsvpcnetworkclient.Client]
}

func NewStateFactory(clientProvider awsclient.SkrClientProvider[awsvpcnetworkclient.Client]) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, baseState types.State) (context.Context, *State, error) {
    roleName := awsutil.RoleArnDefault(baseState.Subscription().Status.SubscriptionInfo.Aws.Account)

    client, err := f.clientProvider(
        ctx,
        baseState.Subscription().Status.SubscriptionInfo.Aws.Account,
        baseState.ObjAsVpcNetwork().Spec.Region,
        awsconfig.AwsConfig.Default.AccessKeyId,
        awsconfig.AwsConfig.Default.SecretAccessKey,
        roleName,
    )
    if err != nil {
        return ctx, nil, err
    }

    return ctx, &State{State: baseState, client: client}, nil
}
```

## Provider New() Pattern

**File**: `pkg/kcp/provider/aws/vpcnetwork/new.go`

```go
package awsvpcnetwork

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

## Individual Action Pattern

```go
// pkg/kcp/provider/aws/vpcnetwork/infraCreateUpdate.go
package awsvpcnetwork

func infraCreateUpdate(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    // Use state.client to interact with AWS
    vpc, err := state.client.DescribeVpc(ctx, vpcId)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error describing VPC", composed.StopWithRequeue, ctx)
    }

    // Update state fields
    state.vpc = vpc

    // Update status
    obj := state.ObjAsVpcNetwork()
    obj.Status.VpcId = *vpc.VpcId

    return composed.PatchStatus(obj).
        SetCondition(metav1.Condition{...}).
        SuccessError(nil).
        Run(ctx, state)
}
```

## Cloud Provider Interaction Styles

### Blocking Style

All provider work done in a single reconciliation. Simpler but longer execution.

```go
func infraCreateUpdate(ctx context.Context, st composed.State) (error, context.Context) {
    // Do all work synchronously
    err := createVpc(ctx, state)
    if err != nil {
        return err, ctx
    }
    err = createSubnets(ctx, state)
    if err != nil {
        return err, ctx
    }
    // ... more operations
    return nil, ctx
}
```

### Progressive Style (Requeue-based)

Uses requeues to poll cloud provider status. Better for long-running operations.

```go
func waitElastiCacheAvailable(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    cluster, err := state.client.DescribeReplicationGroup(ctx, state.clusterId)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error describing cluster", composed.StopWithRequeue, ctx)
    }
    
    if cluster.Status != "available" {
        // Not ready yet, requeue
        return composed.StopWithRequeueDelay(30 * time.Second), nil
    }
    
    // Ready, continue to next action
    return nil, ctx
}
```

## kcpcommonaction.State Reference

**File**: `pkg/kcp/commonAction/state.go`

```go
type State interface {
    composed.State
    ObjAsObjWithStatus() composed.ObjWithStatus
    Subscription() *cloudcontrolv1beta1.Subscription
    VpcNetwork() *cloudcontrolv1beta1.VpcNetwork
    IpRange() *cloudcontrolv1beta1.IpRange
    GcpSubnet() *cloudcontrolv1beta1.GcpSubnet
}
```

**File**: `pkg/kcp/commonAction/new.go`

`kcpcommonaction.New()` loads these resources in order:
1. `composed.LoadObj` - Load the main object
2. `statusStaleProcessing` - Check generation vs observedGeneration
3. `ipRangeLoad` - Load IpRange dependency
4. `gcpSubnetLoad` - Load GcpSubnet dependency
5. `vpcNetworkLoad` - Load VpcNetwork dependency
6. `subscriptionLoad` - Load Subscription
7. `labelObj` - Apply standard labels

## focal.State Reference (LEGACY)

**File**: `pkg/common/actions/focal/state.go`

```go
type State interface {
    composed.State
    Scope() *cloudcontrolv1beta1.Scope
    SetScope(*cloudcontrolv1beta1.Scope)
    ObjAsCommonObj() CommonObject
}
```

**File**: `pkg/common/actions/focal/new.go`

`focal.New()` loads:
1. `composed.LoadObj` - Load the main object
2. `loadScopeFromRef` - Load Scope from ScopeRef field
3. `loadFeatureContext` - Load feature flags

## Example Files in Codebase

| Pattern | Files |
|---------|-------|
| kcpcommonaction reconciler | `pkg/kcp/vpcnetwork/reconciler.go` |
| focal reconciler (legacy) | `pkg/kcp/nfsinstance/reconciler.go` |
| Provider state (AWS) | `pkg/kcp/provider/aws/vpcnetwork/state.go` |
| Provider new (AWS) | `pkg/kcp/provider/aws/vpcnetwork/new.go` |
| Provider state (GCP) | `pkg/kcp/provider/gcp/vpcnetwork/state.go` |
| Multiple clients (GCP) | `pkg/kcp/provider/gcp/iprange/v3/state.go` |
