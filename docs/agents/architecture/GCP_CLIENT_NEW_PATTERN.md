# GCP Client NEW Pattern (GcpClients)

**Authority**: Foundational architecture (GCP-specific)  
**Prerequisite For**: All new GCP KCP reconciler work  
**Must Read Before**: Creating any new GCP KCP reconciler

**Prerequisites**:
- MUST understand: [NEW Reconciler Pattern](RECONCILER_NEW_PATTERN.md)
- MUST have read: [State Pattern](STATE_PATTERN.md)

**Skip This File If**:
- You are working on Azure/AWS resources (not GCP)
- You are maintaining legacy resources (see [OLD Pattern](GCP_CLIENT_OLD_PATTERN.md))

## Pattern Status

**Status**: ✅ **REQUIRED FOR ALL NEW GCP RESOURCES**  
**Location**: `pkg/kcp/provider/gcp/client/gcpClients.go`  
**Examples**: GcpSubnet, GcpRedisCluster, GcpNfsVolume

## Rules: GCP Client NEW Pattern

### MUST

1. MUST use for ALL new GCP resources
2. MUST add client to `GcpClients` struct
3. MUST initialize client in `NewGcpClients()`
4. MUST create typed client interface in resource package
5. MUST use `GcpClientProvider[T]` for injection
6. MUST use Cloud Client Libraries (`cloud.google.com/go/*`)

### MUST NOT

1. MUST NOT create clients on-demand (create once at startup)
2. MUST NOT use legacy REST APIs (`google.golang.org/api/*`) when modern library exists
3. MUST NOT share HTTP client across services (each service has token provider)
4. MUST NOT ignore client creation errors at startup

### ALWAYS

1. ALWAYS use gRPC-based Cloud Client Libraries
2. ALWAYS fail fast at startup if client creation fails
3. ALWAYS use token provider with appropriate scopes per service

## Pattern Characteristics

| Aspect | NEW Pattern |
|--------|-------------|
| Client Creation | Once at startup |
| HTTP Client | Token-based per service |
| API Style | Cloud Client Libraries (gRPC) |
| Package | `cloud.google.com/go/*` |
| Performance | Excellent (pre-initialized) |
| Error Handling | Fail fast at startup |
| Status | Required for new code |

## GcpClients Structure

**Location**: `pkg/kcp/provider/gcp/client/gcpClients.go`

```go
type GcpClients struct {
    // Compute service clients
    ComputeNetworks      *compute.NetworksClient
    ComputeAddresses     *compute.AddressesClient
    ComputeSubnetworks   *compute.SubnetworksClient
    ComputeRouters       *compute.RoutersClient
    RegionOperations     *compute.RegionOperationsClient
    GlobalOperations     *compute.GlobalOperationsClient
    
    // Networking clients
    NetworkConnectivityCrossNetworkAutomation *networkconnectivity.CrossNetworkAutomationClient
    
    // Redis clients
    RedisCluster   *rediscluster.CloudRedisClusterClient
    RedisInstance  *redisinstance.CloudRedisClient
    
    // VPC Peering (special case)
    VpcPeeringClients *VpcPeeringClients
}

func NewGcpClients(ctx context.Context, credentialsFile, peeringCredentialsFile string, logger logr.Logger) (*GcpClients, error) {
    // Create token providers with scopes
    // Initialize all clients
    // Return ready-to-use clients
}
```

## Adding New GCP Client (Step-by-Step)

### Step 1: Add to GcpClients Struct

```go
type GcpClients struct {
    // ... existing clients
    MyNewService *mynewservice.MyServiceClient  // Add here
}
```

### Step 2: Initialize in NewGcpClients()

```go
func NewGcpClients(ctx context.Context, credentialsFile, peeringCredentialsFile string, logger logr.Logger) (*GcpClients, error) {
    // ... existing initialization
    
    // Create token provider with appropriate scopes
    myServiceTokenProvider, err := b.WithScopes(mynewservice.DefaultAuthScopes()).BuildTokenProvider()
    if err != nil {
        return nil, fmt.Errorf("failed to build myservice token provider: %w", err)
    }
    myServiceTokenSource := oauth2adapt.TokenSourceFromTokenProvider(myServiceTokenProvider)
    
    // Create client
    myServiceClient, err := mynewservice.NewMyServiceClient(ctx, option.WithTokenSource(myServiceTokenSource))
    if err != nil {
        return nil, fmt.Errorf("create myservice client: %w", err)
    }
    
    return &GcpClients{
        // ... existing clients
        MyNewService: myServiceClient,
    }, nil
}
```

### Step 3: Create Typed Client Interface

**Location**: `pkg/kcp/provider/gcp/<resource>/client/<service>Client.go`

```go
package client

import (
    "context"
    mynewservice "cloud.google.com/go/mynewservice/apiv1"
    "cloud.google.com/go/mynewservice/apiv1/mynewservicepb"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// Business operations interface
type MyServiceClient interface {
    CreateResource(ctx context.Context, req CreateRequest) (*mynewservicepb.Resource, error)
    GetResource(ctx context.Context, projectId, locationId, resourceId string) (*mynewservicepb.Resource, error)
    UpdateResource(ctx context.Context, resource *mynewservicepb.Resource, updateMask []string) error
    DeleteResource(ctx context.Context, projectId, locationId, resourceId string) error
}

// Provider function returns accessor
func NewMyServiceClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MyServiceClient] {
    return func() MyServiceClient {
        return NewMyServiceClient(gcpClients)
    }
}

// Implementation wrapping real GCP client
type myServiceClient struct {
    client *mynewservice.MyServiceClient
}

func NewMyServiceClient(gcpClients *gcpclient.GcpClients) MyServiceClient {
    return &myServiceClient{client: gcpClients.MyNewService}
}

func (c *myServiceClient) CreateResource(ctx context.Context, req CreateRequest) (*mynewservicepb.Resource, error) {
    grpcReq := &mynewservicepb.CreateResourceRequest{
        Parent:     fmt.Sprintf("projects/%s/locations/%s", req.ProjectId, req.LocationId),
        ResourceId: req.ResourceId,
        Resource:   req.ToProto(),
    }
    
    return c.client.CreateResource(ctx, grpcReq)
}

// ... implement other methods
```

### Step 4: Wire Up in main.go

```go
// cmd/main.go
gcpClients, err := gcpclient.NewGcpClients(ctx, config.GcpConfig.CredentialsFile, config.GcpConfig.PeeringCredentialsFile, rootLogger)
if err != nil {
    setupLog.Error(err, "unable to create GCP clients")
    os.Exit(1)  // Fail fast
}

// Create state factory with client provider
myResourceStateFactory := gcpmyresource.NewStateFactory(
    gcpmyresourceclient.NewMyServiceClientProvider(gcpClients),
    env,
)

// Register reconciler
if err = gcpmyresource.NewReconciler(
    composedStateFactory,
    focalStateFactory,
    myResourceStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "GcpMyResource")
    os.Exit(1)
}
```

### Step 5: Use in State Factory

```go
// pkg/kcp/provider/gcp/myresource/state.go
type State struct {
    focal.State
    myServiceClient client.MyServiceClient
    remoteResource  *mynewservicepb.Resource
}

type stateFactory struct {
    myServiceClientProvider gcpclient.GcpClientProvider[client.MyServiceClient]
    env abstractions.Environment
}

func NewStateFactory(
    myServiceClientProvider gcpclient.GcpClientProvider[client.MyServiceClient],
    env abstractions.Environment,
) StateFactory {
    return &stateFactory{
        myServiceClientProvider: myServiceClientProvider,
        env:                     env,
    }
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:           focalState,
        myServiceClient: f.myServiceClientProvider(),  // Lightweight call
    }, nil
}
```

## Real Examples

### GcpRedisCluster

**Client Provider** (`pkg/kcp/provider/gcp/rediscluster/client/memorystoreClusterClient.go`):
```go
func NewMemorystoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MemorystoreClusterClient] {
    return func() MemorystoreClusterClient {
        return NewMemorystoreClient(gcpClients)
    }
}

func NewMemorystoreClient(gcpClients *gcpclient.GcpClients) MemorystoreClusterClient {
    return &memorystoreClient{
        redisClusterClient: gcpClients.RedisCluster,
    }
}
```

### GcpSubnet

**Client Provider** (`pkg/kcp/provider/gcp/subnet/client/computeClient.go`):
```go
func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
    return func() ComputeClient {
        return NewComputeClient(gcpClients)
    }
}

func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
    return &computeClient{
        subnetworksClient: gcpClients.ComputeSubnetworks,
        networksClient:    gcpClients.ComputeNetworks,
    }
}
```

## Benefits Over OLD Pattern

| Aspect | NEW Pattern | OLD Pattern |
|--------|-------------|-------------|
| Performance | Clients created once | May create multiple times |
| Dependencies | Modern gRPC | Legacy REST |
| Error Handling | Fail fast at startup | Fail during reconciliation |
| Testing | Easy to mock GcpClients | Must mock function |
| Maintenance | All in one place | Scattered |
| Token Refresh | Automatic per service | Manual HTTP renewal |

## Common Pitfalls

### Pitfall 1: Using Legacy REST API

**Frequency**: Occasional  
**Impact**: Performance degradation, missing features  
**Detection**: Import `google.golang.org/api/*` instead of `cloud.google.com/go/*`

❌ **WRONG**:
```go
import "google.golang.org/api/compute/v1"  // Legacy REST
```

✅ **CORRECT**:
```go
import compute "cloud.google.com/go/compute/apiv1"  // Modern gRPC
```

**Why It Fails**: Legacy APIs slower, less maintained  
**How to Fix**: Use Cloud Client Libraries (`cloud.google.com/go/*`)  
**Prevention**: Always check for modern library first

### Pitfall 2: Creating Clients On-Demand

**Frequency**: Rare  
**Impact**: Slower reconciliation, fails during runtime  
**Detection**: Client creation in state factory or action

❌ **WRONG**:
```go
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    // Creating client on-demand (WRONG)
    client, err := mynewservice.NewClient(ctx, ...)
    if err != nil {
        return nil, err
    }
    // ...
}
```

✅ **CORRECT**:
```go
// Client already in GcpClients
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:           focalState,
        myServiceClient: f.myServiceClientProvider(),  // Uses pre-created client
    }, nil
}
```

**Why It Fails**: Client creation expensive, fails at wrong time  
**How to Fix**: Add client to GcpClients, create at startup  
**Prevention**: Always use GcpClients pattern

### Pitfall 3: Missing Token Scopes

**Frequency**: Occasional  
**Impact**: 403 Forbidden errors during API calls  
**Detection**: Permission errors in reconciliation logs

❌ **WRONG**:
```go
// Using wrong or missing scopes
tokenProvider, err := b.WithScopes("https://www.googleapis.com/auth/cloud-platform.read-only").BuildTokenProvider()
```

✅ **CORRECT**:
```go
// Use service's default scopes
tokenProvider, err := b.WithScopes(mynewservice.DefaultAuthScopes()).BuildTokenProvider()
```

**Why It Fails**: Insufficient permissions for API operations  
**How to Fix**: Use `DefaultAuthScopes()` from service package  
**Prevention**: Always use service-provided default scopes

## Summary Checklist

Before adding new GCP client:
- [ ] Verify modern Cloud Client Library exists (`cloud.google.com/go/*`)
- [ ] Understand the API operations needed
- [ ] Check reference implementations (GcpSubnet, GcpRedisCluster)

When adding client:
- [ ] Add to `GcpClients` struct
- [ ] Initialize in `NewGcpClients()` with token provider
- [ ] Create typed interface in resource package
- [ ] Create provider function
- [ ] Wire up in main.go
- [ ] Use in state factory via provider

After adding client:
- [ ] Test startup (client creation should succeed)
- [ ] Verify API calls work
- [ ] Add mock for testing
- [ ] Document any special considerations

## Related Documentation

**MUST READ NEXT**:
- [NEW Reconciler Pattern](RECONCILER_NEW_PATTERN.md) - How to use clients in reconcilers
- [Add KCP Reconciler Guide](../guides/ADD_KCP_RECONCILER.md) - Complete workflow

**REFERENCE**:
- [OLD Pattern](GCP_CLIENT_OLD_PATTERN.md) - Legacy pattern (avoid)
- [Hybrid Pattern](GCP_CLIENT_HYBRID.md) - When mixing required
