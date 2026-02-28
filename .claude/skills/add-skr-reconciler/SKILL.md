---
name: add-skr-reconciler
description: Create a new SKR reconciler for user-facing resources. Use when adding resources to cloud-resources API, creating user-facing CRDs, or implementing SKR-to-KCP bridge controllers.
---

# Add SKR Reconciler

Create user-facing reconcilers that bridge SKR (user cluster) resources to KCP (control plane) resources.

## Quick Start

1. Ensure corresponding KCP reconciler exists
2. Scaffold CRD with kubebuilder:
   ```bash
   kubebuilder create api --group cloud-resources --version v1beta1 --kind Gcp<Resource>
   ```
3. Edit CRD in `api/cloud-resources/v1beta1/gcp<resource>_types.go`
4. Run `make manifests && make generate`
5. Create package `pkg/skr/<resource>/`
6. Implement: reconciler.go, createKcp<Resource>.go, updateStatus.go
7. Register in `cmd/main.go`
8. Run `./config/patchAfterMakeManifests.sh && ./config/sync.sh`

## Rules

### MUST
- Have corresponding KCP reconciler first
- Use provider-specific CRD names (`GcpNfsVolume`)
- Check feature flags before execution
- Create KCP resource with proper RemoteRef
- Propagate status from KCP to SKR

### MUST NOT
- Directly call cloud provider APIs (that's KCP's job)
- Skip feature flag checks
- Forget to sync CRDs after API changes

## Execution Flow

```
User creates SKR resource (cloud-resources.kyma-project.io)
    ↓
SKR reconciler creates KCP resource (cloud-control.kyma-project.io)
    ↓
KCP reconciler provisions cloud infrastructure
    ↓
Status propagates: Cloud → KCP → SKR
```

## Package Structure

```
pkg/skr/<resource>/
├── reconciler.go           # Main reconciler
├── createKcp<Resource>.go  # Create KCP resource
├── loadKcp<Resource>.go    # Load KCP resource
├── updateStatus.go         # Sync status from KCP
└── deleteKcp<Resource>.go  # Delete KCP resource
```

## CRD Template

**File**: `api/cloud-resources/v1beta1/gcp<resource>_types.go`

```go
type Gcp<Resource>Spec struct {
    // User-facing fields (no RemoteRef/Scope - SKR handles that)

    // +kubebuilder:validation:Required
    Location string `json:"location"`

    // Resource-specific fields...
}

type Gcp<Resource>Status struct {
    State      string             `json:"state,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Populated from KCP resource
    Id       string `json:"id,omitempty"`
    Endpoint string `json:"endpoint,omitempty"`
}
```

## Reconciler Pattern

```go
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newState(req.NamespacedName)

    action := composed.ComposeActions(
        "gcpResource-main",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpResource{}),
        composed.LoadObj,
        checkFeatureDisabled,
        loadKcpResource,
        composed.IfElse(
            composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions("create",
                createKcpResource,
                waitKcpReady,
                updateStatus,
            ),
            composed.ComposeActions("delete",
                deleteKcpResource,
                waitKcpDeleted,
                removeFinalizer,
                composed.StopAndForgetAction,
            ),
        ),
        composed.StopAndForgetAction,
    )

    return composed.Handle(action(ctx, state))
}
```

## Creating KCP Resource

```go
func createKcpResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    if state.kcpResource != nil {
        return nil, ctx
    }

    skrObj := state.ObjAsGcpResource()

    kcpResource := &cloudcontrolv1beta1.GcpResource{
        ObjectMeta: metav1.ObjectMeta{
            Name:      skrObj.Name,
            Namespace: state.KcpNamespace(),
        },
        Spec: cloudcontrolv1beta1.GcpResourceSpec{
            RemoteRef: cloudcontrolv1beta1.RemoteRef{
                Name:      skrObj.Name,
                Namespace: skrObj.Namespace,
            },
            Scope: cloudcontrolv1beta1.ScopeRef{
                Name: state.Scope().Name,
            },
            // Map SKR fields to KCP fields
        },
    }

    err := state.KcpCluster().Create(ctx, kcpResource)
    if err != nil {
        return err, ctx
    }

    state.kcpResource = kcpResource
    return nil, ctx
}
```

## Status Propagation

```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    skrObj := state.ObjAsGcpResource()

    if state.kcpResource == nil {
        return nil, ctx
    }

    // Copy status from KCP to SKR
    skrObj.Status.State = string(state.kcpResource.Status.State)
    skrObj.Status.Id = state.kcpResource.Status.Id

    // Copy conditions
    for _, cond := range state.kcpResource.Status.Conditions {
        meta.SetStatusCondition(&skrObj.Status.Conditions, cond)
    }

    return composed.UpdateStatus(skrObj).
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

## Feature Flag Check

```go
func checkFeatureDisabled(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.GcpResource.Value(ctx) == feature.GcpResourceDisabled {
        return composed.StopAndForget, nil
    }
    return nil, ctx
}
```

## Post-API Change Commands

```bash
make manifests
make generate
./config/patchAfterMakeManifests.sh
./config/sync.sh
```

## Checklist

- [ ] KCP reconciler exists and works
- [ ] CRD defined in cloud-resources API
- [ ] Feature flag check implemented
- [ ] KCP resource created with RemoteRef
- [ ] Status propagated from KCP to SKR
- [ ] Deletion cleans up KCP resource
- [ ] Ran patch and sync scripts

## Troubleshooting

| Issue | Solution |
|-------|----------|
| SKR resource stuck in Processing | Check KCP resource status |
| KCP resource not created | Verify RemoteRef and Scope |
| Status not updating | Check status propagation action |
| CRD not in SKR cluster | Run sync.sh |

## Related

- Full guide: [docs/agents/guides/ADD_SKR_RECONCILER.md](../../../docs/agents/guides/ADD_SKR_RECONCILER.md)
- KCP reconciler: `/add-kcp-reconciler`
- Feature flags: `/feature-flags`
