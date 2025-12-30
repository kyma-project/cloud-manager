# Development Workflow

**Authority**: REQUIRED for daily development tasks. All commands and procedures are deterministic.

**Target**: LLM coding agents performing day-to-day Cloud Manager development.

**Related**: [ADD_KCP_RECONCILER.md](guides/ADD_KCP_RECONCILER.md) | [QUICK_REFERENCE.md](reference/QUICK_REFERENCE.md) | [COMMON_PITFALLS.md](reference/COMMON_PITFALLS.md)

---

## Authority: Workflow Rules

### MUST Follow
- API changes require `make manifests` + patch script + sync script
- Tests MUST pass before committing
- NEW pattern for all new resources (post-2024)
- Feature flag checks in all reconcilers
- Status updates with error checking

### MUST NOT
- Commit without running tests
- Skip `patchAfterMakeManifests.sh` after API changes
- Use OLD pattern for new resources
- Modify generated files manually (`zz_generated*.go`)
- Skip sync script after CRD changes

---

## Quick Command Reference

| Task | Command | When to Run |
|------|---------|-------------|
| Generate CRDs | `make manifests` | After API type changes |
| Generate code | `make generate` | After API interface changes |
| Add version annotations | `./config/patchAfterMakeManifests.sh` | After `make manifests` |
| Sync CRDs | `./config/sync.sh` | After patching CRDs |
| Run tests | `make test` | Before committing |
| Build binary | `make build` | Verify compilation |
| Install CRDs | `make install` | Local testing |
| Run locally | `make run` | Development/debugging |

---

## Initial Setup

### Prerequisites Check

```bash
# Go 1.25+ required
go version

# Make required
make --version

# Docker for image building
docker --version

# kubectl for cluster interaction
kubectl version --client

# Git for version control
git --version
```

### Repository Setup

```bash
# Clone repository
git clone https://github.com/kyma-project/cloud-manager.git
cd cloud-manager

# Download dependencies
go mod download

# Install development tools
make envtest

# Verify setup
make test
make build
ls -lh bin/manager
```

---

## Daily Development Workflow

### Before Starting Work

```bash
# Update repository
git pull origin main

# Sync dependencies
go mod download
go mod tidy

# Ensure tools current
make envtest

# Verify tests pass
make test
```

### Creating Branch

```bash
# Feature branch
git checkout -b feature/<feature-name>

# Bug fix branch
git checkout -b fix/<issue-number>
```

---

## Task: Add New KCP Reconciler (NEW Pattern)

### Step 1: Define CRD

**File**: `api/cloud-control/v1beta1/gcprediscluster_types.go`

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GcpRedisCluster struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    
    Spec   GcpRedisClusterSpec   `json:"spec,omitempty"`
    Status GcpRedisClusterStatus `json:"status,omitempty"`
}

type GcpRedisClusterSpec struct {
    // +kubebuilder:validation:Required
    RemoteRef RemoteRef `json:"remoteRef"`
    
    // +kubebuilder:validation:Required
    Scope ScopeRef `json:"scope"`
    
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=3
    ShardCount int32 `json:"shardCount"`
    
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum=REDIS_STANDARD_SMALL;REDIS_HIGHMEM_MEDIUM
    NodeType string `json:"nodeType"`
}

type GcpRedisClusterStatus struct {
    State      StatusState        `json:"state,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    Id         string             `json:"id,omitempty"`
    Endpoint   string             `json:"endpoint,omitempty"`
}
```

**Generate**:
```bash
make manifests
make generate
```

### Step 2: Create Provider Package

```bash
mkdir -p pkg/kcp/provider/gcp/rediscluster
mkdir -p pkg/kcp/provider/gcp/rediscluster/client
```

### Step 3: Implement State

**File**: `pkg/kcp/provider/gcp/rediscluster/state.go`

```go
package rediscluster

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
)

type State struct {
    focal.State
    
    client       RedisClusterClient
    redisCluster *redisclusterpb.Cluster
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    clientProvider func(ctx context.Context, scope *cloudcontrolv1beta1.Scope) RedisClusterClient
}

func NewStateFactory(
    clientProvider func(ctx context.Context, scope *cloudcontrolv1beta1.Scope) RedisClusterClient,
) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:  focalState,
        client: f.clientProvider(ctx, focalState.Scope()),
    }, nil
}

func (s *State) ObjAsGcpRedisCluster() *cloudcontrolv1beta1.GcpRedisCluster {
    return s.Obj().(*cloudcontrolv1beta1.GcpRedisCluster)
}
```

### Step 4: Implement Actions

**File**: `pkg/kcp/provider/gcp/rediscluster/loadRedis.go`

```go
func loadRedis(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    cluster := state.ObjAsGcpRedisCluster()
    if cluster.Status.Id == "" {
        return nil, ctx  // Not created yet
    }
    
    logger.Info("Loading GCP Redis Cluster")
    
    redisCluster, err := state.client.GetCluster(ctx, cluster.Status.Id)
    if err != nil {
        if IsNotFound(err) {
            logger.Info("Redis Cluster not found in GCP")
            state.redisCluster = nil
            return nil, ctx
        }
        return err, ctx
    }
    
    state.redisCluster = redisCluster
    return nil, ctx
}
```

**File**: `pkg/kcp/provider/gcp/rediscluster/createRedis.go`

```go
func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    if state.redisCluster != nil {
        return nil, ctx  // Already exists
    }
    
    cluster := state.ObjAsGcpRedisCluster()
    logger.Info("Creating GCP Redis Cluster")
    
    op, err := state.client.CreateCluster(ctx, &redisclusterpb.CreateClusterRequest{
        Parent:    state.Scope().Spec.Scope.Gcp.Project,
        ClusterId: cluster.Name,
        Cluster: &redisclusterpb.Cluster{
            ShardCount: cluster.Spec.ShardCount,
            NodeType:   redisclusterpb.NodeType(cluster.Spec.NodeType),
        },
    })
    
    if err != nil {
        logger.Error(err, "Failed to create Redis Cluster")
        state.SetCondition(
            cloudcontrolv1beta1.ConditionTypeError,
            metav1.ConditionTrue,
            cloudcontrolv1beta1.ReasonCloudProviderError,
            err.Error(),
        )
        _ = state.UpdateObjStatus(ctx)
        return err, ctx
    }
    
    cluster.Status.Id = op.Name
    err = state.UpdateObjStatus(ctx)
    if err != nil {
        return err, ctx
    }
    
    return nil, ctx
}
```

**File**: `pkg/kcp/provider/gcp/rediscluster/waitRedisReady.go`

```go
func waitRedisReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    if state.redisCluster == nil {
        return fmt.Errorf("redisCluster not loaded"), ctx
    }
    
    if state.redisCluster.State != redisclusterpb.Cluster_ACTIVE {
        logger.Info("Waiting for Redis Cluster to become active",
            "state", state.redisCluster.State,
        )
        return composed.StopWithRequeueDelay(3 * time.Second), nil
    }
    
    logger.Info("Redis Cluster is active")
    return nil, ctx
}
```

**File**: `pkg/kcp/provider/gcp/rediscluster/deleteRedis.go`

```go
func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    if state.redisCluster == nil {
        logger.Info("Redis Cluster already deleted")
        return nil, ctx
    }
    
    logger.Info("Deleting GCP Redis Cluster")
    
    err := state.client.DeleteCluster(ctx, state.redisCluster.Name)
    if err != nil && !IsNotFound(err) {
        logger.Error(err, "Failed to delete Redis Cluster")
        return err, ctx
    }
    
    return nil, ctx
}
```

**File**: `pkg/kcp/provider/gcp/rediscluster/updateStatus.go`

```go
func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    cluster := state.ObjAsGcpRedisCluster()
    
    if state.redisCluster == nil {
        cluster.Status.State = cloudcontrolv1beta1.StateCreating
        state.SetCondition(
            cloudcontrolv1beta1.ConditionTypeProcessing,
            metav1.ConditionTrue,
            "Creating",
            "Redis Cluster is being created",
        )
    } else if state.redisCluster.State == redisclusterpb.Cluster_ACTIVE {
        cluster.Status.State = cloudcontrolv1beta1.StateReady
        cluster.Status.Endpoint = state.redisCluster.DiscoveryEndpoints[0].Address
        state.SetCondition(
            cloudcontrolv1beta1.ConditionTypeReady,
            metav1.ConditionTrue,
            "Ready",
            "Redis Cluster is ready",
        )
    }
    
    return composed.UpdateStatus(state).Run(ctx, state)
}
```

### Step 5: Compose Actions in Reconciler

**File**: `pkg/kcp/provider/gcp/rediscluster/reconcile.go`

```go
package rediscluster

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    ctrl "sigs.k8s.io/controller-runtime"
)

type reconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory
}

func NewGcpRedisClusterReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    stateFactory StateFactory,
) *reconciler {
    return &reconciler{
        composedStateFactory: composedStateFactory,
        focalStateFactory:    focalStateFactory,
        stateFactory:         stateFactory,
    }
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.composedStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.GcpRedisCluster{})
    return composed.Handle(r.newAction()(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster-main",
        composed.LoadObj,
        focal.New(),
        r.newFlow(),
    )
}

func (r *reconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        focalState := st.(focal.State)
        
        state, err := r.stateFactory.NewState(ctx, focalState)
        if err != nil {
            return err, ctx
        }
        
        return composed.ComposeActions(
            "gcpRedisCluster-flow",
            actions.AddCommonFinalizer(),
            loadRedis,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    composed.IfElse(
                        func(ctx context.Context, st composed.State) bool {
                            state := st.(*State)
                            return state.redisCluster == nil
                        },
                        createRedis,
                        composed.ComposeActions("update",
                            // Add update actions if needed
                        ),
                    ),
                    waitRedisReady,
                    updateStatus,
                    composed.StopAndForgetAction,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteRedis,
                    waitRedisDeleted,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
        )(ctx, state)
    }
}

func (r *reconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&cloudcontrolv1beta1.GcpRedisCluster{}).
        Complete(r)
}
```

### Step 6: Wire Up in main.go

**File**: `cmd/main.go`

```go
// Import
import (
    gcprediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster"
    gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
)

// In main() function, after other reconcilers:

// GCP Redis Cluster client provider
redisClusterClientProvider := gcpredisclusterclient.NewRedisClusterClientProvider(gcpClients)

// Create state factory
redisClusterStateFactory := gcprediscluster.NewStateFactory(redisClusterClientProvider)

// Register reconciler
if err = gcprediscluster.NewGcpRedisClusterReconciler(
    composedStateFactory,
    focalStateFactory,
    redisClusterStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "GcpRedisCluster")
    os.Exit(1)
}
```

### Step 7: Create Tests

**File**: `internal/controller/cloud-control/gcprediscluster_test.go`

```go
var _ = Describe("Feature: KCP GcpRedisCluster", func() {
    
    It("Scenario: GcpRedisCluster is created and deleted", func() {
        scope := &cloudcontrolv1beta1.Scope{}
        cluster := &cloudcontrolv1beta1.GcpRedisCluster{}
        
        By("Given Scope exists", func() {
            Eventually(CreateScopeGcp).
                WithArguments(infra.Ctx(), infra, scope).
                Should(Succeed())
        })
        
        By("When GcpRedisCluster is created", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    WithScope(scope.Name),
                    WithShardCount(3),
                ).Should(Succeed())
        })
        
        By("Then Redis Cluster has ID assigned", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    NewObjActions(),
                    HavingRedisClusterStatusId(),
                ).Should(Succeed())
        })
        
        By("And Redis Cluster exists in mock", func() {
            gcpCluster := infra.GcpMock().GetRedisClusterByName(cluster.Status.Id)
            Expect(gcpCluster).NotTo(BeNil())
            Expect(gcpCluster.ShardCount).To(Equal(int32(3)))
        })
        
        By("When GCP marks cluster as ready", func() {
            infra.GcpMock().SetRedisClusterState(cluster.Status.Id, "ACTIVE")
        })
        
        By("Then GcpRedisCluster has Ready condition", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    NewObjActions(),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                    HavingState(cloudcontrolv1beta1.StateReady),
                ).WithTimeout(5 * time.Second).
                Should(Succeed())
        })
        
        // DELETE
        
        By("When GcpRedisCluster is deleted", func() {
            Eventually(Delete).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster).
                Should(Succeed())
        })
        
        By("Then Redis Cluster is deleted from GCP", func() {
            Eventually(func() bool {
                return infra.GcpMock().GetRedisClusterByName(cluster.Status.Id) == nil
            }).Should(BeTrue())
        })
        
        By("And GcpRedisCluster does not exist", func() {
            Eventually(IsDeleted).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster).
                Should(Succeed())
        })
    })
})
```

### Step 8: Run Tests

```bash
# Run all tests
make test

# Run specific test
go test ./internal/controller/cloud-control -v -ginkgo.focus="GcpRedisCluster"
```

---

## Task: Modify Existing API

### Step 1: Update Type Definition

**Example**: Add replica count to GcpRedisCluster

**File**: `api/cloud-control/v1beta1/gcprediscluster_types.go`

```go
type GcpRedisClusterSpec struct {
    // ... existing fields
    
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:validation:Maximum=5
    ReplicaCount int32 `json:"replicaCount,omitempty"`  // NEW FIELD
}
```

### Step 2: Generate Code

```bash
make manifests
make generate
```

### Step 3: Update Reconciler Logic

**File**: `pkg/kcp/provider/gcp/rediscluster/updateRedis.go` (create if needed)

```go
func updateRedis(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    cluster := state.ObjAsGcpRedisCluster()
    
    // Check if replica count changed
    if state.redisCluster.ReplicaCount != cluster.Spec.ReplicaCount {
        logger.Info("Updating replica count",
            "old", state.redisCluster.ReplicaCount,
            "new", cluster.Spec.ReplicaCount,
        )
        
        _, err := state.client.UpdateCluster(ctx, &redisclusterpb.UpdateClusterRequest{
            Cluster: &redisclusterpb.Cluster{
                Name:         state.redisCluster.Name,
                ReplicaCount: cluster.Spec.ReplicaCount,
            },
            UpdateMask: &fieldmaskpb.FieldMask{
                Paths: []string{"replica_count"},
            },
        })
        
        if err != nil {
            logger.Error(err, "Failed to update Redis Cluster")
            return err, ctx
        }
    }
    
    return nil, ctx
}
```

### Step 4: Add to Reconciler Flow

**File**: `pkg/kcp/provider/gcp/rediscluster/reconcile.go`

```go
composed.IfElse(
    func(ctx context.Context, st composed.State) bool {
        state := st.(*State)
        return state.redisCluster == nil
    },
    createRedis,
    updateRedis,  // ✅ Added update action
),
```

### Step 5: Update Tests

```go
It("Scenario: Replica count can be updated", func() {
    cluster := &cloudcontrolv1beta1.GcpRedisCluster{}
    
    By("Given cluster exists with 0 replicas", func() {
        Eventually(CreateGcpRedisCluster).
            WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                WithReplicaCount(0),
            ).Should(Succeed())
    })
    
    By("When replica count is changed to 3", func() {
        Eventually(Update).
            WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                func(obj *cloudcontrolv1beta1.GcpRedisCluster) {
                    obj.Spec.ReplicaCount = 3
                },
            ).Should(Succeed())
    })
    
    By("Then cluster is updated in GCP", func() {
        Eventually(func() int32 {
            gcpCluster := infra.GcpMock().GetRedisClusterByName(cluster.Status.Id)
            return gcpCluster.ReplicaCount
        }).WithTimeout(5 * time.Second).
        Should(Equal(int32(3)))
    })
})
```

---

## Task: Modify SKR API

### Additional Steps for SKR

**After API change**:

```bash
# Generate CRDs
make manifests
make generate

# Add/update version annotation
# Edit config/patchAfterMakeManifests.sh

# Run patch script
./config/patchAfterMakeManifests.sh

# Sync to distribution
./config/sync.sh
```

**Patch script example**:

```bash
# In config/patchAfterMakeManifests.sh
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.2"' \
  $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumes.yaml
```

---

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Test File

```bash
go test ./internal/controller/cloud-control -v -ginkgo.focus="GcpRedisCluster"
```

### Run Single Test Case

```bash
go test ./internal/controller/cloud-control -v \
  -ginkgo.focus="GcpRedisCluster" \
  -ginkgo.focus="is created and deleted"
```

### Verbose Output

```bash
go test ./internal/controller/cloud-control -v -ginkgo.v
```

### Feature Flag Validation

```bash
make test-ff
```

---

## Building and Running

### Build Binary

```bash
make build
./bin/manager --help
```

### Run Locally

```bash
# Install CRDs to cluster
make install

# Run controller (connects to cluster in kubeconfig)
make run
```

### Build Docker Image

```bash
make docker-build IMG=my-registry/cloud-manager:dev
docker push my-registry/cloud-manager:dev
```

### Deploy to Cluster

```bash
make deploy IMG=my-registry/cloud-manager:dev

# Check deployment
kubectl get pods -n kyma-system

# View logs
kubectl logs -n kyma-system deployment/cloud-manager-controller-manager -f
```

---

## Debugging

### Add Logging

```go
func myAction(ctx context.Context, st composed.State) (error, context.Context) {
    logger := composed.LoggerFromCtx(ctx)
    state := st.(*State)
    
    logger.Info("Action executing",
        "resourceName", state.Name(),
        "key", "value",
    )
    
    if err != nil {
        logger.Error(err, "Operation failed",
            "resourceName", state.Name(),
        )
    }
    
    return nil, ctx
}
```

### View Logs

```bash
# Local run
make run

# Deployed controller
kubectl logs -n kyma-system deployment/cloud-manager-controller-manager -f

# Filter by resource
kubectl logs -n kyma-system deployment/cloud-manager-controller-manager | grep GcpRedisCluster
```

### Inspect Resources

```bash
# List all
kubectl get gcpredisclusters -A

# Describe with events
kubectl describe gcprediscluster my-cluster -n my-namespace

# View YAML
kubectl get gcprediscluster my-cluster -n my-namespace -o yaml

# Check conditions
kubectl get gcprediscluster my-cluster -o jsonpath='{.status.conditions}'
```

### Common Issues

| Symptom | Possible Cause | Solution |
|---------|----------------|----------|
| Resource stuck in Creating | Async operation not complete | Check mock state, verify wait action |
| Status not updating | Missing UpdateObjStatus() | Add status update with error checking |
| Resource not deleted | Finalizer not removed | Check delete flow removes finalizer |
| Test fails with timeout | Eventually() timeout too short | Increase timeout or check mock state |
| CRD changes not reflected | Forgot make manifests | Run `make manifests` and `make install` |

---

## Before Committing Checklist

- [ ] Run `make test` - all tests pass
- [ ] Run `make fmt` - code formatted
- [ ] Run `make lint` - no lint errors
- [ ] Run `make manifests` if API changed
- [ ] Run `./config/patchAfterMakeManifests.sh` if SKR API changed
- [ ] Run `./config/sync.sh` if SKR API changed
- [ ] Verify NEW pattern used for new resources
- [ ] Check all actions end with `StopAndForget` or return value
- [ ] Verify status updates check errors
- [ ] Confirm tests use `testinfra` mocks, not real APIs
- [ ] Ensure `Eventually()` wraps reconciliation assertions

---

## Troubleshooting

### Tests Fail with "context canceled"

**Cause**: Context timeout.

**Solution**: Check if async operations complete. Update mock state if needed.

```go
// In test
By("When GCP marks resource as ready", func() {
    infra.GcpMock().SetResourceState(id, "READY")
})
```

### CRD Changes Not Reflected

**Cause**: Forgot `make manifests`.

**Solution**:
```bash
make manifests
make install  # If running locally
```

### Import Cycle Error

**Cause**: Circular dependency.

**Solution**: Restructure imports. Move shared types to common package.

### Build Fails After Dependency Update

**Cause**: Go module cache out of sync.

**Solution**:
```bash
go clean -modcache
go mod download
go mod tidy
make build
```

---

## Getting Help

**Can't find pattern?** → [QUICK_REFERENCE.md](reference/QUICK_REFERENCE.md)

**Need structure info?** → [PROJECT_STRUCTURE.md](reference/PROJECT_STRUCTURE.md)

**Encountering error?** → [COMMON_PITFALLS.md](reference/COMMON_PITFALLS.md)

**Creating reconciler?** → [ADD_KCP_RECONCILER.md](guides/ADD_KCP_RECONCILER.md)

**Writing tests?** → [CONTROLLER_TESTS.md](guides/CONTROLLER_TESTS.md)
