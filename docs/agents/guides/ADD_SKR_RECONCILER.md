# Guide: Adding New SKR Reconciler

**Target Audience**: LLM coding agents  
**Prerequisites**: [ADD_KCP_RECONCILER.md](ADD_KCP_RECONCILER.md), [STATE_PATTERN.md](../architecture/STATE_PATTERN.md), [ACTION_COMPOSITION.md](../architecture/ACTION_COMPOSITION.md)  
**Purpose**: Step-by-step guide to create user-facing SKR reconciler that maps to KCP resources  
**Context**: SKR reconcilers run in user clusters, create/manage corresponding KCP resources in control plane

## Authority: SKR Reconciler Requirements

### MUST

- **ALWAYS extend `common.State`**: SKR state pattern uses common base, not focal.State
- **MUST generate unique ID**: Use UUID-based ID in Status.Id before creating KCP resource
- **MUST set KCP annotations**: LabelKymaName, LabelRemoteName, LabelRemoteNamespace REQUIRED on all KCP resources
- **MUST sync status from KCP**: Copy State, Conditions, and user-relevant fields from KCP to SKR status
- **MUST wait for KCP deletion**: Delete flow waits for KCP resource deletion BEFORE removing finalizer
- **MUST use NEW pattern**: NEW Pattern (provider-specific SKR → provider-specific KCP) REQUIRED for all new code
- **MUST add finalizer**: Required for resources needing deletion cleanup
- **MUST implement deletion flow**: Delete KCP resource, wait for deletion, remove finalizer

### MUST NOT

- **NEVER use OLD pattern**: OLD Pattern (provider-specific SKR → multi-provider KCP) FORBIDDEN for new code
- **NEVER create KCP without ID**: Status.Id MUST be set before creating KCP resource
- **NEVER skip KCP wait on delete**: Removing finalizer before KCP deletion causes orphaned cloud resources
- **NEVER mutate KCP spec directly**: SKR reconciler creates/deletes KCP resources, never modifies spec after creation
- **NEVER hardcode Kyma namespace**: Always use state.KymaRef.Namespace from reconciler configuration

### ALWAYS

- **ALWAYS end with StopAndForget**: Successful flows return `composed.StopAndForget, nil`
- **ALWAYS check feature flags**: Use `feature.LoadFeatureContextFromObj()` in action composition
- **ALWAYS use type-safe getters**: Use `state.ObjAsGcpRedisCluster()` not `state.Obj().(*Type)`
- **ALWAYS validate state factory errors**: Check error from `NewState()` before proceeding

### NEVER

- **NEVER sync from SKR to KCP after creation**: KCP spec is set ONCE during creation, never updated
- **NEVER create multiple KCP resources**: One SKR resource maps to EXACTLY one KCP resource
- **NEVER skip error handling**: All API calls MUST check errors and return appropriate error+context

## Pattern Selection: NEW vs OLD

### Decision Rule

```
Are you creating NEW SKR reconciler (post-2024)?
├─ YES: MUST use NEW Pattern
│  ├─ Provider-specific SKR → Provider-specific KCP
│  ├─ Example: GcpRedisCluster (SKR) → GcpRedisCluster (KCP)
│  └─ 1:1 mapping, clean separation
│
└─ NO: Maintaining existing?
   └─ GcpRedisInstance → RedisInstance (OLD Pattern)
      └─ ONLY maintain, NEVER replicate
```

### ❌ WRONG: Using OLD Pattern for New Code

```go
// NEVER: Creating multi-provider mapping
state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
    Spec: cloudcontrolv1beta1.RedisInstanceSpec{
        Instance: cloudcontrolv1beta1.RedisInstanceInfo{
            Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{  // Provider section
                MemorySizeGb: memorySize,
            },
        },
    },
}
```

### ✅ CORRECT: Using NEW Pattern

```go
// ALWAYS: Direct provider-specific mapping
state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{
    Spec: cloudcontrolv1beta1.GcpRedisClusterSpec{
        ShardCount: redisCluster.Spec.ShardCount,  // Simple 1:1 mapping
    },
}
```

## Implementation Steps

### Step 1: Define SKR API

**Location**: `api/cloud-resources/v1beta1/gcprediscluster_types.go`

#### ❌ WRONG: Cloud Provider Jargon

```go
// NEVER: Technical provider terms in user-facing API
type GcpRedisClusterSpec struct {
    // +kubebuilder:validation:Required
    ManagedInstancesCount int32  // Internal GCP terminology
}
```

#### ✅ CORRECT: User-Friendly Field Names

```go
// ALWAYS: User-friendly field names
type GcpRedisClusterSpec struct {
    // +kubebuilder:validation:Required
    ShardCount int32  // Clear, user-facing terminology
    
    // +kubebuilder:validation:Required
    ReplicaCount int32
}

type GcpRedisClusterStatus struct {
    // REQUIRED: Status fields
    Id         string
    State      string
    Conditions []metav1.Condition
    
    // User-relevant information
    PrimaryEndpoint string
}
```

**Run after defining**:
```bash
make generate  # Generate DeepCopy methods
make manifests # Generate CRD YAML
```

### Step 2: Create State Definition

**Location**: `pkg/skr/gcprediscluster/state.go`

#### ❌ WRONG: Extending focal.State

```go
// NEVER: SKR state does NOT extend focal.State
type State struct {
    focal.State  // WRONG: focal.State is for KCP only
    KcpGcpRedisCluster *cloudcontrolv1beta1.GcpRedisCluster
}
```

#### ✅ CORRECT: Extending common.State

```go
// ALWAYS: SKR state extends common.State
package gcprediscluster

import (
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "k8s.io/apimachinery/pkg/types"
)

type State struct {
    common.State  // MUST extend common.State, not focal.State
    
    KymaRef    types.NamespacedName                      // Kyma environment reference
    KcpCluster composed.StateCluster                     // KCP cluster client
    
    KcpGcpRedisCluster *cloudcontrolv1beta1.GcpRedisCluster  // KCP resource
}

func (s *State) ObjAsGcpRedisCluster() *cloudresourcesv1beta1.GcpRedisCluster {
    return s.Obj().(*cloudresourcesv1beta1.GcpRedisCluster)
}
```

### Step 3: Generate Unique ID

**Location**: `pkg/skr/gcprediscluster/updateId.go`

#### ❌ WRONG: Creating KCP Without ID

```go
// NEVER: Creating KCP resource before setting ID
func createKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{
        // Creating without ID in status is ERROR
    }
    return state.KcpCluster.K8sClient().Create(ctx, state.KcpGcpRedisCluster), ctx
}
```

#### ✅ CORRECT: Generate ID Before KCP Creation

```go
// ALWAYS: Generate and persist ID before creating KCP resource
package gcprediscluster

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    
    // Generate UUID-based ID if not exists
    if redisCluster.Status.Id == "" {
        redisCluster.Status.Id = fmt.Sprintf("%s-%s", 
            redisCluster.Name, 
            uuid.New().String()[:8])
        
        return composed.UpdateStatus(redisCluster).
            ErrorLogMessage("Error updating GcpRedisCluster status with ID").
            SuccessError(composed.StopWithRequeue).
            Run(ctx, state)
    }
    
    return nil, ctx
}
```

### Step 4: Load KCP Resource

**Location**: `pkg/skr/gcprediscluster/loadKcpGcpRedisCluster.go`

#### ❌ WRONG: Not Handling 404 Errors

```go
// NEVER: Failing on 404 errors
func loadKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
    err := state.KcpCluster.K8sClient().Get(ctx, name, state.KcpGcpRedisCluster)
    if err != nil {
        return err, ctx  // WRONG: 404 is normal during creation
    }
    return nil, ctx
}
```

#### ✅ CORRECT: 404 Means Not Created Yet

```go
// ALWAYS: Treat 404 as "not created yet"
package gcprediscluster

import (
    "context"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/types"
)

func loadKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    
    if redisCluster.Status.Id == "" {
        return nil, ctx  // ID not generated yet
    }
    
    state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{}
    err := state.KcpCluster.K8sClient().Get(ctx,
        types.NamespacedName{
            Namespace: state.KymaRef.Namespace,
            Name:      redisCluster.Status.Id,
        },
        state.KcpGcpRedisCluster)
    
    if apierrors.IsNotFound(err) {
        state.KcpGcpRedisCluster = nil  // Not created yet
        return nil, ctx
    }
    
    if err != nil {
        return composed.LogErrorAndReturn(err, 
            "Error loading KCP GcpRedisCluster", 
            composed.StopWithRequeue, ctx)
    }
    
    return nil, ctx
}
```

### Step 5: Create KCP Resource

**Location**: `pkg/skr/gcprediscluster/createKcpGcpRedisCluster.go`

#### ❌ WRONG: Missing Annotations or RemoteRef

```go
// NEVER: Creating KCP without required annotations
state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{
    ObjectMeta: metav1.ObjectMeta{
        Name:      redisCluster.Status.Id,
        Namespace: state.KymaRef.Namespace,
        // Missing Annotations - WRONG!
    },
    Spec: cloudcontrolv1beta1.GcpRedisClusterSpec{
        ShardCount: redisCluster.Spec.ShardCount,
        // Missing RemoteRef - WRONG!
    },
}
```

#### ✅ CORRECT: Complete KCP Resource Creation

```go
// ALWAYS: Set annotations, RemoteRef, and Scope
package gcprediscluster

import (
    "context"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    logger := composed.LoggerFromCtx(ctx)
    
    if state.KcpGcpRedisCluster != nil {
        return nil, ctx  // Already created
    }
    
    logger.Info("Creating KCP GcpRedisCluster")
    
    // MUST: Set all required fields
    state.KcpGcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{
        ObjectMeta: metav1.ObjectMeta{
            Name:      redisCluster.Status.Id,
            Namespace: state.KymaRef.Namespace,
            Labels: map[string]string{
                cloudcontrolv1beta1.LabelKymaName: state.KymaRef.Name,
            },
            // REQUIRED: Annotations for tracking
            Annotations: map[string]string{
                cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
                cloudcontrolv1beta1.LabelRemoteName:      redisCluster.Name,
                cloudcontrolv1beta1.LabelRemoteNamespace: redisCluster.Namespace,
            },
        },
        Spec: cloudcontrolv1beta1.GcpRedisClusterSpec{
            // REQUIRED: RemoteRef back to SKR
            RemoteRef: cloudcontrolv1beta1.RemoteRef{
                Namespace: redisCluster.Namespace,
                Name:      redisCluster.Name,
            },
            // REQUIRED: Scope reference
            Scope: cloudcontrolv1beta1.ScopeRef{
                Name: redisCluster.Spec.Scope.Name,
            },
            // Map spec fields
            ShardCount:   redisCluster.Spec.ShardCount,
            ReplicaCount: redisCluster.Spec.ReplicaCount,
        },
    }
    
    err := state.KcpCluster.K8sClient().Create(ctx, state.KcpGcpRedisCluster)
    if err != nil {
        return composed.LogErrorAndReturn(err, 
            "Error creating KCP GcpRedisCluster", 
            composed.StopWithRequeue, ctx)
    }
    
    logger.Info("KCP GcpRedisCluster created")
    return nil, ctx
}
```

### Step 6: Sync Status from KCP

**Location**: `pkg/skr/gcprediscluster/updateStatus.go`

#### ❌ WRONG: Not Syncing Conditions

```go
// NEVER: Only syncing State, missing Conditions
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    
    if state.KcpGcpRedisCluster != nil {
        redisCluster.Status.State = string(state.KcpGcpRedisCluster.Status.State)
        // Missing Conditions sync - WRONG!
    }
    
    return composed.UpdateStatus(redisCluster).Run(ctx, state)
}
```

#### ✅ CORRECT: Complete Status Synchronization

```go
// ALWAYS: Sync State, Conditions, and user-relevant fields
package gcprediscluster

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    
    if state.KcpGcpRedisCluster != nil {
        // MUST: Copy State
        redisCluster.Status.State = string(state.KcpGcpRedisCluster.Status.State)
        
        // MUST: Copy Conditions
        redisCluster.Status.Conditions = state.KcpGcpRedisCluster.Status.Conditions
        
        // MUST: Copy user-relevant information
        if state.KcpGcpRedisCluster.Status.PrimaryEndpoint != "" {
            redisCluster.Status.PrimaryEndpoint = state.KcpGcpRedisCluster.Status.PrimaryEndpoint
        }
    }
    
    return composed.UpdateStatus(redisCluster).
        ErrorLogMessage("Error updating GcpRedisCluster status").
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

### Step 7: Implement Deletion Flow

**Location**: `pkg/skr/gcprediscluster/deleteKcpGcpRedisCluster.go`, `waitKcpGcpRedisClusterDeleted.go`

#### ❌ WRONG: Removing Finalizer Before KCP Deletion

```go
// NEVER: Removing finalizer while KCP resource still exists
func reconciler() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster",
        composed.IfElse(
            composed.Not(composed.MarkedForDeletionPredicate),
            createFlow,
            composed.ComposeActions(
                "delete",
                deleteKcpGcpRedisCluster,
                actions.RemoveCommonFinalizer(),  // WRONG: No wait for deletion!
            ),
        ),
    )
}
```

#### ✅ CORRECT: Wait for KCP Deletion Before Finalizer Removal

```go
// ALWAYS: Delete KCP → Wait for deletion → Remove finalizer
package gcprediscluster

import (
    "context"
    "time"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/types"
)

// Step 1: Delete KCP resource
func deleteKcpGcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    if state.KcpGcpRedisCluster == nil {
        return nil, ctx  // Already deleted
    }
    
    logger.Info("Deleting KCP GcpRedisCluster")
    
    err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpGcpRedisCluster)
    if apierrors.IsNotFound(err) {
        return nil, ctx
    }
    
    if err != nil {
        return composed.LogErrorAndReturn(err, 
            "Error deleting KCP GcpRedisCluster", 
            composed.StopWithRequeue, ctx)
    }
    
    return nil, ctx
}

// Step 2: Wait for KCP deletion to complete
func waitKcpGcpRedisClusterDeleted(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    redisCluster := state.ObjAsGcpRedisCluster()
    logger := composed.LoggerFromCtx(ctx)
    
    kcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
    err := state.KcpCluster.K8sClient().Get(ctx,
        types.NamespacedName{
            Namespace: state.KymaRef.Namespace,
            Name:      redisCluster.Status.Id,
        },
        kcpRedisCluster)
    
    if apierrors.IsNotFound(err) {
        logger.Info("KCP GcpRedisCluster is deleted")
        return nil, ctx  // Continue to finalizer removal
    }
    
    if err != nil {
        return composed.LogErrorAndReturn(err, 
            "Error checking KCP GcpRedisCluster deletion", 
            composed.StopWithRequeue, ctx)
    }
    
    logger.Info("Waiting for KCP GcpRedisCluster to be deleted")
    return composed.StopWithRequeueDelay(3 * time.Second), nil
}

// Action composition with correct ordering
func reconciler() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster",
        composed.IfElse(
            composed.Not(composed.MarkedForDeletionPredicate),
            createFlow,
            composed.ComposeActions(
                "delete",
                deleteKcpGcpRedisCluster,           // Step 1
                waitKcpGcpRedisClusterDeleted,      // Step 2
                actions.RemoveCommonFinalizer(),     // Step 3: Only after KCP gone
                composed.StopAndForgetAction,
            ),
        ),
    )
}
```

### Step 8: Compose Reconciler Actions

**Location**: `pkg/skr/gcprediscluster/reconciler.go`

#### ❌ WRONG: Missing Feature Flag or StopAndForget

```go
// NEVER: No feature flag check or missing StopAndForget
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster",
        composed.LoadObj,  // Missing feature flag check!
        updateId,
        createKcpGcpRedisCluster,
        updateStatus,
        // Missing StopAndForgetAction!
    )
}
```

#### ✅ CORRECT: Complete Action Composition

```go
// ALWAYS: Feature flags, proper flow control, StopAndForget
package gcprediscluster

import (
    "context"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler interface {
    reconcile.Reconciler
}

type reconciler struct {
    composedStateFactory composed.StateFactory
    kymaRef              types.NamespacedName
    kcpCluster           composed.StateCluster
}

func NewReconciler(
    composedStateFactory composed.StateFactory,
    kymaRef types.NamespacedName,
    kcpCluster composed.StateCluster,
) Reconciler {
    return &reconciler{
        composedStateFactory: composedStateFactory,
        kymaRef:              kymaRef,
        kcpCluster:           kcpCluster,
    }
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newState(req.NamespacedName)
    action := r.newAction()
    return composed.Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster",
        // MUST: Feature flag check first
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),
        composed.LoadObj,
        
        // MUST: Generate ID before KCP operations
        updateId,
        loadKcpGcpRedisCluster,
        
        // MUST: Proper create/delete branching
        composed.IfElse(
            composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "create-update",
                actions.AddCommonFinalizer(),
                createKcpGcpRedisCluster,
                updateStatus,
            ),
            composed.ComposeActions(
                "delete",
                deleteKcpGcpRedisCluster,
                waitKcpGcpRedisClusterDeleted,
                actions.RemoveCommonFinalizer(),
                composed.StopAndForgetAction,
            ),
        ),
        
        // MUST: End with StopAndForget
        composed.StopAndForgetAction,
    )
}

func (r *reconciler) newState(name types.NamespacedName) *State {
    return &State{
        State:      r.composedStateFactory.NewState(name, &cloudresourcesv1beta1.GcpRedisCluster{}),
        KymaRef:    r.kymaRef,
        KcpCluster: r.kcpCluster,
    }
}

func (r *reconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&cloudresourcesv1beta1.GcpRedisCluster{}).
        Complete(r)
}
```

## Common Pitfalls

### Pitfall 1: Creating KCP Without ID

**Symptom**: KCP resource name is empty or invalid

**Cause**:
```go
// WRONG: No ID generation before KCP creation
func createKcpResource(ctx context.Context, st composed.State) (error, context.Context) {
    state.KcpResource = &cloudcontrolv1beta1.Resource{
        ObjectMeta: metav1.ObjectMeta{
            Name: state.Obj().Status.Id,  // Empty if not generated!
        },
    }
}
```

**Solution**: ALWAYS generate ID first in separate action:
```go
func reconciler() composed.Action {
    return composed.ComposeActions(
        updateId,                  // Generate ID FIRST
        loadKcpResource,
        createKcpResource,         // Then create
    )
}
```

### Pitfall 2: Removing Finalizer Before KCP Deletion

**Symptom**: SKR resource deleted but cloud resources remain (orphaned)

**Cause**:
```go
// WRONG: No wait for KCP deletion
composed.ComposeActions(
    "delete",
    deleteKcpResource,
    actions.RemoveCommonFinalizer(),  // Finalizer removed immediately!
)
```

**Solution**: ALWAYS wait for KCP deletion:
```go
// CORRECT: Wait before finalizer removal
composed.ComposeActions(
    "delete",
    deleteKcpResource,
    waitKcpResourceDeleted,           // Wait for actual deletion
    actions.RemoveCommonFinalizer(),  // Then remove finalizer
)
```

### Pitfall 3: Missing KCP Annotations

**Symptom**: Cannot track SKR → KCP relationship, debugging difficult

**Cause**:
```go
// WRONG: Missing required annotations
state.KcpResource = &cloudcontrolv1beta1.Resource{
    ObjectMeta: metav1.ObjectMeta{
        Name:      resourceId,
        Namespace: namespace,
        // No Annotations!
    },
}
```

**Solution**: ALWAYS set three required annotations:
```go
// CORRECT: All required annotations
Annotations: map[string]string{
    cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
    cloudcontrolv1beta1.LabelRemoteName:      skrResource.Name,
    cloudcontrolv1beta1.LabelRemoteNamespace: skrResource.Namespace,
}
```

### Pitfall 4: Not Syncing Conditions from KCP

**Symptom**: SKR resource shows no errors even when KCP resource fails

**Cause**:
```go
// WRONG: Only syncing State
redisCluster.Status.State = string(state.KcpResource.Status.State)
// Missing Conditions!
```

**Solution**: ALWAYS copy Conditions:
```go
// CORRECT: Full status sync
redisCluster.Status.State = string(state.KcpResource.Status.State)
redisCluster.Status.Conditions = state.KcpResource.Status.Conditions  // Copy conditions
```

## Validation Checklist

### API Definition
- [ ] User-friendly field names (not cloud provider jargon)
- [ ] Required fields marked with `+kubebuilder:validation:Required`
- [ ] Status has `Id`, `State`, `Conditions` fields
- [ ] Status includes user-relevant information (endpoints, etc.)
- [ ] Run `make generate && make manifests`

### State Implementation
- [ ] State extends `common.State` (not focal.State)
- [ ] KCP cluster client included
- [ ] KCP resource field typed correctly
- [ ] Type-safe `ObjAs*()` getter implemented

### ID Generation
- [ ] UUID-based ID generated in `updateId` action
- [ ] ID persisted to status with UpdateStatus
- [ ] ID generation happens BEFORE KCP creation
- [ ] `updateId` action first in composition

### KCP Resource Creation
- [ ] Three annotations set (LabelKymaName, LabelRemoteName, LabelRemoteNamespace)
- [ ] RemoteRef points back to SKR resource
- [ ] Scope reference included
- [ ] Spec fields mapped from SKR
- [ ] Only creates if KCP resource doesn't exist

### Status Synchronization
- [ ] State copied from KCP to SKR
- [ ] Conditions copied from KCP to SKR
- [ ] User-relevant information extracted
- [ ] Uses `composed.UpdateStatus()`
- [ ] Ends with StopAndForget

### Deletion Flow
- [ ] Deletes KCP resource first
- [ ] Waits for KCP deletion (waitKcpResourceDeleted action)
- [ ] Removes finalizer AFTER KCP deleted
- [ ] Handles 404 errors (IsNotFound)
- [ ] Returns StopAndForget after finalizer removal

### Action Composition
- [ ] Feature flag check first (`feature.LoadFeatureContextFromObj`)
- [ ] LoadObj after feature flags
- [ ] Finalizer added in create flow
- [ ] Create/delete flows in IfElse branching
- [ ] Ends with `composed.StopAndForgetAction`

### Testing
- [ ] Test creates SKR resource
- [ ] Verifies KCP resource created
- [ ] Verifies status synced from KCP to SKR
- [ ] Tests deletion flow
- [ ] Tests run with `make test`

## Summary: Key Rules

1. **NEW Pattern ONLY**: Provider-specific SKR → provider-specific KCP (GcpRedisCluster → GcpRedisCluster)
2. **Generate ID First**: UUID-based ID BEFORE creating KCP resource
3. **Three Annotations**: LabelKymaName, LabelRemoteName, LabelRemoteNamespace REQUIRED
4. **Sync Status**: Copy State + Conditions + user fields from KCP to SKR
5. **Wait for Deletion**: Delete KCP → Wait for deletion → Remove finalizer (in that order)
6. **End with StopAndForget**: All successful flows return `composed.StopAndForget, nil`
7. **Check Feature Flags**: Use `feature.LoadFeatureContextFromObj` in composition
8. **Extend common.State**: SKR state uses common.State, not focal.State

## Next Steps

- [Add Controller Tests](CONTROLLER_TESTS.md)
- [Configure Feature Flags](FEATURE_FLAGS.md)
- [Add API Validation](API_VALIDATION_TESTS.md)
