# Common Pitfalls and Solutions

**Authority**: MANDATORY for agents to consult when debugging or preventing errors.

**Target**: LLM coding agents working with Cloud Manager reconcilers.

**Related**: [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md) | [CONTROLLER_TESTS.md](../guides/CONTROLLER_TESTS.md)

---

## Authority: Error Prevention Rules

### MUST Prevent
- State type confusion (asserting to wrong state level)
- Missing `StopAndForget` at action flow end
- Silent status update failures (missing error checks)
- Creating resources without existence checks
- Skipping wait actions after async operations
- Using real cloud APIs in tests

### MUST Check Before Submitting
- [ ] State assertions use correct types (`composed.State` → `focal.State` → provider state)
- [ ] All successful flows end with `StopAndForget` or explicit return
- [ ] `UpdateObjStatus()` returns checked for errors
- [ ] Resources loaded before creation attempts
- [ ] Async operations followed by wait actions
- [ ] Tests use `testinfra` mocks exclusively

---

## State Management Pitfalls

### Pitfall 1: State Type Confusion

**Problem**: Asserting state to wrong type causes runtime panic.

**Symptom**: Panic during reconciliation: `interface conversion: composed.State is *gcpsubnet.State, not *focal.State`.

**Cause**: State hierarchy violated - passing provider state where focal state expected.

❌ **BAD - Wrong type assertion**:
```go
func NewStateFactory(baseStateFactory focal.StateFactory) StateFactory {
    return func(ctx context.Context, state composed.State) (*State, error) {
        // ❌ Asserting composed.State directly to *State
        return &State{
            State: state.(*State), // PANIC!
        }, nil
    }
}
```

✅ **GOOD - Correct hierarchy**:
```go
func NewStateFactory(baseStateFactory focal.StateFactory) StateFactory {
    return func(ctx context.Context, state composed.State) (*State, error) {
        // ✅ Assert to focal.State first, then extend
        focalState, err := baseStateFactory.NewState(ctx, state)
        if err != nil {
            return nil, err
        }
        
        return &State{
            State: focalState,
            client: client,
        }, nil
    }
}
```

**State Hierarchy Flow**:
1. `composed.State` (base interface)
2. `focal.State` (adds Scope, Cluster info)
3. Provider state (adds provider-specific fields)

**Validation**: Check state factory initializes from parent state level.

**Location**: State factory in `state.go` files.

---

### Pitfall 2: Mutating State Between Actions

**Problem**: State changes in one action not visible to subsequent actions.

**Symptom**: Action B doesn't see updates made by Action A.

**Cause**: State immutability expectation - each action receives same state.

❌ **BAD - Expecting mutation to persist**:
```go
func loadResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    state.remoteResource = &ResourceFromAPI{}  // ❌ Mutation lost
    return nil, ctx
}

func checkResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    if state.remoteResource == nil {  // ❌ Always nil!
        return errors.New("not found"), ctx
    }
    return nil, ctx
}
```

✅ **GOOD - Set state in factory or use context**:
```go
// Option 1: Set in state factory (preferred for provider state)
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    state := &State{State: focalState}
    
    // ✅ Load during initialization
    remoteResource, err := f.client.Get(ctx, name)
    if err != nil && !IsNotFound(err) {
        return nil, err
    }
    state.remoteResource = remoteResource
    
    return state, nil
}

// Option 2: Use context for passing data
func loadResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    resource, err := state.client.Get(ctx, name)
    if err != nil {
        return err, ctx
    }
    // ✅ Store in context
    ctx = context.WithValue(ctx, "resource", resource)
    return nil, ctx
}

func checkResource(ctx context.Context, st composed.State) (error, context.Context) {
    resource := ctx.Value("resource").(*Resource)  // ✅ Retrieved from context
    if resource == nil {
        return errors.New("not found"), ctx
    }
    return nil, ctx
}
```

**Validation**: State initialization happens in factory, not in actions.

---

## Action Composition Pitfalls

### Pitfall 3: Missing StopAndForget

**Problem**: Actions complete successfully but don't signal reconciliation end.

**Symptom**: Controller requeues indefinitely even when work complete.

**Cause**: No explicit termination signal in action flow.

❌ **BAD - No termination**:
```go
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newState(req)
    
    err := composed.ComposeActions(
        "main",
        loadResource,
        updateResource,
        updateStatus,  // ❌ No StopAndForget - controller requeues!
    )(ctx, state)
    
    return composed.Result(err)
}
```

✅ **GOOD - Explicit termination**:
```go
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newState(req)
    
    err := composed.ComposeActions(
        "main",
        loadResource,
        updateResource,
        updateStatus,
        composed.StopAndForgetAction,  // ✅ Signals success and completion
    )(ctx, state)
    
    return composed.Result(err)
}
```

**Alternative - Return from action**:
```go
func finalAction(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    // ... do final work
    
    return composed.StopAndForget, nil  // ✅ Explicit return
}
```

**Validation**: EVERY successful reconciliation flow MUST end with `StopAndForget`.

---

### Pitfall 4: Wrong Action Ordering

**Problem**: Actions execute before dependencies loaded.

**Symptom**: Nil pointer dereferences, "not found" errors.

**Cause**: Incorrect action sequence - creation before existence check.

❌ **BAD - Create before checking existence**:
```go
composed.ComposeActions(
    "main",
    createResource,     // ❌ Creates even if exists
    loadResource,       // ❌ Too late
    updateStatus,
)
```

✅ **GOOD - Load dependencies first**:
```go
composed.ComposeActions(
    "main",
    loadDependencies,   // ✅ Load Scope, Network, etc.
    loadResource,       // ✅ Check if already exists
    composed.IfElse(
        resourceNotExists,
        createResource,
        updateResource,
    ),
    waitResourceReady,  // ✅ Wait for async completion
    updateStatus,
    composed.StopAndForgetAction,
)
```

**Correct Ordering Pattern**:
1. Load dependencies (Scope, parent resources)
2. Load target resource
3. If not exists → create; else → update
4. Wait for async operations
5. Update status
6. Stop and forget

**Validation**: Actions respect dependency order.

---

### Pitfall 5: Ignoring Async Operations

**Problem**: Proceeding without waiting for cloud resource creation.

**Symptom**: Status shows "Creating" indefinitely, actual resource fails.

**Cause**: Missing wait actions after async create/update.

❌ **BAD - No wait after create**:
```go
composed.ComposeActions(
    "main",
    loadResource,
    composed.IfElse(
        resourceNotExists,
        composed.ComposeActions(
            "create-flow",
            createResource,      // ❌ Returns operation ID immediately
            updateStatusCreating,// ❌ Assumes creation succeeded
        ),
        updateStatusReady,
    ),
    composed.StopAndForgetAction,
)
```

✅ **GOOD - Wait for operation completion**:
```go
composed.ComposeActions(
    "main",
    loadResource,
    composed.IfElse(
        resourceNotExists,
        composed.ComposeActions(
            "create-flow",
            createResource,           // ✅ Starts operation
            waitOperationDone,        // ✅ Polls until complete
            composed.IfElse(
                operationSucceeded,
                updateStatusReady,    // ✅ Only if truly ready
                handleOperationError,
            ),
        ),
        updateStatusReady,
    ),
    composed.StopAndForgetAction,
)
```

**Wait Action Pattern**:
```go
func waitOperationDone(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    op, err := state.client.GetOperation(ctx, state.operationName)
    if err != nil {
        return err, ctx
    }
    
    if !op.Done {
        // ✅ Requeue to check again later
        return composed.StopWithRequeueDelay(3 * time.Second), nil
    }
    
    if op.Error != nil {
        return fmt.Errorf("operation failed: %s", op.Error), ctx
    }
    
    return nil, ctx  // ✅ Continue to next action
}
```

**Validation**: Every create/update/delete followed by wait action.

---

## Status Update Pitfalls

### Pitfall 6: Silent Status Update Failures

**Problem**: Status updates fail silently, user sees stale data.

**Symptom**: Conditions not updated, state stuck at previous value.

**Cause**: Not checking `UpdateObjStatus()` return value.

❌ **BAD - Ignoring errors**:
```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    state.ObjAsGcpSubnet().Status.State = cloudcontrolv1beta1.ReadyState
    state.SetCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionTrue, "Ready", "")
    
    // ❌ Error ignored - status may not persist
    state.UpdateObjStatus(ctx)
    
    return nil, ctx
}
```

✅ **GOOD - Check errors**:
```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    state.ObjAsGcpSubnet().Status.State = cloudcontrolv1beta1.ReadyState
    state.SetCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionTrue, "Ready", "")
    
    // ✅ Propagate errors - triggers requeue
    err := state.UpdateObjStatus(ctx)
    if err != nil {
        return err, ctx
    }
    
    return nil, ctx
}
```

**Validation**: ALWAYS check `UpdateObjStatus()` return value.

---

### Pitfall 7: Missing Condition Updates

**Problem**: Conditions not set when errors occur.

**Symptom**: Status shows no error info, users can't diagnose issues.

**Cause**: Forgetting to set error conditions before returning error.

❌ **BAD - No error condition**:
```go
func createResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    _, err := state.client.Create(ctx, resource)
    if err != nil {
        // ❌ Error returned but status not updated
        return err, ctx
    }
    
    return nil, ctx
}
```

✅ **GOOD - Set error condition**:
```go
func createResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    _, err := state.client.Create(ctx, resource)
    if err != nil {
        // ✅ Set condition before returning error
        state.SetCondition(
            cloudcontrolv1beta1.ConditionTypeError,
            metav1.ConditionTrue,
            "CreateFailed",
            fmt.Sprintf("Failed to create: %s", err),
        )
        
        // ✅ Try to persist error condition (ignore update errors)
        _ = state.UpdateObjStatus(ctx)
        
        return err, ctx
    }
    
    return nil, ctx
}
```

**Condition Setting Pattern**:
- On error: Set `ConditionTypeError` to `True` with message
- On success: Set `ConditionTypeReady` to `True`
- Always persist before returning

**Validation**: Error paths set conditions before returning.

---

## Client Usage Pitfalls

### Pitfall 8: Wrong Client Pattern (GCP)

**Problem**: Using OLD ClientProvider for new resources.

**Symptom**: Complex on-demand client initialization, hard to test.

**Cause**: Not recognizing NEW vs OLD client patterns.

❌ **BAD - OLD pattern for new resource**:
```go
// ❌ Using ClientProvider for new resource
type State struct {
    focal.State
    provider client.ClientProvider
}

func (s *State) computeClient(ctx context.Context) (*compute.Client, error) {
    // ❌ On-demand initialization
    return s.provider.ComputeClient(ctx, s.Scope().Spec.Region)
}
```

✅ **GOOD - NEW pattern with GcpClients**:
```go
// ✅ Using centralized GcpClients
type State struct {
    focal.State
    client SubnetClient  // ✅ Typed interface
}

// In main.go
gcpClients, err := gcpclient.NewGcpClients(ctx, credFile, ...)
subnetClientProvider := subnetclient.NewComputeClientProvider(gcpClients)

// In state factory
state := &State{
    State: focalState,
    client: clientProvider(ctx, focalState.Scope()),  // ✅ Pre-initialized
}
```

**Pattern Decision**:
- **NEW resources** (created after 2024): MUST use `GcpClients`
- **OLD resources** (RedisInstance, IpRange): Continue using `ClientProvider`

**Location**: [GCP_CLIENT_NEW_PATTERN.md](../architecture/GCP_CLIENT_NEW_PATTERN.md)

---

### Pitfall 9: Missing Update Masks (GCP)

**Problem**: Updating GCP resources without specifying changed fields.

**Symptom**: Entire resource replaced, unrelated fields reset.

**Cause**: Not setting `UpdateMask` in update requests.

❌ **BAD - No update mask**:
```go
func updateResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    // ❌ All fields sent - may overwrite unrelated data
    op, err := state.client.PatchSubnet(ctx, &compute.Subnet{
        Name: state.ObjAsGcpSubnet().Spec.Name,
        IpCidrRange: state.ObjAsGcpSubnet().Spec.IpRange,
        // Other fields omitted
    })
    
    return err, ctx
}
```

✅ **GOOD - Specify update mask**:
```go
func updateResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    // ✅ UpdateMask specifies exactly which fields to update
    op, err := state.client.PatchSubnet(ctx, &compute.PatchSubnetRequest{
        Subnet: &compute.Subnet{
            Name: state.ObjAsGcpSubnet().Spec.Name,
            IpCidrRange: state.ObjAsGcpSubnet().Spec.IpRange,
        },
        UpdateMask: &fieldmaskpb.FieldMask{
            Paths: []string{"ipCidrRange"},  // ✅ Only ipCidrRange updated
        },
    })
    
    return err, ctx
}
```

**Validation**: GCP Patch/Update calls include `UpdateMask`.

---

## Testing Pitfalls

### Pitfall 10: Using Real Cloud APIs in Tests

**Problem**: Tests call actual GCP/AWS/Azure APIs.

**Symptom**: Slow tests, flakiness, cloud costs, requires credentials.

**Cause**: Not using `testinfra` mock infrastructure.

❌ **BAD - Real API calls**:
```go
var _ = Describe("GcpSubnet", func() {
    It("Should create subnet", func() {
        // ❌ Calls real GCP API
        client, _ := compute.NewClient(ctx)
        subnet, _ := client.CreateSubnet(ctx, request)
        
        Expect(subnet).NotTo(BeNil())
    })
})
```

✅ **GOOD - Use testinfra mocks**:
```go
var infra testinfra.Infra

var _ = BeforeSuite(func() {
    infra, err = testinfra.Start()  // ✅ Starts mock servers
    Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("GcpSubnet", func() {
    It("Should create subnet", func() {
        subnet := &cloudcontrolv1beta1.GcpSubnet{
            Spec: cloudcontrolv1beta1.GcpSubnetSpec{
                Name: "test-subnet",
                IpRange: "10.0.0.0/24",
            },
        }
        
        // ✅ Uses mock GCP server
        Eventually(CreateGcpSubnet).
            WithArguments(infra.Ctx(), infra.KCP().Client(), subnet).
            Should(Succeed())
        
        // ✅ Verify in mock
        mockSubnet := infra.GcpMock().GetSubnetByName("test-subnet")
        Expect(mockSubnet).NotTo(BeNil())
        Expect(mockSubnet.IpCidrRange).To(Equal("10.0.0.0/24"))
    })
})
```

**testinfra Pattern**:
1. `BeforeSuite`: Initialize `testinfra.Start()`
2. Tests: Use `infra.KCP().Client()` or `infra.SKR().Client()`
3. Assertions: Check mock state with `infra.GcpMock()`, `infra.AwsMock()`, etc.

**Validation**: NO real cloud API calls in controller tests.

**Location**: [CONTROLLER_TESTS.md](../guides/CONTROLLER_TESTS.md)

---

### Pitfall 11: Not Using Eventually() for Async Checks

**Problem**: Synchronous assertions on asynchronous reconciliation.

**Symptom**: Flaky tests - sometimes pass, sometimes fail.

**Cause**: Asserting immediately without waiting for reconciliation.

❌ **BAD - Synchronous assertion**:
```go
It("Should update status", func() {
    CreateGcpSubnet(infra.Ctx(), infra.KCP().Client(), subnet)
    
    // ❌ Reconciliation may not have completed yet
    Expect(subnet.Status.State).To(Equal(cloudcontrolv1beta1.ReadyState))
})
```

✅ **GOOD - Eventually() with timeout**:
```go
It("Should update status", func() {
    Eventually(CreateGcpSubnet).
        WithArguments(infra.Ctx(), infra.KCP().Client(), subnet).
        Should(Succeed())
    
    // ✅ Waits up to 5 seconds for condition
    Eventually(LoadAndCheck).
        WithArguments(
            infra.Ctx(),
            infra.KCP().Client(),
            subnet,
            NewObjActions(HavingState(cloudcontrolv1beta1.ReadyState)),
        ).
        WithTimeout(5 * time.Second).
        WithPolling(200 * time.Millisecond).
        Should(Succeed())
})
```

**Eventually() Rules**:
- ALWAYS use for reconciliation results
- Set reasonable timeout (default: 1 second, typical: 5-10 seconds)
- Set polling interval (typical: 200ms)
- Use `LoadAndCheck` helper for status assertions

**Validation**: ALL reconciliation assertions wrapped in `Eventually()`.

---

## Feature Flag Pitfalls

### Pitfall 12: Missing Feature Flag Checks

**Problem**: Reconciler runs even when feature disabled.

**Symptom**: Resources created when they shouldn't be, feature flag ignored.

**Cause**: Not checking `apiDisabled` or other feature flags.

❌ **BAD - No feature flag check**:
```go
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ❌ Always runs regardless of feature flags
    return composed.Result(r.run(ctx, req))
}
```

✅ **GOOD - Check feature flag early**:
```go
func (r *reconciler) newFlow() composed.Action {
    return composed.ComposeActions(
        "main",
        checkApiDisabled,  // ✅ First action - early exit if disabled
        loadResource,
        // ... rest of flow
    )
}

func checkApiDisabled(ctx context.Context, st composed.State) (error, context.Context) {
    if feature.ApiDisabled.Value(ctx) {
        logger := composed.LoggerFromCtx(ctx)
        logger.Info("API is disabled by feature flag")
        return composed.StopAndForget, nil  // ✅ Exit without error
    }
    return nil, ctx
}
```

**Feature Flag Loading**:
```go
// In reconcile.go SetupWithManager
feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{})
```

**Validation**: ALL reconcilers check `apiDisabled` as first action.

**Location**: [FEATURE_FLAGS.md](../guides/FEATURE_FLAGS.md)

---

## Pattern Confusion Pitfalls

### Pitfall 13: Mixing OLD and NEW Patterns

**Problem**: Using OLD pattern (multi-provider CRD) for new resources instead of NEW pattern (provider-specific CRD).

**Symptom**: Complex state hierarchies for simple resources, unnecessary provider switching.

**Cause**: Following existing OLD pattern code instead of NEW pattern template.

❌ **BAD - OLD pattern for new resource**:
```go
// ❌ Creating multi-provider CRD
type MyNewResource struct {
    Spec MyNewResourceSpec {
        Instance MyNewResourceInstance {
            Gcp   *GcpSpec
            Aws   *AwsSpec
            Azure *AzureSpec
        }
    }
}

// ❌ Shared state layer
type State interface {
    focal.State
    // Shared methods
}

// ❌ Provider switching
composed.BuildSwitchAction("providerSwitch", nil,
    composed.NewCase(GcpProviderPredicate, gcpAction),
    composed.NewCase(AwsProviderPredicate, awsAction),
)
```

✅ **GOOD - NEW pattern with provider-specific CRD**:
```go
// ✅ Provider-specific CRD
type GcpMyNewResource struct {
    Spec GcpMyNewResourceSpec {
        Name string
        // GCP-specific fields only
    }
}

// ✅ Direct state extension
type State struct {
    focal.State
    client     SubnetClient
    remoteResource *gcpResource
}

// ✅ Direct flow, no switching
func (r *reconciler) newFlow() composed.Action {
    return composed.ComposeActions(
        "main",
        loadDependencies,
        loadResource,
        createOrUpdate,
        updateStatus,
        composed.StopAndForgetAction,
    )
}
```

**Pattern Decision Rules**:
- **NEW resources** (created after 2024): MUST use provider-specific CRDs
- **OLD resources** (RedisInstance, NfsInstance, IpRange): Maintain existing pattern
- NEVER mix patterns within single resource

**Location**: [RECONCILER_PATTERN_COMPARISON.md](../architecture/RECONCILER_PATTERN_COMPARISON.md)

---

## Quick Validation Checklist

Before submitting code, verify:

- [ ] State types correctly used (`composed.State` → `focal.State` → provider state)
- [ ] Actions end with `StopAndForget` or appropriate return value
- [ ] Status updates explicitly persisted with `UpdateObjStatus()` error checks
- [ ] Error conditions set when errors occur
- [ ] Resources checked for existence before creation
- [ ] Async operations have corresponding wait actions
- [ ] Feature flags loaded and checked (`apiDisabled` first action)
- [ ] Tests use `testinfra` mocks, not real cloud APIs
- [ ] `Eventually()` wraps all reconciliation assertions
- [ ] Update masks used for GCP resource updates
- [ ] Context passed to all client calls
- [ ] NEW pattern used for new resources (provider-specific CRDs)
- [ ] Tests cover both happy path and error scenarios

---

## Getting Help

**Can't find pattern?** → Check [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

**Need implementation example?** → See [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md)

**Testing issues?** → Consult [CONTROLLER_TESTS.md](../guides/CONTROLLER_TESTS.md)

**Pattern confusion?** → Review [RECONCILER_PATTERN_COMPARISON.md](../architecture/RECONCILER_PATTERN_COMPARISON.md)

**Still stuck?** → Search existing reconcilers in `pkg/kcp/provider/` for similar patterns
