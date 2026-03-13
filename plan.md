# Migration Plan: GcpNfsVolumeBackupDiscovery v1 → v2 Client

## Overview

Migrate `pkg/skr/gcpnfsvolumebackupdiscovery` from the legacy v1 backup client
(`pkg/kcp/provider/gcp/nfsbackup/client/v1`) to the modern v2 backup client
(`pkg/kcp/provider/gcp/nfsbackup/client/v2`).

### Why

- v1 uses the deprecated REST-based `google.golang.org/api/file/v1` library with `*file.Backup` types
- v2 uses the modern gRPC-based `cloud.google.com/go/filestore/apiv1` library with `*filestorepb.Backup` protobuf types
- v2 follows the NEW GcpClients singleton pattern; v1 uses the OLD cached ClientProvider pattern
- Other SKR reconcilers (GcpNfsVolumeBackup, GcpNfsVolumeRestore) have already been migrated to v2

---

## Key Differences Between v1 and v2

| Aspect | v1 (current) | v2 (target) |
|--------|-------------|-------------|
| Package | `pkg/kcp/provider/gcp/nfsbackup/client/v1` | `pkg/kcp/provider/gcp/nfsbackup/client/v2` |
| Import alias | `gcpnfsbackupclientv1` | `gcpnfsbackupclientv2` |
| Backup type | `*file.Backup` (`google.golang.org/api/file/v1`) | `*filestorepb.Backup` (`cloud.google.com/go/filestore/apiv1/filestorepb`) |
| Client interface | `v1.FileBackupClient` | `v2.FileBackupClient` |
| List method | `ListFilesBackups(ctx, projectId, filter)` → `[]*file.Backup` | `ListBackups(ctx, projectId, filter)` → `[]*filestorepb.Backup` |
| Provider type | `gcpclient.ClientProvider[T]` = `func(ctx, credentialsFile) (T, error)` | `gcpclient.GcpClientProvider[T]` = `func() T` |
| Client creation | Needs `ctx` + `credentialsFile` at call time | Pre-initialized via `GcpClients` singleton; no args needed |
| `Backup.Name` | `string` (same) | `string` (same) |
| `Backup.Labels` | `map[string]string` (same) | `map[string]string` (same) |
| `Backup.CreateTime` | `string` (RFC3339 format) | `*timestamppb.Timestamp` (protobuf timestamp) |

---

## Files to Modify

### 1. `pkg/skr/gcpnfsvolumebackupdiscovery/state.go`

**Changes:**

- **Import**: Replace `gcpnfsbackupclientv1 ".../nfsbackup/client/v1"` with `gcpnfsbackupclientv2 ".../nfsbackup/client/v2"`
- **Import**: Replace `"google.golang.org/api/file/v1"` with `"cloud.google.com/go/filestore/apiv1/filestorepb"`
- **Import**: Remove `"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"` (no longer needed for credentials)
- **Import**: Remove `"github.com/kyma-project/cloud-manager/pkg/common/abstractions"` (env no longer needed)
- **State struct**:
  - Change `fileBackupClient gcpnfsbackupclientv1.FileBackupClient` → `fileBackupClient gcpnfsbackupclientv2.FileBackupClient`
  - Change `backups []*file.Backup` → `backups []*filestorepb.Backup`
- **NewStateFactory signature**:
  - Change `fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]` → `fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]`
  - Remove `env abstractions.Environment` parameter (unused)
- **stateFactory struct**:
  - Change `fileBackupClientProvider` field type to `gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]`
  - Remove `env abstractions.Environment` field
- **stateFactory.NewState method**:
  - Replace `f.fileBackupClientProvider(ctx, config.GcpConfig.CredentialsFile)` error-returning call with simple `f.fileBackupClientProvider()` (no error possible)
  - This simplifies the method—no error handling needed for client creation

### 2. `pkg/skr/gcpnfsvolumebackupdiscovery/loadAvailableBackups.go`

**Changes:**

- **Import**: Remove `gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"` if only used for `gcpmeta` (check)
  - Actually `gcpclient` is still used for `gcpclient.GetSharedBackupsFilter()` — keep it
- **Method call**: Change `state.fileBackupClient.ListFilesBackups(...)` → `state.fileBackupClient.ListBackups(...)`
- No other changes needed — the error handling and filter logic remain the same

### 3. `pkg/skr/gcpnfsvolumebackupdiscovery/updateStatus.go`

**Changes:**

- **Import**: Add `"google.golang.org/protobuf/types/known/timestamppb"` or handle `*timestamppb.Timestamp` type
- **Import**: Remove `"time"` if only used for `time.Parse` of CreateTime (replaced by `.AsTime()`)
- **CreateTime handling**: Replace string parsing:
  ```go
  // v1 (current):
  if b.CreateTime != "" {
      if creationTime, err := time.Parse(time.RFC3339, b.CreateTime); err == nil {
          availableBackup.CreationTime = ptr.To(metav1.NewTime(creationTime))
      }
  }

  // v2 (target):
  if b.CreateTime != nil {
      availableBackup.CreationTime = ptr.To(metav1.NewTime(b.CreateTime.AsTime()))
  }
  ```
- **Labels and Name**: No changes — both v1 and v2 use `string` for `Name` and `map[string]string` for `Labels`

### 4. `pkg/skr/gcpnfsvolumebackupdiscovery/reconciler.go`

**Changes:**

- **Import**: Replace `gcpnfsbackupclientv1` with `gcpnfsbackupclientv2`
- **Import**: Remove `"github.com/kyma-project/cloud-manager/pkg/common/abstractions"` (env removed)
- **NewReconciler signature**:
  - Change `fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]` → `fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]`
  - Remove `env abstractions.Environment` parameter
- **NewReconciler body**: Update `NewStateFactory` call to remove `env` argument

### 5. `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_controller.go`

**Changes:**

- **Import**: Replace `gcpnfsbackupclientv1` with `gcpnfsbackupclientv2`
- **Import**: Remove `"github.com/kyma-project/cloud-manager/pkg/common/abstractions"` (env removed)
- **GcpNfsVolumeBackupDiscoveryReconcilerFactory struct**:
  - Change `fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]` → `fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]`
  - Remove `env abstractions.Environment` field
- **Factory.New method**: Update call to remove `env` from `gcpnfsvolumebackupdiscovery.NewReconciler(...)`
- **SetupGcpNfsVolumeBackupDiscoveryReconciler signature**:
  - Change `fileBackupClientProvider gcpclient.ClientProvider[gcpnfsbackupclientv1.FileBackupClient]` → `fileBackupClientProvider gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]`
  - Remove `env abstractions.Environment` parameter
  - Remove `logger logr.Logger` parameter (currently unused inside, only passed through)
- **SetupGcpNfsVolumeBackupDiscoveryReconciler body**: Update factory initialization

### 6. `cmd/main.go` (line ~254)

**Changes:**

- Change:
  ```go
  // Before:
  cloudresourcescontroller.SetupGcpNfsVolumeBackupDiscoveryReconciler(
      skrRegistry, gcpnfsbackupclientv1.NewFileBackupClientProvider(), env, setupLog,
  )

  // After:
  cloudresourcescontroller.SetupGcpNfsVolumeBackupDiscoveryReconciler(
      skrRegistry, gcpnfsbackupclientv2.NewFileBackupClientProvider(gcpClients),
  )
  ```
- Import alias `gcpnfsbackupclientv2` should already exist (used by GcpNfsVolumeBackup setup above)

### 7. `internal/controller/cloud-resources/suite_test.go` (line ~122)

**Changes:**

- Change:
  ```go
  // Before:
  Expect(SetupGcpNfsVolumeBackupDiscoveryReconciler(
      infra.Registry(), infra.GcpMock().FileBackupClientProvider(), env, testSetupLog,
  )).NotTo(HaveOccurred())

  // After:
  Expect(SetupGcpNfsVolumeBackupDiscoveryReconciler(
      infra.Registry(), infra.GcpMock().FileBackupClientProviderV2(),
  )).NotTo(HaveOccurred())
  ```

### 8. `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go`

**Changes:**

- **Import**: Replace `"google.golang.org/api/file/v1"` with `"cloud.google.com/go/filestore/apiv1/filestorepb"` and `"google.golang.org/protobuf/types/known/timestamppb"`
- **Test data**: Replace `infra.GcpMock().CreateFakeBackup(&file.Backup{...})` with `infra.GcpMock().CreateFakeBackupV2(&filestorepb.Backup{...})`
- **Backup field mapping**:
  - `Name` → `Name` (same)
  - `Description` → `Description` (same, but not present in filestorepb — remove if not in proto)
  - `State: "READY"` → `State: filestorepb.Backup_READY` (enum, not string)
  - `CreateTime: "2024-10-30T10:00:00Z"` → `CreateTime: timestamppb.New(time.Date(2024, 10, 30, 10, 0, 0, 0, time.UTC))`
  - `SourceFileShare` → `SourceFileShare` (same)
  - `SourceInstance` → `SourceInstance` (same)
  - `SourceInstanceTier: "STANDARD"` → remove or use `SourceInstanceTier: filestorepb.Instance_STANDARD` (check proto)
  - `Labels` → `Labels` (same `map[string]string`)
  - `CapacityGb: 100` → `CapacityGb: 100` (same `int64`)
  - `StorageBytes: 107374182400` → `StorageBytes: 107374182400` (same `int64`)

---

## Files NOT Requiring Changes

| File | Reason |
|------|--------|
| `pkg/skr/gcpnfsvolumebackupdiscovery/ignorant.go` | No v1 client dependency |
| `pkg/skr/gcpnfsvolumebackupdiscovery/loadScope.go` | No v1 client dependency |
| `pkg/skr/gcpnfsvolumebackupdiscovery/setProcessing.go` | No v1 client dependency |
| `pkg/skr/gcpnfsvolumebackupdiscovery/shortCircuit.go` | No v1 client dependency |
| `pkg/skr/gcpnfsvolumebackupdiscovery/util.go` | Parses backup `Name` strings — same format in both v1 and v2 |
| `pkg/kcp/provider/gcp/nfsbackup/client/v2/*` | Already exists, no modifications needed |
| `pkg/kcp/provider/gcp/mock/nfsBackupStoreV2.go` | Already exists, already implements v2 interface |
| `pkg/kcp/provider/gcp/mock/type.go` | Already has `FileBackupClientProviderV2()` method |
| `pkg/kcp/provider/gcp/mock/server.go` | Already has `FileBackupClientProviderV2()` implementation |

---

## Migration Steps (Execution Order)

1. **Modify `state.go`** — Core type changes (State struct, StateFactory, provider type)
2. **Modify `loadAvailableBackups.go`** — Change `ListFilesBackups` → `ListBackups`
3. **Modify `updateStatus.go`** — Change `CreateTime` from string parsing to `*timestamppb.Timestamp`
4. **Modify `reconciler.go`** — Update NewReconciler signature/imports
5. **Modify `gcpnfsvolumebackupdiscovery_controller.go`** — Update controller factory and setup function
6. **Modify `cmd/main.go`** — Wire v2 provider with `gcpClients`
7. **Modify `suite_test.go`** — Use v2 mock provider
8. **Modify `gcpnfsvolumebackupdiscovery_test.go`** — Use `filestorepb.Backup` in test data
9. **Run `make build`** — Verify compilation
10. **Run controller tests** — Verify `gcpnfsvolumebackupdiscovery_test.go` passes

---

## Progress Tracker

- [x] Step 1: `pkg/skr/gcpnfsvolumebackupdiscovery/state.go`
- [x] Step 2: `pkg/skr/gcpnfsvolumebackupdiscovery/loadAvailableBackups.go`
- [x] Step 3: `pkg/skr/gcpnfsvolumebackupdiscovery/updateStatus.go`
- [x] Step 4: `pkg/skr/gcpnfsvolumebackupdiscovery/reconciler.go`
- [x] Step 5: `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_controller.go`
- [x] Step 6: `cmd/main.go`
- [x] Step 7: `internal/controller/cloud-resources/suite_test.go`
- [x] Step 8: `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go`
- [x] Step 9: `make build`
- [ ] Step 10: Controller tests pass

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| `filestorepb.Backup` missing fields used in test (Description, SourceInstanceTier) | Low | Check proto definition; remove or map to equivalent enum fields |
| Mock v2 `ListBackups` filter matching differs from v1 | Low | Already verified — v2 mock uses same regex-based filter matching |
| `CreateTime` nil-safety | Low | v2 uses `*timestamppb.Timestamp` — add nil check before `.AsTime()` |
| `env` parameter removal is a signature change | Low | The `env` parameter is unused in the discovery reconciler — safe to remove |

---

## Reference Implementation

The `GcpNfsVolumeBackup` SKR v2 reconciler at `pkg/skr/gcpnfsvolumebackup/v2/` serves as the
canonical reference for this migration. Its `state.go` demonstrates the correct v2 pattern:

- Uses `gcpclient.GcpClientProvider[v2client.FileBackupClient]`
- Client provider is `func() T` (no args, no error)
- State stores `*filestorepb.Backup` instead of `*file.Backup`
