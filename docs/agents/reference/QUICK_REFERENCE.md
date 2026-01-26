# Quick Reference

**Authority**: REQUIRED for fast lookup during coding. All tables are deterministic.

**Target**: LLM coding agents needing immediate answers while implementing Cloud Manager reconcilers.

**Related**: [COMMON_PITFALLS.md](COMMON_PITFALLS.md) | [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) | [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md)

---

## Authority: Quick Lookup Rules

### MUST Use For
- Action return values during composition
- Predicate functions for conditionals
- State method signatures
- Test helper patterns
- Make command invocation

### MUST NOT
- Guess action signatures - consult tables
- Assume predicate names - verify here
- Skip error checking patterns
- Use synchronous assertions in tests

---

## Action Return Values

| Return Value | Effect | Requeue | Use Case |
|-------------|--------|---------|----------|
| `nil, ctx` | Continue to next action | No | Normal flow progression |
| `error, ctx` | Stop execution, log error | Yes (exponential backoff) | Unexpected failure |
| `composed.StopAndForget, nil` | Stop execution, success | No | Reconciliation complete |
| `composed.StopWithRequeue, nil` | Stop execution, success | Yes (immediate) | Check status again soon |
| `composed.StopWithRequeueDelay(duration), nil` | Stop execution, success | Yes (after duration) | Wait for async operation |

**Rule**: ALWAYS end successful flows with `StopAndForget`.

---

## Common Predicates

| Predicate | Purpose | Returns True When |
|-----------|---------|------------------|
| `composed.MarkedForDeletionPredicate` | Check deletion timestamp | `.DeletionTimestamp != nil` |
| `composed.Not(predicate)` | Negate predicate | Inner predicate returns false |
| `resourceExists(ctx, st)` | Check remote resource loaded | `state.remoteResource != nil` |
| `resourceNotExists(ctx, st)` | Check remote resource missing | `state.remoteResource == nil` |
| `resourceReady(ctx, st)` | Check resource ready | `state.remoteResource.Status == "READY"` |
| `operationSucceeded(ctx, st)` | Check operation completed | `state.operation.Done && state.operation.Error == nil` |
| `scopeNotReady(ctx, st)` | Check Scope not ready | `!meta.IsStatusConditionTrue(scope.Status.Conditions, "Ready")` |

**Rule**: Predicates NEVER modify state, only inspect.

---

## State Methods (Hierarchy)

### composed.State Interface

| Method | Returns | Purpose |
|--------|---------|---------|
| `Obj()` | `client.Object` | Get underlying K8s resource |
| `LoadObj(ctx)` | `error` | Reload resource from cluster |
| `UpdateObj(ctx)` | `error` | Update resource spec in cluster |
| `UpdateObjStatus(ctx)` | `error` | Update resource status in cluster |
| `K8sClient()` | `client.Client` | Get Kubernetes client |
| `SetCondition(t, s, r, m)` | - | Set status condition |

### focal.State Interface (extends composed.State)

| Method | Returns | Purpose |
|--------|---------|---------|
| `Scope()` | `*cloudcontrolv1beta1.Scope` | Get cloud provider config |
| `Cluster()` | `client.Client` | Get cluster client (SKR or KCP) |
| `Name()` | `types.NamespacedName` | Get resource name/namespace |

### Provider State (extends focal.State)

```go
type State struct {
    focal.State
    client         SubnetClient       // Provider-specific client
    remoteResource *compute.Subnet    // Remote cloud resource
    operation      *compute.Operation // Async operation tracking
}

// Typed getter (replace GcpSubnet with your resource)
func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
    return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}
```

**Rule**: ALWAYS use typed getters, NOT raw `.Obj().(*Type)`.

---

## Action Composition Patterns

### Sequential Actions

```go
composed.ComposeActions(
    "action-flow-name",
    firstAction,
    secondAction,
    thirdAction,
    composed.StopAndForgetAction,
)
```

### Conditional Execution

```go
composed.IfElse(
    predicateFunction,
    actionIfTrue,
    actionIfFalse,
)
```

### Nested Composition

```go
composed.ComposeActions(
    "main",
    loadResource,
    composed.IfElse(
        resourceNotExists,
        composed.ComposeActions(
            "create-flow",
            createResource,
            waitCreationDone,
            updateStatusReady,
        ),
        composed.ComposeActions(
            "update-flow",
            updateResource,
            waitUpdateDone,
            updateStatusReady,
        ),
    ),
    composed.StopAndForgetAction,
)
```

### Early Exit

```go
func checkFeatureDisabled(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil  // Early exit
    }
    return nil, ctx  // Continue
}
```

---

## Test Helpers (Ginkgo/Gomega)

### Test Structure

```go
var infra testinfra.Infra

var _ = BeforeSuite(func() {
    infra, err = testinfra.Start()
    Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Feature", func() {
    It("Should do something", func() {
        // Test body
    })
})

var _ = AfterSuite(func() {
    err = infra.Stop()
    Expect(err).NotTo(HaveOccurred())
})
```

### Eventually Pattern

```go
Eventually(LoadAndCheck).
    WithArguments(
        infra.Ctx(),
        infra.KCP().Client(),
        resource,
        NewObjActions(HavingState(cloudcontrolv1beta1.ReadyState)),
    ).
    WithTimeout(5 * time.Second).
    WithPolling(200 * time.Millisecond).
    Should(Succeed())
```

### Custom Matchers

| Matcher | Purpose | Example |
|---------|---------|---------|
| `HavingState(state)` | Check `.Status.State` | `HavingState(cloudcontrolv1beta1.ReadyState)` |
| `HavingConditionTrue(type)` | Check condition true | `HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)` |
| `HavingDeletionTimestamp()` | Check deletion timestamp set | `HavingDeletionTimestamp()` |
| `HavingFinalizer(name)` | Check finalizer present | `HavingFinalizer(cloudcontrolv1beta1.FinalizerName)` |

### LoadAndCheck Helper

```go
func LoadAndCheck(
    ctx context.Context,
    client client.Client,
    obj client.Object,
    checks ...ObjAction,
) error {
    err := client.Get(ctx, types.NamespacedName{
        Name:      obj.GetName(),
        Namespace: obj.GetNamespace(),
    }, obj)
    if err != nil {
        return err
    }
    
    for _, check := range checks {
        if err := check(obj); err != nil {
            return err
        }
    }
    
    return nil
}
```

---

## Status Condition Helpers

### Setting Conditions

```go
// Success
state.SetCondition(
    cloudcontrolv1beta1.ConditionTypeReady,
    metav1.ConditionTrue,
    "Ready",
    "Resource is ready",
)

// Error
state.SetCondition(
    cloudcontrolv1beta1.ConditionTypeError,
    metav1.ConditionTrue,
    "CreateFailed",
    fmt.Sprintf("Failed to create: %s", err),
)

// Processing
state.SetCondition(
    cloudcontrolv1beta1.ConditionTypeProcessing,
    metav1.ConditionTrue,
    "Creating",
    "Resource is being created",
)
```

### Checking Conditions

```go
import "k8s.io/apimachinery/pkg/api/meta"

// Check if condition exists and is True
if meta.IsStatusConditionTrue(obj.Status.Conditions, "Ready") {
    // Resource is ready
}

// Get condition
condition := meta.FindStatusCondition(obj.Status.Conditions, "Error")
if condition != nil && condition.Status == metav1.ConditionTrue {
    // Error condition present
}
```

---

## Feature Flags

### Loading Context

```go
// In reconcile.go SetupWithManager
feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{})
```

### Checking Flags

```go
// Check if API disabled
if feature.ApiDisabled.Value(ctx) {
    return composed.StopAndForget, nil
}

// Check custom flag
if feature.CustomFlag.Value(ctx) {
    // Custom behavior
}
```

### Flag Locations

| File | Purpose |
|------|---------|
| `pkg/feature/ff_ga.yaml` | Generally Available features (stable) |
| `pkg/feature/ff_edge.yaml` | Edge/experimental features |
| `config/featureToggles/featureToggles.local.yaml` | Local development overrides |

---

## Key Directories

| Path | Purpose | When to Use |
|------|---------|-------------|
| `api/cloud-control/v1beta1/` | KCP CRD definitions | Adding/modifying KCP resources |
| `api/cloud-resources/v1beta1/` | SKR CRD definitions | Adding/modifying SKR resources |
| `pkg/kcp/provider/<provider>/<resource>/` | KCP reconcilers (NEW pattern) | Implementing cloud resource reconciliation |
| `pkg/skr/<resource>/` | SKR reconcilers | Implementing user-facing reconciliation |
| `pkg/composed/` | Action composition framework | Understanding action patterns |
| `pkg/common/actions/` | Shared actions | Reusing common reconciliation logic |
| `pkg/common/actions/focal/` | Scope management | Loading Scope in KCP reconcilers |
| `pkg/testinfra/` | Test infrastructure | Writing controller tests |
| `pkg/kcp/provider/<provider>/mock/` | Mock implementations | Creating test mocks |
| `internal/controller/` | Controller tests | Writing/running tests |
| `config/crd/bases/` | Generated CRDs | Reviewing CRD manifests |

---

## Make Commands

| Command | Purpose | When to Run |
|---------|---------|-------------|
| `make manifests` | Generate CRDs, RBAC, webhooks | After modifying API types |
| `make generate` | Run code generators | After modifying interfaces |
| `make test` | Run unit tests | Before committing |
| `make build` | Compile manager binary | Testing compilation |
| `make install` | Install CRDs to cluster | Local testing |
| `make run` | Run controller locally | Development/debugging |
| `make docker-build` | Build container image | Before deployment |
| `make envtest` | Setup test environment | Before running tests |

**Post-manifests Workflow**:
1. `make manifests`
2. `./config/patchAfterMakeManifests.sh`
3. `./config/sync.sh`

---

## File Naming Conventions

| Pattern | Example | Purpose |
|---------|---------|---------|
| `<resource>_types.go` | `gcpsubnet_types.go` | CRD type definition |
| `reconcile.go` | `reconcile.go` | Reconciler setup, SetupWithManager |
| `state.go` | `state.go` | State struct and factory |
| `load<Resource>.go` | `loadSubnet.go` | Load remote resource action |
| `create<Resource>.go` | `createSubnet.go` | Create resource action |
| `update<Resource>.go` | `updateSubnet.go` | Update resource action |
| `delete<Resource>.go` | `deleteSubnet.go` | Delete resource action |
| `wait<Resource><State>.go` | `waitSubnetReady.go` | Wait for state transition action |
| `updateStatus.go` | `updateStatus.go` | Update status action |
| `<resource>_test.go` | `gcpsubnet_test.go` | Controller tests |
| `<resource>_gcp_test.go` | `redisinstance_gcp_test.go` | Provider-specific controller tests |

---

## Pattern Recognition Cheatsheet

### Identifying NEW vs OLD Pattern

| Characteristic | NEW Pattern | OLD Pattern |
|---------------|-------------|-------------|
| **CRD Name** | Provider-specific (`GcpSubnet`) | Multi-provider (`RedisInstance`) |
| **CRD Spec** | Direct fields | `.Spec.Instance.Gcp/Aws/Azure` |
| **State Hierarchy** | 2 layers (focal → provider) | 3 layers (focal → shared → provider) |
| **Location** | `pkg/kcp/provider/<provider>/<resource>/` | `pkg/kcp/<resource>/` + `pkg/kcp/provider/` |
| **Reconciler** | Direct action composition | Provider switching with `BuildSwitchAction` |
| **When to Use** | All new resources (post-2024) | Only maintain existing (RedisInstance, IpRange) |

### Identifying GCP Client Pattern

| Characteristic | NEW Pattern | OLD Pattern |
|---------------|-------------|-------------|
| **Initialization** | `NewGcpClients()` in main.go | `ClientProvider` on-demand |
| **State Field** | Typed client interface | `ClientProvider` interface |
| **Client Access** | Direct method call | `.provider.ComputeClient(ctx, region)` |
| **When to Use** | All new GCP resources | Only maintain existing (RedisInstance) |

---

## Context and Logging

### Adding Values to Context

```go
ctx = context.WithValue(ctx, "key", value)
value := ctx.Value("key").(Type)
```

### Logging

```go
logger := composed.LoggerFromCtx(ctx)
logger.Info("Message", "key", value)
logger.Error(err, "Error message", "key", value)

// With values
ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("resource", name))
```

---

## Kubebuilder Markers

### CRD Validation

```go
// +kubebuilder:validation:Required
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
// +kubebuilder:validation:Enum=value1;value2;value3
```

### Status Subresource

```go
// +kubebuilder:subresource:status
type GcpSubnet struct {
    // ...
}
```

### RBAC

```go
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=gcpsubnets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=gcpsubnets/status,verbs=get;update;patch
```

---

## Error Handling Pattern

```go
func actionName(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    // Validate prerequisites
    if state.remoteResource == nil {
        return fmt.Errorf("resource not loaded"), ctx
    }
    
    // Perform operation
    result, err := state.client.DoSomething(ctx, request)
    if err != nil {
        // Set error condition
        state.SetCondition(
            cloudcontrolv1beta1.ConditionTypeError,
            metav1.ConditionTrue,
            "OperationFailed",
            fmt.Sprintf("Failed: %s", err),
        )
        
        // Try to persist status
        _ = state.UpdateObjStatus(ctx)
        
        // Return error (triggers requeue)
        return err, ctx
    }
    
    // Store result in state or context
    state.operation = result
    
    // Continue to next action
    return nil, ctx
}
```

---

## Decision Trees

### Should I Create New Action or Reuse?

```
Is the logic specific to one resource type?
├─ YES: Create in resource package (pkg/kcp/provider/<provider>/<resource>/)
└─ NO: Is it used by multiple resources?
   ├─ YES: Is it provider-specific?
   │  ├─ YES: Create in pkg/kcp/provider/<provider>/common/
   │  └─ NO: Create in pkg/common/actions/
   └─ NO: Keep in resource package
```

### Which State Type Should I Use?

```
Where is this code executing?
├─ Action function: Use provider state (type State struct)
├─ State factory: Return provider state, accept focal.State
├─ Reconciler Reconcile(): Use composed.State
└─ Helper function: Accept appropriate level (usually provider state)
```

### Should I Use Eventually()?

```
Is this assertion checking reconciliation result?
├─ YES: MUST use Eventually()
│  └─ Set timeout (5-10 seconds typical)
│
└─ NO: Is it checking immediate operation?
   ├─ YES: Direct assertion OK
   └─ NO: Use Eventually() (safer)
```

---

## Getting Help

**Can't find specific pitfall?** → [COMMON_PITFALLS.md](COMMON_PITFALLS.md)

**Need full navigation?** → [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)

**Creating new reconciler?** → [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md)

**Writing tests?** → [CONTROLLER_TESTS.md](../guides/CONTROLLER_TESTS.md)

**Feature flags?** → [FEATURE_FLAGS.md](../guides/FEATURE_FLAGS.md)
