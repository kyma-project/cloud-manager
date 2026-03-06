---
name: gcp-client
description: Implement GCP cloud provider clients in Cloud Manager. Use when creating Google Cloud API clients, working with GCP resources, or implementing GCP-specific reconcilers.
---

# GCP Client Implementation

Create GCP cloud provider clients using the NEW pattern.

## Quick Start

1. Add client to `pkg/kcp/provider/gcp/client/gcpClients.go`
2. Create typed client interface in resource's `client/` directory
3. Implement client wrapping GCP SDK
4. Create provider function
5. Wire in `cmd/main.go`

## NEW Pattern (Required)

Centralized client initialization in `main.go`, injected via typed providers.

### GcpClients Structure

**File**: `pkg/kcp/provider/gcp/client/gcpClients.go`

```go
type GcpClients struct {
    Compute       *compute.Service
    Filestore     *file.CloudFilestoreManagerClient
    RedisCluster  *redis.CloudRedisClusterClient
    // Add new clients here
}

func NewGcpClients(ctx context.Context, saJsonKeyPath string) (*GcpClients, error) {
    creds, err := google.CredentialsFromJSON(ctx, []byte(saJsonKeyPath),
        compute.CloudPlatformScope,
    )
    if err != nil {
        return nil, err
    }

    computeService, err := compute.NewService(ctx, option.WithCredentials(creds))
    if err != nil {
        return nil, err
    }

    // Initialize other clients...

    return &GcpClients{
        Compute:      computeService,
        RedisCluster: redisClusterClient,
    }, nil
}
```

### Typed Client Interface

**File**: `pkg/kcp/provider/gcp/<resource>/client/client.go`

```go
package client

import (
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// Business operations interface
type RedisClusterClient interface {
    CreateCluster(ctx context.Context, req CreateClusterRequest) (string, error)
    GetCluster(ctx context.Context, project, location, id string) (*Cluster, error)
    UpdateCluster(ctx context.Context, cluster *Cluster, mask []string) (string, error)
    DeleteCluster(ctx context.Context, project, location, id string) error
}

// Provider function type
type GcpClientProvider[T any] func() T

// Create provider
func NewRedisClusterClientProvider(gcpClients *gcpclient.GcpClients) GcpClientProvider[RedisClusterClient] {
    return func() RedisClusterClient {
        return NewRedisClusterClient(gcpClients)
    }
}

// Implementation
type redisClusterClient struct {
    client *redis.CloudRedisClusterClient
}

func NewRedisClusterClient(gcpClients *gcpclient.GcpClients) RedisClusterClient {
    return &redisClusterClient{
        client: gcpClients.RedisCluster,
    }
}

func (c *redisClusterClient) CreateCluster(ctx context.Context, req CreateClusterRequest) (string, error) {
    parent := fmt.Sprintf("projects/%s/locations/%s", req.Project, req.Location)

    op, err := c.client.CreateCluster(ctx, &clusterpb.CreateClusterRequest{
        Parent:    parent,
        ClusterId: req.ClusterId,
        Cluster: &clusterpb.Cluster{
            // Map fields
        },
    })
    if err != nil {
        return "", err
    }

    return op.Name(), nil
}

func (c *redisClusterClient) GetCluster(ctx context.Context, project, location, id string) (*Cluster, error) {
    name := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, id)

    cluster, err := c.client.GetCluster(ctx, &clusterpb.GetClusterRequest{Name: name})
    if err != nil {
        return nil, err
    }

    return &Cluster{
        Name:   cluster.Name,
        State:  cluster.State.String(),
        // Map other fields
    }, nil
}
```

### Wiring in main.go

```go
// Initialize GCP clients
gcpClients, err := gcpclient.NewGcpClients(ctx, gcpSaJsonKeyPath)
if err != nil {
    setupLog.Error(err, "failed to create GCP clients")
    os.Exit(1)
}

// Create provider for specific resource
redisClusterClientProvider := gcpredisclusterclient.NewRedisClusterClientProvider(gcpClients)

// Create state factory
redisClusterStateFactory := gcprediscluster.NewStateFactory(redisClusterClientProvider)
```

## Error Handling

```go
import "google.golang.org/api/googleapi"

func IsNotFound(err error) bool {
    var gerr *googleapi.Error
    if errors.As(err, &gerr) {
        return gerr.Code == 404
    }
    return false
}

// Usage
resource, err := client.Get(ctx, id)
if err != nil {
    if IsNotFound(err) {
        return nil, ctx  // OK - doesn't exist
    }
    return err, ctx  // Real error
}
```

## Async Operations

```go
// Create returns operation ID
opId, err := client.CreateCluster(ctx, req)
if err != nil {
    return err, ctx
}

// Store for tracking
obj.Status.OpIdentifier = opId
state.UpdateObjStatus(ctx)

// Wait action checks operation
func waitClusterReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    if state.ObjAsGcpCluster().Status.OpIdentifier == "" {
        return nil, ctx
    }

    op, err := state.client.GetOperation(ctx, state.ObjAsGcpCluster().Status.OpIdentifier)
    if err != nil {
        return err, ctx
    }

    if !op.Done {
        return composed.StopWithRequeueDelay(10 * time.Second), nil
    }

    // Clear operation, reload resource
    state.ObjAsGcpCluster().Status.OpIdentifier = ""
    state.UpdateObjStatus(ctx)

    return loadCluster(ctx, st)
}
```

## Adding New GCP Client

1. Add to `GcpClients` struct:
```go
type GcpClients struct {
    // existing...
    NewService *newservice.Client
}
```

2. Initialize in `NewGcpClients`:
```go
newServiceClient, err := newservice.NewClient(ctx, option.WithCredentials(creds))
if err != nil {
    return nil, err
}
```

3. Create typed interface in resource's client package
4. Create provider function
5. Wire in main.go

## Checklist

- [ ] Client added to GcpClients struct
- [ ] Client initialized in NewGcpClients
- [ ] Typed interface defined
- [ ] Provider function created
- [ ] Implementation wraps GCP SDK
- [ ] Error handling with IsNotFound
- [ ] Async operations tracked
- [ ] Wired in main.go

## Related

- Full guide: [docs/agents/architecture/GCP_CLIENT_NEW_PATTERN.md](../../../docs/agents/architecture/GCP_CLIENT_NEW_PATTERN.md)
- Hybrid pattern: [docs/agents/architecture/GCP_CLIENT_HYBRID.md](../../../docs/agents/architecture/GCP_CLIENT_HYBRID.md)
- Add reconciler: `/add-kcp-reconciler`
