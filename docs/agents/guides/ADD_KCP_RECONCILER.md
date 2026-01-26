# Guide: Adding a New KCP Reconciler

**Authority**: Step-by-step procedure  
**Prerequisite For**: Creating cloud provider resource reconcilers  
**Must Read Before**: Adding new GCP/Azure/AWS resource support

**Prerequisites**:
- MUST have read: [NEW Reconciler Pattern](../architecture/RECONCILER_NEW_PATTERN.md)
- MUST have read: [Action Composition](../architecture/ACTION_COMPOSITION.md)
- MUST have read: [State Pattern](../architecture/STATE_PATTERN.md)
- MUST have: Go 1.25+, Kubebuilder installed
- MUST understand: Kubernetes controllers, target cloud provider API

**Skip This Guide If**:
- You are creating user-facing resources (see [ADD_SKR_RECONCILER.md](ADD_SKR_RECONCILER.md))
- You are maintaining existing reconcilers (not creating new)

**Estimated Time**: 2-4 hours

## What You Will Create

- Provider-specific CRD (e.g., `GcpRedisCluster`, `AzureRedisEnterprise`)
- KCP reconciler managing cloud provider resources
- State, client, and actions following NEW pattern
- Unit tests with mocked provider
- Controller registration

## Rules: Creating KCP Reconcilers

### MUST

1. MUST use NEW Pattern (provider-specific CRD)
2. MUST extend `focal.State` directly (no intermediate layer)
3. MUST place all code in `pkg/kcp/provider/<provider>/<resource>/`
4. MUST implement one action per file
5. MUST add finalizer for resources requiring cleanup
6. MUST end successful flows with `StopAndForgetAction`
7. MUST handle 404 as success case (resource doesn't exist yet)
8. MUST update status on errors
9. MUST add wait actions for async cloud operations

### MUST NOT

1. MUST NOT use OLD Pattern (multi-provider CRDs)
2. MUST NOT add shared intermediate state layer
3. MUST NOT create clients on-demand (use provider pattern)
4. MUST NOT skip error handling
5. MUST NOT forget to remove finalizer on deletion

## Step 1: Define API (CRD)

**Duration**: 15-20 minutes

### Command: Scaffold CRD

```bash
cd /path/to/cloud-manager

kubebuilder create api --group cloud-control --version v1beta1 --kind GcpRedisCluster --resource --controller
```

**Output**: Creates `api/cloud-control/v1beta1/gcprediscluster_types.go`

**Note**: Delete generated controller in `internal/controller/` - you'll create your own in `pkg/kcp/provider/`

### Template: CRD Definition

**File**: `api/cloud-control/v1beta1/gcprediscluster_types.go`

```go
package v1beta1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GcpRedisClusterSpec defines desired state
type GcpRedisClusterSpec struct {
    // REQUIRED: Links to SKR resource
    // +kubebuilder:validation:Required
    RemoteRef RemoteRef `json:"remoteRef"`
    
    // REQUIRED: Cloud provider project
    // +kubebuilder:validation:Required
    Scope ScopeRef `json:"scope"`
    
    // Provider-specific fields
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=100
    ShardCount int32 `json:"shardCount"`
    
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    ReplicaCount int32 `json:"replicaCount"`
}

// GcpRedisClusterStatus defines observed state
type GcpRedisClusterStatus struct {
    // +optional
    State StatusState `json:"state,omitempty"`
    
    // +optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    
    // +optional
    Id string `json:"id,omitempty"`
    
    // +optional
    DiscoveryEndpoint string `json:"discoveryEndpoint,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// GcpRedisCluster is the Schema
type GcpRedisCluster struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   GcpRedisClusterSpec   `json:"spec,omitempty"`
    Status GcpRedisClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type GcpRedisClusterList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []GcpRedisCluster `json:"items"`
}

// REQUIRED: Implement ScopeRef interface
func (in *GcpRedisCluster) ScopeRef() ScopeRef {
    return in.Spec.Scope
}

func (in *GcpRedisCluster) SetScopeRef(scopeRef ScopeRef) {
    in.Spec.Scope = scopeRef
}

func init() {
    SchemeBuilder.Register(&GcpRedisCluster{}, &GcpRedisClusterList{})
}
```

### Generate Manifests

```bash
make manifests  # Generate CRD YAML
make generate   # Generate DeepCopy methods
```

**Verification**:
- File created: `config/crd/bases/cloud-control.kyma-project.io_gcpredisclusters.yaml`
- File updated: `api/cloud-control/v1beta1/zz_generated.deepcopy.go`

## Step 2: Create Package Structure

**Duration**: 5 minutes

```bash
mkdir -p pkg/kcp/provider/gcp/rediscluster/client
```

**Required Files**:
```
pkg/kcp/provider/gcp/rediscluster/
├── reconcile.go              # Reconciler
├── state.go                  # State
├── client/client.go          # Cloud API client
├── loadCluster.go            # Load action
├── createCluster.go          # Create action
├── deleteCluster.go          # Delete action
├── waitClusterAvailable.go   # Wait action
├── waitClusterDeleted.go     # Wait action
└── updateStatus.go           # Status action
```

## Step 3: Implement State

**Duration**: 10 minutes

**File**: `pkg/kcp/provider/gcp/rediscluster/state.go`

```go
package rediscluster

import (
    "context"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    rediscluster "cloud.google.com/go/redis/cluster/apiv1"
    redisclusterpb "cloud.google.com/go/redis/cluster/apiv1/clusterpb"
)

// State extends focal.State DIRECTLY (NEW pattern)
type State struct {
    focal.State  // Direct extension
    
    // Cloud provider client
    redisClusterClient client.RedisClusterClient
    
    // Remote resource
    cluster *redisclusterpb.Cluster
    
    // Modification tracking
    updateMask []string
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    redisClusterClientProvider gcpclient.GcpClientProvider[client.RedisClusterClient]
}

func NewStateFactory(
    redisClusterClientProvider gcpclient.GcpClientProvider[client.RedisClusterClient],
) StateFactory {
    return &stateFactory{
        redisClusterClientProvider: redisClusterClientProvider,
    }
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:              focalState,
        redisClusterClient: f.redisClusterClientProvider(),
    }, nil
}

// REQUIRED: Type-safe accessor
func (s *State) ObjAsGcpRedisCluster() *cloudcontrolv1beta1.GcpRedisCluster {
    return s.Obj().(*cloudcontrolv1beta1.GcpRedisCluster)
}
```

## Step 4: Implement Cloud Client

**Duration**: 20-30 minutes

**File**: `pkg/kcp/provider/gcp/rediscluster/client/client.go`

```go
package client

import (
    "context"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    rediscluster "cloud.google.com/go/redis/cluster/apiv1"
    redisclusterpb "cloud.google.com/go/redis/cluster/apiv1/clusterpb"
)

// Business operations interface
type RedisClusterClient interface {
    CreateCluster(ctx context.Context, req CreateClusterRequest) (string, error)
    GetCluster(ctx context.Context, projectId, locationId, clusterId string) (*redisclusterpb.Cluster, error)
    UpdateCluster(ctx context.Context, cluster *redisclusterpb.Cluster, updateMask []string) (string, error)
    DeleteCluster(ctx context.Context, projectId, locationId, clusterId string) (string, error)
    GetOperation(ctx context.Context, operation string) (*redisclusterpb.OperationMetadata, error)
}

type CreateClusterRequest struct {
    ProjectId    string
    LocationId   string
    ClusterId    string
    ShardCount   int32
    ReplicaCount int32
}

// Provider function
func NewRedisClusterClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[RedisClusterClient] {
    return func() RedisClusterClient {
        return NewRedisClusterClient(gcpClients)
    }
}

// Implementation
type redisClusterClient struct {
    client *rediscluster.CloudRedisClusterClient
}

func NewRedisClusterClient(gcpClients *gcpclient.GcpClients) RedisClusterClient {
    return &redisClusterClient{
        client: gcpClients.RedisCluster,
    }
}

func (c *redisClusterClient) CreateCluster(ctx context.Context, req CreateClusterRequest) (string, error) {
    parent := fmt.Sprintf("projects/%s/locations/%s", req.ProjectId, req.LocationId)
    
    op, err := c.client.CreateCluster(ctx, &redisclusterpb.CreateClusterRequest{
        Parent:    parent,
        ClusterId: req.ClusterId,
        Cluster: &redisclusterpb.Cluster{
            ShardCount:   req.ShardCount,
            ReplicaCount: req.ReplicaCount,
        },
    })
    if err != nil {
        return "", err
    }
    
    return op.Name(), nil  // Return operation ID
}

func (c *redisClusterClient) GetCluster(ctx context.Context, projectId, locationId, clusterId string) (*redisclusterpb.Cluster, error) {
    name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectId, locationId, clusterId)
    
    return c.client.GetCluster(ctx, &redisclusterpb.GetClusterRequest{
        Name: name,
    })
}

// Implement UpdateCluster, DeleteCluster, GetOperation...
```

## Step 5: Implement Actions

**Duration**: 40-60 minutes

### Load Action

**File**: `loadCluster.go`

```go
package rediscluster

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "google.golang.org/api/googleapi"
)

func loadCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    cluster := state.ObjAsGcpRedisCluster()
    
    // Skip if no ID yet
    if cluster.Status.Id == "" {
        return nil, ctx
    }
    
    gcpCluster, err := state.redisClusterClient.GetCluster(ctx,
        state.Scope().Spec.Gcp.Project,
        state.Scope().Spec.Gcp.Region,
        cluster.Status.Id,
    )
    
    if err != nil {
        // 404 is OK - resource doesn't exist
        if googleapi.IsNotFound(err) {
            return nil, ctx
        }
        return err, ctx
    }
    
    state.cluster = gcpCluster
    logger.Info("GCP Redis Cluster loaded", "id", cluster.Status.Id)
    
    return nil, ctx
}
```

### Create Action

**File**: `createCluster.go`

```go
package rediscluster

import (
    "context"
    "time"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "k8s.io/apimachinery/pkg/api/meta"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createCluster(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Skip if already exists
    if state.cluster != nil {
        return nil, ctx
    }
    
    cluster := state.ObjAsGcpRedisCluster()
    
    logger.Info("Creating GCP Redis Cluster")
    
    // Generate ID if not set
    if cluster.Status.Id == "" {
        cluster.Status.Id = cluster.Name
        state.UpdateObjStatus(ctx)
    }
    
    operationId, err := state.redisClusterClient.CreateCluster(ctx, client.CreateClusterRequest{
        ProjectId:    state.Scope().Spec.Gcp.Project,
        LocationId:   state.Scope().Spec.Gcp.Region,
        ClusterId:    cluster.Status.Id,
        ShardCount:   cluster.Spec.ShardCount,
        ReplicaCount: cluster.Spec.ReplicaCount,
    })
    
    if err != nil {
        meta.SetStatusCondition(cluster.Conditions(), metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
            Message: err.Error(),
        })
        cluster.Status.State = cloudcontrolv1beta1.StateError
        state.UpdateObjStatus(ctx)
        return composed.StopWithRequeueDelay(time.Minute), nil
    }
    
    cluster.Status.OpIdentifier = operationId
    cluster.Status.State = cloudcontrolv1beta1.StateProcessing
    state.UpdateObjStatus(ctx)
    
    logger.Info("Cluster creation initiated", "operationId", operationId)
    
    return nil, ctx
}
```

### Wait Action

**File**: `waitClusterAvailable.go`

```go
package rediscluster

import (
    "context"
    "time"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    redisclusterpb "cloud.google.com/go/redis/cluster/apiv1/clusterpb"
)

func waitClusterAvailable(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    cluster := state.ObjAsGcpRedisCluster()
    
    // Skip if no operation
    if cluster.Status.OpIdentifier == "" {
        return nil, ctx
    }
    
    opMetadata, err := state.redisClusterClient.GetOperation(ctx, cluster.Status.OpIdentifier)
    if err != nil {
        return err, ctx
    }
    
    // Check if done
    if opMetadata.EndTime == nil {
        logger.Info("Waiting for cluster creation")
        return composed.StopWithRequeueDelay(10 * time.Second), nil
    }
    
    // Clear operation ID
    cluster.Status.OpIdentifier = ""
    state.UpdateObjStatus(ctx)
    
    logger.Info("Cluster creation complete")
    
    // Reload cluster
    return loadCluster(ctx, state)
}
```

### Status Update Action

**File**: `updateStatus.go`

```go
package rediscluster

import (
    "context"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    cluster := state.ObjAsGcpRedisCluster()
    
    // Skip if not loaded
    if state.cluster == nil {
        return nil, ctx
    }
    
    cluster.Status.Id = state.cluster.Name
    cluster.Status.DiscoveryEndpoint = state.cluster.DiscoveryEndpoints[0].Address
    cluster.Status.State = cloudcontrolv1beta1.StateReady
    
    return composed.UpdateStatus(cluster).
        SetExclusiveConditions(metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeReady,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonReady,
            Message: "Cluster is ready",
        }).
        SuccessError(composed.StopAndForget).
        Run(ctx, state)
}
```

### Delete Actions

Create `deleteCluster.go` and `waitClusterDeleted.go` following similar patterns.

## Step 6: Compose Reconciler

**Duration**: 15 minutes

**File**: `reconcile.go`

```go
package rediscluster

import (
    "context"
    "fmt"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type GcpRedisClusterReconciler interface {
    reconcile.Reconciler
}

type gcpRedisClusterReconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory
}

func NewGcpRedisClusterReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    stateFactory StateFactory,
) GcpRedisClusterReconciler {
    return &gcpRedisClusterReconciler{
        composedStateFactory: composedStateFactory,
        focalStateFactory:    focalStateFactory,
        stateFactory:         stateFactory,
    }
}

func (r *gcpRedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newFocalState(req.NamespacedName)
    action := r.newAction()
    
    return composed.Handle(action(ctx, state))
}

func (r *gcpRedisClusterReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpRedisCluster{}),
        focal.New(),
        r.newFlow(),
    )
}

func (r *gcpRedisClusterReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Create provider state
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap state")
            cluster := st.Obj().(*cloudcontrolv1beta1.GcpRedisCluster)
            cluster.Status.State = cloudcontrolv1beta1.StateError
            return composed.UpdateStatus(cluster).
                SetExclusiveConditions(metav1.Condition{
                    Type:    cloudcontrolv1beta1.ConditionTypeError,
                    Status:  metav1.ConditionTrue,
                    Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
                    Message: "Failed to create state",
                }).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }
        
        return composed.ComposeActions(
            "gcpRedisCluster",
            actions.AddCommonFinalizer(),
            loadCluster,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    createCluster,
                    waitClusterAvailable,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteCluster,
                    waitClusterDeleted,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)
    }
}

func (r *gcpRedisClusterReconciler) newFocalState(name types.NamespacedName) focal.State {
    return r.focalStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.GcpRedisCluster{}),
    )
}

func (r *gcpRedisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&cloudcontrolv1beta1.GcpRedisCluster{}).
        Complete(r)
}
```

## Step 7: Register Controller

**Duration**: 5 minutes

**File**: `cmd/main.go`

```go
// Import
import (
    gcprediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster"
    gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
)

// In main() after other controllers:

// Create state factory
gcpRedisClusterStateFactory := gcprediscluster.NewStateFactory(
    gcpredisclusterclient.NewRedisClusterClientProvider(gcpClients),
)

// Register controller
if err = gcprediscluster.NewGcpRedisClusterReconciler(
    composedStateFactory,
    focalStateFactory,
    gcpRedisClusterStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "GcpRedisCluster")
    os.Exit(1)
}
```

## Validation Checklist

Before committing:

### API
- [ ] Provider-specific CRD name (`GcpRedisCluster`)
- [ ] RemoteRef field present
- [ ] Scope field present
- [ ] State and Conditions in status
- [ ] ScopeRef() method implemented
- [ ] CRD manifests generated (`make manifests`)
- [ ] DeepCopy methods generated (`make generate`)

### State
- [ ] Extends `focal.State` directly
- [ ] No intermediate state layer
- [ ] Client injected via provider
- [ ] Type-safe `ObjAs*()` method

### Reconciler
- [ ] Uses `composed.ComposeActions`
- [ ] Finalizer actions included
- [ ] Deletion path implemented
- [ ] Ends with `StopAndForgetAction`
- [ ] State factory errors handled

### Actions
- [ ] One action per file
- [ ] Proper return values
- [ ] 404 handled as success
- [ ] Status updated on errors
- [ ] Wait actions for async operations

### Client
- [ ] Business operations interface
- [ ] Provider function created
- [ ] Implementation wraps GCP client

### Registration
- [ ] Controller registered in main.go
- [ ] State factory created
- [ ] Client provider configured

## Troubleshooting

| Issue | Solution |
|-------|----------|
| CRD not generated | Run `make manifests && make generate` |
| State creation fails | Check client provider in main.go |
| Reconciler never completes | Add `StopAndForgetAction` at end |
| Status not updating | Call `state.UpdateObjStatus(ctx)` |
| 404 errors cause failure | Handle 404 as success case |

## Related Documentation

**MUST READ NEXT**:
- [Controller Tests Guide](CONTROLLER_TESTS.md) - Add tests
- [Creating Mocks](CREATING_MOCKS.md) - Mock cloud providers

**REFERENCE**:
- [NEW Reconciler Pattern](../architecture/RECONCILER_NEW_PATTERN.md) - Pattern details
- [Action Composition](../architecture/ACTION_COMPOSITION.md) - Action patterns
- [Feature Flags](FEATURE_FLAGS.md) - Add feature flags
