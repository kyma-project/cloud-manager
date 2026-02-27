---
name: azure-client
description: Implement Azure cloud provider clients in Cloud Manager. Use when creating Azure API clients, working with Azure resources, or implementing Azure-specific reconcilers.
---

# Azure Client Implementation

Create Azure cloud provider clients for Cloud Manager.

## Quick Start

1. Create client interface in `pkg/kcp/provider/azure/<resource>/client/`
2. Implement using Azure SDK for Go
3. Create provider function
4. Wire in `cmd/main.go`

## Client Interface Pattern

**File**: `pkg/kcp/provider/azure/<resource>/client/client.go`

```go
package client

import (
    "context"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
)

// Business operations interface
type RedisClient interface {
    CreateRedis(ctx context.Context, req CreateRedisRequest) error
    GetRedis(ctx context.Context, resourceGroup, name string) (*Redis, error)
    UpdateRedis(ctx context.Context, resourceGroup, name string, update UpdateRedisRequest) error
    DeleteRedis(ctx context.Context, resourceGroup, name string) error
}

type CreateRedisRequest struct {
    ResourceGroup string
    Name          string
    Location      string
    Sku           string
    Capacity      int32
}

type Redis struct {
    Name              string
    ProvisioningState string
    HostName          string
    Port              int32
}
```

## Implementation

```go
type redisClient struct {
    client *armredis.Client
}

func NewRedisClient(subscriptionId string, cred azcore.TokenCredential) (RedisClient, error) {
    client, err := armredis.NewClient(subscriptionId, cred, nil)
    if err != nil {
        return nil, err
    }
    return &redisClient{client: client}, nil
}

func (c *redisClient) CreateRedis(ctx context.Context, req CreateRedisRequest) error {
    poller, err := c.client.BeginCreate(ctx, req.ResourceGroup, req.Name,
        armredis.CreateParameters{
            Location: to.Ptr(req.Location),
            Properties: &armredis.CreateProperties{
                Sku: &armredis.Sku{
                    Name:     to.Ptr(armredis.SkuName(req.Sku)),
                    Capacity: to.Ptr(req.Capacity),
                },
            },
        }, nil)
    if err != nil {
        return err
    }

    // Don't wait - return immediately, track via provisioning state
    return nil
}

func (c *redisClient) GetRedis(ctx context.Context, resourceGroup, name string) (*Redis, error) {
    resp, err := c.client.Get(ctx, resourceGroup, name, nil)
    if err != nil {
        return nil, err
    }

    return &Redis{
        Name:              *resp.Name,
        ProvisioningState: string(*resp.Properties.ProvisioningState),
        HostName:          *resp.Properties.HostName,
        Port:              *resp.Properties.Port,
    }, nil
}

func (c *redisClient) DeleteRedis(ctx context.Context, resourceGroup, name string) error {
    poller, err := c.client.BeginDelete(ctx, resourceGroup, name, nil)
    if err != nil {
        return err
    }
    // Don't wait
    return nil
}
```

## Error Handling

```go
import "github.com/Azure/azure-sdk-for-go/sdk/azcore"

func IsNotFound(err error) bool {
    var respErr *azcore.ResponseError
    if errors.As(err, &respErr) {
        return respErr.StatusCode == 404
    }
    return false
}

// Usage
resource, err := client.GetRedis(ctx, rg, name)
if err != nil {
    if IsNotFound(err) {
        return nil, ctx
    }
    return err, ctx
}
```

## Provider Pattern

```go
// Provider function
func NewRedisClientProvider(subscriptionId string, cred azcore.TokenCredential) func() RedisClient {
    return func() RedisClient {
        client, _ := NewRedisClient(subscriptionId, cred)
        return client
    }
}

// In state factory
type stateFactory struct {
    redisClientProvider func() RedisClient
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:       focalState,
        redisClient: f.redisClientProvider(),
    }, nil
}
```

## Async Operations

Azure operations return pollers. Don't wait for completion - check provisioning state:

```go
func waitRedisReady(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)

    redis, err := state.redisClient.GetRedis(ctx, state.resourceGroup(), state.name())
    if err != nil {
        return err, ctx
    }

    switch redis.ProvisioningState {
    case "Succeeded":
        state.redis = redis
        return nil, ctx
    case "Failed":
        return fmt.Errorf("provisioning failed"), ctx
    default:
        // Still provisioning
        return composed.StopWithRequeueDelay(30 * time.Second), nil
    }
}
```

## Checklist

- [ ] Client interface defined
- [ ] Implementation uses Azure SDK
- [ ] Error handling with IsNotFound
- [ ] Provider function created
- [ ] Async operations check provisioning state
- [ ] Wired in main.go

## Related

- Add reconciler: `/add-kcp-reconciler`
- Azure SDK: https://github.com/Azure/azure-sdk-for-go
