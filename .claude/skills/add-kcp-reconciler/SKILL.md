---
name: add-kcp-reconciler
description: Create a new KCP reconciler for cloud provider resources (GCP, Azure, AWS). Use when adding a new CRD to cloud-control API, implementing cloud resource provisioning, or creating a new controller in pkg/kcp/provider/.
---

# Add KCP Reconciler

Create cloud provider reconcilers that provision infrastructure resources (GCP/Azure/AWS).

## Quick Start

1. Scaffold CRD with kubebuilder:
   ```bash
   kubebuilder create api --group cloud-control --version v1beta1 --kind Gcp<Resource>
   ```
2. Edit CRD in `api/cloud-control/v1beta1/gcp<resource>_types.go`
3. Run `make manifests && make generate`
4. Create package `pkg/kcp/provider/<provider>/<resource>/`
5. Implement: state.go, reconcile.go, client/, actions
6. Register in `cmd/main.go`
7. Write tests in `internal/controller/cloud-control/`

## Rules

### MUST
- Use NEW Pattern (provider-specific CRD like `GcpSubnet`)
- Extend `focal.State` directly (no intermediate layer)
- Place code in `pkg/kcp/provider/<provider>/<resource>/`
- One action per file
- Add finalizer for resources requiring cleanup
- End successful flows with `StopAndForgetAction`
- Handle 404 as success (resource doesn't exist yet)
- Add wait actions for async cloud operations

### MUST NOT
- Use OLD Pattern (multi-provider CRDs like `RedisInstance`)
- Create clients on-demand (use provider pattern)
- Skip error handling or status updates
- Forget finalizer removal on deletion

## Package Structure

```
pkg/kcp/provider/gcp/<resource>/
├── reconcile.go          # Reconciler, SetupWithManager
├── state.go              # State struct + factory
├── client/client.go      # Cloud API client interface
├── load<Resource>.go     # Load remote resource
├── create<Resource>.go   # Create resource
├── delete<Resource>.go   # Delete resource
├── wait<State>.go        # Wait for async operation
└── updateStatus.go       # Status update action
```

## CRD Template

**File**: `api/cloud-control/v1beta1/gcp<resource>_types.go`

```go
type Gcp<Resource>Spec struct {
    // +kubebuilder:validation:Required
    RemoteRef RemoteRef `json:"remoteRef"`

    // +kubebuilder:validation:Required
    Scope ScopeRef `json:"scope"`

    // Provider-specific fields...
}

type Gcp<Resource>Status struct {
    State      StatusState        `json:"state,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    Id         string             `json:"id,omitempty"`
}

func (in *Gcp<Resource>) ScopeRef() ScopeRef { return in.Spec.Scope }
```

## State Template

```go
type State struct {
    focal.State
    client         <Resource>Client
    remoteResource *<CloudType>
}

func (s *State) ObjAsGcp<Resource>() *cloudcontrolv1beta1.Gcp<Resource> {
    return s.Obj().(*cloudcontrolv1beta1.Gcp<Resource>)
}
```

## Action Pattern

```go
func load<Resource>(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    obj := state.ObjAsGcp<Resource>()

    if obj.Status.Id == "" {
        return nil, ctx  // Not created yet
    }

    resource, err := state.client.Get(ctx, obj.Status.Id)
    if err != nil {
        if IsNotFound(err) {
            return nil, ctx  // 404 is OK
        }
        return err, ctx
    }

    state.remoteResource = resource
    return nil, ctx
}
```

## Reconciler Flow

```go
composed.ComposeActions(
    "main",
    actions.AddCommonFinalizer(),
    load<Resource>,
    composed.IfElse(
        composed.Not(composed.MarkedForDeletionPredicate),
        composed.ComposeActions("create-update",
            create<Resource>,
            wait<Resource>Ready,
            updateStatus,
        ),
        composed.ComposeActions("delete",
            delete<Resource>,
            wait<Resource>Deleted,
            actions.RemoveCommonFinalizer(),
            composed.StopAndForgetAction,
        ),
    ),
    composed.StopAndForgetAction,
)
```

## Registration (main.go)

```go
stateFactory := <resource>.NewStateFactory(
    <client>.NewClientProvider(gcpClients),
)

if err = <resource>.NewReconciler(
    composedStateFactory,
    focalStateFactory,
    stateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller")
    os.Exit(1)
}
```

## Checklist

- [ ] CRD name is provider-specific (`GcpSubnet` not `Subnet`)
- [ ] RemoteRef and Scope fields present
- [ ] ScopeRef() method implemented
- [ ] `make manifests && make generate` successful
- [ ] State extends focal.State directly
- [ ] Client uses provider pattern (not on-demand)
- [ ] Actions handle 404 as success
- [ ] Finalizer added and removed correctly
- [ ] Flow ends with StopAndForgetAction
- [ ] Tests written with mocks

## Troubleshooting

| Issue | Solution |
|-------|----------|
| CRD not generated | Run `make manifests && make generate` |
| State creation fails | Check client provider in main.go |
| Reconciler never completes | Add `StopAndForgetAction` at end |
| Status not updating | Call `state.UpdateObjStatus(ctx)` |

## Related

- Full guide: [docs/agents/guides/ADD_KCP_RECONCILER.md](../../../docs/agents/guides/ADD_KCP_RECONCILER.md)
- Architecture: [docs/agents/architecture/RECONCILER_NEW_PATTERN.md](../../../docs/agents/architecture/RECONCILER_NEW_PATTERN.md)
- Actions: [docs/agents/architecture/ACTION_COMPOSITION.md](../../../docs/agents/architecture/ACTION_COMPOSITION.md)
- GCP client: `/gcp-client`
- Testing: `/write-tests`
