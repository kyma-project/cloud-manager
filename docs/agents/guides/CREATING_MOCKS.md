# Guide: Creating Cloud Provider Mocks

**Target Audience**: LLM coding agents  
**Prerequisites**: [CONTROLLER_TESTS.md](CONTROLLER_TESTS.md), [ADD_KCP_RECONCILER.md](ADD_KCP_RECONCILER.md)  
**Purpose**: Step-by-step guide for creating mock implementations of cloud provider clients  
**Context**: Cloud Manager uses dual-interface pattern (Client + Utils) for testing without real API calls

## Authority: Mock Implementation Requirements

### MUST

- **MUST use dual-interface pattern**: Client interface (for reconcilers) + Utils interface (for tests)
- **MUST be thread-safe**: All map operations MUST use `sync.Mutex`
- **MUST check context cancellation**: EVERY method checks `isContextCanceled(ctx)` first
- **MUST return realistic errors**: Use `googleapi.Error` with proper HTTP codes (404, 409, 500)
- **MUST model async operations**: Use separate operations mock for long-running tasks
- **MUST use realistic state transitions**: CREATING → READY, UPDATING, DELETING (not immediate removal)
- **MUST match real API naming**: Use same name construction helpers as real client
- **MUST embed in Server**: All mocks embed in `server` struct, return server from providers

### MUST NOT

- **NEVER skip mutex locking**: All map reads/writes MUST be protected
- **NEVER immediately delete resources**: Set DELETING state, let test explicitly remove
- **NEVER expose Utils through Client interface**: Keep test utilities separate
- **NEVER use unrealistic mock data**: IPs, ports, endpoints must look realistic
- **NEVER skip context checks**: Always check `isContextCanceled()` at method start
- **NEVER mutate without locks**: All state changes require mutex

### ALWAYS

- **ALWAYS implement both interfaces**: Client (reconciler calls) + Utils (test calls)
- **ALWAYS lock/unlock with defer**: Pattern `m.mutex.Lock(); defer m.mutex.Unlock()`
- **ALWAYS return 404 for not found**: Use `&googleapi.Error{Code: 404, Message: "Not Found"}`
- **ALWAYS initialize in New()**: Create mock with empty maps in `New()` function
- **ALWAYS use realistic values**: Mock endpoints like `192.168.0.1:6379`, not `localhost`

### NEVER

- **NEVER mix test and client methods**: Utils methods for tests only, Client methods for reconcilers
- **NEVER bypass state transitions**: Model real cloud provider behavior (async states)
- **NEVER hardcode resource keys**: Use name construction helpers from real client
- **NEVER skip error returns**: All operations check errors and return appropriate types

## Mock Architecture: Dual-Interface Pattern

### Client Interface vs Utils Interface

```
┌─────────────────────────────────────────────────────────────┐
│                         Server                               │
│  (Aggregates all mocks + providers + utilities)            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Client Interface              Utils Interface              │
│  (Reconciler calls)            (Test manipulates)           │
│  ┌─────────────────────┐      ┌──────────────────────┐     │
│  │ CreateRedisInstance │      │ GetRedisInstanceByName│     │
│  │ GetRedisInstance    │      │ SetRedisState        │     │
│  │ UpdateRedisInstance │      │ DeleteRedisInstance  │     │
│  │ DeleteRedisInstance │      └──────────────────────┘     │
│  └─────────────────────┘                                    │
│         ▲                              ▲                     │
│         │                              │                     │
│    Reconciler                       Test                    │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

**Location**: `pkg/kcp/provider/<provider>/mock/`

```
pkg/kcp/provider/gcp/mock/
├── type.go                          # Server interface aggregation
├── server.go                        # Server implementation
├── memoryStoreClientFake.go        # Redis mock
├── computeClientFake.go            # Compute mock
└── regionalOperationsClientFake.go # Operations mock
```

## Implementation Steps

### Step 1: Define Utils Interface

**Location**: `<service>ClientFake.go`

#### ❌ WRONG: Missing Utils Interface

```go
// NEVER: Only implementing Client interface without Utils
type memoryStoreClientFake struct {
    mutex          sync.Mutex
    redisInstances map[string]*redispb.Instance
}

// WRONG: No test utilities!
// Tests cannot inspect or manipulate mock state
```

#### ✅ CORRECT: Utils Interface for Test Manipulation

```go
// ALWAYS: Define Utils interface first
package mock

import (
    "context"
    "sync"
    "cloud.google.com/go/redis/apiv1/redispb"
    gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
    "google.golang.org/api/googleapi"
)

// Utils interface for test manipulation
type MemoryStoreClientFakeUtils interface {
    // Get methods - retrieve mock resources by identifier
    GetMemoryStoreRedisByName(name string) *redispb.Instance
    
    // Set methods - modify mock resource state
    SetMemoryStoreRedisLifeCycleState(name string, state redispb.Instance_State)
    SetMemoryStoreRedisError(name string, errorMsg string)
    
    // Delete methods - remove mock resources
    DeleteMemorStoreRedisByName(name string)
}
```

### Step 2: Implement Mock Struct

#### ❌ WRONG: Not Thread-Safe

```go
// NEVER: No mutex for concurrent access
type memoryStoreClientFake struct {
    redisInstances map[string]*redispb.Instance  // WRONG: Race conditions!
}
```

#### ✅ CORRECT: Thread-Safe with Mutex

```go
// ALWAYS: Use sync.Mutex for thread safety
type memoryStoreClientFake struct {
    mutex          sync.Mutex                        // REQUIRED
    redisInstances map[string]*redispb.Instance      // In-memory storage
}
```

### Step 3: Implement Utils Methods

#### ❌ WRONG: No Locking

```go
// NEVER: Accessing map without mutex
func (m *memoryStoreClientFake) GetMemoryStoreRedisByName(name string) *redispb.Instance {
    return m.redisInstances[name]  // WRONG: Race condition!
}
```

#### ✅ CORRECT: Locked Utils Methods

```go
// ALWAYS: Lock with defer pattern
func (m *memoryStoreClientFake) GetMemoryStoreRedisByName(name string) *redispb.Instance {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    return m.redisInstances[name]
}

func (m *memoryStoreClientFake) SetMemoryStoreRedisLifeCycleState(
    name string,
    state redispb.Instance_State,
) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    if instance, ok := m.redisInstances[name]; ok {
        instance.State = state
    }
}

func (m *memoryStoreClientFake) DeleteMemorStoreRedisByName(name string) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    delete(m.redisInstances, name)
}
```

### Step 4: Implement Client Interface - Create

#### ❌ WRONG: Immediate READY State

```go
// NEVER: Skipping state transitions
func (m *memoryStoreClientFake) CreateRedisInstance(...) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    redisInstance := &redispb.Instance{
        State: redispb.Instance_READY,  // WRONG: Should be CREATING!
    }
    m.redisInstances[name] = redisInstance
    return nil
}
```

#### ✅ CORRECT: Realistic State Transitions

```go
// ALWAYS: Model async behavior with CREATING state
func (m *memoryStoreClientFake) CreateRedisInstance(
    ctx context.Context,
    projectId string,
    locationId string,
    instanceId string,
    options gcpredisinstanceclient.CreateRedisInstanceOptions,
) error {
    // MUST: Check context cancellation first
    if isContextCanceled(ctx) {
        return context.Canceled
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    // MUST: Use real client's name construction
    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(
        projectId,
        locationId,
        instanceId,
    )
    
    // MUST: Create with CREATING state
    redisInstance := &redispb.Instance{
        Name:              name,
        State:             redispb.Instance_CREATING,  // Initial state
        Host:              "192.168.0.1",              // Realistic endpoint
        Port:              6379,
        ReadEndpoint:      "192.168.0.2",
        ReadEndpointPort:  6379,
        MemorySizeGb:      options.MemorySizeGb,
        RedisConfigs:      options.RedisConfigs,
        MaintenancePolicy: options.MaintenancePolicy,
        AuthEnabled:       options.AuthEnabled,
        RedisVersion:      options.RedisVersion,
    }
    
    m.redisInstances[name] = redisInstance
    return nil
}

// Helper function - add to file
func isContextCanceled(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return true
    default:
        return false
    }
}
```

### Step 5: Implement Client Interface - Get

#### ❌ WRONG: Not Returning 404

```go
// NEVER: Returning nil without error
func (m *memoryStoreClientFake) GetRedisInstance(...) (*redispb.Instance, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    instance, ok := m.redisInstances[name]
    if !ok {
        return nil, nil  // WRONG: Should return 404 error!
    }
    return instance, nil
}
```

#### ✅ CORRECT: Return 404 for Not Found

```go
// ALWAYS: Match real API error behavior
func (m *memoryStoreClientFake) GetRedisInstance(
    ctx context.Context,
    projectId string,
    locationId string,
    instanceId string,
) (*redispb.Instance, *redispb.InstanceAuthString, error) {
    if isContextCanceled(ctx) {
        return nil, nil, context.Canceled
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(
        projectId,
        locationId,
        instanceId,
    )

    // MUST: Return 404 error if not found
    instance, ok := m.redisInstances[name]
    if !ok {
        return nil, nil, &googleapi.Error{
            Code:    404,
            Message: "Not Found",
        }
    }

    // Return mock auth string
    return instance, &redispb.InstanceAuthString{
        AuthString: "mock-auth-token-12345",
    }, nil
}
```

### Step 6: Implement Client Interface - Update

#### ❌ WRONG: Immediate Update Without State

```go
// NEVER: Updating directly without UPDATING state
func (m *memoryStoreClientFake) UpdateRedisInstance(...) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    if instance, ok := m.redisInstances[name]; ok {
        instance.MemorySizeGb = newSize  // WRONG: No UPDATING state!
    }
    return nil
}
```

#### ✅ CORRECT: Set UPDATING State

```go
// ALWAYS: Model update as async operation
func (m *memoryStoreClientFake) UpdateRedisInstance(
    ctx context.Context,
    redisInstance *redispb.Instance,
    updateMask []string,
) error {
    if isContextCanceled(ctx) {
        return context.Canceled
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    if instance, ok := m.redisInstances[redisInstance.Name]; ok {
        // MUST: Set UPDATING state
        instance.State = redispb.Instance_UPDATING
        
        // Apply updates based on updateMask
        for _, field := range updateMask {
            switch field {
            case "memory_size_gb":
                instance.MemorySizeGb = redisInstance.MemorySizeGb
            case "redis_configs":
                instance.RedisConfigs = redisInstance.RedisConfigs
            case "maintenance_policy":
                instance.MaintenancePolicy = redisInstance.MaintenancePolicy
            case "auth_enabled":
                instance.AuthEnabled = redisInstance.AuthEnabled
            }
        }
    }

    return nil
}
```

### Step 7: Implement Client Interface - Delete

#### ❌ WRONG: Immediate Deletion

```go
// NEVER: Removing resource immediately
func (m *memoryStoreClientFake) DeleteRedisInstance(...) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    delete(m.redisInstances, name)  // WRONG: No DELETING state!
    return nil
}
```

#### ✅ CORRECT: Set DELETING State

```go
// ALWAYS: Set DELETING state, let test remove explicitly
func (m *memoryStoreClientFake) DeleteRedisInstance(
    ctx context.Context,
    projectId string,
    locationId string,
    instanceId string,
) error {
    if isContextCanceled(ctx) {
        return context.Canceled
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(
        projectId,
        locationId,
        instanceId,
    )

    // MUST: Set DELETING state instead of removing
    if instance, ok := m.redisInstances[name]; ok {
        instance.State = redispb.Instance_DELETING
        return nil
    }

    // Return 404 if not found
    return &googleapi.Error{
        Code:    404,
        Message: "Not Found",
    }
}
```

### Step 8: Add to Server Interface

**Location**: `type.go`

#### ❌ WRONG: Not Adding Utils Interface

```go
// NEVER: Missing Utils interface in Server
type Server interface {
    Clients
    Providers
    // WRONG: No MemoryStoreClientFakeUtils!
}
```

#### ✅ CORRECT: Include Utils Interface

```go
// ALWAYS: Add Utils interface to Server
type Server interface {
    Clients
    Providers
    ClientErrors
    
    MemoryStoreClientFakeUtils  // Add your Utils interface
    ComputeClientFakeUtils
    RegionalOperationsClientFakeUtils
    // ... other utils
}
```

### Step 9: Integrate into Server

**Location**: `server.go`

#### ❌ WRONG: Not Initializing or Embedding

```go
// NEVER: Not creating mock in New()
func New() Server {
    return &server{}  // WRONG: Mock not initialized!
}

type server struct {
    // WRONG: Mock not embedded!
}
```

#### ✅ CORRECT: Initialize and Embed

```go
// ALWAYS: Initialize mock in New() and embed in server
func New() Server {
    regionalOperationsClientfake := &regionalOperationsClientFake{
        mutex:      sync.Mutex{},
        operations: map[string]*computepb.Operation{},
    }
    
    return &server{
        memoryStoreClientFake: &memoryStoreClientFake{
            mutex:          sync.Mutex{},
            redisInstances: map[string]*redispb.Instance{},
        },
        regionalOperationsClientFake: regionalOperationsClientfake,
        // ... other mocks
    }
}

// MUST: Embed mock struct
type server struct {
    *memoryStoreClientFake              // Embeds Client + Utils interfaces
    *regionalOperationsClientFake
    // ... other mocks
}

// MUST: Add provider method
func (s *server) MemoryStoreProviderFake() client.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient] {
    return func() gcpredisinstanceclient.MemorystoreClient {
        return s  // Server embeds memoryStoreClientFake
    }
}
```

## Mocking Async Operations

### Operations Mock Pattern

Many cloud APIs return operation IDs for long-running tasks. Create separate operations mock:

#### ✅ Operations Mock Implementation

**Location**: `regionalOperationsClientFake.go`

```go
package mock

import (
    "context"
    "sync"
    "cloud.google.com/go/compute/apiv1/computepb"
    "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
    "google.golang.org/api/googleapi"
)

// Utils interface for operations
type RegionalOperationsClientFakeUtils interface {
    AddRegionOperation(name string) string
    GetRegionOperationById(operationId string) *computepb.Operation
    SetRegionOperationDone(operationId string)
    SetRegionOperationError(operationId string, errorMsg string)
}

type regionalOperationsClientFake struct {
    mutex      sync.Mutex
    operations map[string]*computepb.Operation
}

// Utils method - test creates operation
func (c *regionalOperationsClientFake) AddRegionOperation(name string) string {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    key := name
    c.operations[key] = &computepb.Operation{
        Status: computepb.Operation_PENDING.Enum(),
        Name:   &key,
    }
    return name
}

// Client method - reconciler checks operation
func (c *regionalOperationsClientFake) GetRegionOperation(
    ctx context.Context,
    request client.GetRegionOperationRequest,
) (*computepb.Operation, error) {
    if isContextCanceled(ctx) {
        return nil, context.Canceled
    }

    c.mutex.Lock()
    defer c.mutex.Unlock()

    op, ok := c.operations[request.Name]
    if !ok {
        return nil, &googleapi.Error{Code: 404, Message: "Not Found"}
    }

    return op, nil
}

// Utils method - test completes operation
func (c *regionalOperationsClientFake) SetRegionOperationDone(operationId string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if op := c.operations[operationId]; op != nil {
        op.Status = computepb.Operation_DONE.Enum()
    }
}

// Utils method - test sets operation error
func (c *regionalOperationsClientFake) SetRegionOperationError(
    operationId string,
    errorMsg string,
) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if op := c.operations[operationId]; op != nil {
        op.Status = computepb.Operation_DONE.Enum()
        op.Error = &computepb.Error{
            Errors: []*computepb.Errors{
                {Message: &errorMsg},
            },
        }
    }
}
```

### Using Operations Mock in Resource Mock

```go
type computeClientFake struct {
    mutex   sync.Mutex
    subnets map[string]*computepb.Subnetwork

    // Reference to operations mock
    operationsClientUtils RegionalOperationsClientFakeUtils
}

func (c *computeClientFake) CreateSubnet(
    ctx context.Context,
    request client.CreateSubnetRequest,
) (string, error) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    name := subnet.GetSubnetFullName(request.ProjectId, request.Region, request.Name)

    // Create subnet
    c.subnets[name] = &computepb.Subnetwork{
        Name:        &name,
        Region:      &request.Region,
        IpCidrRange: &request.Cidr,
    }

    // Create async operation via operations mock
    opKey := c.operationsClientUtils.AddRegionOperation(request.Name)
    return opKey, nil
}
```

## Using Mocks in Tests

### Basic Pattern

```go
var _ = Describe("Feature: KCP RedisInstance", func() {
    It("Scenario: Creates and deletes Redis instance", func() {
        redisInstance := &cloudcontrolv1beta1.RedisInstance{}

        By("When RedisInstance is created", func() {
            Eventually(CreateRedisInstance).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
                Should(Succeed())
        })

        var gcpRedis *redispb.Instance
        
        By("Then GCP Redis is created with CREATING state", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
                    NewObjActions(),
                    HavingRedisInstanceStatusId()).
                Should(Succeed())
            
            // MUST: Use Utils interface to inspect mock
            gcpRedis = infra.GcpMock().GetMemoryStoreRedisByName(
                redisInstance.Status.Id,
            )
            Expect(gcpRedis).NotTo(BeNil())
            Expect(gcpRedis.State).To(Equal(redispb.Instance_CREATING))
        })

        By("When GCP Redis becomes available", func() {
            // MUST: Use Utils interface to transition state
            infra.GcpMock().SetMemoryStoreRedisLifeCycleState(
                gcpRedis.Name,
                redispb.Instance_READY,
            )
        })

        By("Then RedisInstance has Ready condition", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
                    NewObjActions(),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).
                Should(Succeed())
        })

        By("When RedisInstance is deleted", func() {
            Eventually(Delete).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
                Should(Succeed())
        })

        By("Then GCP Redis enters DELETING state", func() {
            Eventually(func() bool {
                redis := infra.GcpMock().GetMemoryStoreRedisByName(gcpRedis.Name)
                return redis != nil && redis.State == redispb.Instance_DELETING
            }, "5s").Should(BeTrue())
        })

        By("When GCP Redis deletion completes", func() {
            // MUST: Use Utils interface to remove resource
            infra.GcpMock().DeleteMemorStoreRedisByName(gcpRedis.Name)
        })

        By("Then RedisInstance is removed", func() {
            Eventually(IsDeleted).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
                Should(Succeed())
        })
    })
})
```

### Testing Async Operations

```go
It("Scenario: Handles async operations", func() {
    subnet := &cloudcontrolv1beta1.GcpSubnet{}
    
    By("When GcpSubnet is created", func() {
        Eventually(CreateGcpSubnet).
            WithArguments(infra.Ctx(), infra.KCP().Client(), subnet).
            Should(Succeed())
    })
    
    var operationId string
    
    By("Then creation operation is initiated", func() {
        Eventually(LoadAndCheck).
            WithArguments(infra.Ctx(), infra.KCP().Client(), subnet,
                NewObjActions(),
                func(obj *cloudcontrolv1beta1.GcpSubnet) {
                    operationId = obj.Status.OperationId
                    Expect(operationId).NotTo(BeEmpty())
                },
            ).
            Should(Succeed())
    })
    
    By("When GCP operation completes", func() {
        // Use operations Utils interface
        infra.GcpMock().SetRegionOperationDone(operationId)
    })
    
    By("Then GcpSubnet becomes ready", func() {
        Eventually(LoadAndCheck).
            WithArguments(infra.Ctx(), infra.KCP().Client(), subnet,
                NewObjActions(),
                HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
            ).
            Should(Succeed())
    })
})
```

## Common Pitfalls

### Pitfall 1: Missing Mutex Locks

**Symptom**: Random test failures, race detector errors

**Cause**:
```go
// WRONG: No mutex lock
func (m *mock) GetResource(name string) *Resource {
    return m.resources[name]  // Race condition!
}
```

**Solution**: ALWAYS lock:
```go
// CORRECT: Locked access
func (m *mock) GetResource(name string) *Resource {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    return m.resources[name]
}
```

### Pitfall 2: Immediate Deletion

**Symptom**: Deletion tests fail, reconciler doesn't see DELETING state

**Cause**:
```go
// WRONG: Immediate removal
func (m *mock) DeleteResource(...) error {
    delete(m.resources, name)  // Gone immediately!
    return nil
}
```

**Solution**: Set DELETING state first:
```go
// CORRECT: Set DELETING, let test remove
func (m *mock) DeleteResource(...) error {
    if resource, ok := m.resources[name]; ok {
        resource.State = STATE_DELETING  // Test removes later
    }
    return nil
}
```

### Pitfall 3: Mock Not Registered

**Symptom**: Reconciler fails with nil client

**Cause**: Mock provider not passed to reconciler setup

**Solution**: Verify provider registration in test suite:
```go
infra.StartReconcilers(
    SetupGcpRedisClusterReconciler,
)

// In reconciler setup:
func SetupGcpRedisClusterReconciler(ctx context.Context, mgr manager.Manager) error {
    return (&reconciler{
        client: infra.GcpMock().MemoryStoreProviderFake()(),  // Use mock
    }).SetupWithManager(mgr)
}
```

### Pitfall 4: Wrong Error Types

**Symptom**: Reconciler doesn't handle "not found" correctly

**Cause**:
```go
// WRONG: Generic error
if !found {
    return nil, errors.New("not found")
}
```

**Solution**: Use googleapi.Error:
```go
// CORRECT: Cloud provider error type
if !found {
    return nil, &googleapi.Error{Code: 404, Message: "Not Found"}
}
```

## Validation Checklist

### Mock Structure
- [ ] Utils interface defined with Get/Set/Delete methods
- [ ] Mock struct has `sync.Mutex`
- [ ] Mock struct has map storage (string keys)
- [ ] Embedded in `server` struct
- [ ] Initialized in `New()` function

### Client Interface Implementation
- [ ] All CRUD methods implemented (Create, Get, Update, Delete)
- [ ] Context cancellation checked in every method
- [ ] All map operations use mutex lock/unlock
- [ ] Returns 404 error for not found resources
- [ ] Uses realistic state transitions (CREATING, UPDATING, DELETING)
- [ ] Uses name construction helpers from real client

### Utils Interface Implementation
- [ ] Get methods return resources by name
- [ ] Set methods modify resource state
- [ ] Delete methods remove resources from map
- [ ] All methods are thread-safe (mutex locked)

### Server Integration
- [ ] Utils interface added to `Server` in type.go
- [ ] Mock embedded in `server` struct in server.go
- [ ] Provider method returns server (which embeds mock)
- [ ] Mock initialized with empty maps in `New()`

### Mock Behavior
- [ ] Create sets CREATING state
- [ ] Delete sets DELETING state (doesn't remove)
- [ ] Get returns 404 for non-existent resources
- [ ] Update sets UPDATING state
- [ ] Realistic mock data (IPs, ports, endpoints)

### Testing Integration
- [ ] Tests use Utils interface to inspect state
- [ ] Tests use Utils interface to transition states
- [ ] Tests use Utils interface to explicitly remove resources
- [ ] Tests verify Client interface called by reconciler

## Summary: Key Rules

1. **Dual Interface**: Client (reconciler) + Utils (test) REQUIRED
2. **Thread Safety**: All map operations MUST use `sync.Mutex`
3. **Context Checks**: EVERY method checks `isContextCanceled()` first
4. **Realistic Errors**: Use `&googleapi.Error{Code: 404}` for not found
5. **State Transitions**: CREATING → READY → UPDATING → DELETING (model async)
6. **No Immediate Delete**: Set DELETING state, let test explicitly remove
7. **Embed in Server**: All mocks embed in `server`, return server from providers
8. **Realistic Data**: Mock endpoints like `192.168.0.1:6379`, not `localhost`

## Next Steps

- [Write Controller Tests](CONTROLLER_TESTS.md)
- [Add API Validation Tests](API_VALIDATION_TESTS.md)
- [Configure Feature Flags](FEATURE_FLAGS.md)
