# SKR Reconciler Patterns Reference

This document provides detailed patterns for SKR (cloud-resources) reconcilers.

## Overview

SKR reconcilers run in SAP BTP Kyma Runtimes (user's clusters) and:
- **Do NOT branch by provider** - simpler linear flows
- Sync SKR object spec to corresponding KCP object
- Sync KCP object status back to SKR
- May create local K8s resources (PV, PVC, Secrets)
- Access both KcpCluster and SkrCluster (local)

## Architecture

```
User's Kyma Runtime (SKR)              Kyma Control Plane (KCP)
┌──────────────────────────┐           ┌──────────────────────────┐
│  cloudresourcesv1beta1   │           │  cloudcontrolv1beta1     │
│  AwsRedisInstance        │──spec────▶│  RedisInstance           │
│  GcpNfsVolume            │◀─status───│  NfsInstance             │
│  ...                     │           │  ...                     │
└──────────────────────────┘           └──────────────────────────┘
         │
         │ local actions
         ▼
   PV, PVC, Secrets
```

## Directory Structure

```
internal/controller/cloud-resources/
└── {resource}_controller.go          # Factory, registers with SKR runtime

pkg/skr/{resource}/
├── reconciler.go                     # Main reconciler and newAction()
├── state.go                          # State with KcpCluster + SkrCluster
├── loadKcp{Resource}.go              # Load corresponding KCP object
├── createKcp{Resource}.go            # Create/modify KCP object
├── deleteKcp{Resource}.go            # Delete KCP object
├── updateStatus.go                   # Sync KCP status to SKR
├── wait{Something}.go                # Wait for conditions
└── {localAction}.go                  # Local resource creation (if needed)
```

## Controller Factory Pattern

SKR reconcilers use a factory pattern because the SKR runtime creates new reconciler instances for each Kyma runtime.

**File**: `internal/controller/cloud-resources/{resource}_controller.go`

```go
package cloudresources

type {Resource}ReconcilerFactory struct{}

func (f *{Resource}ReconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
    return &{Resource}Reconciler{
        reconciler: {resource}.NewReconcilerFactory().New(args),
    }
}

type {Resource}Reconciler struct {
    reconciler reconcile.Reconciler
}

func (r *{Resource}Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    return r.reconciler.Reconcile(ctx, req)
}

func Setup{Resource}Reconciler(reg skrruntime.SkrRegistry) error {
    return reg.Register().
        WithFactory(&{Resource}ReconcilerFactory{}).
        For(&cloudresourcesv1beta1.{Resource}{}).
        Complete()
}
```

### With Watches (e.g., for PV/PVC)

```go
func SetupGcpNfsVolumeReconciler(
    reg skrruntime.SkrRegistry,
    fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient],
    env abstractions.Environment,
) error {
    return reg.Register().
        WithFactory(&GcpNfsVolumeReconcilerFactory{...}).
        For(&cloudresourcesv1beta1.GcpNfsVolume{}).
        Watches(&corev1.PersistentVolume{}, gcpnfsvolume.PVEventHandler).
        Complete()
}
```

## State Pattern

SKR state holds references to **both clusters**.

**File**: `pkg/skr/{resource}/state.go`

```go
package {resource}

type State struct {
    composed.State                                    // Base state for SKR object
    KymaRef       klog.ObjectRef                      // Reference to Kyma (scope)
    KcpCluster    composed.StateCluster               // KCP cluster client
    Kcp{Resource} *cloudcontrolv1beta1.{Resource}     // KCP mirror object
    SkrIpRange    *cloudresourcesv1beta1.IpRange      // Related SKR resources
    // ... other fields
}

type StateFactory interface {
    NewState(ctx context.Context, req ctrl.Request) (*State, error)
}

type stateFactory struct {
    baseStateFactory composed.StateFactory
    scopeProvider    scopeprovider.ScopeProvider
    kcpCluster       composed.StateCluster
}

func NewStateFactory(
    baseStateFactory composed.StateFactory,
    scopeProvider scopeprovider.ScopeProvider,
    kcpCluster composed.StateCluster,
) StateFactory {
    return &stateFactory{
        baseStateFactory: baseStateFactory,
        scopeProvider:    scopeProvider,
        kcpCluster:       kcpCluster,
    }
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

// Type-safe accessor
func (s *State) ObjAs{Resource}() *cloudresourcesv1beta1.{Resource} {
    return s.Obj().(*cloudresourcesv1beta1.{Resource})
}
```

## Reconciler Pattern

**File**: `pkg/skr/{resource}/reconciler.go`

```go
package {resource}

type reconciler struct {
    factory *stateFactory
}

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
    return &reconcilerFactory{}
}

type reconcilerFactory struct{}

func (f *reconcilerFactory) New(args skrreconciler.ReconcilerArguments) reconcile.Reconciler {
    return &reconciler{
        factory: &stateFactory{
            baseStateFactory: composed.NewStateFactory(
                composed.NewStateClusterFromCluster(args.SkrCluster),
            ),
            scopeProvider: args.ScopeProvider,
            kcpCluster:    composed.NewStateClusterFromCluster(args.KcpCluster),
        },
    }
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state, err := r.factory.NewState(ctx, req)
    if err != nil {
        return ctrl.Result{}, fmt.Errorf("error creating state: %w", err)
    }
    action := r.newAction()

    return composed.Handling().
        WithMetrics("{resource}", util.RequestObjToString(req)).
        WithNoLog().
        Handle(action(ctx, state))
}

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

## Core Action Patterns

### Load KCP Object

```go
// pkg/skr/{resource}/loadKcp{Resource}.go
func loadKcp{Resource}(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    obj := state.ObjAs{Resource}()

    // Skip if no ID yet (not created)
    if obj.Status.Id == "" {
        return nil, ctx
    }

    kcp{Resource} := &cloudcontrolv1beta1.{Resource}{}
    err := state.KcpCluster.K8sClient().Get(ctx, types.NamespacedName{
        Name:      obj.Status.Id,
        Namespace: state.KymaRef.Namespace,
    }, kcp{Resource})

    if apierrors.IsNotFound(err) {
        return nil, ctx  // Not found, will be created
    }
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error loading KCP {Resource}", composed.StopWithRequeue, ctx)
    }

    state.Kcp{Resource} = kcp{Resource}
    logger.Info("KCP {Resource} loaded")

    return nil, ctx
}
```

### Create KCP Object (Spec Projection)

```go
// pkg/skr/{resource}/createKcp{Resource}.go
func createKcp{Resource}(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    if state.Kcp{Resource} != nil {
        return nil, ctx  // Already exists
    }

    skrObj := state.ObjAs{Resource}()

    // Generate ID if needed
    if skrObj.Status.Id == "" {
        skrObj.Status.Id = uuid.NewString()
        err := state.UpdateObjStatus(ctx)
        if err != nil {
            return composed.LogErrorAndReturn(err, "Error updating SKR status with ID", composed.StopWithRequeue, ctx)
        }
    }

    // Project SKR spec to KCP object
    state.Kcp{Resource} = &cloudcontrolv1beta1.{Resource}{
        ObjectMeta: metav1.ObjectMeta{
            Name:      skrObj.Status.Id,
            Namespace: state.KymaRef.Namespace,
            Labels: map[string]string{
                cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
                cloudcontrolv1beta1.LabelRemoteName:      skrObj.Name,
                cloudcontrolv1beta1.LabelRemoteNamespace: skrObj.Namespace,
            },
        },
        Spec: cloudcontrolv1beta1.{Resource}Spec{
            RemoteRef: cloudcontrolv1beta1.RemoteRef{
                Namespace: skrObj.Namespace,
                Name:      skrObj.Name,
            },
            Scope: cloudcontrolv1beta1.ScopeRef{
                Name: state.KymaRef.Name,
            },
            // Map SKR spec fields to KCP spec
            // ...
        },
    }

    err := state.KcpCluster.K8sClient().Create(ctx, state.Kcp{Resource})
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error creating KCP {Resource}", composed.StopWithRequeue, ctx)
    }

    logger.Info("KCP {Resource} created")

    // Update SKR status to Creating
    skrObj.Status.State = cloudresourcesv1beta1.StateCreating
    return composed.UpdateStatus(skrObj).
        SuccessError(composed.StopWithRequeue).
        Run(ctx, state)
}
```

### Update Status (KCP → SKR)

```go
// pkg/skr/{resource}/updateStatus.go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    skrObj := state.ObjAs{Resource}()

    // Read KCP conditions
    kcpCondErr := meta.FindStatusCondition(state.Kcp{Resource}.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
    kcpCondReady := meta.FindStatusCondition(state.Kcp{Resource}.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady)

    // Read SKR conditions
    skrCondErr := meta.FindStatusCondition(skrObj.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
    skrCondReady := meta.FindStatusCondition(skrObj.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

    // Propagate Error condition
    if kcpCondErr != nil && skrCondErr == nil {
        skrObj.Status.State = cloudresourcesv1beta1.StateError
        return composed.UpdateStatus(skrObj).
            SetCondition(metav1.Condition{
                Type:    cloudresourcesv1beta1.ConditionTypeError,
                Status:  metav1.ConditionTrue,
                Reason:  cloudresourcesv1beta1.ConditionReasonError,
                Message: kcpCondErr.Message,
            }).
            RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
            ErrorLogMessage("Error updating SKR status with error condition").
            SuccessError(composed.StopAndForget).
            Run(ctx, state)
    }

    // Propagate Ready condition
    if kcpCondReady != nil && skrCondReady == nil {
        logger.Info("Updating SKR status with Ready condition")
        skrObj.Status.State = cloudresourcesv1beta1.StateReady
        return composed.UpdateStatus(skrObj).
            SetCondition(metav1.Condition{
                Type:    cloudresourcesv1beta1.ConditionTypeReady,
                Status:  metav1.ConditionTrue,
                Reason:  cloudresourcesv1beta1.ConditionTypeReady,
                Message: kcpCondReady.Message,
            }).
            RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
            ErrorLogMessage("Error updating SKR status with ready condition").
            SuccessError(composed.StopWithRequeue).
            Run(ctx, state)
    }

    return nil, ctx
}
```

### Delete KCP Object

```go
// pkg/skr/{resource}/deleteKcp{Resource}.go
func deleteKcp{Resource}(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    if state.Kcp{Resource} == nil {
        return nil, ctx  // Already deleted
    }

    err := state.KcpCluster.K8sClient().Delete(ctx, state.Kcp{Resource})
    if apierrors.IsNotFound(err) {
        return nil, ctx  // Already deleted
    }
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error deleting KCP {Resource}", composed.StopWithRequeue, ctx)
    }

    logger.Info("KCP {Resource} deleted")

    // Requeue to verify deletion
    return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
```

## Local Resource Creation Patterns

### Create PersistentVolume

```go
// pkg/skr/gcpnfsvolume/createPersistenceVolume.go
func createPersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    if composed.MarkedForDeletionPredicate(ctx, st) {
        return nil, ctx
    }

    nfsVolume := state.ObjAsGcpNfsVolume()

    // Skip if not ready
    if !meta.IsStatusConditionTrue(nfsVolume.Status.Conditions, v1beta1.ConditionTypeReady) {
        return nil, ctx
    }

    // Skip if already exists
    if state.PV != nil {
        return nil, ctx
    }

    // Create PV on SKR (local cluster)
    pv := &corev1.PersistentVolume{
        ObjectMeta: metav1.ObjectMeta{
            Name:       getVolumeName(nfsVolume),
            Labels:     getVolumeLabels(nfsVolume),
            Finalizers: []string{api.CommonFinalizerDeletionHook},
        },
        Spec: corev1.PersistentVolumeSpec{
            Capacity: corev1.ResourceList{
                "storage": *gcpNfsVolumeCapacityToResourceQuantity(nfsVolume),
            },
            AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
            StorageClassName: "",
            PersistentVolumeSource: corev1.PersistentVolumeSource{
                NFS: &corev1.NFSVolumeSource{
                    Server: nfsVolume.Status.Hosts[0],  // From KCP status
                    Path:   fmt.Sprintf("/%s", nfsVolume.Spec.FileShareName),
                },
            },
        },
    }

    // Use local cluster client (not KcpCluster)
    err := state.Cluster().K8sClient().Create(ctx, pv)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error creating PV", composed.StopWithRequeue, ctx)
    }

    logger.Info("PV created")
    return composed.StopWithRequeueDelay(3 * util.Timing.T1000ms()), ctx
}
```

### Create Auth Secret (for Redis)

```go
// pkg/skr/awsredisinstance/createAuthSecret.go
func createAuthSecret(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)

    if state.AuthSecret != nil {
        return nil, ctx  // Already exists
    }

    // Create secret with connection info from KCP status
    secret := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Namespace:  state.Obj().GetNamespace(),
            Name:       getAuthSecretName(state.ObjAsAwsRedisInstance()),
            Labels:     getAuthSecretLabels(state.ObjAsAwsRedisInstance()),
            Finalizers: []string{api.CommonFinalizerDeletionHook},
        },
        Data: state.GetAuthSecretData(),  // Extract from KCP status
    }

    err := state.Cluster().K8sClient().Create(ctx, secret)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error creating auth secret", composed.StopWithRequeue, ctx)
    }

    logger.Info("Auth secret created")
    return nil, ctx
}

// Extract auth data from KCP RedisInstance status
func getAuthSecretBaseData(kcpRedis *cloudcontrolv1beta1.RedisInstance) map[string][]byte {
    result := map[string][]byte{}

    if len(kcpRedis.Status.PrimaryEndpoint) > 0 {
        result["primaryEndpoint"] = []byte(kcpRedis.Status.PrimaryEndpoint)
        splitEndpoint := strings.Split(kcpRedis.Status.PrimaryEndpoint, ":")
        if len(splitEndpoint) >= 2 {
            result["host"] = []byte(splitEndpoint[0])
            result["port"] = []byte(splitEndpoint[1])
        }
    }

    if len(kcpRedis.Status.AuthString) > 0 {
        result["authString"] = []byte(kcpRedis.Status.AuthString)
    }

    return result
}
```

## Differences from KCP Reconcilers

| Aspect | SKR | KCP |
|--------|-----|-----|
| Provider branching | NO | YES (Switch by provider) |
| State factories | One | Multiple (per provider) |
| Cluster access | SKR (local) + KCP | KCP only |
| Cloud API calls | NO (done by KCP) | YES |
| Local K8s resources | YES (PV, PVC, Secrets) | NO |
| Complexity | Linear pipeline | Branched by provider |

## Example Files in Codebase

| Pattern | Files |
|---------|-------|
| SKR Controller | `internal/controller/cloud-resources/gcpnfsvolume_controller.go` |
| SKR Reconciler | `pkg/skr/gcpnfsvolume/reconciler.go` |
| SKR State | `pkg/skr/awsredisinstance/state.go` |
| Load KCP | `pkg/skr/awsredisinstance/loadKcpRedisInstance.go` |
| Create KCP | `pkg/skr/awsredisinstance/createKcpRedisInstance.go` |
| Update Status | `pkg/skr/awsredisinstance/updateStatus.go` |
| Create PV | `pkg/skr/gcpnfsvolume/createPersistenceVolume.go` |
| Create PVC | `pkg/skr/gcpnfsvolume/createPersistentVolumeClaim.go` |
| Create Secret | `pkg/skr/awsredisinstance/createAuthSecret.go` |
