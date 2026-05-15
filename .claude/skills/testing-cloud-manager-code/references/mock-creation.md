# Cloud Provider Mock Creation

## Overview

Each provider has its own mock structure under `pkg/kcp/provider/<provider>/mock/`. They are **not** interchangeable — each provider has a different architecture. Below is what each looks like; extend the relevant one when adding a new resource.

Accessed via `infra` in tests:
- `infra.GcpMock()` — GCP old mock (pkg/kcp/provider/gcp/mock/)
- `infra.GcpMock2()` — GCP new mock (pkg/kcp/provider/gcp/mock2/) — use for new resources
- `infra.AwsMock()` — AWS mock (pkg/kcp/provider/aws/mock/)
- `infra.AzureMock()` — Azure mock (pkg/kcp/provider/azure/mock/)

---

## GCP mock2 — New Pattern (pkg/kcp/provider/gcp/mock2/)

**Use this for all new GCP resources.** The `Server` manages isolated subscriptions; each test scenario creates one.

### Architecture

```
pkg/kcp/provider/gcp/mock2/
├── types.go                   # Server, Store, Clients, Providers, Configs interfaces
├── server.go                  # Server implementation, NewSubscription/GetSubscription/DeleteSubscription
├── store.go                   # Store implementation (per-subscription state)
├── storeFilestore.go          # Filestore client impl (per subscription)
├── storeRedisCluster.go       # RedisCluster client impl
├── storeNetworks.go           # VPC network client impl
└── ...                        # one file per resource type
```

### Key interfaces

```go
type Server interface {
    Providers                     // provider factory methods for reconciler wiring
    NewSubscription(prefix string) Store
    GetSubscription(projectId string) Store   // used internally by providers
    DeleteSubscription(projectId string)
}

type Store interface {
    ProjectId() string
    Delete()           // shortcut for DeleteSubscription(this.ProjectId())
    Clients            // direct client calls (GCP API methods)
    Configs            // test control: set states, resolve operations, inject errors
}
```

### Adding a new resource to mock2

**1. Add to Clients interface (types.go)**

```go
type Clients interface {
    // existing...
    gcpclient.MyNewResourceClient
}
```

**2. Add a Configs interface for test control (types.go)**

```go
type MyNewResourceConfig interface {
    // Methods that tests call to control mock behavior
    SetMyResourceState(id string, state string)
    GetMyResourceByName(name string) *mypb.Resource
    ResolveMyOperation(ctx context.Context, opName string) error
}

type Configs interface {
    // existing...
    MyNewResourceConfig
}
```

**3. Create storeMyResource.go** — implement both `gcpclient.MyNewResourceClient` (real API methods) and `MyNewResourceConfig` (test-control methods) on the `store` struct:

```go
// storeMyResource.go
package mock2

import (
    "context"
    "sync"
)

type myNewResourceStore struct {
    mu        sync.Mutex
    resources map[string]*mypb.Resource
}

// Client method — called by reconciler
func (s *store) CreateMyResource(ctx context.Context, req *mypb.CreateRequest) (*longrunningpb.Operation, error) {
    s.myResource.mu.Lock()
    defer s.myResource.mu.Unlock()

    r := &mypb.Resource{Name: req.Name, State: mypb.Resource_CREATING}
    s.myResource.resources[req.Name] = r
    opName := "operations/create-" + req.Name
    s.addOperation(opName)
    return &longrunningpb.Operation{Name: opName}, nil
}

// Config method — called by tests
func (s *store) SetMyResourceState(id string, state string) {
    s.myResource.mu.Lock()
    defer s.myResource.mu.Unlock()
    if r, ok := s.myResource.resources[id]; ok {
        r.State = mypb.Resource_State(mypb.Resource_State_value[state])
    }
}
```

**4. Add a Provider factory to types.go + server.go**

```go
// types.go
type Providers interface {
    // existing...
    MyNewResourceProvider() gcpclient.GcpClientProvider[mypb.MyNewResourceClient]
}

// server.go — provider returns client from the subscription's Store
func (srv *server) MyNewResourceProvider() gcpclient.GcpClientProvider[mypb.MyNewResourceClient] {
    return func(ctx context.Context, projectId string) (mypb.MyNewResourceClient, error) {
        store := srv.GetSubscription(projectId)
        if store == nil {
            return nil, fmt.Errorf("no subscription for project %s", projectId)
        }
        return store, nil
    }
}
```

---

## GCP mock (old pattern) — pkg/kcp/provider/gcp/mock/

Used for older resources (VpcPeering, NfsInstance v1, IpRange v2, Scope, etc.). Flat global state — no subscriptions. State is mutated globally:

```go
infra.GcpMock().SetRedisLifeCycleState(instanceName, redispb.Instance_READY)
infra.GcpMock().GetFilestoreInstanceByName(name)
```

When adding to this mock: follow the existing dual-interface pattern in `memoryStoreClientFake.go`. Each fake implements a `*Client` interface (for reconcilers) and a `*Utils` interface (for tests), both aggregated in `Server` via `type.go`.

---

## AWS mock — pkg/kcp/provider/aws/mock/

AWS uses an **Account + Region** model. Each test scenario creates an account:

```go
awsAccount := infra.AwsMock().NewAccount()
defer awsAccount.Delete()

// Account exposes: AccountId(), Region(), and all client methods
Eventually(CreateScopeAws).
    WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), ...).
    Should(Succeed())
```

Structure:
```go
type Server interface {
    Providers
    NewAccount() Account
    GetAccount(accountId string) Account
    Login(key, secret string) (Account, error)
}

type Account interface {
    AccountId() string
    Delete()
    Clients   // all AWS client methods for this account
    Configs   // test control: VpcConfig, NfsConfig, AwsElastiCacheMockUtils, etc.
}
```

Adding a new AWS resource: add client interface to `Clients`, add config interface to `Configs`, implement on the account struct, add a `SkrClientProvider` method to `Providers`.

---

## SAP/OpenStack mock — pkg/kcp/provider/sap/mock/

SAP uses a **Project** model. Each test scenario creates a project with randomized credentials. There is **no `Delete()`** — projects accumulate in the server for the test suite's lifetime (isolation is achieved via unique random credentials).

```go
sapMock := infra.SapMock().NewProject()
// sapMock.DomainName(), sapMock.ProjectName(), sapMock.RegionName() — random per call
// sapMock.ProviderParams() — convenience wrapper for all three

Eventually(CreateScopeOpenStack).
    WithArguments(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(name)).
    Should(Succeed())
```

Structure:
```go
type Server interface {
    Providers                                        // provider factory methods
    NewProject() Project                             // creates isolated project
    GetProject(domainName, project, region string) Project
}

type Project interface {
    DomainName() string
    ProjectName() string
    RegionName() string
    ProviderParams() sapclient.ProviderParams        // all three in one struct
    Clients                                          // OpenStack client methods (Network, Subnet, Router, Share, Snapshot, Port)
    Config                                           // NfsConfig: AddNetwork, AddRouter, SetShareStatus, SetSnapshotStatus
}
```

Test-control methods (`Config`):
```go
// Pre-seed infra that the reconciler expects to find
sapMock.AddNetwork(networkId, networkName)
sapMock.AddRouter(routerId, routerName, "10.0.0.1")
sapMock.SetShareStatus(shareId, "available")
sapMock.SetSnapshotStatus(snapshotId, "available")
```

Direct client calls work too (sapMock implements full Clients interface):
```go
subnet, err := sapMock.GetSubnetByName(infra.Ctx(), vpcId, subnetName)
arr, err := sapMock.ListRouterSubnetInterfaces(infra.Ctx(), routerId)
net, err := sapMock.GetNetwork(infra.Ctx(), networkId)
```

**Key differences from GCP/AWS:**
- Not-found: `sapmeta.NewNotFoundError("...")` and `gophercloud.ErrUnexpectedResponseCode{Actual: http.StatusNotFound}`
- Context cancel: `util.IsContextDone(ctx)` (not `isContextCanceled`)
- All client methods add `time.Sleep(time.Millisecond)` to simulate OpenStack latency
- No async operations — calls are synchronous (no operation IDs)

Adding a new SAP resource: add client interface to `Clients` in `types.go`, add test-control methods to a new `MyResourceConfig` interface under `Config`, implement both on `mainStore` in a new `mainStoreMyResource.go` file, add provider method to `Providers`.

---

## Azure mock — pkg/kcp/provider/azure/mock/

Azure uses a large flat interface with many store types. No account/subscription isolation — state is global per mock instance.

```go
type Server interface {
    Providers
    // Many store interfaces: RedisInstanceStore, VpcPeeringStore, NetworkStore, ...
    // Each store has Get/Set/Delete methods for test control
}
```

Example pattern from tests:
```go
infra.AzureMock().SetRedisInstanceState(name, armredis.ProvisioningStateSucceeded)
```

Providers are injected into reconcilers:
```go
infra.AzureMock().RedisClientProvider()
infra.AzureMock().VpcPeeringProvider()
```

Adding a new Azure resource: add a store interface to the type list in `type.go`, implement on the server struct, add provider method.

---

## Checklist (mock2 additions)

- [ ] New `Clients` entry in `types.go`
- [ ] New `Configs` interface in `types.go` with test-control methods
- [ ] `storeMyResource.go` implements both client and config methods on `store`
- [ ] All client methods check `isContextCanceled(ctx)` and return `context.Canceled` if true
- [ ] Create → resource in transitional state (CREATING, not READY/ACTIVE)
- [ ] Operations returned as `*longrunningpb.Operation` with non-empty `Name`
- [ ] `ResolveMyOperation` marks operation `Done=true` and transitions resource to final state
- [ ] Provider factory in `server.go` returns `store` via `GetSubscription(projectId)`
- [ ] Provider method added to `Providers` interface in `types.go`
- [ ] Mock store initialized in `store.go` `newStore()` constructor
