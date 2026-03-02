# GCP NFS Volume Restore - Refactor Plan (v1 → v2)

**Date**: 2026-03-02  
**Scope**: `pkg/skr/gcpnfsvolumerestore`, `pkg/kcp/provider/gcp/nfsrestore/client`, controller, mock, feature flag, tests

## Overview

Refactor `GcpNfsVolumeRestore` from the old `google.golang.org/api/file/v1` library to the modern `cloud.google.com/go/filestore/apiv1` library, following the same pattern established by the `GcpNfsVolumeBackup` v1→v2 refactor.

This is an **SKR-only component** (no backing KCP resource). It directly calls GCP Filestore APIs from the SKR reconciler using clients located in `pkg/kcp/provider/gcp/nfsrestore/client`.

### Key Differences from Backup Refactor

- Restore is a **one-shot job** (status transitions: `""` → `Processing` → `InProgress` → `Done`/`Failed`) — it does not have a lifecycle like backup (create/update/delete). No deletion of cloud resources is needed.
- Restore uses **two GCP clients**: `FileRestoreClient` (for restore operations) and `FileBackupClient` (for loading backup details / permission checks on `backupUrl` source).
- Restore uses a **lease mechanism** to prevent concurrent restores to the same NFS volume.
- The v2 backup client (`pkg/kcp/provider/gcp/nfsbackup/client/v2`) already exists and can be reused for backup loading.

## Progress Checklist

- [x] Step 1: Create Feature Flag `gcpNfsRestoreV2`
- [x] Step 2: Create v2 Restore Client
- [x] Step 3: Move Current Restore Implementation to v1
- [ ] Step 4: Create v2 Restore Implementation
- [ ] Step 5: Create v2 Restore Mock
- [ ] Step 6: Update Controller
- [ ] Step 7: Update `cmd/main.go`
- [ ] Step 8: Update Test Suite Setup
- [ ] Step 9: Update Existing Controller Tests
- [ ] Step 10: Create v2 Controller Tests
- [ ] Step 11: Update Import References Across Codebase
- [ ] Step 12: Verify & Clean Up

---

## Step 1: Create Feature Flag `gcpNfsRestoreV2`

**Files to create:**
- `pkg/feature/ffGcpNfsRestoreV2.go`

**Files to modify:**
- `pkg/feature/ff_edge.yaml` — add `gcpNfsRestoreV2` entry (default: `disabled`)
- `pkg/feature/ff_ga.yaml` — add `gcpNfsRestoreV2` entry (default: `disabled`)

**Pattern** (copy from `ffGcpBackupV2.go`):
```go
package feature

import "context"

const GcpNfsRestoreV2FlagName = "gcpNfsRestoreV2"

var GcpNfsRestoreV2 = &gcpNfsRestoreV2Info{}

type gcpNfsRestoreV2Info struct{}

func (k *gcpNfsRestoreV2Info) Value(ctx context.Context) bool {
    return provider.BoolVariation(ctx, GcpNfsRestoreV2FlagName, false)
}
```

YAML entries (both files):
```yaml
gcpNfsRestoreV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled
```

---

## Step 2: Create v2 Restore Client (`pkg/kcp/provider/gcp/nfsrestore/client/v2`)

**Files to create:**
- `pkg/kcp/provider/gcp/nfsrestore/client/v2/fileRestoreClient.go`
- `pkg/kcp/provider/gcp/nfsrestore/client/v2/util.go`

**What changes from v1:**

| Aspect | v1 (current) | v2 (new) |
|--------|-------------|----------|
| Library | `google.golang.org/api/file/v1` | `cloud.google.com/go/filestore/apiv1` |
| Client init | `ClientProvider[T]` (old pattern: `func(ctx, credFile) (T, error)`) | `GcpClientProvider[T]` (new pattern: `func() T`) |
| Operation type | `*file.Operation` | `*longrunningpb.Operation` |
| Restore request | `file.RestoreInstanceRequest` | `filestorepb.RestoreInstanceRequest` |
| Error handling | `*googleapi.Error` with `.Code` | `gcpmeta.IsNotFound()` / `status.FromProto()` |
| Client backing | Creates own `file.Service` via HTTP | Reuses `GcpClients.Filestore` (`*filestore.CloudFilestoreManagerClient`) |

**v2 `FileRestoreClient` interface:**
```go
type FileRestoreClient interface {
    // RestoreFile triggers a restore from a backup to a Filestore instance.
    // Returns the operation name for tracking.
    RestoreFile(ctx context.Context, projectId, destFileFullPath, destFileShareName, srcBackupFullPath string) (string, error)

    // GetRestoreOperation retrieves status of a long-running restore operation.
    GetRestoreOperation(ctx context.Context, operationName string) (*longrunningpb.Operation, error)

    // FindRestoreOperation lists operations for an instance to find a running restore.
    FindRestoreOperation(ctx context.Context, projectId, location, instanceId string) (*longrunningpb.Operation, error)
}
```

**Key implementation notes:**
- `RestoreFile` uses `GcpClients.Filestore.RestoreInstance()` which returns a `RestoreInstanceOperation`. Call `.Name()` on it to get the operation name.
- `GetRestoreOperation` uses `GcpClients.Filestore.LROClient.GetOperation()` (same pattern as v2 backup's `GetBackupLROperation`).
- `FindRestoreOperation` uses `GcpClients.Filestore.LROClient.ListOperations()` with metadata filter.
- Return `(string, error)` for operations (operation name, not the full operation object) — consistent with v2 backup pattern.
- Exception: `GetRestoreOperation` returns `(*longrunningpb.Operation, error)` to check `.Done` and `.GetError()`.
- Use `NewFileRestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FileRestoreClient]`.

**`util.go`** — helper functions for path construction:
- `GetFilestoreInstancePath(projectId, location, name)` — can reuse from `nfsbackup/client/v2/util.go` or define locally.
- `GetFilestoreParentPath(projectId, location)`.

**Rename current client dir:**
- Move `pkg/kcp/provider/gcp/nfsrestore/client/fileRestoreClient.go` → `pkg/kcp/provider/gcp/nfsrestore/client/v1/fileRestoreClient.go`
- Update the `package` declaration to `package v1`
- Update all v1 imports referencing the old path

---

## Step 3: Move Current Restore Implementation to v1

**Action:** Restructure `pkg/skr/gcpnfsvolumerestore/` to use `v1/` and `v2/` sub-packages.

**Files to move** (all files from `pkg/skr/gcpnfsvolumerestore/` → `pkg/skr/gcpnfsvolumerestore/v1/`):

| Current file | Destination |
|-------------|-------------|
| `reconciler.go` | `v1/reconciler.go` |
| `state.go` | `v1/state.go` |
| `util.go` | `v1/util.go` |
| `loadScope.go` | `v1/loadScope.go` |
| `loadGcpNfsVolume.go` | `v1/loadGcpNfsVolume.go` |
| `loadBackup.go` | `v1/loadBackup.go` |
| `populateBackupUrl.go` | `v1/populateBackupUrl.go` |
| `runNfsRestore.go` | `v1/runNfsRestore.go` |
| `checkRestoreOperation.go` | `v1/checkRestoreOperation.go` |
| `checkRestorePermissions.go` | `v1/checkRestorePermissions.go` |
| `findRestoreOperation.go` | `v1/findRestoreOperation.go` |
| `addFinalizer.go` | `v1/addFinalizer.go` |
| `removeFinalizer.go` | `v1/removeFinalizer.go` |
| `acquireLease.go` | `v1/acquireLease.go` |
| `releaseLease.go` | `v1/releaseLease.go` |
| `setProcessing.go` | `v1/setProcessing.go` |

**Changes:**
- Update `package` declaration from `gcpnfsvolumerestore` to `v1`
- Add `Ignore` variable (ignorant pattern): `var Ignore = ignorant.New()`
- Add `Ignore.ShouldIgnoreKey(req)` check at the start of `Run()` (same pattern as backup v2)
- Update import of old nfsrestore client: `gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"` → `gcpnfsrestoreclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v1"`
- Keep using old `ClientProvider[T]` pattern (v1 remains unchanged functionally)

**Test files:** Move existing unit tests to `v1/` as well. They are mostly trivial but keeping them won't hurt.

---

## Step 4: Create v2 Restore Implementation (`pkg/skr/gcpnfsvolumerestore/v2`)

**Files to create:**

| File | Purpose |
|------|---------|
| `v2/reconcile.go` | Reconciler + action composition (create/delete flow) |
| `v2/state.go` | State struct using `filestorepb.Backup` and v2 clients |
| `v2/ignorant.go` | `var Ignore = ignorant.New()` |
| `v2/shortCircuitCompleted.go` | Early exit when Done/Failed and not deleting |
| `v2/loadScope.go` | Load KCP Scope (same logic, new package) |
| `v2/loadGcpNfsVolume.go` | Load destination GcpNfsVolume (same logic) |
| `v2/populateBackupUrl.go` | Load backup URL from source ref or backupUrl spec |
| `v2/loadBackup.go` | Load GCP backup for permission checks (uses v2 backup client) |
| `v2/checkRestorePermissions.go` | Permission check on backupUrl source |
| `v2/acquireLease.go` | Acquire NFS volume lease |
| `v2/releaseLease.go` | Release NFS volume lease |
| `v2/findRestoreOperation.go` | Find existing running restore operation |
| `v2/runNfsRestore.go` | Submit GCP restore request |
| `v2/checkRestoreOperation.go` | Check LRO status |
| `v2/setProcessing.go` | Set initial Processing state |
| `v2/util.go` | Utility functions |

### 4a. State (`v2/state.go`)

```go
type State struct {
    composed.State
    KymaRef    klog.ObjectRef
    KcpCluster composed.StateCluster
    SkrCluster composed.StateCluster

    Scope             *cloudcontrolv1beta1.Scope
    GcpNfsVolume      *cloudresourcesv1beta1.GcpNfsVolume
    SrcBackupFullPath string

    fileBackup        *filestorepb.Backup  // modern protobuf type (for permission checks)

    fileRestoreClient v2client.FileRestoreClient
    fileBackupClient  v2backupclient.FileBackupClient  // reuse nfsbackup v2 client
}
```

**Key differences from v1 State:**
- `fileBackup` type changes from `*file.Backup` to `*filestorepb.Backup`
- `fileRestoreClient` type changes to v2 FileRestoreClient
- `fileBackupClient` type changes from v1 to v2 (`gcpnfsbackupclientv2.FileBackupClient`)
- `StateFactory.NewState()` uses `GcpClientProvider[T]` (simple `func() T` call, no `ctx`/`credFile` args)
- No `abstractions.Environment` needed (that was a v1 legacy pattern)

### 4b. Reconciler (`v2/reconcile.go`)

```go
func NewReconciler(kymaRef klog.ObjectRef, kcpCluster cluster.Cluster, skrCluster cluster.Cluster,
    fileRestoreClientProvider gcpclient.GcpClientProvider[v2client.FileRestoreClient],
    fileBackupClientProvider gcpclient.GcpClientProvider[v2backupclient.FileBackupClient],
) Reconciler
```

Action composition — **must use create/delete flow** regulated by `composed.MarkedForDeletionPredicate`, following the v2 backup pattern:
```go
func composeActions() composed.Action {
    return composed.ComposeActions(
        "gcpNfsVolumeRestoreV2",
        loadScope,
        shortCircuitCompleted,     // stop if already Done/Failed and not deleting
        actions.AddCommonFinalizer(),

        loadGcpNfsVolume,
        populateBackupUrl,
        loadBackup,

        checkRestoreOperation,     // check running LRO before branching

        composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "gcpNfsVolumeRestoreV2-create",
                setProcessing,
                checkRestorePermissions,
                acquireLease,
                findRestoreOperation,
                runNfsRestore,
                // checkRestoreOperation already ran above;
                // after runNfsRestore submits the op, requeue will re-enter
                // and checkRestoreOperation will track it to completion.
                releaseLease,
            ),
            composed.ComposeActions(
                "gcpNfsVolumeRestoreV2-delete",
                releaseLease,
                actions.RemoveCommonFinalizer(),
            ),
        ),
        composed.StopAndForgetAction,
    )
}
```

**Key design points (following backup v2 pattern):**
- `shortCircuitCompleted` — early exit when restore is `Done`/`Failed` and object is NOT marked for deletion. When it IS marked for deletion, let it fall through to the delete branch.
- `checkRestoreOperation` runs **before** the `IfElse` branch so it handles LRO polling regardless of whether deletion is pending (same pattern as backup v2's `checkBackupOperation`).
- The **create branch** contains the restore-specific flow (processing → permissions → lease → find/run op).
- The **delete branch** releases the lease (if held) and removes the finalizer. Restore doesn't create a cloud resource that needs explicit deletion — the only cleanup is the lease.
- `actions.AddCommonFinalizer()` / `actions.RemoveCommonFinalizer()` replace the old inline implementations.
- Individual actions **no longer need** `if composed.MarkedForDeletionPredicate(ctx, st) { return nil, nil }` guards — the `IfElse` branching handles this at the composition level.

### 4c. Key Action Changes (v2 vs v1)

**`checkRestoreOperation.go`:**
- Operation type: `*longrunningpb.Operation` instead of `*file.Operation`
- Check `op.Done` (same field name)
- Error check: `op.GetError()` → `status.FromProto(op.GetError())` instead of `*googleapi.Error`
- Use `gcpmeta.IsNotFound(err)` instead of `errors.As(err, &googleapi.Error{})` for 404 checks
- Runs **before** the `IfElse` branch (same position as `checkBackupOperation` in backup v2)

**`shortCircuitCompleted.go`** (new, replaces old `setProcessing` short-circuit):
- If restore state is `Done` or `Failed` **and** object is not marked for deletion → `composed.StopAndForget`
- If object IS marked for deletion → continue (let delete branch handle cleanup)
- Pattern matches backup v2's `shortCircuitCompleted`

**`findRestoreOperation.go`:**
- Returns `*longrunningpb.Operation` instead of `*file.Operation`
- Use `gcpmeta.IsNotFound(err)` instead of `meta.IsNotFound(err)` (check which meta package is correct)

**`runNfsRestore.go`:**
- `state.fileRestoreClient.RestoreFile()` returns `(string, error)` instead of `(*file.Operation, error)`
- Store operation name directly: `restore.Status.OpIdentifier = opName`

**`loadBackup.go`:**
- Uses v2 backup client: `state.fileBackupClient.GetBackup()` returns `*filestorepb.Backup`
- `state.fileBackup` is `*filestorepb.Backup`

**`checkRestorePermissions.go` / `IsAllowedToRestoreBackup()`:**
- Same logic, but `state.fileBackup.Labels` is `map[string]string` on `*filestorepb.Backup` (same as before — labels field is the same concept)

**`populateBackupUrl.go`:**
- Uses v2 backup path utilities from `nfsbackup/client/v2/util.go` instead of `gcpclient.GetFileBackupPath`

**`setProcessing.go`:**
- Only runs inside the **create branch** — no need for `MarkedForDeletionPredicate` guard
- Simplified: just sets `Processing` state if state is empty, or continues if already in progress

**Finalizer handling:**
- `actions.AddCommonFinalizer()` called **before** the `IfElse` branch (runs on both create and delete paths, is a no-op if already present)
- `actions.RemoveCommonFinalizer()` called as the last step in the **delete branch** only

**`loadGcpNfsVolume.go`, `populateBackupUrl.go`, `loadBackup.go`, `checkRestorePermissions.go`, `acquireLease.go`:**
- No longer need `if composed.MarkedForDeletionPredicate(ctx, st) { return nil, nil }` guards — the `IfElse` composition handles branching at a higher level

---

## Step 5: Create v2 Restore Mock (`pkg/kcp/provider/gcp/mock`)

**Files to create:**
- `pkg/kcp/provider/gcp/mock/nfsRestoreStoreV2.go`

**Files to modify:**
- `pkg/kcp/provider/gcp/mock/type.go` — add v2 restore provider and utils interfaces
- `pkg/kcp/provider/gcp/mock/server.go` — add v2 restore store instance and provider creation

**Mock implementation:** `nfsRestoreStoreV2` implementing `v2client.FileRestoreClient`:
```go
type nfsRestoreStoreV2 struct {
    restoreFileError         error
    restoreOperationError    error
    operations               []*longrunningpb.Operation
}
```

Methods:
- `RestoreFile()` → returns `("mock-op-name", nil)` wrapped as `longrunningpb.Operation`
- `GetRestoreOperation()` → returns done operation
- `FindRestoreOperation()` → returns running/nil operation

**Additions to `type.go`:**
```go
// In Providers interface:
FileRestoreClientProviderV2() client.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient]

// New utility interface:
type FileRestoreClientFakeUtilsV2 interface {
    // Add methods as needed for test control
}
```

**Additions to `Server` interface:**
```go
FileRestoreClientFakeUtilsV2
```

---

## Step 6: Update Controller (`internal/controller/cloud-resources`)

**Files to modify:**
- `internal/controller/cloud-resources/gcpnfsvolumerestore_controller.go`

**Reference:** `gcpnfsvolumebackup_controller.go` — follow this pattern exactly.

**Current code** stores a concrete `gcpnfsvolumerestore.Reconciler` struct field. The refactored version must:
1. Introduce a private `gcpNfsVolumeRestoreRunner` interface (same as backup's `gcpNfsVolumeBackupRunner`)
2. Change `GcpNfsVolumeRestoreReconciler` to hold the interface instead of a concrete type
3. Add both v1 and v2 providers to the factory
4. Use feature flag in `New()` to select reconciler
5. Remove the `logr.Logger` param from `SetupGcpNfsVolumeRestoreReconciler` (not used in backup pattern)

```go
package cloudresources

import (
    "context"

    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/common/abstractions"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
    gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
    gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
    gcpnfsrestoreclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v1"
    gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
    gcpnfsvolumerestorev1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v1"
    gcpnfsvolumerestorev2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v2"
    skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
    reconcile2 "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    ctrl "sigs.k8s.io/controller-runtime"
)

// gcpNfsVolumeRestoreRunner is a common interface for v1 and v2 reconcilers
type gcpNfsVolumeRestoreRunner interface {
    Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}

// GcpNfsVolumeRestoreReconciler reconciles a GcpNfsVolumeRestore object
type GcpNfsVolumeRestoreReconciler struct {
    reconciler gcpNfsVolumeRestoreRunner
}

func (r *GcpNfsVolumeRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    return r.reconciler.Run(ctx, req)
}

type GcpNfsVolumeRestoreReconcilerFactory struct {
    // v1 providers (old pattern: func(ctx, credFile) (T, error))
    fileRestoreClientProviderV1 gcpclient.ClientProvider[gcpnfsrestoreclientv1.FileRestoreClient]
    fileBackupClientProviderV1  gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]
    env                         abstractions.Environment
    // v2 providers (new pattern: func() T)
    fileRestoreClientProviderV2 gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient]
    fileBackupClientProviderV2  gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]
}

func (f *GcpNfsVolumeRestoreReconcilerFactory) New(args reconcile2.ReconcilerArguments) reconcile.Reconciler {
    // Check feature flag at reconciler creation time (after feature.Initialize has run)
    if feature.GcpNfsRestoreV2.Value(context.Background()) {
        reconciler := gcpnfsvolumerestorev2.NewReconciler(
            args.KymaRef,
            args.KcpCluster,
            args.SkrCluster,
            f.fileRestoreClientProviderV2,
            f.fileBackupClientProviderV2,
        )
        return &GcpNfsVolumeRestoreReconciler{reconciler: &reconciler}
    }

    reconciler := gcpnfsvolumerestorev1.NewReconciler(
        args.KymaRef,
        args.KcpCluster,
        args.SkrCluster,
        f.fileRestoreClientProviderV1,
        f.fileBackupClientProviderV1,
        f.env,
    )
    return &GcpNfsVolumeRestoreReconciler{reconciler: &reconciler}
}

func SetupGcpNfsVolumeRestoreReconciler(
    reg skrruntime.SkrRegistry,
    fileRestoreClientProviderV1 gcpclient.ClientProvider[gcpnfsrestoreclientv1.FileRestoreClient],
    fileBackupClientProviderV1 gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient],
    fileRestoreClientProviderV2 gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient],
    fileBackupClientProviderV2 gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient],
    env abstractions.Environment,
) error {
    return reg.Register().
        WithFactory(&GcpNfsVolumeRestoreReconcilerFactory{
            fileRestoreClientProviderV1: fileRestoreClientProviderV1,
            fileBackupClientProviderV1:  fileBackupClientProviderV1,
            fileRestoreClientProviderV2: fileRestoreClientProviderV2,
            fileBackupClientProviderV2:  fileBackupClientProviderV2,
            env:                         env,
        }).
        For(&cloudresourcesv1beta1.GcpNfsVolumeRestore{}).
        Complete()
}
```

**Key differences from current restore controller:**
- `Reconciler` field becomes private `reconciler` with interface type (not exported concrete struct)
- `logr.Logger` parameter removed from `SetupGcpNfsVolumeRestoreReconciler` (unused in backup pattern)
- Factory holds both v1 and v2 providers for restore client AND backup client
- `New()` checks `feature.GcpNfsRestoreV2.Value(context.Background())` to branch

---

## Step 7: Update `cmd/main.go`

**Changes:**
- Import v2 restore client package
- Pass both v1 and v2 providers to `SetupGcpNfsVolumeRestoreReconciler`:
  - v1: `gcpnfsrestoreclientv1.NewFileRestoreClientProvider()` (existing)
  - v2: `gcpnfsrestoreclientv2.NewFileRestoreClientProvider(gcpClients)` (new)
  - v1 backup: `gcpnfsbackupclientv1.NewFileBackupClientProvider()` (existing)
  - v2 backup: `gcpnfsbackupclientv2.NewFileBackupClientProvider(gcpClients)` (existing)

---

## Step 8: Update Test Suite Setup (`internal/controller/cloud-resources/suite_test.go`)

**Changes:**
- Update `SetupGcpNfsVolumeRestoreReconciler` call to pass v2 providers from mock:
  ```go
  Expect(SetupGcpNfsVolumeRestoreReconciler(
      infra.Registry(),
      infra.GcpMock().FilerestoreClientProvider(),     // v1
      infra.GcpMock().FileBackupClientProvider(),       // v1
      infra.GcpMock().FileRestoreClientProviderV2(),    // v2 (new)
      infra.GcpMock().FileBackupClientProviderV2(),     // v2 (already exists)
      env,
      testSetupLog,
  )).NotTo(HaveOccurred())
  ```

---

## Step 9: Update Existing Controller Tests

**Files to modify:**
- `internal/controller/cloud-resources/gcpnfsvolumerestore_test.go`

**Changes:**
- Add feature flag skip guard (like backup v1 tests):
  ```go
  BeforeEach(func() {
      if feature.GcpNfsRestoreV2.Value(context.Background()) {
          Skip("Skipping v1 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is enabled")
      }
  })
  ```
- Add `Ignore` imports for both v1 and v2:
  ```go
  skrgcpnfsvolrestorev1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v1"
  skrgcpnfsvolrestorev2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumerestore/v2"
  ```

---

## Step 10: Create v2 Controller Tests

**Files to create:**
- `internal/controller/cloud-resources/gcpnfsvolumerestore_v2_test.go`

**Test scenarios** (mandatory):
1. **Create**: GcpNfsVolumeRestore created with backup ref → reaches Done state with Ready condition
2. **Create with backupUrl**: GcpNfsVolumeRestore created with backupUrl → reaches Done state
3. **Delete**: GcpNfsVolumeRestore in Done state → deleted successfully (finalizer removed)

**Pattern** (follow backup v2 test pattern):
```go
var _ = Describe("Feature: SKR GcpNfsVolumeRestore V2", func() {

    It("Scenario: SKR GcpNfsVolumeRestore V2 is created and completed", func() {
        if !feature.GcpNfsRestoreV2.Value(context.Background()) {
            Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
        }
        // Given KCP Scope exists
        // And Given SKR GcpNfsVolume exists in Ready state
        // And Given SKR GcpNfsVolumeBackup exists in Ready state
        // When GcpNfsVolumeRestore is created
        // Then GcpNfsVolumeRestore reaches Done state with Ready condition
    })

    It("Scenario: SKR GcpNfsVolumeRestore V2 is deleted", func() {
        if !feature.GcpNfsRestoreV2.Value(context.Background()) {
            Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
        }
        // Given GcpNfsVolumeRestore in Done state
        // When Delete is called
        // Then GcpNfsVolumeRestore is removed
    })
})
```

---

## Step 11: Update Import References Across Codebase

All files that currently import `pkg/skr/gcpnfsvolumerestore` need to be updated to import the v1 sub-package. Similarly, files importing `pkg/kcp/provider/gcp/nfsrestore/client` need to be updated to import v1.

**Key files to check:**
- `cmd/main.go`
- `internal/controller/cloud-resources/gcpnfsvolumerestore_controller.go`
- `internal/controller/cloud-resources/suite_test.go`
- `internal/controller/cloud-resources/gcpnfsvolumerestore_test.go`
- `pkg/kcp/provider/gcp/mock/server.go`
- `pkg/kcp/provider/gcp/mock/type.go`
- `pkg/kcp/provider/gcp/mock/nfsRestoreStore.go`

---

## Step 12: Verify & Clean Up

1. **Build**: `make build` — verify all imports resolve
2. **Generate**: `make manifests` — verify CRD generation (no API changes expected)
3. **Test**: `make test` — run full test suite
4. **Verify feature flag behavior**:
   - Default (disabled): v1 reconciler instantiated, v1 tests pass, v2 tests skipped
   - Enabled: v2 reconciler instantiated, v2 tests pass, v1 tests skipped

---

## Summary of All Files

### New Files
| File | Purpose |
|------|---------|
| `pkg/feature/ffGcpNfsRestoreV2.go` | Feature flag definition |
| `pkg/kcp/provider/gcp/nfsrestore/client/v2/fileRestoreClient.go` | v2 restore client (cloud.google.com) |
| `pkg/kcp/provider/gcp/nfsrestore/client/v2/util.go` | Path helper functions |
| `pkg/skr/gcpnfsvolumerestore/v2/reconcile.go` | v2 reconciler (create/delete flow) |
| `pkg/skr/gcpnfsvolumerestore/v2/state.go` | v2 state |
| `pkg/skr/gcpnfsvolumerestore/v2/ignorant.go` | Ignore pattern for tests |
| `pkg/skr/gcpnfsvolumerestore/v2/shortCircuitCompleted.go` | Short-circuit Done/Failed |
| `pkg/skr/gcpnfsvolumerestore/v2/loadScope.go` | Load KCP Scope |
| `pkg/skr/gcpnfsvolumerestore/v2/loadGcpNfsVolume.go` | Load destination NFS volume |
| `pkg/skr/gcpnfsvolumerestore/v2/populateBackupUrl.go` | Populate backup source path |
| `pkg/skr/gcpnfsvolumerestore/v2/loadBackup.go` | Load backup for permission check |
| `pkg/skr/gcpnfsvolumerestore/v2/checkRestorePermissions.go` | Permission validation |
| `pkg/skr/gcpnfsvolumerestore/v2/acquireLease.go` | Lease acquisition |
| `pkg/skr/gcpnfsvolumerestore/v2/releaseLease.go` | Lease release |
| `pkg/skr/gcpnfsvolumerestore/v2/findRestoreOperation.go` | Find running restore op |
| `pkg/skr/gcpnfsvolumerestore/v2/runNfsRestore.go` | Submit restore request |
| `pkg/skr/gcpnfsvolumerestore/v2/checkRestoreOperation.go` | Check LRO status |
| `pkg/skr/gcpnfsvolumerestore/v2/setProcessing.go` | Initial state transition |
| `pkg/skr/gcpnfsvolumerestore/v2/util.go` | Utility functions |
| `pkg/kcp/provider/gcp/mock/nfsRestoreStoreV2.go` | v2 mock restore store |
| `internal/controller/cloud-resources/gcpnfsvolumerestore_v2_test.go` | v2 controller tests |

### Moved Files (v1)
| From | To |
|------|------|
| `pkg/skr/gcpnfsvolumerestore/*.go` | `pkg/skr/gcpnfsvolumerestore/v1/*.go` |
| `pkg/kcp/provider/gcp/nfsrestore/client/fileRestoreClient.go` | `pkg/kcp/provider/gcp/nfsrestore/client/v1/fileRestoreClient.go` |

### Modified Files
| File | Change |
|------|--------|
| `pkg/feature/ff_edge.yaml` | Add `gcpNfsRestoreV2` entry |
| `pkg/feature/ff_ga.yaml` | Add `gcpNfsRestoreV2` entry |
| `internal/controller/cloud-resources/gcpnfsvolumerestore_controller.go` | Feature flag switch v1/v2 |
| `internal/controller/cloud-resources/suite_test.go` | Pass v2 providers |
| `internal/controller/cloud-resources/gcpnfsvolumerestore_test.go` | Add v1 skip guard, update imports |
| `cmd/main.go` | Pass v2 providers |
| `pkg/kcp/provider/gcp/mock/type.go` | Add v2 restore interfaces |
| `pkg/kcp/provider/gcp/mock/server.go` | Add v2 restore provider |
| `pkg/kcp/provider/gcp/mock/nfsRestoreStore.go` | Move to acknowledge as v1 (package stays `mock`) |

---

## Execution Order

The recommended implementation order minimizes broken intermediate states:

1. **Step 1** — Feature flag (no functional change)
2. **Step 2** — v2 restore client (no consumers yet)
3. **Step 3** — Move current code to v1 sub-packages + update all imports (functionally equivalent)
4. **Step 5** — v2 mock (needed before v2 impl can be tested)
5. **Step 4** — v2 restore implementation
6. **Step 6** — Controller feature flag switch
7. **Step 7** — main.go wiring
8. **Step 8** — Test suite setup
9. **Step 9** — Update existing v1 tests with skip guard
10. **Step 10** — v2 controller tests
11. **Step 11** — Verify all import references
12. **Step 12** — Full build + test verification
