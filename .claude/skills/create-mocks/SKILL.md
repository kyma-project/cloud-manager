---
name: create-mocks
description: Create mock implementations for cloud provider APIs. Use when building test infrastructure, implementing fake GCP/Azure/AWS clients, or setting up test doubles.
---

# Create Mocks

Implement mock cloud provider clients for testing reconcilers.

## Quick Start

1. Create mock in `pkg/kcp/provider/<provider>/mock/`
2. Implement Client interface (what reconcilers call)
3. Implement Utils interface (what tests call)
4. Add realistic state transitions
5. Wire up in testinfra

## Rules

### MUST
- Implement dual interfaces: Client (for reconcilers) + Utils (for tests)
- Model realistic async state transitions
- Track operations (create/update/delete)
- Use thread-safe state management
- Implement NotFound errors correctly

### MUST NOT
- Make mocks synchronous when real API is async
- Skip state transitions (CREATING → READY)
- Allow race conditions in state access
- Return success without storing state

## Mock Structure

```
pkg/kcp/provider/gcp/mock/
├── server.go          # Mock server setup
├── redisCluster.go    # Resource-specific mock
└── util.go            # Shared utilities
```

## Dual Interface Pattern

```go
// Client interface - what reconcilers call
type RedisClusterClient interface {
    CreateCluster(ctx context.Context, req CreateRequest) (string, error)
    GetCluster(ctx context.Context, id string) (*Cluster, error)
    UpdateCluster(ctx context.Context, cluster *Cluster) (string, error)
    DeleteCluster(ctx context.Context, id string) error
}

// Utils interface - what tests call
type RedisClusterUtils interface {
    GetClusterByName(name string) *Cluster
    SetClusterState(id string, state string)
    SetError(id string, err error)
    DeleteClusterDirect(id string)
}
```

## Mock Implementation Template

```go
type redisClusterMock struct {
    mu       sync.RWMutex
    clusters map[string]*Cluster
    errors   map[string]error
}

func NewRedisClusterMock() *redisClusterMock {
    return &redisClusterMock{
        clusters: make(map[string]*Cluster),
        errors:   make(map[string]error),
    }
}

// Client interface implementation

func (m *redisClusterMock) CreateCluster(ctx context.Context, req CreateRequest) (string, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    id := fmt.Sprintf("projects/%s/locations/%s/clusters/%s",
        req.Project, req.Location, req.ClusterId)

    if m.errors[id] != nil {
        return "", m.errors[id]
    }

    m.clusters[id] = &Cluster{
        Name:   id,
        State:  "CREATING",  // Start in CREATING state
        Spec:   req,
    }

    return id, nil
}

func (m *redisClusterMock) GetCluster(ctx context.Context, id string) (*Cluster, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    if m.errors[id] != nil {
        return nil, m.errors[id]
    }

    cluster, ok := m.clusters[id]
    if !ok {
        return nil, &googleapi.Error{Code: 404, Message: "not found"}
    }

    return cluster, nil
}

func (m *redisClusterMock) DeleteCluster(ctx context.Context, id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.errors[id] != nil {
        return m.errors[id]
    }

    cluster, ok := m.clusters[id]
    if !ok {
        return &googleapi.Error{Code: 404, Message: "not found"}
    }

    cluster.State = "DELETING"
    return nil
}

// Utils interface implementation

func (m *redisClusterMock) GetClusterByName(name string) *Cluster {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.clusters[name]
}

func (m *redisClusterMock) SetClusterState(id string, state string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if cluster, ok := m.clusters[id]; ok {
        cluster.State = state
    }
}

func (m *redisClusterMock) SetError(id string, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.errors[id] = err
}

func (m *redisClusterMock) DeleteClusterDirect(id string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    delete(m.clusters, id)
}
```

## State Transitions

```go
// Realistic async state machine:
// Create: CREATING → READY
// Update: UPDATING → READY
// Delete: DELETING → (removed)

// In tests, trigger transitions:
By("When GCP marks cluster as ready", func() {
    infra.GcpMock().SetClusterState(id, "READY")
})

// Simulate async completion:
By("When delete operation completes", func() {
    infra.GcpMock().DeleteClusterDirect(id)
})
```

## Error Simulation

```go
// Set error for specific resource
infra.GcpMock().SetError(id, &googleapi.Error{
    Code:    503,
    Message: "Service unavailable",
})

// Clear error
infra.GcpMock().SetError(id, nil)

// NotFound error
return nil, &googleapi.Error{Code: 404, Message: "not found"}
```

## Wiring in TestInfra

```go
// In pkg/testinfra/gcpMock.go

type GcpMock struct {
    redisCluster *redisClusterMock
    // other mocks...
}

func NewGcpMock() *GcpMock {
    return &GcpMock{
        redisCluster: NewRedisClusterMock(),
    }
}

func (m *GcpMock) RedisCluster() RedisClusterClient {
    return m.redisCluster
}

func (m *GcpMock) RedisClusterUtils() RedisClusterUtils {
    return m.redisCluster
}
```

## Checklist

- [ ] Implements Client interface (for reconcilers)
- [ ] Implements Utils interface (for tests)
- [ ] Thread-safe with sync.RWMutex
- [ ] Realistic state transitions
- [ ] NotFound errors return correct type
- [ ] Error injection supported
- [ ] Wired into testinfra

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Race condition | Add mutex locking |
| Wrong error type | Use provider's error type |
| Test hangs | Set mock state to trigger transition |
| State not persisted | Check mutex unlock order |

## Related

- Full guide: [docs/agents/guides/CREATING_MOCKS.md](../../../docs/agents/guides/CREATING_MOCKS.md)
- Writing tests: `/write-tests`
