# GCP Client Refactoring Plan

## Context

The project is migrating GCP clients from an **OLD pattern** (feature-local clients directly store concrete GCP SDK types) to a **NEW pattern** (feature-local clients use wrapped interfaces from the central `gcpclient` package).

Each feature migration has **two parts**:
1. **Client code refactoring** — switch the feature-local client to use wrapped interfaces
2. **Mock migration** — switch controller tests from mock v1 (`pkg/kcp/provider/gcp/mock`) to mock v2 (`pkg/kcp/provider/gcp/mock2`)

A feature migration is only complete when **both parts** are done.

### Reference Implementation

**Client code:** `pkg/kcp/provider/gcp/exposedData/client/client.go`  
**Mock usage:** `pkg/kcp/provider/gcp/vpcnetwork/` (uses mock2 in controller tests)

```
OLD: client stores *redis.CloudRedisClient (concrete SDK type)
NEW: client stores gcpclient.RoutersClient (wrapped interface from central package)
```

### Key Differences

| Aspect | OLD (current) | NEW (target) |
|--------|--------------|--------------|
| **Client field type** | Concrete SDK type (e.g., `*redis.CloudRedisClient`) | Wrapped interface from `gcpclient` (e.g., `gcpclient.RedisInstanceClient`) |
| **Client initialization** | `gcpClients.RedisInstance` (raw field access) | `gcpClients.RedisInstanceWrapped()` (method returning interface) |
| **Mock version** | mock v1 (`pkg/kcp/provider/gcp/mock`) | mock v2 (`pkg/kcp/provider/gcp/mock2`) |
| **Test data setup** | Global helpers like `infra.GcpMock().GetMemoryStoreRedisByName()` | Subscription-scoped: `gcpMock := infra.GcpMock2().NewSubscription("prefix")` |
| **Test isolation** | Shared global state across all tests | Per-test subscription with `defer gcpMock.Delete()` |

### What Already Exists

#### Wrapped interfaces in `pkg/kcp/provider/gcp/client/`

| Wrapped Interface | File | `*Wrapped()` method exists? |
|---|---|---|
| `RedisInstanceClient` | `clientRedisInstance.go` | **NO** - needs to be added |
| `RedisClusterClient` | `clientRedisCluster.go` | **NO** - needs to be added |
| `FilestoreClient` | `clientFilestore.go` | **NO** - needs to be added |
| `SubnetClient` | `clientSubnet.go` | **NO** - needs to be added |
| `ComputeRegionalOperationsClient` | `clientComputeRegionalOperations.go` | **NO** - needs to be added |
| `NetworkConnectivityClient` | `clientNetworkConnectivity.go` | **NO** - needs to be added |
| `GlobalAddressesClient` | `clientGlobalAddresses.go` | **NO** - needs to be added |
| `ComputeGlobalOperationsClient` | `clientComputeGlobalOperations.go` | **NO** - needs to be added |
| `ServiceNetworkingClient` | `clientServiceNetworking.go` | **NO** - needs to be added |
| `ResourceManagerClient` | `clientResourceManager.go` | **NO** - needs to be added |
| `RoutersClient` | `clientRouters.go` | YES ✅ |
| `RegionalAddressesClient` | `clientRegionalAddresses.go` | YES ✅ |
| `VpcNetworkClient` | `clientVpcNetwork.go` | YES ✅ |

#### Mock v2 (`pkg/kcp/provider/gcp/mock2/`)

- `Store` already implements ALL `gcpclient.*Client` interfaces (RedisInstanceClient, RedisClusterClient, FilestoreClient, SubnetClient, NetworkConnectivityClient, etc.)
- `Server.Providers` interface currently only has: `ExposedDataProvider()`, `VpcNetworkProvider()`
- **New provider methods** must be added per feature (e.g., `RedisInstanceProvider()`)
- `Store` also implements `Configs` interface for operation resolution (e.g., `ResolveRedisInstanceOperation()`)

---

## Mock Migration Architecture

### Mock v1 (OLD) — how it works today

```
suite_test.go:
  SetupRedisInstanceReconciler(
      infra.GcpMock().MemoryStoreProviderFake(),  // returns GcpClientProvider[MemorystoreClient]
      ...
  )

test file:
  infra.GcpMock().GetMemoryStoreRedisByName(name)             // global helper
  infra.GcpMock().SetMemoryStoreRedisLifeCycleState(name, st)  // global helper
  infra.GcpMock().DeleteMemorStoreRedisByName(name)            // global helper
```

- Mock v1 implements the **feature-local client interface** (e.g., `MemorystoreClient`)
- Tests use feature-specific helper methods (e.g., `GetMemoryStoreRedisByName`)
- All tests share a single global mock with no isolation

### Mock v2 (NEW) — target pattern

```
suite_test.go:
  SetupRedisInstanceReconciler(
      infra.GcpMock2().RedisInstanceProvider(),  // returns GcpClientProvider[MemorystoreClient]
      ...
  )

test file:
  gcpMock := infra.GcpMock2().NewSubscription("redis-instance")
  defer gcpMock.Delete()
  // use gcpMock (type Store) directly — it implements gcpclient.RedisInstanceClient
  // use gcpMock.ResolveRedisInstanceOperation(ctx, opName) to complete async ops
  // use gcpMock.GetRedisInstance(ctx, req) to inspect state
```

- Mock v2 implements the **wrapped `gcpclient.*Client` interfaces** (e.g., `gcpclient.RedisInstanceClient`)
- Tests use the same wrapped interface methods for assertions (e.g., `gcpMock.GetRedisInstance()`)
- Each test gets its own subscription (isolated project) via `NewSubscription()`
- Async operations are resolved via `ResolveRedisInstanceOperation()` config methods

### Key changes per feature for mock migration

1. **`pkg/kcp/provider/gcp/mock2/server.go`** — Add provider method (e.g., `RedisInstanceProvider()`)
2. **`pkg/kcp/provider/gcp/mock2/types.go`** — Add provider method to `Providers` interface
3. **`internal/controller/cloud-control/suite_test.go`** — Switch reconciler setup from `infra.GcpMock().XxxProvider()` to `infra.GcpMock2().XxxProvider()`
4. **`internal/controller/cloud-control/*_test.go`** — Rewrite test to use subscription pattern and mock2 APIs

---

## Resources to Migrate (Priority Order)

### 1. RedisInstance (`pkg/kcp/provider/gcp/redisinstance`) ✅ COMPLETE

**Status**: Both Part A (client code) and Part B (mock migration) are complete. Build and tests passing.

#### Part A: Client Code

**Current state:** Feature-local client (`client/memorystoreClient.go`) stores `*redis.CloudRedisClient` directly.

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add `RedisInstanceWrapped()` method returning `RedisInstanceClient` interface

**b. `pkg/kcp/provider/gcp/redisinstance/client/memorystoreClient.go`**
- Change `memorystoreClient.redisInstanceClient` field type from `*redis.CloudRedisClient` to `gcpclient.RedisInstanceClient`
- Update `NewMemorystoreClient()` to use `gcpClients.RedisInstanceWrapped()` instead of `gcpClients.RedisInstance`
- Update all method implementations to call the wrapped interface methods (e.g., `c.redisInstanceClient.CreateRedisInstance(ctx, req)` instead of `c.redisInstanceClient.CreateInstance(ctx, req)`)
- Note: The wrapped interface uses request/response protobuf types directly (e.g., `*redispb.CreateInstanceRequest`), so method bodies need to construct request objects and handle `ResultOperation`/`VoidOperation` return types
- Remove the direct import of `cloud.google.com/go/redis/apiv1`
- The `GetRedisInstance` method currently also calls `GetInstanceAuthString` — verify the wrapped `RedisInstanceClient` does not expose this, meaning the feature client may need to add a call through the wrapped client or extend the interface. Check `clientRedisInstance.go` for `GetInstanceAuthString` support.

**Impact:** `state.go` and action files (e.g., `loadRedis.go`, `createRedis.go`) do NOT need changes since they call the feature-local `MemorystoreClient` interface which stays the same.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/redisinstance_gcp_test.go`
- Uses `infra.GcpMock().GetMemoryStoreRedisByName()`, `SetMemoryStoreRedisLifeCycleState()`, `DeleteMemorStoreRedisByName()`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go`**
- Add `RedisInstanceProvider() gcpclient.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient]` to `Providers` interface

**b. `pkg/kcp/provider/gcp/mock2/server.go`**
- Add `RedisInstanceProvider()` method — returns a `GcpClientProvider[MemorystoreClient]` that creates a feature-local `MemorystoreClient` wrapping the subscription's `Store` (which already implements `gcpclient.RedisInstanceClient`)

**c. `internal/controller/cloud-control/suite_test.go`**
- Change `infra.GcpMock().MemoryStoreProviderFake()` → `infra.GcpMock2().RedisInstanceProvider()`

**d. `internal/controller/cloud-control/redisinstance_gcp_test.go`**
- Add `gcpMock := infra.GcpMock2().NewSubscription("redis-instance")` at test start
- Replace `infra.GcpMock().GetMemoryStoreRedisByName(name)` → `gcpMock.GetRedisInstance(ctx, req)`
- Replace `infra.GcpMock().SetMemoryStoreRedisLifeCycleState(name, state)` → `gcpMock.ResolveRedisInstanceOperation(ctx, opName)`
- Replace `infra.GcpMock().DeleteMemorStoreRedisByName(name)` → `gcpMock.ResolveRedisInstanceOperation(ctx, opName)` (for delete operations)
- The test needs to work with the subscription-based project ID instead of the Scope's project

---

### 2. Subnet (`pkg/kcp/provider/gcp/subnet`) ✅ COMPLETE

**Status**: Both Part A (client code) and Part B (mock migration) are complete. Build and tests passing.

#### Part A: Client Code

**Current state:** Three feature-local clients:
- `client/computeClient.go` — stores `*compute.SubnetworksClient`
- `client/networkConnectivityClient.go` — stores `networkconnectivity.CrossNetworkAutomationClient` (by value, not pointer)
- `client/regionOperationsClient.go` — stores `*compute.RegionOperationsClient`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add `SubnetWrapped()` method returning `SubnetClient`
- Add `RegionOperationsWrapped()` method returning `ComputeRegionalOperationsClient`
- Add `NetworkConnectivityWrapped()` method returning `NetworkConnectivityClient`

**b. `pkg/kcp/provider/gcp/subnet/client/computeClient.go`**
- Change field type to `gcpclient.SubnetClient`
- Update `NewComputeClient()` to use `gcpClients.SubnetWrapped()`
- Update method implementations

**c. `pkg/kcp/provider/gcp/subnet/client/networkConnectivityClient.go`**
- Change field type to `gcpclient.NetworkConnectivityClient`
- Update `NewNetworkConnectivityClient()` to use `gcpClients.NetworkConnectivityWrapped()`
- Update method implementations

**d. `pkg/kcp/provider/gcp/subnet/client/regionOperationsClient.go`**
- Change field type to `gcpclient.ComputeRegionalOperationsClient`
- Update `NewRegionOperationsClient()` to use `gcpClients.RegionOperationsWrapped()`
- Update method implementations

**Impact:** `state.go` and action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/gcpsubnet_test.go`
- Uses `infra.GcpMock().SetRegionOperationDone()`, `infra.GcpMock().GetSubnet()`, `infra.GcpMock().GetServiceConnectionPolicy()`, etc.

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go` & `server.go`**
- Add `SubnetComputeProvider()`, `SubnetNetworkConnectivityProvider()`, `SubnetRegionOperationsProvider()` to `Providers` interface and implement

**b. `internal/controller/cloud-control/suite_test.go`**
- Change GcpSubnet setup from `infra.GcpMock().SubnetComputeClientProvider()` etc. → mock2 equivalents

**c. `internal/controller/cloud-control/gcpsubnet_test.go`**
- Rewrite to subscription pattern using wrapped interface methods (e.g., `gcpMock.GetSubnet()`, `gcpMock.GetServiceConnectionPolicy()`, compute operation resolution)

---

### 3. RedisCluster (`pkg/kcp/provider/gcp/rediscluster`) ✅ COMPLETE

**Status**: Both Part A (client code) and Part B (mock migration) are complete.

#### Part A: Client Code

**Current state:** Feature-local client (`client/memorystoreClusterClient.go`) stores `*cluster.CloudRedisClusterClient` directly.

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add `RedisClusterWrapped()` method returning `RedisClusterClient`

**b. `pkg/kcp/provider/gcp/rediscluster/client/memorystoreClusterClient.go`**
- Change field type from `*cluster.CloudRedisClusterClient` to `gcpclient.RedisClusterClient`
- Update `NewMemorystoreClient()` to use `gcpClients.RedisClusterWrapped()`
- Update method implementations to use wrapped interface methods
- Remove direct import of `cloud.google.com/go/redis/cluster/apiv1`

**Impact:** `state.go` and action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/gcprediscluster_test.go`
- Uses `infra.GcpMock().GetMemoryStoreRedisClusterByName()`, `SetMemoryStoreRedisClusterLifeCycleState()`, `DeleteMemorStoreRedisClusterByName()`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go` & `server.go`**
- Add `RedisClusterProvider()` to `Providers` interface and implement

**b. `internal/controller/cloud-control/suite_test.go`**
- Change `infra.GcpMock().MemoryStoreClusterProviderFake()` → `infra.GcpMock2().RedisClusterProvider()`

**c. `internal/controller/cloud-control/gcprediscluster_test.go`**
- Rewrite to subscription pattern with `gcpMock.GetRedisCluster()`, `gcpMock.ResolveRedisClusterOperation()`, etc.

---

### 4. NfsInstance v2 (`pkg/kcp/provider/gcp/nfsinstance/v2`) ✅ COMPLETE

**Status**: Both Part A (client code) and Part B (mock migration) are complete.

#### Part A: Client Code

**Current state:** Feature-local client (`client/filestoreClient.go`) stores `*filestore.CloudFilestoreManagerClient` directly.

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add `FilestoreWrapped()` method returning `FilestoreClient`

**b. `pkg/kcp/provider/gcp/nfsinstance/v2/client/filestoreClient.go`**
- Change field type from `*filestore.CloudFilestoreManagerClient` to `gcpclient.FilestoreClient`
- Update `NewFilestoreClient()` to use `gcpClients.FilestoreWrapped()`
- Update method implementations to use wrapped interface methods
- The `GetOperation` method currently accesses `c.cloudFilestoreManager.LROClient.GetOperation()` — this needs to map to `gcpclient.FilestoreClient.GetFilestoreOperation()`
- Remove direct import of `cloud.google.com/go/filestore/apiv1`

**Impact:** `state.go` and action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/nfsinstance_gcp_v2_test.go`
- Uses `infra.GcpMock()` with `FilestoreClientFakeUtils` methods

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go` & `server.go`**
- Add `NfsInstanceV2Provider()` to `Providers` interface and implement

**b. `internal/controller/cloud-control/suite_test.go`**
- Change NfsInstance v2 setup from `infra.GcpMock().FilestoreClientProviderV2()` → mock2 equivalent

**c. `internal/controller/cloud-control/nfsinstance_gcp_v2_test.go`**
- Rewrite to use subscription pattern and mock2 APIs (e.g., `gcpMock.GetFilestoreInstance()`, `gcpMock.ResolveFilestoreOperation()`)

---

### 5. VpcPeering (`pkg/kcp/provider/gcp/vpcpeering`) — PENDING

#### Part A: Client Code

**Current state:** Feature-local client (`client/vpcPeeringClient.go`) stores three concrete SDK types:
- `*compute.NetworksClient`
- `*compute.GlobalOperationsClient`
- `*resourcemanager.TagBindingsClient`

These come from `gcpClients.VpcPeeringClients` (separate service account).

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add wrapped methods on `VpcPeeringClients` struct:
  - `NetworkWrapped()` → `VpcNetworkClient`
  - `GlobalOperationsWrapped()` → `ComputeGlobalOperationsClient`
  - `ResourceManagerWrapped()` → `ResourceManagerClient`

**b. `pkg/kcp/provider/gcp/vpcpeering/client/vpcPeeringClient.go`**
- Change field types from concrete SDK types to wrapped interfaces
- Update `NewVpcPeeringClient()` to use wrapped methods
- Update method body implementations
- The `CreateVpcPeeringRequest` standalone function currently takes `*compute.NetworksClient` — it needs to accept the wrapped interface instead
- Remove direct imports of `cloud.google.com/go/compute/apiv1` and `cloud.google.com/go/resourcemanager/apiv3`

**Impact:** `state.go` and action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/vpcpeering_gcp_test.go`
- Uses `infra.GcpMock().SetMockVpcPeeringTags()`, `GetMockVpcPeering()`, `SetMockVpcPeeringLifeCycleState()`, `SetMockVpcPeeringError()`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go` & `server.go`**
- Add `VpcPeeringProvider()` to `Providers` interface and implement

**b. `internal/controller/cloud-control/suite_test.go`**
- Change `infra.GcpMock().VpcPeeringProvider()` → `infra.GcpMock2().VpcPeeringProvider()`

**c. `internal/controller/cloud-control/vpcpeering_gcp_test.go`**
- Rewrite to subscription pattern using wrapped interface methods for network peering operations and resource manager tag queries

---

### 6. IpRange v3 (`pkg/kcp/provider/gcp/iprange/v3`) — PENDING

#### Part A: Client Code

**Current state:** Two feature-local clients:
- `client/computeClient.go` — stores `*compute.GlobalAddressesClient` and `*compute.GlobalOperationsClient` directly
- `client/serviceNetworkingClient.go` — stores `*servicenetworking.APIService` and `*cloudresourcemanager.Service` directly (OLD API, no modern Cloud Client Library available)

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/gcpClients.go`**
- Add `GlobalAddressesWrapped()` method returning `GlobalAddressesClient`
- Add `GlobalOperationsWrapped()` method returning `ComputeGlobalOperationsClient`
- Add `ServiceNetworkingWrapped()` method returning `ServiceNetworkingClient`

**b. `pkg/kcp/provider/gcp/iprange/client/computeClient.go`**
- Change field types from concrete SDK types to wrapped interfaces (`gcpclient.GlobalAddressesClient`, `gcpclient.ComputeGlobalOperationsClient`)
- Update `NewComputeClient()` to use `gcpClients.GlobalAddressesWrapped()` and `gcpClients.GlobalOperationsWrapped()`
- Update method implementations to use wrapped interface methods
- Remove direct import of `cloud.google.com/go/compute/apiv1`

**c. `pkg/kcp/provider/gcp/iprange/client/serviceNetworkingClient.go`**
- Change field type from `*servicenetworking.APIService` / `*cloudresourcemanager.Service` to `gcpclient.ServiceNetworkingClient`
- Update `NewServiceNetworkingClientForService()` and `NewServiceNetworkingClientProvider()` to use wrapped interface
- Update method implementations
- Note: The `ServiceNetworkingClient` wrapped interface may need method adjustments for the OLD API pattern

**Impact:** `state.go` and action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-control/iprange_gcp_refactored_test.go`
- Uses `infra.GcpMock().GetIpRangeDiscovery()`, `infra.GcpMock().ListServiceConnections()`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go`**
- Add `IpRangeComputeProvider()` and `IpRangeServiceNetworkingProvider()` to `Providers` interface

**b. `pkg/kcp/provider/gcp/mock2/server.go`**
- Add provider methods

**c. `internal/controller/cloud-control/suite_test.go`**
- Change IpRange setup from `infra.GcpMock().ServiceNetworkingClientProviderGcp()` / `ComputeClientProviderGcp()` → mock2 equivalents

**d. `internal/controller/cloud-control/iprange_gcp_refactored_test.go`**
- Rewrite to use subscription pattern and mock2 wrapped interface methods

---

## SKR Reconciler Migrations

These SKR reconcilers use GCP clients directly (not through KCP), so they also need client refactoring and mock migration.

### 7. GcpNfsVolumeBackup v2 (`pkg/skr/gcpnfsvolumebackup/v2`) ✅ COMPLETE

**Status**: Both Part A (client code) and Part B (mock migration) are complete. Location-fallback logic extracted to `util.go` unit test. Feature flag skip restored in integration test.

#### Part A: Client Code

**Current state:** Feature-local client (`pkg/kcp/provider/gcp/nfsbackup/client/v2/fileBackupClient.go`) stores `*filestore.CloudFilestoreManagerClient` directly (from `cloud.google.com/go/filestore/apiv1`).

**Methods that use the concrete SDK type:**
- `GetBackup` → calls `c.cloudFilestoreManager.GetBackup(ctx, req)`
- `ListBackups` → calls `c.cloudFilestoreManager.ListBackups(ctx, req)` with iterator
- `CreateBackup` → calls `c.cloudFilestoreManager.CreateBackup(ctx, req)` → returns `op.Name()`
- `DeleteBackup` → calls `c.cloudFilestoreManager.DeleteBackup(ctx, req)` → returns `op.Name()`
- `GetBackupLROperation` → calls `c.cloudFilestoreManager.LROClient.GetOperation(ctx, req)`
- `UpdateBackup` → calls `c.cloudFilestoreManager.UpdateBackup(ctx, req)` → returns `op.Name()`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/nfsbackup/client/v2/fileBackupClient.go`**
- Change `fileBackupClient.cloudFilestoreManager` field type from `*filestore.CloudFilestoreManagerClient` to `gcpclient.FilestoreClient`
- Update `NewFileBackupClient()` to use `gcpClients.FilestoreWrapped()` instead of `gcpClients.Filestore`
- Add `NewFileBackupClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FileBackupClient` constructor (needed for mock2 wiring)
- Update all method implementations to call the wrapped interface methods:
  - `GetBackup` → `c.filestoreClient.GetFilestoreBackup(ctx, &filestorepb.GetBackupRequest{Name: ...})`
  - `ListBackups` → `c.filestoreClient.ListFilestoreBackups(ctx, &filestorepb.ListBackupsRequest{...})` — returns `Iterator[*filestorepb.Backup]` instead of paged iterator, so iteration changes to use `it.All()` range loop
  - `CreateBackup` → `c.filestoreClient.CreateFilestoreBackup(ctx, &filestorepb.CreateBackupRequest{...})` — returns `ResultOperation`, extract `op.Name()`
  - `DeleteBackup` → `c.filestoreClient.DeleteFilestoreBackup(ctx, &filestorepb.DeleteBackupRequest{...})` — returns `VoidOperation`, extract `op.Name()`
  - `GetBackupLROperation` → `c.filestoreClient.GetFilestoreOperation(ctx, &longrunningpb.GetOperationRequest{Name: ...})`
  - `UpdateBackup` → `c.filestoreClient.UpdateFilestoreBackup(ctx, &filestorepb.UpdateBackupRequest{...})` — returns `ResultOperation`, extract `op.Name()`
- Remove direct import of `cloud.google.com/go/filestore/apiv1`
- Remove import of `google.golang.org/api/iterator` (replaced by `gcpclient.Iterator` range-based iteration)

**Note:** `FilestoreWrapped()` method already exists on `GcpClients`. The wrapped `gcpclient.FilestoreClient` already has all needed methods: `GetFilestoreBackup`, `ListFilestoreBackups`, `CreateFilestoreBackup`, `DeleteFilestoreBackup`, `UpdateFilestoreBackup`, `GetFilestoreOperation`.

**Impact:** The `FileBackupClient` interface does NOT change. SKR reconciler `state.go`, `clientCreate.go`, and all action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-resources/gcpnfsvolumebackup_v2_test.go`
- Uses `infra.GcpMock().SetNfsBackupV2State(gcpBackupPath, state)` to simulate backup ready
- Uses `infra.GcpMock().GetNfsBackupV2ByName(gcpBackupPath)` to inspect backup state 
- Uses `infra.GcpMock().DeleteNfsBackupV2ByName(gcpBackupPath)` to simulate GCP deletion

**Current suite wiring:** `internal/controller/cloud-resources/suite_test.go`
```go
Expect(SetupGcpNfsVolumeBackupReconciler(infra.Registry(), infra.GcpMock().FileBackupClientProvider(), infra.GcpMock().FileBackupClientProviderV2(), env)).
```

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go`**
- Add `NfsBackupV2Provider() gcpclient.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient]` to `Providers` interface

**b. `pkg/kcp/provider/gcp/mock2/server.go`**
- Add `NfsBackupV2Provider()` method — returns a `GcpClientProvider[FileBackupClient]` that creates a feature-local `FileBackupClient` wrapping the subscription's `Store` (which already implements `gcpclient.FilestoreClient` with all backup methods)
- Pattern: `return gcpnfsbackupclientv2.NewFileBackupClientFromFilestoreClient(sub)`

**c. `internal/controller/cloud-resources/suite_test.go`**
- Change `infra.GcpMock().FileBackupClientProviderV2()` → `infra.GcpMock2().NfsBackupV2Provider()` in `SetupGcpNfsVolumeBackupReconciler` call
- Keep `infra.GcpMock().FileBackupClientProvider()` for v1 provider parameter (v1 stays on mock v1)

**d. `internal/controller/cloud-resources/gcpnfsvolumebackup_v2_test.go`**
- Add `gcpMock := infra.GcpMock2().NewSubscription("nfs-backup-v2")` at test start
- Add `defer gcpMock.Delete()` for cleanup
- Replace `infra.GcpMock().SetNfsBackupV2State(path, state)` → `gcpMock.ResolveFilestoreBackupOperation(ctx, opName)` (resolving the create operation makes backup READY)
- Replace `infra.GcpMock().GetNfsBackupV2ByName(path)` → `gcpMock.GetFilestoreBackup(ctx, &filestorepb.GetBackupRequest{Name: path})`
- Replace `infra.GcpMock().DeleteNfsBackupV2ByName(path)` → `gcpMock.ResolveFilestoreBackupOperation(ctx, opName)` (resolving the delete operation removes backup)
- The test needs to work with `gcpMock.ProjectId()` instead of the Scope's project (or the Scope project must match the subscription's project)
- The mock2 `CreateFilestoreBackup` requires a valid source instance (validates `SourceInstance` and `SourceFileShare`), so the test may need to create a filestore instance first via `gcpMock.CreateFilestoreInstance()` and resolve it, or the backup creation flow needs to be adapted

---

### 8. GcpNfsVolumeRestore v2 (`pkg/skr/gcpnfsvolumerestore/v2`) ✅ COMPLETE

**Uses:** `pkg/kcp/provider/gcp/nfsrestore/client/v2` (FileRestoreClient) AND `pkg/kcp/provider/gcp/nfsbackup/client/v2` (FileBackupClient)

#### Part A: Client Code

**Current state:** Feature-local client (`pkg/kcp/provider/gcp/nfsrestore/client/v2/fileRestoreClient.go`) stores `*filestore.CloudFilestoreManagerClient` directly.

**Methods that use the concrete SDK type:**
- `RestoreFile` → calls `c.cloudFilestoreManager.RestoreInstance(ctx, req)` → returns `op.Name()`
- `GetRestoreOperation` → calls `c.cloudFilestoreManager.LROClient.GetOperation(ctx, req)`
- `FindRestoreOperation` → calls `c.cloudFilestoreManager.LROClient.ListOperations(ctx, req)` with iterator, filters for `metadata.verb="restore"` + `metadata.target=...`

**Prerequisite — `gcpclient.FilestoreClient` interface must be extended:**
- The wrapped `gcpclient.FilestoreClient` interface (`pkg/kcp/provider/gcp/client/clientFilestore.go`) does NOT currently include `RestoreFilestoreInstance` in the interface definition (it exists only on the concrete `filestoreClient` struct)
- **Must add** `RestoreFilestoreInstance(ctx context.Context, req *filestorepb.RestoreInstanceRequest, opts ...gax.CallOption) (ResultOperation[*filestorepb.Instance], error)` to the `FilestoreClient` interface

**Prerequisite — mock2 Store must implement RestoreFilestoreInstance:**
- The mock2 `Store` currently does NOT implement `RestoreFilestoreInstance`
- Must add `RestoreFilestoreInstance` to `pkg/kcp/provider/gcp/mock2/storeFilestore.go` — creates/registers a long-running restore operation targeting the filestore instance
- Must add `FileStoreRestoreOperationsConfig` interface to `pkg/kcp/provider/gcp/mock2/` with `ResolveFilestoreRestoreOperation()` method for tests to complete restore operations
- Add `FileStoreRestoreOperationsConfig` to the `Configs` interface in `types.go`

**Changes needed:**

**a. `pkg/kcp/provider/gcp/client/clientFilestore.go`**
- Add `RestoreFilestoreInstance` to `FilestoreClient` interface

**b. `pkg/kcp/provider/gcp/mock2/storeFilestore.go`**
- Implement `RestoreFilestoreInstance` on `store` — validate instance exists, create long-running operation with verb "restore"

**c. `pkg/kcp/provider/gcp/mock2/storeOperationsLongRunning_filestore.go`** (or new file `storeOperationsLongRunning_restore.go`)
- Add `FileStoreRestoreOperationsConfig` interface with `ResolveFilestoreRestoreOperation(ctx, operationName, opts...)` method
- Implement resolution: find operation, set filestore instance to READY, mark operation done

**d. `pkg/kcp/provider/gcp/mock2/types.go`**
- Add `FileStoreRestoreOperationsConfig` to `Configs` interface

**e. `pkg/kcp/provider/gcp/nfsrestore/client/v2/fileRestoreClient.go`**
- Change `fileRestoreClient.cloudFilestoreManager` field type from `*filestore.CloudFilestoreManagerClient` to `gcpclient.FilestoreClient`
- Update `NewFileRestoreClient()` to use `gcpClients.FilestoreWrapped()` instead of `gcpClients.Filestore`
- Add `NewFileRestoreClientFromFilestoreClient(filestoreClient gcpclient.FilestoreClient) FileRestoreClient` constructor (needed for mock2 wiring)
- Update method implementations:
  - `RestoreFile` → `c.filestoreClient.RestoreFilestoreInstance(ctx, &filestorepb.RestoreInstanceRequest{...})` — returns `ResultOperation`, extract `op.Name()`
  - `GetRestoreOperation` → `c.filestoreClient.GetFilestoreOperation(ctx, &longrunningpb.GetOperationRequest{Name: ...})`
  - `FindRestoreOperation` → `c.filestoreClient.ListFilestoreOperations(ctx, &longrunningpb.ListOperationsRequest{...})` — iterate with `it.All()` range loop
- Remove direct import of `cloud.google.com/go/filestore/apiv1`
- Remove import of `google.golang.org/api/iterator` (replaced by `gcpclient.Iterator` range-based iteration)
- Note: `FindFilestoreRestoreOperation` high-level function already exists on `clientFilestore.go` concrete type — but since we're wrapping at the feature level, the feature client should implement its own version using the interface methods

**Impact:** The `FileRestoreClient` interface does NOT change. SKR reconciler `state.go`, `clientCreate.go`, and all action files do NOT need changes.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-resources/gcpnfsvolumerestore_v2_test.go`
- Uses `infra.GcpMock().SetRestoreOperationDoneV2(opName)` to complete restore operation

**Current suite wiring:** `internal/controller/cloud-resources/suite_test.go`
```go
Expect(SetupGcpNfsVolumeRestoreReconciler(infra.Registry(), infra.GcpMock().FilerestoreClientProvider(), infra.GcpMock().FileBackupClientProvider(), infra.GcpMock().FileRestoreClientProviderV2(), infra.GcpMock().FileBackupClientProviderV2(), env)).
```

**Changes needed:**

**a. `pkg/kcp/provider/gcp/mock2/types.go`**
- Add `NfsRestoreV2Provider() gcpclient.GcpClientProvider[gcpnfsrestoreclientv2.FileRestoreClient]` to `Providers` interface

**b. `pkg/kcp/provider/gcp/mock2/server.go`**
- Add `NfsRestoreV2Provider()` method — returns a `GcpClientProvider[FileRestoreClient]` wrapping the subscription's Store
- Pattern: `return gcpnfsrestoreclientv2.NewFileRestoreClientFromFilestoreClient(sub)`

**c. `internal/controller/cloud-resources/suite_test.go`**
- Change `infra.GcpMock().FileRestoreClientProviderV2()` → `infra.GcpMock2().NfsRestoreV2Provider()` in `SetupGcpNfsVolumeRestoreReconciler` call
- Change `infra.GcpMock().FileBackupClientProviderV2()` → `infra.GcpMock2().NfsBackupV2Provider()` in `SetupGcpNfsVolumeRestoreReconciler` call
- Keep v1 providers on mock v1

**d. `internal/controller/cloud-resources/gcpnfsvolumerestore_v2_test.go`**
- Add `gcpMock := infra.GcpMock2().NewSubscription("nfs-restore-v2")` at test start
- Add `defer gcpMock.Delete()` for cleanup
- Replace `infra.GcpMock().SetRestoreOperationDoneV2(opName)` → `gcpMock.ResolveFilestoreRestoreOperation(ctx, opName)` (or `gcpMock.ResolveFilestoreOperation(ctx, opName)` if restore ops use the same resolution path)
- Restore tests need a filestore instance to exist in mock2 (the restore targets an instance), so setup must create one
- Restore tests may need a backup to exist (as the source), so setup must also create a backup in mock2

---

### 9. GcpNfsVolumeBackupDiscovery (`pkg/skr/gcpnfsvolumebackupdiscovery`) ✅ COMPLETE

**Status**: Both Part A (no client changes needed — reuses FileBackupClient from #7) and Part B (mock migration) are complete.

**Uses:** `pkg/kcp/provider/gcp/nfsbackup/client/v2` (FileBackupClient) — same client as GcpNfsVolumeBackup v2

#### Part A: Client Code

**No additional client changes needed.** This feature reuses the same `FileBackupClient` from `pkg/kcp/provider/gcp/nfsbackup/client/v2/` that is migrated in item #7 above. The main method used is `ListBackups` which calls the wrapped `ListFilestoreBackups`.

**Impact:** None — automatically benefits from the client migration in item #7.

#### Part B: Mock Migration

**Current test:** `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go`
- Uses `infra.GcpMock().CreateFakeBackupV2(&filestorepb.Backup{...})` to pre-populate backups

**Current suite wiring:** `internal/controller/cloud-resources/suite_test.go`
```go
Expect(SetupGcpNfsVolumeBackupDiscoveryReconciler(infra.Registry(), infra.GcpMock().FileBackupClientProviderV2())).
```

**Changes needed:**

**a. `internal/controller/cloud-resources/suite_test.go`**
- Change `infra.GcpMock().FileBackupClientProviderV2()` → `infra.GcpMock2().NfsBackupV2Provider()` in `SetupGcpNfsVolumeBackupDiscoveryReconciler` call
- Note: `NfsBackupV2Provider` is already defined in item #7

**b. `internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go`**
- Add `gcpMock := infra.GcpMock2().NewSubscription("backup-discovery")` at test start
- Add `defer gcpMock.Delete()` for cleanup
- Replace `infra.GcpMock().CreateFakeBackupV2(&filestorepb.Backup{...})` → use `gcpMock.CreateFilestoreBackup(ctx, &filestorepb.CreateBackupRequest{...})` followed by `gcpMock.ResolveFilestoreBackupOperation(ctx, opName)` to make backups READY
- **Important:** Mock2's `CreateFilestoreBackup` validates source instance and file share — the test must first create a filestore instance in mock2 before creating backups. Alternatively, backups can be pre-populated differently if mock2 provides a direct insertion method.
- The backup labels (managed-by, scope-name, cm-allow-*) must be set in the `CreateBackupRequest.Backup.Labels` so that `ListFilestoreBackups` filter expressions match them
- The test needs to work with `gcpMock.ProjectId()` for the project — the backups' `Name` field must use this project, and the Scope project must match

**Note on filter compatibility:** Mock2's `ListFilestoreBackups` uses `FilterByExpression` which parses label-based filter expressions. The discovery reconciler calls `ListBackups` with `gcpclient.GetSharedBackupsFilter(shootName, subaccountId)` — verify that mock2's filter expression parser handles this filter format (e.g., `labels.cm-allow-{shootName}:*`). If not, the filter parser may need extension.

---

## Implementation Checklist Per Feature

For each resource, the complete steps are:

### Client Code (Part A)
1. Add `*Wrapped()` method to `gcpClients.go` (if not already present)
2. Update feature-local client to use wrapped interface instead of concrete SDK type
3. Verify compilation (`make build`)

### Mock Migration (Part B)
4. Add provider method to `mock2/types.go` (`Providers` interface)
5. Implement provider method in `mock2/server.go`
6. Update `suite_test.go` to wire mock2 provider into reconciler setup
7. Rewrite controller test file to use subscription pattern + mock2 APIs
8. Run tests (`make test`)

### Suggested PR strategy

**One PR per resource** (recommended) — each PR contains both Part A (client) and Part B (mock/tests):
- Easier to review, isolated blast radius
- Each PR is self-contained and the feature works end-to-end

---

## Notes

- The feature-local `Client` **interface** (e.g., `MemorystoreClient`) does NOT change — only the internal implementation changes
- `state.go` files do NOT change — they already use `GcpClientProvider[T]` pattern
- `cmd/main.go` does NOT change — the `NewXxxClientProvider(gcpClients)` calls remain the same
- Action files do NOT change — they call the feature-local client interface
- The wrapped interfaces in `pkg/kcp/provider/gcp/client/client*.go` use request/response protobuf types with `ResultOperation`/`VoidOperation` return types. The feature-local client methods need to handle these (e.g., extracting results from `ResultOperation`, handling `VoidOperation`)
- For `GetInstanceAuthString` (RedisInstance): check if the wrapped `RedisInstanceClient` interface includes it; if not, the interface needs extending in `clientRedisInstance.go`
- For `ServiceNetworking` (IpRange): this uses Google's OLD pattern API. The wrapped `ServiceNetworkingClient` interface in `clientServiceNetworking.go` may need methods that match the old API patterns
- The mock v1 (`pkg/kcp/provider/gcp/mock/`) should **NOT be deleted** — it is still used by features not yet migrated (e.g., NfsInstance v1, NfsBackup v1, NfsRestore v1, Scope/ExposedData via old path, IpRange v2 legacy). Only remove mock v1 references for the specific features being migrated.
- The `nfsinstance_gcp_test.go` (v1) test should remain on mock v1 — only `nfsinstance_gcp_v2_test.go` migrates to mock v2
- SKR reconcilers (items 7–9) use the same `gcpclient.FilestoreClient` wrapped interface as KCP NfsInstance v2 — the wrapped interface already covers backup operations (Get/List/Create/Delete/Update) but **must be extended** with `RestoreFilestoreInstance` for item #8 (NfsRestore v2)
- Items 7 and 9 share the same `FileBackupClient` — the client migration (Part A) only needs to be done once (in item #7), and item #9 benefits automatically
- The `SetupGcpNfsVolumeBackupReconciler` and `SetupGcpNfsVolumeRestoreReconciler` take both v1 and v2 provider parameters — only the v2 parameters switch to mock2, v1 parameters stay on mock v1
- Mock2's `CreateFilestoreBackup` validates `SourceInstance` and `SourceFileShare` exist — tests that create backups must first set up a filestore instance in mock2 (unlike mock v1 which accepted any backup data without validation)
