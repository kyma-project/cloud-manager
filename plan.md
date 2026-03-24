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

### 2. IpRange v3 (`pkg/kcp/provider/gcp/iprange/v3`)

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

### 3. NfsInstance v2 (`pkg/kcp/provider/gcp/nfsinstance/v2`)

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

### 4. Subnet (`pkg/kcp/provider/gcp/subnet`)

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

### 5. RedisCluster (`pkg/kcp/provider/gcp/rediscluster`)

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

### 6. VpcPeering (`pkg/kcp/provider/gcp/vpcpeering`)

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
- The mock v1 (`pkg/kcp/provider/gcp/mock/`) should **NOT be deleted** — it is still used by features not yet migrated (e.g., NfsInstance v1, NfsBackup, NfsRestore, Scope/ExposedData via old path, IpRange v2 legacy). Only remove mock v1 references for the specific features being migrated.
- The `nfsinstance_gcp_test.go` (v1) test should remain on mock v1 — only `nfsinstance_gcp_v2_test.go` migrates to mock v2
