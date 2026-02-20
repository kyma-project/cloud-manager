# GCP NFS Volume Backup Refactoring Plan

**Target Audience**: LLM coding agents and developers  
**Last Updated**: 2026-02-13  
**Estimated Effort**: High complexity (SKR-only pattern with client refactoring)

---

## Executive Summary

Refactor `pkg/skr/gcpnfsvolumebackup` to support dual implementations:
- **v1**: Current implementation using `google.golang.org/api/file/v1` (legacy)
- **v2**: New implementation using `cloud.google.com/go/filestore/apiv1` (modern)

Feature flag `gcpBackupV2` controls instantiation (defaults to v1).

### Special Considerations

This reconciler is a **SKR-only implementation** — it has no backing KCP component. It directly uses GCP client from `pkg/kcp/provider/gcp/nfsbackup/client`. This is different from standard patterns where SKR creates KCP resources, and KCP reconcilers manage cloud providers.

---

## Current Architecture Analysis

### Files to Refactor

**SKR Reconciler** (`pkg/skr/gcpnfsvolumebackup/`):
| File | Purpose | v1/v2 Notes |
|------|---------|-------------|
| `reconciler.go` | Entry point, action composition | Split to route v1/v2 |
| `state.go` | State struct with FileBackupClient | Version-specific states |
| `createNfsBackup.go` | Create backup in GCP | Uses `file.Backup` types |
| `deleteNfsBackup.go` | Delete backup from GCP | Uses `file.Backup` types |
| `loadNfsBackup.go` | Load backup from GCP | Uses `file.Backup` types |
| `checkBackupOperation.go` | Check async operation status | Uses `file.Operation` types |
| `addLabelsToNfsBackup.go` | Add labels to GCP backup | Uses `file.Backup` types |
| `loadScope.go` | Load Scope from KCP | Shared (version-agnostic) |
| `loadGcpNfsVolume.go` | Load GcpNfsVolume from SKR | Shared (version-agnostic) |
| `addFinalizer.go` | Add finalizer | Shared (version-agnostic) |
| `removeFinalizer.go` | Remove finalizer | Shared (version-agnostic) |
| `updateStatus.go` | Update SKR status | Version-specific (types differ) |
| `updateCapacity.go` | Update capacity in status | Version-specific (types differ) |
| `shortCircuit.go` | Short-circuit optimization | Shared (version-agnostic) |
| `markFailed.go` | Mark backup as failed | Shared (version-agnostic) |
| `mirrorLabelsToStatus.go` | Mirror labels to status | Version-specific (types differ) |
| `ignorant.go` | Test ignore list | Shared |

**Client** (`pkg/kcp/provider/gcp/nfsbackup/client/`):
| File | Purpose | Notes |
|------|---------|-------|
| `fileBackupClient.go` | FileBackupClient using old library | Move to v1, create v2 |

---

## Refactoring Steps

### Phase 1: Create Feature Flag

**Files to create:**
- `pkg/feature/ffGcpBackupV2.go`

**Files to modify:**
- `pkg/feature/ff_ga.yaml`
- `pkg/feature/ff_edge.yaml`
- `config/featureToggles/featureToggles.local.yaml`

#### Step 1.1: Create Feature Flag Definition

Create `pkg/feature/ffGcpBackupV2.go`:
```go
package feature

import (
    "context"
)

const GcpBackupV2FlagName = "gcpBackupV2"

var GcpBackupV2 = &gcpBackupV2Info{}

type gcpBackupV2Info struct{}

func (k *gcpBackupV2Info) Value(ctx context.Context) bool {
    return provider.BoolVariation(ctx, GcpBackupV2FlagName, false)
}
```

#### Step 1.2: Add to Feature Flag Configuration Files

Add to `pkg/feature/ff_ga.yaml`:
```yaml
gcpBackupV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
```

Add to `pkg/feature/ff_edge.yaml`:
```yaml
gcpBackupV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
```

Add to `config/featureToggles/featureToggles.local.yaml`:
```yaml
gcpBackupV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: enabled  # Enable for local development
```

---

### Phase 2: Create v1 Directory Structure (Move Current Implementation)

**Goal**: Move existing implementation to `v1/` folder without behavioral changes.

#### Step 2.1: Create v1 Client Directory

Create `pkg/kcp/provider/gcp/nfsbackup/client/v1/` directory and move:
- Move `pkg/kcp/provider/gcp/nfsbackup/client/fileBackupClient.go` → `pkg/kcp/provider/gcp/nfsbackup/client/v1/fileBackupClient.go`
- Update package declaration to `package v1`
- Update imports to reflect new location

#### Step 2.2: Create v1 SKR Directory

Create `pkg/skr/gcpnfsvolumebackup/v1/` directory containing all actions (no shared package):
- `reconciler.go` - Action composition for v1
- `state.go` - State with v1 FileBackupClient
- `createNfsBackup.go` - Create using old types
- `deleteNfsBackup.go` - Delete using old types
- `loadNfsBackup.go` - Load using old types  
- `checkBackupOperation.go` - Check operation using old types
- `addLabelsToNfsBackup.go` - Add labels using old types
- `loadScope.go` - Load Scope from KCP
- `loadGcpNfsVolume.go` - Load GcpNfsVolume from SKR
- `updateStatus.go` - Update status using old types
- `updateCapacity.go` - Update capacity using old types
- `mirrorLabelsToStatus.go` - Mirror labels using old types
- `shortCircuit.go` - Short-circuit logic
- `markFailed.go` - Mark failed logic
- `ignorant.go` - Test ignore list

**Note**: Each version (v1, v2) has its own complete set of actions. No shared/common package exists. This simplifies the architecture and avoids interface complexity.

---

### Phase 3: Create v2 Client Implementation

**Goal**: Create FileBackupClient using `cloud.google.com/go/filestore/apiv1`.

#### Step 3.1: Create v2 Client

Create `pkg/kcp/provider/gcp/nfsbackup/client/v2/fileBackupClient.go`:

```go
package v2

import (
    "context"
    "fmt"

    filestore "cloud.google.com/go/filestore/apiv1"
    "cloud.google.com/go/filestore/apiv1/filestorepb"
    "cloud.google.com/go/longrunning/autogen/longrunningpb"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
)

// FileBackupClient defines operations for GCP Filestore backups.
type FileBackupClient interface {
    GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error)
    ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error)
    CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error)
    DeleteBackup(ctx context.Context, projectId, location, name string) (string, error)
    GetOperation(ctx context.Context, operationName string) (bool, error)
    UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error)
}

// NewFileBackupClientProvider creates a provider for FileBackupClient.
// Follows NEW pattern - accesses clients from GcpClients singleton.
func NewFileBackupClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileBackupClient] {
    return func() FileBackupClient {
        return NewFileBackupClient(gcpClients)
    }
}

// NewFileBackupClient creates a new FileBackupClient wrapping GcpClients.
func NewFileBackupClient(gcpClients *gcpclient.GcpClients) FileBackupClient {
    return &fileBackupClient{
        filestoreManager: gcpClients.Filestore,
    }
}

type fileBackupClient struct {
    filestoreManager *filestore.CloudFilestoreManagerClient
}

var _ FileBackupClient = &fileBackupClient{}

func (c *fileBackupClient) GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error) {
    logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "name", name)

    req := &filestorepb.GetBackupRequest{
        Name: formatBackupName(projectId, location, name),
    }

    backup, err := c.filestoreManager.GetBackup(ctx, req)
    if err != nil {
        if gcpmeta.IsNotFound(err) {
            logger.Info("target Filestore backup not found")
            return nil, err
        }
        logger.Error(err, "Failed to get Filestore backup")
        return nil, err
    }

    return backup, nil
}

func (c *fileBackupClient) ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error) {
    logger := composed.LoggerFromCtx(ctx)

    req := &filestorepb.ListBackupsRequest{
        Parent: formatParentPath(projectId, "-"),  // "-" for all locations
        Filter: filter,
    }

    var backups []*filestorepb.Backup
    it := c.filestoreManager.ListBackups(ctx, req)
    for {
        backup, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            logger.Error(err, "Failed to list backups")
            return nil, err
        }
        backups = append(backups, backup)
    }

    return backups, nil
}

func (c *fileBackupClient) CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error) {
    logger := composed.LoggerFromCtx(ctx)

    req := &filestorepb.CreateBackupRequest{
        Parent:   formatParentPath(projectId, location),
        BackupId: name,
        Backup:   backup,
    }

    op, err := c.filestoreManager.CreateBackup(ctx, req)
    if err != nil {
        logger.Error(err, "Failed to create Filestore backup",
            "projectId", projectId,
            "location", location,
            "name", name)
        return "", err
    }

    return op.Name(), nil
}

func (c *fileBackupClient) DeleteBackup(ctx context.Context, projectId, location, name string) (string, error) {
    logger := composed.LoggerFromCtx(ctx).WithValues("projectId", projectId, "location", location, "name", name)

    req := &filestorepb.DeleteBackupRequest{
        Name: formatBackupName(projectId, location, name),
    }

    op, err := c.filestoreManager.DeleteBackup(ctx, req)
    if err != nil {
        if gcpmeta.IsNotFound(err) {
            logger.Info("target Filestore backup not found")
            return "", err
        }
        logger.Error(err, "Failed to delete Filestore backup")
        return "", err
    }

    return op.Name(), nil
}

func (c *fileBackupClient) GetOperation(ctx context.Context, operationName string) (bool, error) {
    logger := composed.LoggerFromCtx(ctx)

    req := &longrunningpb.GetOperationRequest{
        Name: operationName,
    }

    op, err := c.filestoreManager.LROClient.GetOperation(ctx, req)
    if err != nil {
        logger.Error(err, "Failed to get operation", "operationName", operationName)
        return false, err
    }

    return op.Done, nil
}

func (c *fileBackupClient) UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error) {
    logger := composed.LoggerFromCtx(ctx)

    backup.Name = formatBackupName(projectId, location, name)

    req := &filestorepb.UpdateBackupRequest{
        Backup: backup,
        UpdateMask: &fieldmaskpb.FieldMask{
            Paths: updateMask,
        },
    }

    op, err := c.filestoreManager.UpdateBackup(ctx, req)
    if err != nil {
        logger.Error(err, "Failed to update Filestore backup")
        return "", err
    }

    return op.Name(), nil
}

// Helper functions
func formatBackupName(projectId, location, name string) string {
    return fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
}

func formatParentPath(projectId, location string) string {
    return fmt.Sprintf("projects/%s/locations/%s", projectId, location)
}
```

#### Step 3.2: Create v2 Client Utility

Create `pkg/kcp/provider/gcp/nfsbackup/client/v2/util.go`:
```go
package v2

import (
    "regexp"
)

var backupPathRegex = regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/backups/([^/]+)$`)

// GetProjectLocationNameFromBackupPath extracts components from a backup path
func GetProjectLocationNameFromBackupPath(path string) (projectId, location, name string) {
    matches := backupPathRegex.FindStringSubmatch(path)
    if len(matches) != 4 {
        return "", "", ""
    }
    return matches[1], matches[2], matches[3]
}
```

---

### Phase 4: Create v2 SKR Implementation

**Goal**: Create SKR reconciler logic using v2 client with protobuf types.

#### Step 4.1: Create v2 State

Create `pkg/skr/gcpnfsvolumebackup/v2/state.go`:

```go
package v2

import (
    "context"
    "fmt"
    "slices"
    "sort"
    "strings"
    "time"

    "cloud.google.com/go/filestore/apiv1/filestorepb"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
    v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
    "k8s.io/klog/v2"
)

// State represents the v2 state using modern GCP protobuf types
type State struct {
    composed.State
    KymaRef    klog.ObjectRef
    KcpCluster composed.StateCluster
    SkrCluster composed.StateCluster

    Scope        *cloudcontrolv1beta1.Scope
    GcpNfsVolume *cloudresourcesv1beta1.GcpNfsVolume

    fileBackup       *filestorepb.Backup  // Modern protobuf type
    fileBackupClient v2client.FileBackupClient
}

// State accessors
func (s *State) GetKymaRef() klog.ObjectRef { return s.KymaRef }
func (s *State) GetKcpCluster() composed.StateCluster { return s.KcpCluster }
func (s *State) GetSkrCluster() composed.StateCluster { return s.SkrCluster }
func (s *State) GetScope() *cloudcontrolv1beta1.Scope { return s.Scope }
func (s *State) SetScope(scope *cloudcontrolv1beta1.Scope) { s.Scope = scope }
func (s *State) GetGcpNfsVolume() *cloudresourcesv1beta1.GcpNfsVolume { return s.GcpNfsVolume }
func (s *State) SetGcpNfsVolume(vol *cloudresourcesv1beta1.GcpNfsVolume) { s.GcpNfsVolume = vol }

func (s *State) ObjAsGcpNfsVolumeBackup() *cloudresourcesv1beta1.GcpNfsVolumeBackup {
    return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
}

// Backup state accessors (abstract away protobuf details)
func (s *State) HasFileBackup() bool { return s.fileBackup != nil }
func (s *State) GetFileBackupState() string {
    if s.fileBackup == nil {
        return ""
    }
    return s.fileBackup.State.String()
}
func (s *State) GetFileBackupStorageBytes() int64 {
    if s.fileBackup == nil {
        return 0
    }
    return s.fileBackup.StorageBytes
}
func (s *State) GetFileBackupLabels() map[string]string {
    if s.fileBackup == nil {
        return nil
    }
    return s.fileBackup.Labels
}

// GetFileBackup returns the raw backup object (for v2-specific usage)
func (s *State) GetFileBackup() *filestorepb.Backup { return s.fileBackup }

// SetFileBackup sets the backup object
func (s *State) SetFileBackup(backup *filestorepb.Backup) { s.fileBackup = backup }

// GetFileBackupClient returns the v2 client
func (s *State) GetFileBackupClient() v2client.FileBackupClient { return s.fileBackupClient }

// ... (additional helper methods similar to v1 but using protobuf types)

type StateFactory interface {
    NewState(ctx context.Context, baseState composed.State) (*State, error)
}

// stateFactory creates State instances
type stateFactory struct {
    kymaRef                  klog.ObjectRef
    kcpCluster               composed.StateCluster
    skrCluster               composed.StateCluster
    fileBackupClientProvider gcpclient.GcpClientProvider[v2client.FileBackupClient]
}

func NewStateFactory(
    kymaRef klog.ObjectRef,
    kcpCluster, skrCluster composed.StateCluster,
    fileBackupClientProvider gcpclient.GcpClientProvider[v2client.FileBackupClient],
) StateFactory {
    return &stateFactory{
        kymaRef:                  kymaRef,
        kcpCluster:               kcpCluster,
        skrCluster:               skrCluster,
        fileBackupClientProvider: fileBackupClientProvider,
    }
}

func (f *stateFactory) NewState(ctx context.Context, baseState composed.State) (*State, error) {
    return &State{
        State:            baseState,
        KymaRef:          f.kymaRef,
        KcpCluster:       f.kcpCluster,
        SkrCluster:       f.skrCluster,
        fileBackupClient: f.fileBackupClientProvider(),
    }, nil
}
```

#### Step 4.2: Create v2 Reconciler

Create `pkg/skr/gcpnfsvolumebackup/v2/reconcile.go`:

```go
package v2

import (
    "context"
    "fmt"

    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// New creates the v2 reconciliation action
func New(stateFactory StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        logger := composed.LoggerFromCtx(ctx)
        logger.Info("Using GcpNfsVolumeBackup v2 implementation")

        state, err := stateFactory.NewState(ctx, st)
        if err != nil {
            logger.Error(err, "Error creating v2 state")
            backup := st.Obj().(*cloudresourcesv1beta1.GcpNfsVolumeBackup)
            return composed.PatchStatus(backup).
                SetExclusiveConditions(metav1.Condition{
                    Type:    cloudresourcesv1beta1.ConditionTypeError,
                    Status:  metav1.ConditionTrue,
                    Reason:  cloudresourcesv1beta1.ConditionReasonError,
                    Message: fmt.Sprintf("Failed to initialize GCP client: %s", err.Error()),
                }).
                SuccessError(composed.StopAndForget).
                Run(ctx, st)
        }

        return composeActions()(ctx, state)
    }
}

func composeActions() composed.Action {
    return composed.ComposeActions(
        "gcpNfsVolumeBackupV2",
        loadScope,
        shortCircuitCompleted,
        markFailed,
        actions.AddCommonFinalizer(),
        loadNfsBackup,
        loadGcpNfsVolume,
        // Check pending operation (shared - works for both create and delete operations)
        checkBackupOperation,
        composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "gcpNfsVolumeBackupV2-create",
                // Create backup if not exists (idempotent - skips if backup already exists)
                createNfsBackup,
                // Wait for backup to be READY (redundancy if OpIdentifier wasn't saved)
                waitBackupReady,
                // Add labels to existing ready backup
                addLabelsToNfsBackup,
                // Update status (Ready condition, capacity, labels mirror)
                updateStatus,
            ),
            composed.ComposeActions(
                "gcpNfsVolumeBackupV2-delete",
                // Delete backup (idempotent - skips if backup doesn't exist)
                deleteNfsBackup,
                // Wait for backup to be deleted (redundancy if OpIdentifier wasn't saved)
                waitBackupDeleted,
                actions.RemoveCommonFinalizer(),
            ),
        ),
        composed.StopAndForgetAction,
    )
}
```

**Design Notes:**

1. **`checkBackupOperation` moved before branching**: Since this action is needed in both create and delete flows, it's placed before the `IfElse`. The logic works because:
   - If no pending operation (`Status.OpIdentifier` empty), it continues to next action
   - If operation pending, it requeues with delay
   - When operation completes, it clears `OpIdentifier` and continues
   - The subsequent create/delete actions are **idempotent** (they check backup state and skip if already done)

2. **`waitBackupReady` / `waitBackupDeleted` for redundancy**: These actions provide a safety net in case `OpIdentifier` wasn't properly saved:
   - `waitBackupReady` - Checks if `backup.State == READY`, otherwise requeues with delay (similar to `waitRedisAvailable` in redis)
   - `waitBackupDeleted` - Checks if backup is nil (NotFound), otherwise requeues if state is `DELETING` (similar to `waitRedisDeleted` in redis)
   - See [pkg/kcp/provider/gcp/redisinstance/waitRedisAvailable.go](pkg/kcp/provider/gcp/redisinstance/waitRedisAvailable.go) for the pattern

3. **Action Order (Create Flow)**: The flow reads top-to-bottom as the behavior:
   - `loadNfsBackup` - Load current backup state from GCP
   - `checkBackupOperation` - If operation pending, wait; otherwise continue
   - `createNfsBackup` - Create backup in GCP (skips if `state.fileBackup != nil`)
   - `waitBackupReady` - Verify backup is READY (redundancy if OpIdentifier lost)
   - `addLabelsToNfsBackup` - Add labels to ready backup
   - `updateStatus` - Single action combining: Ready condition, capacity update, and labels mirror to status

4. **Action Order (Delete Flow)**:
   - `loadNfsBackup` - Load current backup state from GCP
   - `checkBackupOperation` - If operation pending, wait; otherwise continue
   - `deleteNfsBackup` - Delete backup from GCP (skips if `state.fileBackup == nil`)
   - `waitBackupDeleted` - Verify backup is deleted (redundancy if OpIdentifier lost)
   - `RemoveCommonFinalizer` - Remove finalizer to complete deletion

5. **Consolidated `updateStatus`**: Instead of separate `mirrorLabelsToStatus`, `updateCapacity` (which should be named `updateStatusCapacity`), and `updateStatus`, we combine all status updates into one action. This is cleaner and reduces API calls.

6. **Removed `stopAndRequeueForCapacity`**: The v1 implementation has `StopAndRequeueForCapacityAction()` which returns `composed.StopWithRequeueDelay(GcpCapacityCheckInterval)` for periodic capacity monitoring. For v2, this can be handled within `updateStatus` by returning appropriate requeue delay when capacity needs refreshing, or we can keep it as a separate concern if the periodic check is critical:

   ```go
   // Option A: Handle in updateStatus
   func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
       // ... update status logic ...
       
       // If ready and capacity monitoring needed, requeue with delay
       if state.isTimeForCapacityUpdate() {
           return composed.StopWithRequeueDelay(GcpCapacityCheckInterval), nil
       }
       return nil, nil  // Continue to StopAndForget
   }
   
   // Option B: Keep separate action (if periodic check always needed)
   func requeueForCapacityCheck(ctx context.Context, st composed.State) (error, context.Context) {
       return composed.StopWithRequeueDelay(GcpCapacityCheckInterval), nil
   }
   ```

7. **Delete Flow**: No redundant `StopAndForgetAction` - the outer one handles termination.

**Note**: The `composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate), ...)` pattern is the standard way to branch between create and delete flows. See [pkg/kcp/provider/gcp/redisinstance/new.go](pkg/kcp/provider/gcp/redisinstance/new.go#L38) for reference.

#### Step 4.3: Create v2 Actions

Create v2-specific actions using protobuf types (complete set, no shared package):
- `pkg/skr/gcpnfsvolumebackup/v2/loadNfsBackup.go` - Load backup from GCP
- `pkg/skr/gcpnfsvolumebackup/v2/createNfsBackup.go` - Create backup in GCP
- `pkg/skr/gcpnfsvolumebackup/v2/deleteNfsBackup.go` - Delete backup from GCP
- `pkg/skr/gcpnfsvolumebackup/v2/checkBackupOperation.go` - Check async operation status
- `pkg/skr/gcpnfsvolumebackup/v2/addLabelsToNfsBackup.go` - Add labels to existing ready backup
- `pkg/skr/gcpnfsvolumebackup/v2/updateStatus.go` - Combined status update (Ready condition, capacity, labels mirror)
- `pkg/skr/gcpnfsvolumebackup/v2/waitBackupReady.go` - Wait for backup state READY (redundancy for create)
- `pkg/skr/gcpnfsvolumebackup/v2/waitBackupDeleted.go` - Wait for backup deleted (redundancy for delete)
- `pkg/skr/gcpnfsvolumebackup/v2/loadScope.go` - Load Scope from KCP
- `pkg/skr/gcpnfsvolumebackup/v2/loadGcpNfsVolume.go` - Load GcpNfsVolume from SKR
- `pkg/skr/gcpnfsvolumebackup/v2/shortCircuit.go` - Short-circuit for already ready backups
- `pkg/skr/gcpnfsvolumebackup/v2/markFailed.go` - Mark backup as failed
- `pkg/skr/gcpnfsvolumebackup/v2/ignorant.go` - Test ignore list

**Removed from v1 (consolidated into `updateStatus`):**
- ~~`mirrorLabelsToStatus.go`~~ - Merged into `updateStatus`
- ~~`updateCapacity.go`~~ - Merged into `updateStatus` (was misnamed, actually updates status.Capacity)
- ~~`stopAndRequeueForCapacity.go`~~ - Handled as return value in `updateStatus` when capacity check needed

**Key Changes from v1**:
1. Use `*filestorepb.Backup` instead of `*file.Backup`
2. Use `State.String()` for backup state comparison instead of string literals
3. Use `gcpmeta.IsNotFound()` for error checking
4. Return operation names (strings) instead of `*file.Operation`
5. Use `GetOperation()` returning `(bool, error)` instead of full operation object

---

### Phase 5: Update Controller with Feature Flag Routing at Startup

**Implementation Approach Changed**: Instead of per-request routing in a root reconciler, feature flag is checked at startup in the controller factory. This eliminates per-request overhead and keeps v1 unchanged.

#### Step 5.1: Add Reconciler to v2

Add `Reconciler` struct to `pkg/skr/gcpnfsvolumebackup/v2/reconcile.go` matching v1's interface:

```go
type Reconciler struct {
    composedStateFactory composed.StateFactory
    stateFactory         StateFactory
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Same pattern as v1
}

func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
    fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]) Reconciler {
    // Initialize v2 reconciler
}
```

#### Step 5.2: Update Controller

Modify `internal/controller/cloud-resources/gcpnfsvolumebackup_controller.go`:

```go
// gcpNfsVolumeBackupRunner is a common interface for v1 and v2 reconcilers
type gcpNfsVolumeBackupRunner interface {
    Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

type GcpNfsVolumeBackupReconcilerFactory struct {
    fileBackupClientProviderV1 gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
    fileBackupClientProviderV2 gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
    env                        abstractions.Environment
    useV2                      bool
}

func (f *GcpNfsVolumeBackupReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
    if f.useV2 {
        reconciler := gcpnfsvolumebackupv2.NewReconciler(...)
        return &GcpNfsVolumeBackupReconciler{reconciler: &reconciler}
    }
    reconciler := gcpnfsvolumebackupv1.NewReconciler(...)
    return &GcpNfsVolumeBackupReconciler{reconciler: &reconciler}
}

func SetupGcpNfsVolumeBackupReconciler(...) error {
    // Check feature flag at startup
    useV2 := feature.GcpBackupV2.Value(context.Background())
    // Log and create factory with appropriate flag
}
```

#### Step 5.3: Update main.go

Pass both v1 and v2 client providers:

```go
if err = cloudresourcescontroller.SetupGcpNfsVolumeBackupReconciler(
    skrRegistry,
    gcpnfsbackupclientv1.NewFileBackupClientProvider(),
    gcpnfsbackupclientv2.NewFileBackupClientProvider(gcpClients),
    env,
    setupLog,
); err != nil {
    // ...
}
```

---

### Phase 6: Create v2 Mock Implementation

#### Step 6.1: Create v2 Mock Store

Create `pkg/kcp/provider/gcp/mock/nfsBackupStoreV2.go`:

```go
package mock

import (
    "context"
    "fmt"
    "regexp"

    "cloud.google.com/go/filestore/apiv1/filestorepb"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
    v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
)

type FileBackupClientFakeUtilsV2 interface {
    CreateFakeBackup(backup *filestorepb.Backup)
    ClearAllBackups()
}

type nfsBackupStoreV2 struct {
    backups []*filestorepb.Backup
}

var _ v2client.FileBackupClient = &nfsBackupStoreV2{}

func (s *nfsBackupStoreV2) CreateFakeBackup(backup *filestorepb.Backup) {
    s.backups = append(s.backups, backup)
}

func (s *nfsBackupStoreV2) ClearAllBackups() {
    s.backups = []*filestorepb.Backup{}
}

func (s *nfsBackupStoreV2) GetBackup(ctx context.Context, projectId, location, name string) (*filestorepb.Backup, error) {
    if isContextCanceled(ctx) {
        return nil, context.Canceled
    }
    
    completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
    logger := composed.LoggerFromCtx(ctx)
    
    for _, backup := range s.backups {
        if backup.Name == completeName {
            logger.WithName("GetBackup - mock v2").Info("Got Nfs Backup", "backup", backup.Name)
            backup.State = filestorepb.Backup_READY
            return backup, nil
        }
    }
    
    logger.WithName("GetBackup - mock v2").Info(fmt.Sprintf("Backup not found, total: %d", len(s.backups)))
    return nil, gcpmeta.NewNotFoundError()
}

func (s *nfsBackupStoreV2) ListBackups(ctx context.Context, projectId, filter string) ([]*filestorepb.Backup, error) {
    if isContextCanceled(ctx) {
        return nil, context.Canceled
    }

    // Similar filtering logic as v1, adapted for protobuf types
    result := make([]*filestorepb.Backup, 0)
    
    for _, backup := range s.backups {
        projectIdFromPath, _, _ := v2client.GetProjectLocationNameFromBackupPath(backup.Name)
        if projectIdFromPath != projectId {
            continue
        }
        // Apply filter logic...
        result = append(result, backup)
    }

    return result, nil
}

func (s *nfsBackupStoreV2) CreateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup) (string, error) {
    if isContextCanceled(ctx) {
        return "", context.Canceled
    }

    logger := composed.LoggerFromCtx(ctx)
    completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
    backup.Name = completeName
    
    for _, existing := range s.backups {
        if existing.Name == completeName {
            return "", gcpmeta.NewAlreadyExistsError()
        }
    }
    
    s.backups = append(s.backups, backup)
    logger.WithName("CreateBackup - mock v2").Info(fmt.Sprintf("Created backup, total: %d", len(s.backups)))
    
    return fmt.Sprintf("operations/create-%s", name), nil
}

func (s *nfsBackupStoreV2) DeleteBackup(ctx context.Context, projectId, location, name string) (string, error) {
    if isContextCanceled(ctx) {
        return "", context.Canceled
    }

    logger := composed.LoggerFromCtx(ctx)
    completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
    
    for i, backup := range s.backups {
        if backup.Name == completeName {
            s.backups = append(s.backups[:i], s.backups[i+1:]...)
            logger.WithName("DeleteBackup - mock v2").Info(fmt.Sprintf("Deleted backup, total: %d", len(s.backups)))
            return fmt.Sprintf("operations/delete-%s", name), nil
        }
    }
    
    return "", gcpmeta.NewNotFoundError()
}

func (s *nfsBackupStoreV2) GetOperation(ctx context.Context, operationName string) (bool, error) {
    if isContextCanceled(ctx) {
        return false, context.Canceled
    }
    // Mock always returns done
    return true, nil
}

func (s *nfsBackupStoreV2) UpdateBackup(ctx context.Context, projectId, location, name string, backup *filestorepb.Backup, updateMask []string) (string, error) {
    if isContextCanceled(ctx) {
        return "", context.Canceled
    }

    logger := composed.LoggerFromCtx(ctx)
    completeName := fmt.Sprintf("projects/%s/locations/%s/backups/%s", projectId, location, name)
    
    for i, existing := range s.backups {
        if existing.Name == completeName {
            s.backups[i] = backup
            s.backups[i].Name = completeName
            logger.WithName("UpdateBackup - mock v2").Info("Updated backup")
            return fmt.Sprintf("operations/update-%s", name), nil
        }
    }
    
    return "", gcpmeta.NewNotFoundError()
}
```

#### Step 6.2: Update Mock Server

Add v2 provider to `pkg/kcp/provider/gcp/mock/server.go`:

```go
// Add field to server struct
nfsBackupStoreV2 *nfsBackupStoreV2

// Add initialization in New()
nfsBackupStoreV2: &nfsBackupStoreV2{},

// Add provider method
func (s *server) FileBackupClientProviderV2() gcpclient.GcpClientProvider[v2client.FileBackupClient] {
    return func() v2client.FileBackupClient {
        return s.nfsBackupStoreV2
    }
}
```

---

### Phase 7: Update Controller and Tests

#### Step 7.1: Update Controller Setup

Modify `internal/controller/cloud-resources/gcpnfsvolumebackup_controller.go`:

```go
func (f *GcpNfsVolumeBackupReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
    return &GcpNfsVolumeBackupReconciler{
        Reconciler: gcpnfsvolumebackup.NewReconciler(
            args.KymaRef,
            args.KcpCluster,
            args.SkrCluster,
            f.fileBackupClientProviderV1,
            f.fileBackupClientProviderV2,
            f.env,
        ),
    }
}
```

#### Step 7.2: Update Test Suite Setup

Modify `internal/controller/cloud-resources/suite_test.go`:

```go
// GcpNfsVolumeBackup - provide both v1 and v2 providers
Expect(SetupGcpNfsVolumeBackupReconciler(
    infra.Registry(),
    infra.GcpMock().FileBackupClientProvider(),    // v1
    infra.GcpMock().FileBackupClientProviderV2(),  // v2
    env,
    testSetupLog,
)).NotTo(HaveOccurred())
```

#### Step 7.3: Create v2 Controller Tests

Create `internal/controller/cloud-resources/gcpnfsvolumebackup_v2_test.go`:

```go
package cloudresources

import (
    "time"

    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
    . "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackup V2", func() {

    const (
        interval = time.Millisecond * 50
    )
    var (
        timeout = time.Second * 20
    )

    BeforeEach(func() {
        // Enable v2 feature flag for these tests
        // (implementation depends on test infrastructure capabilities)
    })

    Describe("Scenario: SKR GcpNfsVolumeBackup V2 is created", func() {
        gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
        gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-v2-1"

        It("When GcpNfsVolumeBackup V2 Create is called", func() {
            // ... test implementation following existing patterns
        })
    })

    Describe("Scenario: SKR GcpNfsVolumeBackup V2 is deleted", func() {
        // ... delete test
    })
})
```

---

### Phase 8: Cleanup and Validation

#### Step 8.1: Remove Unit Tests That Don't Test Business Logic

Per requirements, remove unit tests that:
- Only test mocks/fakes
- Test simple getters/setters
- Test action composition/orchestration
- Don't test concrete business behavior

Keep unit tests that:
- Test validation logic
- Test data conversion functions
- Test label parsing/generation logic (e.g., `ConvertToAccessibleFromKey`, `IsAccessibleFromKey`)

#### Step 8.2: Run Validation

```bash
# Generate manifests
make manifests

# Run tests
make test

# Verify build
make build
```

---

## Summary: File Changes

### New Files

| Path | Purpose |
|------|---------|
| `pkg/feature/ffGcpBackupV2.go` | Feature flag definition |
| `pkg/kcp/provider/gcp/nfsbackup/client/v1/fileBackupClient.go` | v1 client (moved) |
| `pkg/kcp/provider/gcp/nfsbackup/client/v2/fileBackupClient.go` | v2 client (new library) |
| `pkg/kcp/provider/gcp/nfsbackup/client/v2/util.go` | v2 client utilities |
| `pkg/skr/gcpnfsvolumebackup/v1/reconciler.go` | v1 reconciler (moved from original location) |
| `pkg/skr/gcpnfsvolumebackup/v1/state.go` | v1 state |
| `pkg/skr/gcpnfsvolumebackup/v1/*.go` | v1 complete action set |
| `pkg/skr/gcpnfsvolumebackup/v2/reconcile.go` | v2 reconciler with Reconciler struct |
| `pkg/skr/gcpnfsvolumebackup/v2/state.go` | v2 state with protobuf types |
| `pkg/skr/gcpnfsvolumebackup/v2/*.go` | v2 complete action set |
| `pkg/kcp/provider/gcp/mock/nfsBackupStoreV2.go` | v2 mock implementation |
| `internal/controller/cloud-resources/gcpnfsvolumebackup_v2_test.go` | v2 controller tests |

### Modified Files

| Path | Changes |
|------|---------|
| `pkg/feature/ff_ga.yaml` | Add gcpBackupV2 flag |
| `pkg/feature/ff_edge.yaml` | Add gcpBackupV2 flag |
| `config/featureToggles/featureToggles.local.yaml` | Add gcpBackupV2 flag |
| `internal/controller/cloud-resources/gcpnfsvolumebackup_controller.go` | Check feature flag at startup, route to v1 or v2 |
| `cmd/main.go` | Pass both v1 and v2 client providers |
| `pkg/kcp/provider/gcp/mock/server.go` | Add v2 provider |
| `pkg/kcp/provider/gcp/mock/type.go` | Add v2 interface |
| `internal/controller/cloud-resources/suite_test.go` | Initialize v2 provider |

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking existing functionality | Feature flag defaults to v1 |
| Code duplication between v1/v2 | Acceptable trade-off for simplicity and independence |
| Missing test coverage | Mandatory controller tests for v2 |
| GCP API behavior differences | Test against mock, document differences |

---

## Success Criteria

1. ✅ v1 works identically to current implementation
2. ✅ v2 uses `cloud.google.com/go/filestore/apiv1` library
3. ✅ Feature flag `gcpBackupV2` controls version selection
4. ✅ Default is v1 (flag disabled)
5. ✅ Controller tests pass for both v1 and v2
6. ✅ No unit tests unless testing concrete business logic
7. ✅ `make test` passes
8. ✅ `make manifests` generates correctly

---

## Implementation Order

1. **Phase 1**: Create feature flag (foundation) ✅ **COMPLETED**
2. **Phase 2**: Create v1 directory structure (safe move) ✅ **COMPLETED**
3. **Phase 3**: Create v2 client (new library integration) ✅ **COMPLETED**
4. **Phase 4**: Create v2 SKR implementation (new reconciler) ✅ **COMPLETED**
5. **Phase 5**: Update controller with startup flag routing ✅ **COMPLETED**
6. **Phase 6**: Create v2 mock (test infrastructure) ✅ **COMPLETED**
7. **Phase 7**: Update controller and tests (validation)
8. **Phase 8**: Cleanup and validation (quality)

Each phase should be completed and tested before proceeding to the next.

---

## Implementation Guidelines

**IMPORTANT**: During implementation:
- **DO NOT** create additional `.md` documentation files
- If any context, notes, or implementation details need to be recorded, add them to **this file** (`GCP_BACKUP_REFACTOR.md`)
- Keep all implementation progress, decisions, and changes documented in the existing sections or add new sections to this file as needed
