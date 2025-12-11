# Cloud Manager - AI Agent Guide

## Project Overview

Cloud Manager is a Kubernetes controller manager built with **Kubebuilder** that manages cloud provider resources for SAP BTP Kyma runtime. It acts as a bridge between Kyma clusters (SKR - SAP Kyma Runtime) and cloud provider APIs (AWS, Azure, GCP), enabling users to provision and manage cloud resources like NFS volumes, Redis instances, and VPC peering through Kubernetes Custom Resources.

**Core Concept**: Cloud Manager operates with two reconciliation loops:
- **SKR Loop**: Reconciles user-facing `cloud-resources.kyma-project.io` CRDs in remote Kyma clusters
- **KCP Loop**: Reconciles low-level `cloud-control.kyma-project.io` CRDs into actual cloud provider resources

### API Groups and Their Purpose

| API Group | Location | Purpose | Reconciler Location |
|-----------|----------|---------|-------------------|
| **cloud-resources.kyma-project.io** | `api/cloud-resources/v1beta1/` | User-facing SKR resources | `pkg/skr/` |
| **cloud-control.kyma-project.io** | `api/cloud-control/v1beta1/` | Low-level KCP resources | `pkg/kcp/` |

**Flow**: User creates SKR resource → SKR reconciler creates KCP resource → KCP reconciler provisions cloud provider resource

## Architecture Patterns

### 1. The Three-State Pattern (State, Focal, Composed)

This is the foundational pattern for all reconcilers. Understanding this is **critical** for working on any reconciler.

#### **Composed State** (`pkg/composed/state.go`)
- **Purpose**: Base state providing Kubernetes client access, object management, and common utilities
- **Contains**:
  - K8s client and API reader
  - The object being reconciled
  - Name and namespace
  - Status update helpers
  - Cluster information (client, scheme, event recorder)

```go
type State interface {
    Name() types.NamespacedName
    Obj() client.Object
    K8sClient() client.Client
    Cluster() StateCluster
    UpdateObj(ctx context.Context) error
    UpdateObjStatus(ctx context.Context) error
    // ... more methods
}
```

#### **Focal State** (`pkg/common/actions/focal/state.go`)
- **Purpose**: Adds cloud provider scope management to composed state
- **Used by**: All KCP reconcilers that need cloud provider context
- **Contains**:
  - Scope resource reference (defines cloud provider project/subscription/account)
  - Methods to work with scoped resources
  - Helper for accessing common object fields

```go
type State interface {
    composed.State
    Scope() *cloudcontrolv1beta1.Scope
    SetScope(*cloudcontrolv1beta1.Scope)
    ObjAsCommonObj() CommonObject
}
```

#### **Provider-Specific State** (e.g., `pkg/kcp/provider/gcp/redisinstance/state.go`)
- **Purpose**: Extends focal state with provider-specific context and API clients
- **Contains**:
  - Cloud provider API clients
  - Remote resource representation (e.g., `gcpRedisInstance *redispb.Instance`)
  - Update masks and modification tracking
  - Provider-specific helper methods

```go
type State struct {
    types.State  // types.State is a domain-specific interface extending focal.State

    gcpRedisInstance     *redispb.Instance
    gcpRedisInstanceAuth *redispb.InstanceAuthString
    memorystoreClient    client.MemorystoreClient
    updateMask           []string
}
```

**Key Insight**: States are layered like an onion. Each layer adds more specific context as you move from generic Kubernetes operations to cloud provider-specific API calls.

### 2. The Composed Action Pattern

Actions are the building blocks of reconciliation logic. They follow a functional composition pattern.

#### **Action Definition**
```go
type Action func(ctx context.Context, state State) (error, context.Context)
```

#### **Action Composition** (`pkg/composed/action.go`)
Actions are composed together to form reconciliation flows:

```go
composed.ComposeActions(
    "actionName",
    action1,
    action2,
    action3,
)
```

Actions execute **sequentially** and:
- Stop on first error
- Pass context forward (can be modified by actions)
- Support flow control via special error types (see `pkg/composed/errors.go`)

#### **Common Action Patterns**

**Conditional Execution**:
```go
composed.IfElse(
    predicate,  // function returning bool
    actionWhenTrue,
    actionWhenFalse,
)
```

**Provider Switching**:
```go
composed.BuildSwitchAction(
    "providerSwitch",
    defaultAction,
    composed.NewCase(GcpProviderPredicate, gcpAction),
    composed.NewCase(AzureProviderPredicate, azureAction),
    composed.NewCase(AwsProviderPredicate, awsAction),
)
```

**Marker-based Deletion**:
```go
composed.IfElse(
    composed.Not(composed.MarkedForDeletionPredicate),
    createUpdateActions,
    deleteActions,
)
```

### 3. Reconciler Structure Patterns

There are **two patterns** in the codebase. Understanding both is important for maintaining existing code and creating new reconcilers.

#### **NEW PATTERN (RECOMMENDED): GcpSubnet** - Provider-Specific CRD

**Location**: `pkg/kcp/provider/gcp/subnet/`

**Use this pattern for all new reconcilers**, especially for provider-specific CRDs.

**Key Difference**: No shared intermediate layer. State directly extends focal.State in the provider package.

**CRD Structure**:
```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: GcpSubnet  # Provider-specific CRD name
spec:
  scope:
    name: scope-name
  cidr: 10.0.0.0/24
  # Only GCP-specific fields
```

**Directory Structure**:
```
pkg/kcp/provider/gcp/subnet/
├── reconcile.go                    # Reconciler with action composition
├── state.go                        # State directly extends focal.State
├── client/                         # Cloud provider API client
│   └── client.go
├── loadSubnet.go                   # Load remote resource
├── createSubnet.go                 # Create resource
├── deleteSubnet.go                 # Delete resource
├── updateStatus.go                 # Update CR status
├── waitCreationOperationDone.go    # Wait for async operations
└── ... other actions
```

**State Implementation** (`state.go`):
```go
type State struct {
    focal.State  // DIRECTLY extends focal.State (no intermediate layer)
    
    computeClient              client.ComputeClient
    networkConnectivityClient  client.NetworkConnectivityClient
    
    subnet                  *computepb.Subnetwork  // Remote resource
    serviceConnectionPolicy *networkconnectivitypb.ServiceConnectionPolicy
    
    updateMask []string
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    computeClient := f.computeClientProvider()
    // ... initialize clients
    
    return &State{
        State:         focalState,  // Embed focal.State directly
        computeClient: computeClient,
        // ...
    }, nil
}

func (s *State) ObjAsGcpSubnet() *cloudcontrolv1beta1.GcpSubnet {
    return s.Obj().(*cloudcontrolv1beta1.GcpSubnet)
}
```

**Reconciler Structure** (`reconcile.go`):
```go
type gcpSubnetReconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory  // Provider-specific state factory
}

func (r *gcpSubnetReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.GcpSubnet{}),
        focal.New(),
        r.newFlow(),  // Provider-specific flow
    )
}

func (r *gcpSubnetReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Create provider-specific state from focal state
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            // Handle state creation error
        }
        
        return composed.ComposeActions(
            "privateSubnet",
            actions.AddCommonFinalizer(),
            loadNetwork,
            loadSubnet,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    createSubnet,
                    updateStatusId,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteSubnet,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)  // Pass provider-specific state
    }
}
```

**Pattern Characteristics**:
- ✅ Provider-specific CRD (GcpSubnet, AzureRedisEnterprise, etc.)
- ✅ State directly extends focal.State (no intermediate layer)
- ✅ All reconciliation logic in provider package
- ✅ One file per action
- ✅ Clean separation, easier to version independently

---

#### **OLD PATTERN (LEGACY): RedisInstance** - Multi-Provider CRD

**Location**: 
- Shared: `pkg/kcp/redisinstance/`
- Provider-specific: `pkg/kcp/provider/gcp/redisinstance/`

**Used for existing multi-provider CRDs**. Do not use for new resources.

**Key Difference**: Has a shared intermediate state layer between focal and provider-specific state.

**CRD Structure**:
```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance  # Multi-provider CRD
spec:
  scope:
    name: scope-name
  ipRange:
    name: iprange-name
  instance:
    gcp:              # GCP-specific spec
      memorySize: 5
      tier: BASIC
    aws:              # AWS-specific spec
      cacheNodeType: cache.t3.micro
    azure:            # Azure-specific spec
      sku: Basic
```

**Directory Structure**:
```
pkg/kcp/redisinstance/              # SHARED LAYER
├── reconciler.go                    # Main reconciler with provider switching
├── state.go                         # Shared state (extends focal.State)
├── types/
│   └── state.go                     # Shared state interface
├── loadIpRange.go                   # Shared actions
└── ...

pkg/kcp/provider/gcp/redisinstance/  # PROVIDER-SPECIFIC LAYER
├── new.go                           # Provider action composition
├── state.go                         # Provider state (extends shared state)
├── client/
│   └── client.go
├── loadRedis.go
├── createRedis.go
├── modifyMemorySizeGb.go
└── ...
```

**Shared State Layer** (`pkg/kcp/redisinstance/types/state.go`):
```go
type State interface {
    focal.State  // Extends focal.State
    ObjAsRedisInstance() *v1beta1.RedisInstance
    
    IpRange() *v1beta1.IpRange
    SetIpRange(r *v1beta1.IpRange)
}
```

**Shared State Implementation** (`pkg/kcp/redisinstance/state.go`):
```go
type state struct {
    focal.State  // Embeds focal.State
    
    ipRange *cloudcontrolv1beta1.IpRange  // Shared across all providers
}

func newState(focalState focal.State) types.State {
    return &state{State: focalState}
}
```

**Provider-Specific State** (`pkg/kcp/provider/gcp/redisinstance/state.go`):
```go
type State struct {
    types.State  // Extends shared redisinstance state (which extends focal.State)
    
    gcpRedisInstance     *redispb.Instance     // GCP-specific
    gcpRedisInstanceAuth *redispb.InstanceAuthString
    memorystoreClient    client.MemorystoreClient
    updateMask           []string
}

func (f *stateFactory) NewState(ctx context.Context, redisInstanceState types.State) (*State, error) {
    memorystoreClient := f.memorystoreClientProvider()
    
    return &State{
        State:             redisInstanceState,  // Embed shared state
        memorystoreClient: memorystoreClient,
        // ...
    }, nil
}
```

**Main Reconciler with Provider Switching** (`pkg/kcp/redisinstance/reconciler.go`):
```go
func (r *redisInstanceReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{}),
        focal.New(),
        func(ctx context.Context, st composed.State) (error, context.Context) {
            return composed.ComposeActions(
                "redisInstanceCommon",
                loadIpRange,  // Shared action
                composed.BuildSwitchAction(
                    "providerSwitch",
                    nil,
                    composed.NewCase(statewithscope.GcpProviderPredicate, 
                        gcpredisinstance.New(r.gcpStateFactory)),
                    composed.NewCase(statewithscope.AzureProviderPredicate, 
                        azureredisinstance.New(r.azureStateFactory)),
                    composed.NewCase(statewithscope.AwsProviderPredicate, 
                        awsredisinstance.New(r.awsStateFactory)),
                ),
            )(ctx, newState(st.(focal.State)))
        },
    )
}
```

**Provider-Specific Action Composition** (`pkg/kcp/provider/gcp/redisinstance/new.go`):
```go
func New(stateFactory StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Convert shared state to provider-specific state
        state, err := stateFactory.NewState(ctx, st.(types.State))
        if err != nil {
            // Handle state creation error
        }

        return composed.ComposeActions(
            "redisInstance",
            actions.AddCommonFinalizer(),
            loadRedis,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "redisInstance-create",
                    createRedis,
                    updateStatusId,
                    addUpdatingCondition,
                    waitRedisAvailable,
                    modifyMemorySizeGb,
                    modifyMemoryReplicaCount,
                    modifyRedisConfigs,
                    modifyMaintenancePolicy,
                    modifyAuthEnabled,
                    updateRedis,
                    upgradeRedis,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "redisInstance-delete",
                    removeReadyCondition,
                    deleteRedis,
                    waitRedisDeleted,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)
    }
}
```

**Pattern Characteristics**:
- ⚠️ Multi-provider CRD (single CRD for all providers)
- ⚠️ Three-layer state: composed → focal → shared → provider-specific
- ⚠️ Shared reconciler with provider switching
- ⚠️ Harder to version provider-specific fields independently
- ⚠️ Used only for legacy resources

---

### Pattern Comparison Summary

| Aspect | NEW Pattern (GcpSubnet) | OLD Pattern (RedisInstance) |
|--------|------------------------|----------------------------|
| **CRD** | Provider-specific (GcpSubnet) | Multi-provider (RedisInstance) |
| **State Layers** | 2 layers: composed → focal → provider | 3 layers: composed → focal → shared → provider |
| **Package Structure** | Single provider package | Shared + provider packages |
| **Reconciler** | Direct flow to provider logic | Provider switching in reconciler |
| **Use For** | **All new resources** | Legacy resources only |
| **Example** | GcpSubnet, AzureVNetLink | RedisInstance, NfsInstance |

#### Visual State Hierarchy

**NEW Pattern (GcpSubnet)**:
```
┌─────────────────────────────┐
│   composed.State            │  ← Base K8s operations
│   (pkg/composed/state.go)   │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   focal.State               │  ← Adds Scope management
│   (pkg/common/actions/      │
│    focal/state.go)          │
└──────────────┬──────────────┘
               │ embeds (DIRECTLY)
               ↓
┌─────────────────────────────┐
│   Provider State            │  ← Provider-specific
│   (pkg/kcp/provider/gcp/    │     clients & resources
│    subnet/state.go)         │
│                             │
│   + computeClient           │
│   + subnet                  │
│   + updateMask              │
└─────────────────────────────┘
```

**OLD Pattern (RedisInstance)**:
```
┌─────────────────────────────┐
│   composed.State            │  ← Base K8s operations
│   (pkg/composed/state.go)   │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   focal.State               │  ← Adds Scope management
│   (pkg/common/actions/      │
│    focal/state.go)          │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   Shared State (types.State)│  ← Shared domain logic
│   (pkg/kcp/redisinstance/   │     (e.g., IpRange)
│    types/state.go)          │
│                             │
│   + ipRange                 │
└──────────────┬──────────────┘
               │ embeds
               ↓
┌─────────────────────────────┐
│   Provider State            │  ← Provider-specific
│   (pkg/kcp/provider/gcp/    │     clients & resources
│    redisinstance/state.go)  │
│                             │
│   + gcpRedisInstance        │
│   + memorystoreClient       │
│   + updateMask              │
└─────────────────────────────┘
```

### When to Use Which Pattern

**Use NEW Pattern (GcpSubnet)**:
- ✅ Creating any new resource
- ✅ Adding new provider-specific features (e.g., AzureRedisEnterprise)
- ✅ When provider has unique capabilities

**Use OLD Pattern (RedisInstance)**:
- ⚠️ Only when maintaining existing multi-provider CRDs
- ⚠️ Do not create new multi-provider CRDs

## CRD Architecture Evolution

### Current State (Multi-Provider in Single CRD)

**KCP (cloud-control)**: Resources like `RedisInstance` contain specs for multiple providers in a single CRD:

```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: RedisInstance
spec:
  scope:
    name: scope-name
  ipRange:
    name: iprange-name
  remoteRef:
    name: skr-resource-name
  instance:
    gcp:              # GCP-specific spec
      memorySize: 5
      tier: BASIC
    aws:              # AWS-specific spec
      cacheNodeType: cache.t3.micro
    azure:            # Azure-specific spec
      sku: Basic
```

**Problem**: This design couples multiple providers in one CRD, making it harder to:
- Version provider-specific fields independently
- Add provider-specific features
- Maintain clear separation of concerns

### Future Direction (Provider-Specific CRDs)

**New approach**: Each provider gets its own CRD for new resource types.

**Example** - When adding Azure Redis Enterprise support:
```yaml
apiVersion: cloud-control.kyma-project.io/v1beta1
kind: AzureRedisEnterprise  # Provider-specific CRD
spec:
  scope:
    name: scope-name
  # Azure-specific fields only
  sku: Enterprise_E10
  capacity: 2
  zones: ["1", "2"]
```

**When creating new reconcilers**: Use provider-specific CRD names (e.g., `AzureRedisEnterprise`, `GcpCloudSQL`, `AwsRDS`) rather than adding to multi-provider CRDs.

## GCP Client Architecture

Cloud Manager has two patterns for GCP API client management: the **OLD pattern** and the **NEW pattern**. Understanding the difference is critical when working with GCP reconcilers.

### Pattern Overview

| Aspect | OLD Pattern | NEW Pattern |
|--------|------------|-------------|
| **Location** | `pkg/kcp/provider/gcp/client/provider.go` | `pkg/kcp/provider/gcp/client/gcpClients.go` |
| **Used By** | IpRange, NfsInstance, NfsBackup, NfsRestore | Subnet, RedisCluster, RedisInstance, VpcPeering |
| **Client Creation** | Per-request via `ClientProvider[T]` | Singleton via `NewGcpClients()` |
| **HTTP Client** | Cached `*http.Client` with periodic renewal | Token-based auth per service |
| **API Style** | Google API Discovery (REST) | Google Cloud Client Libraries (gRPC) |
| **Use For** | **Legacy resources only** | **All new resources** |

### NEW Pattern (Recommended) - GcpClients

**Philosophy**: Create all GCP API clients once at startup in `cmd/main.go`, then provide lightweight accessor functions to reconcilers.

#### Architecture

**Centralized Client Initialization** (`pkg/kcp/provider/gcp/client/gcpClients.go`):

```go
type GcpClients struct {
    ComputeNetworks                           *compute.NetworksClient
    ComputeAddresses                          *compute.AddressesClient
    ComputeRouters                            *compute.RoutersClient
    ComputeSubnetworks                        *compute.SubnetworksClient
    RegionOperations                          *compute.RegionOperationsClient
    NetworkConnectivityCrossNetworkAutomation *networkconnectivity.CrossNetworkAutomationClient
    RedisCluster                              *rediscluster.CloudRedisClusterClient
    RedisInstance                             *redisinstance.CloudRedisClient
    VpcPeeringClients                         *VpcPeeringClients
}

func NewGcpClients(ctx context.Context, credentialsFile, peeringCredentialsFile string, logger logr.Logger) (*GcpClients, error) {
    // Creates token providers with appropriate scopes for each service
    // Returns all clients initialized and ready to use
}
```

**Key Features**:
- Uses **Cloud Client Libraries** (modern gRPC-based SDKs from `cloud.google.com/go`)
- Token-based authentication with `oauth2adapt.TokenSourceFromTokenProvider`
- Each service has its own token provider with appropriate scopes
- All clients created once at startup
- Implements `Close()` method for cleanup via reflection

#### Adding a New Client (NEW Pattern)

**Step 1: Add Client to GcpClients struct** (`gcpClients.go`):

```go
type GcpClients struct {
    // ... existing clients
    MyNewService *mynewservice.MyServiceClient  // Add your client here
}
```

**Step 2: Initialize in NewGcpClients()** (`gcpClients.go`):

```go
func NewGcpClients(ctx context.Context, credentialsFile, peeringCredentialsFile, logger) (*GcpClients, error) {
    // ... existing initialization
    
    // Add your service initialization
    myServiceTokenProvider, err := b.WithScopes(mynewservice.DefaultAuthScopes()).BuildTokenProvider()
    if err != nil {
        return nil, fmt.Errorf("failed to build myservice token provider: %w", err)
    }
    myServiceTokenSource := oauth2adapt.TokenSourceFromTokenProvider(myServiceTokenProvider)
    
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

**Step 3: Create Typed Client Interface** (`pkg/kcp/provider/gcp/<resource>/client/<service>Client.go`):

```go
package client

import (
    "context"
    mynewservice "cloud.google.com/go/mynewservice/apiv1"
    "cloud.google.com/go/mynewservice/apiv1/mynewservicepb"
    gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
)

// Define your interface with business operations
type MyServiceClient interface {
    CreateResource(ctx context.Context, projectId, locationId, resourceId string, opts CreateOptions) error
    GetResource(ctx context.Context, projectId, locationId, resourceId string) (*mynewservicepb.Resource, error)
    UpdateResource(ctx context.Context, resource *mynewservicepb.Resource, updateMask []string) error
    DeleteResource(ctx context.Context, projectId, locationId, resourceId string) error
}

// Create provider function that returns accessor
func NewMyServiceClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MyServiceClient] {
    return func() MyServiceClient {
        return NewMyServiceClient(gcpClients)
    }
}

// Implement the client wrapping the real GCP client
func NewMyServiceClient(gcpClients *gcpclient.GcpClients) MyServiceClient {
    return &myServiceClient{client: gcpClients.MyNewService}
}

type myServiceClient struct {
    client *mynewservice.MyServiceClient
}

func (c *myServiceClient) CreateResource(ctx context.Context, projectId, locationId, resourceId string, opts CreateOptions) error {
    req := &mynewservicepb.CreateResourceRequest{
        Parent: fmt.Sprintf("projects/%s/locations/%s", projectId, locationId),
        ResourceId: resourceId,
        Resource: &mynewservicepb.Resource{
            // ... build request from opts
        },
    }
    
    _, err := c.client.CreateResource(ctx, req)
    return err
}

// ... implement other methods
```

**Step 4: Wire Up in main.go**:

```go
// In cmd/main.go
gcpClients, err := gcpclient.NewGcpClients(ctx, config.GcpConfig.CredentialsFile, config.GcpConfig.PeeringCredentialsFile, rootLogger)
// ... error handling

// Create state factory with client provider
myResourceStateFactory := gcpmyresource.NewStateFactory(
    gcpmyresourceclient.NewMyServiceClientProvider(gcpClients),  // Pass client provider
    env,
)
```

**Step 5: Use in State Factory**:

```go
// In pkg/kcp/provider/gcp/myresource/state.go
type State struct {
    focal.State
    
    myServiceClient client.MyServiceClient
    remoteResource  *mynewservicepb.Resource
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    myServiceClientProvider gcpclient.GcpClientProvider[client.MyServiceClient]
    env                     abstractions.Environment
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
    myServiceClient := f.myServiceClientProvider()  // Get client instance
    
    return &State{
        State:           focalState,
        myServiceClient: myServiceClient,
    }, nil
}
```

#### NEW Pattern Examples

**RedisCluster** (`pkg/kcp/provider/gcp/rediscluster/client/memorystoreClusterClient.go`):
```go
func NewMemorystoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MemorystoreClusterClient] {
    return func() MemorystoreClusterClient {
        return NewMemorystoreClient(gcpClients)
    }
}

func NewMemorystoreClient(gcpClients *gcpclient.GcpClients) MemorystoreClusterClient {
    return &memorystoreClient{redisClusterClient: gcpClients.RedisCluster}
}
```

**Subnet** (`pkg/kcp/provider/gcp/subnet/client/computeClient.go`):
```go
func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
    return func() ComputeClient {
        return NewComputeClient(gcpClients)
    }
}

func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
    return &computeClient{subnetworksClient: gcpClients.ComputeSubnetworks}
}
```

### OLD Pattern (Legacy) - ClientProvider

**Philosophy**: Create HTTP client once, reuse it for multiple service clients created on-demand.

#### Architecture

**Cached HTTP Client** (`pkg/kcp/provider/gcp/client/provider.go`):

```go
type ClientProvider[T any] func(ctx context.Context, credentialsFile string) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
    // Caches the result of first call
    // Subsequent calls return cached client
}

func GetCachedGcpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
    // Returns cached *http.Client
    // Periodically renewed in background
}
```

**Key Features**:
- Uses **Google API Discovery Libraries** (older REST-based SDKs from `google.golang.org/api`)
- Single shared `*http.Client` with periodic renewal (every 6 hours by default)
- Generic `ClientProvider[T]` pattern
- Clients created on-demand, then cached

#### OLD Pattern Example

**FilestoreClient** (`pkg/kcp/provider/gcp/nfsinstance/client/filestoreClient.go`):

```go
func NewFilestoreClientProvider() client.ClientProvider[FilestoreClient] {
    return client.NewCachedClientProvider(
        func(ctx context.Context, credentialsFile string) (FilestoreClient, error) {
            httpClient, err := client.GetCachedGcpClient(ctx, credentialsFile)  // OLD: Get HTTP client
            if err != nil {
                return nil, err
            }

            fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))  // OLD: Discovery API
            if err != nil {
                return nil, fmt.Errorf("error obtaining GCP File Client: [%w]", err)
            }
            return NewFilestoreClient(fsClient), nil
        },
    )
}
```

### Pattern Comparison: Code Walkthrough

**NEW Pattern** (Subnet):
```go
// 1. GcpClients struct holds all clients (gcpClients.go)
type GcpClients struct {
    ComputeSubnetworks *compute.SubnetworksClient  // Created at startup
}

// 2. Provider function returns accessor (subnet/client/computeClient.go)
func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
    return func() ComputeClient {
        return NewComputeClient(gcpClients)  // Just wraps existing client
    }
}

// 3. State factory accepts provider (subnet/state.go)
type stateFactory struct {
    computeClientProvider gcpclient.GcpClientProvider[client.ComputeClient]
}

// 4. Get client in NewState
func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    computeClient := f.computeClientProvider()  // Lightweight call, returns wrapper
    return &State{
        State:         focalState,
        computeClient: computeClient,
    }, nil
}
```

**OLD Pattern** (NfsInstance):
```go
// 1. Provider creates client on-demand (nfsinstance/client/filestoreClient.go)
func NewFilestoreClientProvider() client.ClientProvider[FilestoreClient] {
    return client.NewCachedClientProvider(
        func(ctx context.Context, credentialsFile string) (FilestoreClient, error) {
            httpClient, err := client.GetCachedGcpClient(ctx, credentialsFile)  // Get shared HTTP client
            fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))  // Create service
            return NewFilestoreClient(fsClient), nil
        },
    )
}

// 2. State factory calls provider with credentials
func NewStateFactory(filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient]) StateFactory {
    return &stateFactory{filestoreClientProvider: filestoreClientProvider}
}

// 3. Get client in NewState (heavier call, may create service)
func (f *stateFactory) NewState(ctx context.Context, nfsState types.State) (*State, error) {
    credentialsFile := getCredentialsFile(nfsState)
    filestoreClient, err := f.filestoreClientProvider(ctx, credentialsFile)  // May create client
    // ...
}
```

### Migration Guide (OLD → NEW)

If you need to modernize an OLD pattern client to NEW pattern:

1. **Find modern Cloud Client Library**:
   - OLD: `google.golang.org/api/file/v1`
   - NEW: `cloud.google.com/go/filestore/apiv1`

2. **Add to GcpClients**:
   ```go
   type GcpClients struct {
       // ...
       Filestore *filestore.Client
   }
   ```

3. **Initialize in NewGcpClients()**:
   ```go
   filestoreTokenProvider, err := b.WithScopes(filestore.DefaultAuthScopes()).BuildTokenProvider()
   filestoreTokenSource := oauth2adapt.TokenSourceFromTokenProvider(filestoreTokenProvider)
   filestoreClient, err := filestore.NewClient(ctx, option.WithTokenSource(filestoreTokenSource))
   ```

4. **Update client provider**:
   ```go
   func NewFilestoreClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[FilestoreClient] {
       return func() FilestoreClient {
           return NewFilestoreClient(gcpClients)
       }
   }
   ```

5. **Update state factory signature** (remove `ctx` and `credentialsFile` parameters from NewState).

### Why NEW Pattern is Better

| Aspect | NEW Pattern | OLD Pattern |
|--------|-------------|-------------|
| **Performance** | Clients created once at startup | Clients may be created multiple times |
| **Dependencies** | Modern gRPC libraries | Legacy REST libraries |
| **Error Handling** | Fail fast at startup | Fail during reconciliation |
| **Testing** | Easy to mock GcpClients | Must mock ClientProvider function |
| **Maintenance** | All clients visible in one place | Scattered across packages |
| **Token Refresh** | Automatic per service | Manual HTTP client renewal |

**Use NEW pattern for all new GCP resources.**

### IpRange-Specific Pattern: Hybrid Client Approach

The **IpRange** reconciler uses a **hybrid approach** combining both NEW and OLD client patterns due to GCP API limitations:

**Why Hybrid?**
- **Service Networking API**: No modern Cloud Client Library exists (`cloud.google.com/go/servicenetworking` does not exist)
- **Compute API**: Modern Cloud Client Library available (`cloud.google.com/go/compute/apiv1`)
- **Decision**: Use best available client for each API

**IpRange Client Structure** (`pkg/kcp/provider/gcp/iprange/client/`):

```go
// computeClient.go - Uses NEW pattern (gRPC)
type ComputeClient interface {
    CreatePscIpRange(ctx, project, vpc, name, desc, ipAddress string, prefix int64) (string, error)
    GetIpRange(ctx, project, name string) (*computepb.Address, error)
    DeleteGlobalAddress(ctx, project, name string) (string, error)
    GetComputeOperation(ctx, project, operation string) (*computepb.Operation, error)
}

func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
    return func() ComputeClient {
        return NewComputeClient(gcpClients)  // Uses GcpClients.ComputeGlobalAddresses
    }
}

// serviceNetworkingClient.go - Uses OLD pattern (REST)
type ServiceNetworkingClient interface {
    CreateServiceConnection(ctx, project, vpc string, reservedIpRanges []string) (*servicenetworking.Operation, error)
    ListServiceConnections(ctx, project, vpc string) ([]*servicenetworking.Connection, error)
    PatchServiceConnection(ctx, project, vpc string, reservedIpRanges []string, force bool) (*servicenetworking.Operation, error)
    DeleteServiceConnection(ctx, project, vpc string) (*servicenetworking.Operation, error)
    GetServiceNetworkingOperation(ctx, operation string) (*servicenetworking.Operation, error)
}

func NewServiceNetworkingClientProvider() gcpclient.GcpClientProvider[ServiceNetworkingClient] {
    // Uses OLD pattern with google.golang.org/api/servicenetworking/v1
    // No alternative available - keep until GCP provides modern SDK
}
```

**Key Characteristics**:
- ✅ Hybrid approach is acceptable and common when cloud provider doesn't offer modern SDK
- ✅ Both clients implement clean interfaces for testing
- ✅ Compute client follows NEW pattern (GlobalAddresses, GlobalOperations)
- ✅ Service Networking client follows OLD pattern (cached HTTP client)
- ✅ State factory accepts both `GcpClientProvider` types consistently

**IpRange State Structure**:
```go
// pkg/kcp/provider/gcp/iprange/state.go
type State struct {
    types.State  // Extends shared iprange state (multi-provider pattern)
    
    computeClient            client.ComputeClient           // NEW pattern (gRPC)
    serviceNetworkingClient  client.ServiceNetworkingClient // OLD pattern (REST)
    
    address        *computepb.Address               // Remote GCP address
    psaConnection  *servicenetworking.Connection    // Remote PSA connection
}
```

**When to Use Each Client**:
- **ComputeClient**: Address create/get/delete operations (uses NEW pattern gRPC)
- **ServiceNetworkingClient**: PSA connection create/update/delete (uses OLD pattern REST)

**IpRange-Specific Actions**:
- `loadAddress.go` - Load global address (backward compatible with old name format)
- `createAddress.go` - Create PSC (Private Service Connect) global address
- `deleteAddress.go` - Delete global address
- `loadPsaConnection.go` - Load PSA (Private Service Access) connection
- `createPsaConnection.go` - Create PSA connection for services
- `updatePsaConnection.go` - Update PSA connection reserved IP ranges
- `deletePsaConnection.go` - Delete PSA connection
- `waitOperationDone.go` - Poll both compute and servicenetworking operations
- `identifyPeeringIpRanges.go` - Extract IP ranges from VPC peering

**Multi-Provider Pattern**:
IpRange follows the **OLD Pattern (RedisInstance)** for multi-provider support:
- Shared state in `pkg/kcp/iprange/types/` (extends focal.State)
- GCP-specific state in `pkg/kcp/provider/gcp/iprange/` (extends shared state)
- Provider switching in `pkg/kcp/iprange/reconciler.go`

**Feature Flag Support**:
IpRange has dual implementation controlled by `ipRangeRefactored` feature flag:
- **Legacy (v2)**: Original implementation (default)
- **Refactored**: NEW pattern with clean action composition
- Both implementations coexist for safe rollout

## Common Development Tasks

### Important Note: Kubebuilder Usage

Cloud Manager uses **Kubebuilder** for scaffolding. When creating new CRDs:

- **`*_types.go` files** are generated via Kubebuilder commands
- **Controller bootstrap code** is generated but must be **completely rewritten** to follow Cloud Manager patterns
- Always run **both** `make manifests` and `make generate` after modifying types:
  - `make manifests` - Generates CRDs, webhooks, and RBAC
  - `make generate` - Generates DeepCopy, DeepCopyInto, and DeepCopyObject methods

**Kubebuilder command example**:
```bash
kubebuilder create api --group cloud-control --version v1beta1 --kind AzureRedisEnterprise
```

Then rewrite the generated controller to follow the patterns in this document.

---

### Task 1: Adding a New KCP Reconciler (Cloud Provider Resources)

**Purpose**: KCP reconcilers manage actual cloud provider resources (AWS, Azure, GCP).

**Important**: Follow the **NEW Pattern (GcpSubnet)** for all new reconcilers. Use provider-specific CRD names.

#### Step 1: Define the API (KCP)

**Location**: `api/cloud-control/v1beta1/`

**Use Kubebuilder** to scaffold the CRD:
```bash
kubebuilder create api --group cloud-control --version v1beta1 --kind AzureRedisEnterprise --resource --controller
```

**Modify** `azureredisenterprise_types.go` with **provider-specific naming**:
```go
// azureredisenterprise_types.go
type AzureRedisEnterpriseSpec struct {
    // +kubebuilder:validation:Required
    RemoteRef RemoteRef `json:"remoteRef"`
    
    // +kubebuilder:validation:Required
    Scope ScopeRef `json:"scope"`
    
    // Provider-specific fields only
    Sku      string `json:"sku"`
    Capacity int    `json:"capacity"`
    Zones    []string `json:"zones,omitempty"`
}

type AzureRedisEnterpriseStatus struct {
    State      StatusState        `json:"state,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    // Other status fields
    Id              string `json:"id,omitempty"`
    PrimaryEndpoint string `json:"primaryEndpoint,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
type AzureRedisEnterprise struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   AzureRedisEnterpriseSpec   `json:"spec,omitempty"`
    Status AzureRedisEnterpriseStatus `json:"status,omitempty"`
}

// Implement ScopeRef interface
func (in *AzureRedisEnterprise) ScopeRef() ScopeRef {
    return in.Spec.Scope
}

func (in *AzureRedisEnterprise) SetScopeRef(scopeRef ScopeRef) {
    in.Spec.Scope = scopeRef
}
```

**Run codegen**:
```bash
make manifests  # Generate CRDs, webhooks, RBAC
make generate   # Generate DeepCopy methods
```

**Post-generation steps for SKR resources**:

After running `make manifests`, two additional steps are needed for SKR resources:

1. **Add version annotation** (for new SKR resources):
   - Edit `config/patchAfterMakeManifests.sh`
   - Add a line to set the version annotation for your new CRD:
   ```bash
   yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.1"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_yournewresource.yaml
   ```
   - **Important**: When you modify the API of an existing SKR resource, increment its version in this file

2. **Sync CRDs to distribution directories**:
   - Run `config/sync.sh` to copy generated CRDs to their proper locations:
   ```bash
   ./config/sync.sh
   ```
   - This script copies:
     - KCP CRDs → `config/dist/kcp/crd/bases/`
     - SKR CRDs → `config/dist/skr/crd/bases/providers/{aws,gcp,azure,openstack}/`
     - UI extensions → respective provider directories
   - If you add a new SKR resource, update `config/sync.sh` to include the copy commands for your CRD

#### Step 2: Create Provider-Specific Package Structure

**Location**: `pkg/kcp/provider/<provider>/<resource>/`

**Files to create** (following GcpSubnet pattern):
```
pkg/kcp/provider/azure/redisenterprise/
├── reconcile.go              # Reconciler with action composition
├── state.go                  # State extends focal.State directly
├── client/
│   └── client.go            # Cloud provider API client
├── loadRedis.go             # Load remote resource
├── createRedis.go           # Create logic
├── deleteRedis.go           # Delete logic
├── updateRedis.go           # Update logic
├── waitRedisAvailable.go    # Wait for provisioning
├── waitRedisDeleted.go      # Wait for deletion
└── updateStatus.go          # Status management
```

#### Step 3: Implement State (NEW Pattern)

**state.go** - State directly extends focal.State:
```go
package redisenterprise

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisenterprise/client"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type State struct {
    focal.State  // DIRECTLY extends focal.State (no intermediate layer)
    
    azureRedis *armredis.EnterpriseCluster  // Remote resource representation
    client     client.RedisEnterpriseClient  // Provider API client
}

type StateFactory interface {
    NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
    clientProvider azureclient.ClientProvider[client.RedisEnterpriseClient]
}

func NewStateFactory(clientProvider azureclient.ClientProvider[client.RedisEnterpriseClient]) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    client := f.clientProvider()
    
    return &State{
        State:  focalState,  // Embed focal.State directly
        client: client,
    }, nil
}

func (s *State) ObjAsAzureRedisEnterprise() *cloudcontrolv1beta1.AzureRedisEnterprise {
    return s.Obj().(*cloudcontrolv1beta1.AzureRedisEnterprise)
}
```

#### Step 4: Implement Actions

**Follow the GcpSubnet pattern**: One action per file.

**createRedis.go**:
```go
func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    logger := composed.LoggerFromCtx(ctx)
    
    // Check if already exists
    if state.azureRedis != nil {
        return nil, ctx
    }
    
    logger.Info("Creating Azure Redis Enterprise")
    
    // Build create parameters from CRD spec
    redis := state.ObjAsAzureRedisEnterprise()
    params := buildCreateParams(state, redis)
    
    // Call Azure API
    err := state.client.CreateRedisEnterprise(ctx, params)
    if err != nil {
        // Handle error, update status
        meta.SetStatusCondition(redis.Conditions(), metav1.Condition{
            Type:    cloudcontrolv1beta1.ConditionTypeError,
            Status:  metav1.ConditionTrue,
            Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
            Message: err.Error(),
        })
        redis.Status.State = cloudcontrolv1beta1.StateError
        state.UpdateObjStatus(ctx)
        return composed.StopWithRequeueDelay(time.Minute), nil
    }
    
    // Requeue to check provisioning status
    return composed.StopWithRequeue, nil
}
```

#### Step 5: Compose Actions in reconcile.go (NEW Pattern)

**reconcile.go** - Similar to GcpSubnet:
```go
package redisenterprise

import (
    "context"
    "github.com/kyma-project/cloud-manager/pkg/common/actions"
    "github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
    "github.com/kyma-project/cloud-manager/pkg/composed"
    "github.com/kyma-project/cloud-manager/pkg/feature"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    ctrl "sigs.k8s.io/controller-runtime"
)

type AzureRedisEnterpriseReconciler interface {
    reconcile.Reconciler
}

type azureRedisEnterpriseReconciler struct {
    composedStateFactory composed.StateFactory
    focalStateFactory    focal.StateFactory
    stateFactory         StateFactory  // Provider-specific state factory
}

func NewAzureRedisEnterpriseReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    stateFactory StateFactory,
) AzureRedisEnterpriseReconciler {
    return &azureRedisEnterpriseReconciler{
        composedStateFactory: composedStateFactory,
        focalStateFactory:    focalStateFactory,
        stateFactory:         stateFactory,
    }
}

func (r *azureRedisEnterpriseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    state := r.newFocalState(req.NamespacedName)
    action := r.newAction()
    
    return composed.Handling().
        WithMetrics("azureredisenterprise", util.RequestObjToString(req)).
        Handle(action(ctx, state))
}

func (r *azureRedisEnterpriseReconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "main",
        feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.AzureRedisEnterprise{}),
        focal.New(),
        r.newFlow(),  // Provider-specific flow
    )
}

func (r *azureRedisEnterpriseReconciler) newFlow() composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        // Create provider-specific state from focal state
        state, err := r.stateFactory.NewState(ctx, st.(focal.State))
        if err != nil {
            composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap Azure RedisEnterprise state")
            redis := st.Obj().(*cloudcontrolv1beta1.AzureRedisEnterprise)
            redis.Status.State = cloudcontrolv1beta1.StateError
            return composed.UpdateStatus(redis).
                SetExclusiveConditions(metav1.Condition{
                    Type:    cloudcontrolv1beta1.ConditionTypeError,
                    Status:  metav1.ConditionTrue,
                    Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
                    Message: "Failed to create RedisEnterprise state",
                }).
                SuccessError(composed.StopAndForget).
                SuccessLogMsg(fmt.Sprintf("Error creating new Azure RedisEnterprise state: %s", err)).
                Run(ctx, st)
        }
        
        return composed.ComposeActions(
            "azureRedisEnterprise",
            actions.AddCommonFinalizer(),
            loadRedis,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                composed.ComposeActions(
                    "create-update",
                    createRedis,
                    waitRedisAvailable,
                    updateStatus,
                ),
                composed.ComposeActions(
                    "delete",
                    deleteRedis,
                    waitRedisDeleted,
                    actions.RemoveCommonFinalizer(),
                    composed.StopAndForgetAction,
                ),
            ),
            composed.StopAndForgetAction,
        )(ctx, state)  // Pass provider-specific state
    }
}

func (r *azureRedisEnterpriseReconciler) newFocalState(name types.NamespacedName) focal.State {
    return r.focalStateFactory.NewState(
        r.composedStateFactory.NewState(name, &cloudcontrolv1beta1.AzureRedisEnterprise{}),
    )
}
```

#### Step 6: Register Controller in main.go

**cmd/main.go**:
```go
if err = redisenterprise.NewAzureRedisEnterpriseReconciler(
    composedStateFactory,
    focalStateFactory,
    azureRedisEnterpriseStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "AzureRedisEnterprise")
    os.Exit(1)
}
```

---

### Task 2: Adding a New SKR Reconciler (User-Facing Resources)

**Purpose**: SKR reconcilers project user-facing resources in Kyma clusters to KCP resources. They act as a bridge between the user's cluster and the control plane.

**Key Difference from KCP**: SKR reconcilers don't talk to cloud provider APIs directly. They create/manage KCP resources which then handle the actual cloud provisioning.

#### Architecture: Two Patterns for SKR Reconcilers

Just like KCP reconcilers, SKR reconcilers follow two patterns based on whether they're working with legacy multi-provider CRDs or new provider-specific CRDs.

---

#### **NEW PATTERN: SKR GcpSubnet** (Recommended)

**Location**: `pkg/skr/gcpsubnet/`

**Flow**: SKR `GcpSubnet` → creates KCP `GcpSubnet` → provisions GCP subnet

**Characteristics**:
- Provider-specific SKR CRD (`GcpSubnet`)
- Creates provider-specific KCP CRD (`GcpSubnet`)
- Direct 1:1 mapping between SKR and KCP resource
- Simpler, cleaner code

**Directory Structure**:
```
pkg/skr/gcpsubnet/
├── reconciler.go                # Reconciler with action composition
├── state.go                     # State with SKR and KCP cluster access
├── createKcpGcpSubnet.go       # Create KCP GcpSubnet
├── loadKcpGcpSubnet.go         # Load KCP GcpSubnet
├── deleteKcpGcpSubnet.go       # Delete KCP GcpSubnet
├── waitKcpGcpSubnetDeleted.go  # Wait for deletion
├── updateStatus.go              # Update SKR status from KCP status
└── updateId.go                  # Generate/update ID
```

**SKR CRD** (`api/cloud-resources/v1beta1/gcpsubnet_types.go`):
```go
// GcpSubnet - Provider-specific user-facing resource
type GcpSubnet struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   GcpSubnetSpec   `json:"spec,omitempty"`
    Status GcpSubnetStatus `json:"status,omitempty"`
}

type GcpSubnetSpec struct {
    // User-provided fields
    Cidr string `json:"cidr"`
}

type GcpSubnetStatus struct {
    Id string `json:"id,omitempty"`  // Reference to KCP resource
    State string `json:"state,omitempty"`
    // ... other status fields
}
```

**Reconciler Structure** (`pkg/skr/gcpsubnet/reconciler.go`):
```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpSubnet",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpSubnet{}),
        composed.LoadObj,

        updateId,           // Generate ID for KCP resource
        loadKcpGcpSubnet,   // Load KCP GcpSubnet

        composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "gcpSubnet-create",
                actions.AddCommonFinalizer(),
                createKcpGcpSubnet,      // Create KCP resource
                waitKcpStatusUpdate,     // Wait for KCP reconciliation
                updateStatus,            // Sync status to SKR
            ),
            composed.ComposeActions(
                "gcpSubnet-delete",
                deleteKcpGcpSubnet,      // Delete KCP resource
                waitKcpGcpSubnetDeleted, // Wait for deletion
                actions.RemoveCommonFinalizer(),
                composed.StopAndForgetAction,
            ),
        ),

        composed.StopAndForgetAction,
    )
}
```

**Key Action: Create KCP Resource** (`createKcpGcpSubnet.go`):
```go
func createKcpGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    if state.KcpGcpSubnet != nil {
        return nil, ctx  // Already exists
    }
    
    gcpSubnet := state.ObjAsGcpSubnet()  // SKR resource
    
    // Create KCP GcpSubnet with same provider-specific type
    state.KcpGcpSubnet = &cloudcontrolv1beta1.GcpSubnet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      gcpSubnet.Status.Id,  // Generated ID
            Namespace: state.KymaRef.Namespace,
            Annotations: map[string]string{
                cloudcontrolv1beta1.LabelKymaName:   state.KymaRef.Name,
                cloudcontrolv1beta1.LabelRemoteName: gcpSubnet.Name,
            },
        },
        Spec: cloudcontrolv1beta1.GcpSubnetSpec{
            RemoteRef: cloudcontrolv1beta1.RemoteRef{
                Namespace: gcpSubnet.Namespace,
                Name:      gcpSubnet.Name,
            },
            Scope: cloudcontrolv1beta1.ScopeRef{
                Name: state.KymaRef.Name,
            },
            Cidr:    gcpSubnet.Spec.Cidr,  // Copy from SKR spec
            Purpose: cloudcontrolv1beta1.GcpSubnetPurpose_PRIVATE,
        },
    }
    
    // Create in KCP cluster
    err := state.KcpCluster.K8sClient().Create(ctx, state.KcpGcpSubnet)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error creating KCP GcpSubnet", composed.StopWithRequeue, ctx)
    }
    
    return nil, ctx
}
```

---

#### **OLD PATTERN: SKR GcpRedisInstance** (Legacy)

**Location**: `pkg/skr/gcpredisinstance/`

**Flow**: SKR `GcpRedisInstance` → creates KCP `RedisInstance` (multi-provider) → GCP provider handles it

**Characteristics**:
- Provider-specific SKR CRD (`GcpRedisInstance`)
- Creates **multi-provider** KCP CRD (`RedisInstance`)
- Must set provider-specific fields in KCP spec (`.spec.instance.gcp`)
- More complex mapping logic
- Also manages auth secrets

**Directory Structure**:
```
pkg/skr/gcpredisinstance/
├── reconciler.go                  # Reconciler with action composition
├── state.go                       # State with SKR and KCP cluster access
├── createKcpRedisInstance.go     # Create KCP RedisInstance
├── modifyKcpRedisInstance.go     # Modify KCP RedisInstance
├── loadKcpRedisInstance.go       # Load KCP RedisInstance
├── deleteKcpRedisInstance.go     # Delete KCP RedisInstance
├── createAuthSecret.go           # Manage auth secret in SKR
├── loadAuthSecret.go
├── deleteAuthSecret.go
└── updateStatus.go                # Update SKR status
```

**Key Action: Create KCP Resource** (`createKcpRedisInstance.go`):
```go
func createKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    
    if state.KcpRedisInstance != nil {
        return nil, ctx
    }
    
    gcpRedisInstance := state.ObjAsGcpRedisInstance()  // SKR resource
    
    // Create multi-provider KCP RedisInstance
    state.KcpRedisInstance = &cloudcontrolv1beta1.RedisInstance{
        ObjectMeta: metav1.ObjectMeta{
            Name:      gcpRedisInstance.Status.Id,
            Namespace: state.KymaRef.Namespace,
            Annotations: map[string]string{
                cloudcontrolv1beta1.LabelKymaName:        state.KymaRef.Name,
                cloudcontrolv1beta1.LabelRemoteName:      gcpRedisInstance.Name,
                cloudcontrolv1beta1.LabelRemoteNamespace: gcpRedisInstance.Namespace,
            },
        },
        Spec: cloudcontrolv1beta1.RedisInstanceSpec{
            RemoteRef: cloudcontrolv1beta1.RemoteRef{
                Namespace: gcpRedisInstance.Namespace,
                Name:      gcpRedisInstance.Name,
            },
            IpRange: cloudcontrolv1beta1.IpRangeRef{
                Name: state.SkrIpRange.Status.Id,
            },
            Scope: cloudcontrolv1beta1.ScopeRef{
                Name: state.KymaRef.Name,
            },
            Instance: cloudcontrolv1beta1.RedisInstanceInfo{
                Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{  // GCP-specific section
                    Tier:              tier,
                    MemorySizeGb:      memorySizeGb,
                    RedisVersion:      gcpRedisInstance.Spec.RedisVersion,
                    AuthEnabled:       gcpRedisInstance.Spec.AuthEnabled,
                    // ... more GCP-specific fields
                },
            },
        },
    }
    
    err := state.KcpCluster.K8sClient().Create(ctx, state.KcpRedisInstance)
    if err != nil {
        return composed.LogErrorAndReturn(err, "Error creating KCP RedisInstance", composed.StopWithRequeue, ctx)
    }
    
    return nil, ctx
}
```

---

#### SKR Pattern Comparison

| Aspect | NEW Pattern (GcpSubnet) | OLD Pattern (GcpRedisInstance) |
|--------|------------------------|-------------------------------|
| **SKR CRD** | Provider-specific (GcpSubnet) | Provider-specific (GcpRedisInstance) |
| **KCP CRD** | Provider-specific (GcpSubnet) | Multi-provider (RedisInstance) |
| **Mapping** | 1:1, same structure | Complex, set `.instance.gcp` |
| **Use For** | **All new resources** | Legacy resources only |

---

#### Steps to Add a New SKR Reconciler (NEW Pattern)

##### Step 1: Define the SKR API

**Location**: `api/cloud-resources/v1beta1/`

**Use Kubebuilder**:
```bash
kubebuilder create api --group cloud-resources --version v1beta1 --kind GcpRedisCluster --resource --controller
```

**Modify** `gcprediscluster_types.go`:
```go
type GcpRedisClusterSpec struct {
    // User-provided fields
    ShardCount int32 `json:"shardCount"`
    // ... other user-facing fields
}

type GcpRedisClusterStatus struct {
    Id string `json:"id,omitempty"`  // Reference to KCP resource
    State string `json:"state,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

**Run codegen and post-processing**:
```bash
make manifests  # Generate CRDs, webhooks, RBAC
make generate   # Generate DeepCopy methods

# Add version annotation to config/patchAfterMakeManifests.sh
echo 'yq -i '"'"'.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.1"'"'"' $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpredisclusters.yaml' >> config/patchAfterMakeManifests.sh

# Copy CRDs to distribution directories
./config/sync.sh
```

**Note**: Also update `config/sync.sh` to include copy commands for your new CRD if it's not already there.

##### Step 2: Ensure Matching KCP API Exists

Make sure the corresponding KCP API already exists at `api/cloud-control/v1beta1/gcprediscluster_types.go`.

If not, create it first (see Task 1).

##### Step 3: Create SKR Reconciler Package

**Location**: `pkg/skr/gcprediscluster/`

**Files**:
```
pkg/skr/gcprediscluster/
├── reconciler.go
├── state.go
├── createKcpGcpRedisCluster.go
├── loadKcpGcpRedisCluster.go
├── deleteKcpGcpRedisCluster.go
├── waitKcpGcpRedisClusterDeleted.go
├── updateStatus.go
└── updateId.go
```

##### Step 4: Implement State

**state.go**:
```go
package gcprediscluster

import (
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/composed"
)

type State struct {
    composed.State  // SKR cluster access
    
    KymaRef   klog.ObjectRef       // Reference to Kyma/Shoot
    KcpCluster composed.StateCluster  // KCP cluster access
    
    KcpGcpRedisCluster *cloudcontrolv1beta1.GcpRedisCluster  // KCP resource
}

func (s *State) ObjAsGcpRedisCluster() *cloudresourcesv1beta1.GcpRedisCluster {
    return s.Obj().(*cloudresourcesv1beta1.GcpRedisCluster)
}
```

##### Step 5: Implement Reconciler

**reconciler.go**:
```go
func (r *reconciler) newAction() composed.Action {
    return composed.ComposeActions(
        "gcpRedisCluster",
        feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.GcpRedisCluster{}),
        composed.LoadObj,

        updateId,
        loadKcpGcpRedisCluster,

        composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
            composed.ComposeActions(
                "create-update",
                actions.AddCommonFinalizer(),
                createKcpGcpRedisCluster,
                waitKcpStatusUpdate,
                updateStatus,
            ),
            composed.ComposeActions(
                "delete",
                deleteKcpGcpRedisCluster,
                waitKcpGcpRedisClusterDeleted,
                actions.RemoveCommonFinalizer(),
                composed.StopAndForgetAction,
            ),
        ),

        composed.StopAndForgetAction,
    )
}
```

##### Step 6: Register in SKR Runtime

The SKR runtime will automatically discover and register reconcilers via `ReconcilerFactory` pattern.

---

### Task 3: Writing Controller Tests

Cloud Manager uses **Ginkgo** (BDD-style) and **testinfra** framework with mocked cloud provider APIs.

#### Test Structure

**Location**: 
- KCP tests: `internal/controller/cloud-control/<resource>_test.go`
- SKR tests: `internal/controller/cloud-resources/<resource>_test.go`

**Basic Template**:
```go
package cloudcontrol

import (
    . "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP AzureRedisEnterprise", func() {

    It("Scenario: KCP Azure RedisEnterprise is created and deleted", func() {
        name := "test-redis-enterprise"
        scope := &cloudcontrolv1beta1.Scope{}
        
        By("Given Scope exists", func() {
            Eventually(CreateScopeAzure).
                WithArguments(infra.Ctx(), infra, scope, WithName(name)).
                Should(Succeed())
        })
        
        redisEnterprise := &cloudcontrolv1beta1.AzureRedisEnterprise{}
        
        By("When AzureRedisEnterprise is created", func() {
            Eventually(CreateAzureRedisEnterprise).
                WithArguments(
                    infra.Ctx(), 
                    infra.KCP().Client(), 
                    redisEnterprise,
                    WithName(name),
                    WithScope(scope.Name),
                    // Add more options
                ).
                Should(Succeed())
        })
        
        By("Then Azure Redis Enterprise is provisioned", func() {
            Eventually(LoadAndCheck).
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redisEnterprise,
                    NewObjActions(),
                    HavingRedisEnterpriseStatusId(),
                ).
                Should(Succeed())
        })
        
        By("When Azure marks resource as ready", func() {
            azureRedis := infra.AzureMock().GetRedisEnterprise(redisEnterprise.Status.Id)
            infra.AzureMock().SetRedisEnterpriseState(azureRedis.ID, "Succeeded")
        })
        
        By("Then AzureRedisEnterprise has Ready condition", func() {
            Eventually(LoadAndCheck).
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redisEnterprise,
                    NewObjActions(),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).
                Should(Succeed())
        })
        
        // DELETE
        
        By("When AzureRedisEnterprise is deleted", func() {
            Eventually(Delete).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisEnterprise).
                Should(Succeed())
        })
        
        By("And When Azure Redis is deleted", func() {
            infra.AzureMock().DeleteRedisEnterprise(redisEnterprise.Status.Id)
        })
        
        By("Then AzureRedisEnterprise does not exist", func() {
            Eventually(IsDeleted).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisEnterprise).
                Should(Succeed())
        })
    })
})
```

#### Test Infrastructure

**Test Setup** (`suite_test.go`):
```go
var infra testinfra.Infra

var _ = BeforeSuite(func() {
    var err error
    infra, err = testinfra.Start()
    Expect(err).NotTo(HaveOccurred())
    
    // Create namespaces
    Expect(infra.KCP().GivenNamespaceExists(infra.KCP().Namespace())).
        NotTo(HaveOccurred())
    
    // Setup controllers
    Expect(SetupAzureRedisEnterpriseReconciler(
        infra.Ctx(),
        infra.KcpManager(),
        infra.AzureMock().ClientProvider(),
    )).NotTo(HaveOccurred())
})
```

#### Mock Providers

Tests use mocked cloud provider clients:
- `infra.GcpMock()` - GCP API mock
- `infra.AwsMock()` - AWS API mock  
- `infra.AzureMock()` - Azure API mock

**Common mock operations**:
```go
// Create mock resource
infra.AzureMock().CreateRedisEnterprise(params)

// Get mock resource
redis := infra.AzureMock().GetRedisEnterprise(id)

// Update mock state
infra.AzureMock().SetRedisEnterpriseState(id, "Succeeded")

// Delete mock resource
infra.AzureMock().DeleteRedisEnterprise(id)
```

---

#### Creating Cloud Provider Mocks

Cloud Manager uses a **dual-interface pattern** for mocking cloud provider APIs. Each mock client implements:
1. **Client Interface** - The actual API operations (matches the real client interface)
2. **Utils Interface** - Test utilities for manipulating mock state

**Location**: `pkg/kcp/provider/<provider>/mock/`

##### Mock Architecture Overview

**Key Files**:
- `type.go` - Defines the `Server` interface that aggregates all mocks
- `server.go` - Implements the `Server` interface, instantiates and composes all mock clients
- `<service>ClientFake.go` - Individual mock implementations (e.g., `memoryStoreClientFake.go`)

**The Server Interface Pattern** (`type.go`):
```go
// Server aggregates all mock functionality
type Server interface {
    Clients        // Implements all client interfaces
    Providers      // Provides client provider functions
    ClientErrors   // Error injection for testing error scenarios
    
    // Utils interfaces for each service (test helpers)
    MemoryStoreClientFakeUtils
    MemoryStoreClusterClientFakeUtils
    RegionalOperationsClientFakeUtils
    // ... more utils
}
```

**Server Implementation Pattern** (`server.go`):
```go
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
        // ... other mock clients
    }
}

// server struct embeds all mock clients
type server struct {
    *memoryStoreClientFake
    *regionalOperationsClientFake
    // ... other mock clients
}

// Provider methods return functions that access embedded mocks
func (s *server) MemoryStoreProviderFake() client.GcpClientProvider[MemorystoreClient] {
    return func() MemorystoreClient {
        return s  // Returns server which embeds memoryStoreClientFake
    }
}
```

##### Creating a New Mock Client

**Step 1: Define the Utils Interface**

In your `<service>ClientFake.go` file, define test utilities:

```go
// memoryStoreClientFake.go
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
    
    // Delete methods - remove mock resources
    DeleteMemorStoreRedisByName(name string)
}
```

**Step 2: Implement the Mock Client Struct**

Create a struct with:
- Thread-safe storage (use `sync.Mutex`)
- In-memory map to store mock resources
- References to other mocks if needed (e.g., operations client)

```go
type memoryStoreClientFake struct {
    mutex          sync.Mutex
    redisInstances map[string]*redispb.Instance  // In-memory storage
}
```

**Step 3: Implement Utils Interface Methods**

These are **test-only** methods for manipulating mock state:

```go
func (m *memoryStoreClientFake) GetMemoryStoreRedisByName(name string) *redispb.Instance {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    return m.redisInstances[name]
}

func (m *memoryStoreClientFake) SetMemoryStoreRedisLifeCycleState(name string, state redispb.Instance_State) {
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

**Step 4: Implement Client Interface Methods**

These methods **simulate the real cloud provider API**:

```go
func (m *memoryStoreClientFake) CreateRedisInstance(
    ctx context.Context, 
    projectId string, 
    locationId string, 
    instanceId string, 
    options gcpredisinstanceclient.CreateRedisInstanceOptions,
) error {
    // Always check for context cancellation first
    if isContextCanceled(ctx) {
        return context.Canceled
    }

    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Build resource identifier (match real API naming)
    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)
    
    // Create mock resource with initial state
    redisInstance := &redispb.Instance{
        Name:              name,
        State:             redispb.Instance_CREATING,  // Initial state
        Host:              "192.168.0.1",              // Mock endpoint
        Port:              6093,
        ReadEndpoint:      "192.168.24.1",
        ReadEndpointPort:  5093,
        MemorySizeGb:      options.MemorySizeGb,
        RedisConfigs:      options.RedisConfigs,
        MaintenancePolicy: gcpredisinstanceclient.ToMaintenancePolicy(options.MaintenancePolicy),
        AuthEnabled:       options.AuthEnabled,
        RedisVersion:      options.RedisVersion,
    }
    
    m.redisInstances[name] = redisInstance
    return nil
}

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

    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)

    // Return 404 if not found (match real API behavior)
    instance, ok := m.redisInstances[name]
    if !ok {
        return nil, nil, &googleapi.Error{
            Code:    404,
            Message: "Not Found",
        }
    }

    // Return mock auth string
    return instance, &redispb.InstanceAuthString{
        AuthString: "0df0aea4-2cd6-4b9a-900f-a650661e1740",
    }, nil
}

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

    // Update only if exists
    if instance, ok := m.redisInstances[redisInstance.Name]; ok {
        instance.State = redispb.Instance_UPDATING  // Set updating state
        
        // Apply updates based on updateMask (simplified here)
        instance.MemorySizeGb = redisInstance.MemorySizeGb
        instance.RedisConfigs = redisInstance.RedisConfigs
        instance.MaintenancePolicy = redisInstance.MaintenancePolicy
        instance.AuthEnabled = redisInstance.AuthEnabled
    }

    return nil
}

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

    name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)

    // Set deleting state instead of immediately removing
    if instance, ok := m.redisInstances[name]; ok {
        instance.State = redispb.Instance_DELETING
        return nil
    }

    return &googleapi.Error{
        Code:    404,
        Message: "Not Found",
    }
}
```

**Step 5: Add to Server Interface** (`type.go`)

```go
type Server interface {
    Clients
    Providers
    ClientErrors
    
    MemoryStoreClientFakeUtils  // Add your utils interface
    // ... other utils
}
```

**Step 6: Integrate into Server** (`server.go`)

```go
func New() Server {
    return &server{
        memoryStoreClientFake: &memoryStoreClientFake{
            mutex:          sync.Mutex{},
            redisInstances: map[string]*redispb.Instance{},
        },
        // ... other mocks
    }
}

type server struct {
    *memoryStoreClientFake  // Embed the mock
    // ... other mocks
}

// Add provider method
func (s *server) MemoryStoreProviderFake() client.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient] {
    return func() gcpredisinstanceclient.MemorystoreClient {
        return s  // Server embeds memoryStoreClientFake
    }
}
```

##### Mocking Asynchronous Operations

Many cloud APIs are asynchronous (return operation IDs). Mock these with a separate operations client:

**Example: Regional Operations Mock** (`regionOperationsClientFake.go`):

```go
type RegionalOperationsClientFakeUtils interface {
    AddRegionOperation(name string) string
    GetRegionOperationById(operationId string) *computepb.Operation
    SetRegionOperationDone(operationId string)
    SetRegionOperationError(operationId string)
}

type regionalOperationsClientFake struct {
    mutex      sync.Mutex
    operations map[string]*computepb.Operation
}

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

func (c *regionalOperationsClientFake) SetRegionOperationDone(operationId string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    if op := c.operations[operationId]; op != nil {
        op.Status = computepb.Operation_DONE.Enum()
    }
}
```

**Using Operations Mock in Resource Mock**:

```go
type computeClientFake struct {
    mutex   sync.Mutex
    subnets map[string]*computepb.Subnetwork

    operationsClientUtils RegionalOperationsClientFakeUtils  // Reference to operations mock
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

    // Create async operation
    opKey := c.operationsClientUtils.AddRegionOperation(request.Name)
    return opKey, nil
}
```

##### Using Mocks in Tests

**Example Test Flow**:

```go
var _ = Describe("Feature: KCP RedisInstance", func() {
    It("Scenario: KCP GCP RedisInstance is created and deleted", func() {
        redisInstance := &cloudcontrolv1beta1.RedisInstance{}

        By("When RedisInstance is created", func() {
            Eventually(CreateRedisInstance).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
                    WithName(name),
                    WithKcpGcpRedisInstanceMemorySizeGb(5),
                ).Should(Succeed())
        })

        var memorystoreRedisInstance *redispb.Instance
        
        By("Then GCP Redis is created", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
                    NewObjActions(),
                    HavingRedisInstanceStatusId()).
                Should(Succeed())
            
            // Use Utils interface to get mock resource
            memorystoreRedisInstance = infra.GcpMock().GetMemoryStoreRedisByName(
                redisInstance.Status.Id,
            )
            Expect(memorystoreRedisInstance).NotTo(BeNil())
        })

        By("When GCP Redis becomes Available", func() {
            // Use Utils interface to modify mock state
            infra.GcpMock().SetMemoryStoreRedisLifeCycleState(
                memorystoreRedisInstance.Name, 
                redispb.Instance_READY,
            )
        })

        By("Then RedisInstance has Ready condition", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
                    NewObjActions(),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).Should(Succeed())
        })

        // DELETE

        By("When RedisInstance is deleted", func() {
            Eventually(Delete).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
                Should(Succeed())
        })

        By("And When GCP Redis is deleted", func() {
            // Use Utils interface to remove mock resource
            infra.GcpMock().DeleteMemorStoreRedisByName(memorystoreRedisInstance.Name)
        })

        By("Then RedisInstance does not exist", func() {
            Eventually(IsDeleted).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
                Should(Succeed())
        })
    })
})
```

##### Mock Design Best Practices

1. **Thread Safety**: Always use `sync.Mutex` and lock/unlock around map operations
   ```go
   func (m *mock) Operation() {
       m.mutex.Lock()
       defer m.mutex.Unlock()
       // ... operations
   }
   ```

2. **Context Cancellation**: Check context at the start of every method
   ```go
   if isContextCanceled(ctx) {
       return context.Canceled
   }
   ```

3. **Match Real API Behavior**:
   - Return same error types (`googleapi.Error` with proper codes)
   - Use realistic state transitions (CREATING → READY → UPDATING → DELETING)
   - Return 404 for non-existent resources
   - Generate mock endpoints/IPs that look realistic

4. **State Transitions**: Model asynchronous behavior correctly
   ```go
   // On Create: Set CREATING state
   instance.State = redispb.Instance_CREATING
   
   // Test manually transitions to READY
   infra.GcpMock().SetMemoryStoreRedisLifeCycleState(name, redispb.Instance_READY)
   
   // On Delete: Set DELETING state (don't immediately remove)
   instance.State = redispb.Instance_DELETING
   
   // Test manually removes the resource
   infra.GcpMock().DeleteMemorStoreRedisByName(name)
   ```

5. **Naming Consistency**: Use the same name construction as real client
   ```go
   name := gcpredisinstanceclient.GetGcpMemoryStoreRedisName(projectId, locationId, instanceId)
   ```

6. **Utils vs Client Interface Separation**:
   - **Client Interface**: Methods the reconciler calls (Create, Get, Update, Delete)
   - **Utils Interface**: Methods tests call to manipulate mock state (GetByName, SetState, Delete)
   - Keep them separate - don't expose test utilities through the client interface

7. **Async Operations Pattern**: For resources with long-running operations:
   - Create separate operations mock
   - Resource creation returns operation ID
   - Test waits for operation completion by checking operation status
   - Test explicitly marks operation as done

##### Checklist for New Mock Client

- [ ] Create `<service>ClientFake.go` file in `pkg/kcp/provider/<provider>/mock/`
- [ ] Define `<Service>ClientFakeUtils` interface with test utilities
- [ ] Implement mock struct with `sync.Mutex` and storage map
- [ ] Implement all Utils interface methods (Get, Set, Delete)
- [ ] Implement all Client interface methods (CRUD operations)
- [ ] Check context cancellation in every method
- [ ] Return appropriate errors (404 for not found, etc.)
- [ ] Model async state transitions correctly
- [ ] Add Utils interface to `Server` interface in `type.go`
- [ ] Embed mock struct in `server` struct in `server.go`
- [ ] Initialize mock in `New()` function in `server.go`
- [ ] Add provider method to `server` struct
- [ ] Write tests using both client and utils interfaces

---

#### DSL Helpers

**Location**: `pkg/testinfra/dsl/`

Common helpers for test assertions:
```go
LoadAndCheck          // Load object and run checks
HavingConditionTrue   // Assert condition is true
HavingState          // Assert specific state
IsDeleted            // Assert object is deleted
Eventually           // Retry with timeout
```

#### Running Tests

```bash
# Run all tests
make test

# Run specific test file
go test ./internal/controller/cloud-control -run TestControllers

# Run with specific focus
go test ./internal/controller/cloud-control -ginkgo.focus="AzureRedisEnterprise"
```

---

### Task 4: Fixing Bugs in Existing Reconcilers

#### Debugging Strategy

1. **Understand the reconciliation flow**:
   - Start with `reconciler.go` → `newAction()` 
   - Follow the action composition to understand execution order
   - Identify which action is likely failing

2. **Check state management**:
   - Look at provider-specific `state.go`
   - Verify state is correctly initialized in `StateFactory.NewState()`
   - Check if state fields are being populated correctly

3. **Review cloud API interactions**:
   - Check client implementations in `client/` directory
   - Verify API parameters match cloud provider requirements
   - Look for API version mismatches

4. **Status and conditions**:
   - Verify conditions are being set correctly
   - Check if status updates are persisted
   - Look for race conditions in status updates

#### Common Bug Patterns

**Pattern 1: Status not updating**
```go
// BAD - not persisting status
redisInstance.Status.State = cloudcontrolv1beta1.StateReady
return nil, ctx

// GOOD - explicitly update status
redisInstance.Status.State = cloudcontrolv1beta1.StateReady
err := state.UpdateObjStatus(ctx)
if err != nil {
    return err, ctx
}
```

**Pattern 2: Missing error handling**
```go
// BAD - error not handled
err := cloudProviderAPI.Call()
return nil, ctx

// GOOD - handle and update status
err := cloudProviderAPI.Call()
if err != nil {
    meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
        Type:    cloudcontrolv1beta1.ConditionTypeError,
        Status:  metav1.ConditionTrue,
        Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
        Message: err.Error(),
    })
    state.UpdateObjStatus(ctx)
    return composed.StopWithRequeueDelay(time.Minute), nil
}
```

**Pattern 3: Not checking if resource exists**
```go
// BAD - always trying to create
err := state.client.CreateResource(ctx, params)

// GOOD - check if exists first
if state.remoteResource != nil {
    return nil, ctx  // Already exists, skip creation
}
err := state.client.CreateResource(ctx, params)
```

**Pattern 4: Wrong update mask**
```go
// BAD - updating without tracking changes
state.gcpRedis.MemorySizeGb = newSize
state.client.UpdateRedis(ctx, state.gcpRedis)

// GOOD - track changes with update mask
func (s *State) UpdateMemorySizeGb(size int32) {
    s.updateMask = append(s.updateMask, "memory_size_gb")
    s.gcpRedis.MemorySizeGb = size
}
```

## Feature Flags

Cloud Manager uses a feature flag system for controlling feature availability across landscapes and providers.

### Feature Flag Configuration

**Files**:
- `pkg/feature/ff_ga.yaml` - Generally Available features
- `pkg/feature/ff_edge.yaml` - Edge/experimental features

**Structure**:
```yaml
apiDisabled:
  variations:
    enabled: false
    disabled: true
  targeting:
    - name: Disable NfsBackup
      query: feature == "nfsBackup"
      variation: disabled
    
    - name: Disable NFS On CCEE
      query: provider == "openstack" and feature == "nfs"
      variation: disabled
      
    - name: Disable all on trial
      query: brokerPlan == "trial"
      variation: disabled
  defaultRule:
    variation: enabled
```

### Feature Flag Context

**Context Keys** (defined in `pkg/feature/types/types.go`):
```go
const (
    KeyLandscape       = "landscape"    // dev, stage, prod
    KeyFeature         = "feature"      // nfs, redis, peering, etc.
    KeyProvider        = "provider"     // aws, gcp, azure
    KeyBrokerPlan      = "brokerPlan"   // trial, standard, premium
    KeyPlane           = "plane"        // skr, kcp
)
```

### Using Feature Flags in Code

**Loading feature context**:
```go
// In reconciler action composition
feature.LoadFeatureContextFromObj(&cloudcontrolv1beta1.RedisInstance{})
```

This loads feature flags based on:
- Resource type (determines feature)
- Scope (determines provider and landscape)
- Labels (determines broker plan)

**Checking feature flags**:
```go
import "github.com/kyma-project/cloud-manager/pkg/feature"

func someAction(ctx context.Context, state composed.State) (error, context.Context) {
    // Check if API is disabled
    if feature.ApiDisabled.Value(ctx) {
        return composed.StopAndForget, nil
    }
    
    // Check specific feature flag
    if feature.NukeBackupsGcp.Value(ctx) {
        // Perform cleanup
    }
    
    // Continue with normal logic
    return nil, ctx
}
```

### Adding New Feature Flags

1. **Define in YAML** (`pkg/feature/ff_ga.yaml` or `ff_edge.yaml`):
```yaml
myNewFeature:
  variations:
    enabled: true
    disabled: false
  targeting:
    - name: Disable on trial
      query: brokerPlan == "trial"
      variation: disabled
  defaultRule:
    variation: enabled
```

2. **Add constant** (`pkg/feature/flags.go`):
```go
var MyNewFeature = Flag[bool]("myNewFeature")
```

3. **Use in code**:
```go
if feature.MyNewFeature.Value(ctx) {
    // Feature enabled logic
}
```

4. **Validate** with schema validation:
```bash
make test-ff
```

## Key Files and Directories

### Critical Core Files

| Path | Purpose |
|------|---------|
| `pkg/composed/action.go` | Action composition, the heart of reconciliation logic |
| `pkg/composed/state.go` | Base state implementation |
| `pkg/common/actions/focal/` | Focal state for scope management |
| `pkg/feature/` | Feature flag system |
| `api/cloud-control/v1beta1/` | KCP (low-level) API definitions |
| `api/cloud-resources/v1beta1/` | SKR (user-facing) API definitions |

### Provider Implementations

| Path | Purpose |
|------|---------|
| `pkg/kcp/provider/aws/` | AWS-specific KCP implementations |
| `pkg/kcp/provider/azure/` | Azure-specific KCP implementations |
| `pkg/kcp/provider/gcp/` | GCP-specific KCP implementations |
| `pkg/skr/` | SKR reconcilers (remote reconciliation) |

### Testing Infrastructure

| Path | Purpose |
|------|---------|
| `pkg/testinfra/` | Test infrastructure and utilities |
| `pkg/testinfra/dsl/` | DSL helpers for tests |
| `internal/controller/cloud-control/` | KCP controller tests |
| `internal/controller/cloud-resources/` | SKR controller tests |

### Documentation

| Path | Purpose |
|------|---------|
| `docs/contributor/architecture/` | Architecture documentation |
| `docs/user/` | User-facing documentation |

## Development Workflow

### Setting Up Development Environment

1. **Prerequisites**:
```bash
# Go 1.25+
go version

# Controller-gen, kustomize (installed via make)
make mod-download
```

2. **Install tools**:
```bash
make envtest
make controller-gen
```

3. **Run tests**:
```bash
# Run all tests
make test

# Run specific controller tests
go test ./internal/controller/cloud-control -run TestControllers
```

### Making Changes

1. **Update API** (if needed):
```bash
# After modifying *_types.go files
make manifests  # Generate CRDs, webhooks, RBAC
make generate   # Generate DeepCopy methods
```

2. **Add/modify reconciler logic**:
   - Follow GCP RedisInstance pattern
   - One action per file
   - Update tests

3. **Run tests**:
```bash
make test
```

4. **Check linting**:
```bash
make lint
```

### Building and Running

```bash
# Build binary
make build

# Build UI (if needed)
make build_ui

# Run locally (development)
go run cmd/main.go
```

## Common Pitfalls and Solutions

### Pitfall 1: Confusing State Types

**Problem**: Not understanding which state type to use.

**Solution**:
- Use `composed.State` only for generic utilities
- Use `focal.State` for KCP reconcilers that need scope
- Use provider-specific `State` for actual cloud operations
- Always type-assert correctly: `state := st.(*ProviderState)`

### Pitfall 2: Forgetting StopAndForget

**Problem**: Reconciler keeps re-queuing even after successful completion.

**Solution**: Always end successful reconciliation flows with:
```go
composed.StopAndForgetAction  // or
composed.StopAndForget       // or
return composed.StopAndForget, nil
```

### Pitfall 3: Not Handling Async Operations

**Problem**: Cloud resources take time to provision, but reconciler assumes immediate availability.

**Solution**: Use the wait pattern:
```go
composed.ComposeActions(
    "create-flow",
    createRedis,           // Initiates creation
    waitRedisAvailable,    // Polls until ready
    updateStatus,          // Updates CR status
)
```

### Pitfall 4: Status Update Races

**Problem**: Status updates getting overwritten or lost.

**Solution**:
- Use `UpdateObjStatus()` not `UpdateObj()`
- Update status at end of action chains
- Use conditions properly with `meta.SetStatusCondition()`

### Pitfall 5: Not Testing with Mocks

**Problem**: Tests try to call real cloud APIs.

**Solution**: Always use mock providers from testinfra:
```go
// Don't create real clients in tests
// Use provided mocks:
infra.AzureMock().ClientProvider()
infra.GcpMock().MemorystoreClientProvider()
```

### Pitfall 6: Ignoring Feature Flags

**Problem**: Feature works in dev but fails in prod due to feature flags.

**Solution**:
- Always load feature context with `feature.LoadFeatureContextFromObj()`
- Check `feature.ApiDisabled.Value(ctx)` early in reconciliation
- Test with different feature flag configurations

## Testing Strategy Summary

### Unit Tests
- Test individual actions in isolation
- Mock cloud provider clients
- Focus on business logic correctness

### Integration Tests (Controller Tests)
- Use testinfra framework
- Test full reconciliation flows
- Use Ginkgo BDD style
- Mock cloud providers with testinfra mocks

### API Validation Tests
- **Location**: `internal/api-tests/`
- Test CRD validation rules (webhooks, OpenAPI validation)
- Verify field constraints, immutability, and business rules
- Use Ginkgo BDD style with declarative test helpers

**Good Example**: `skr_gcpnfsvolume_test.go`

#### Structure of API Validation Tests

**Test File Pattern**: `<skr|kcp>_<resource>_test.go`

Each test file should:
1. Define a builder for the resource
2. Test creation validation (valid and invalid cases)
3. Test update validation (allowed and disallowed changes)
4. Test field constraints specific to the resource

**Builder Pattern**:
```go
type testGcpNfsVolumeBuilder struct {
    instance cloudresourcesv1beta1.GcpNfsVolume
}

func newTestGcpNfsVolumeBuilder() *testGcpNfsVolumeBuilder {
    return &testGcpNfsVolumeBuilder{
        instance: cloudresourcesv1beta1.GcpNfsVolume{
            Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
        },
    }
}

func (b *testGcpNfsVolumeBuilder) Build() *cloudresourcesv1beta1.GcpNfsVolume {
    return &b.instance
}

func (b *testGcpNfsVolumeBuilder) WithTier(tier cloudresourcesv1beta1.GcpFileTier) *testGcpNfsVolumeBuilder {
    b.instance.Spec.Tier = tier
    return b
}

func (b *testGcpNfsVolumeBuilder) WithCapacityGb(capacityGb int) *testGcpNfsVolumeBuilder {
    b.instance.Spec.CapacityGb = capacityGb
    return b
}
```

#### Test Helper Functions

**Location**: `internal/api-tests/builder_test.go`

These declarative helpers test resource creation and modification:

**For SKR Resources**:
- `canCreateSkr(title, builder)` - Assert resource can be created successfully
- `canNotCreateSkr(title, builder, expectedErrorMsg)` - Assert creation fails with expected error
- `canChangeSkr(title, builder, modifyFunc)` - Assert field can be updated
- `canNotChangeSkr(title, builder, modifyFunc, expectedErrorMsg)` - Assert update fails (immutable fields)

**For KCP Resources**:
- `canCreateKcp(title, builder)` - Assert KCP resource can be created
- `canNotCreateKcp(title, builder, expectedErrorMsg)` - Assert KCP creation fails
- `canNotChangeKcp(title, builder, modifyFunc, expectedErrorMsg)` - Assert KCP field is immutable

#### Example Test Cases

```go
var _ = Describe("Feature: SKR GcpNfsVolume", Ordered, func() {

    // Test valid creation
    canCreateSkr(
        "GcpNfsVolume REGIONAL tier instance can be created with valid capacity: 1024",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithValidFileShareName(),
    )
    
    // Test invalid creation
    canNotCreateSkr(
        "GcpNfsVolume REGIONAL tier instance can not be created with invalid capacity: 1023",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1023).
            WithValidFileShareName(),
        "REGIONAL tier capacityGb must be between 1024 and 9984, and it must be divisble by 256",
    )
    
    // Test allowed update
    canChangeSkr(
        "GcpNfsVolume REGIONAL tier instance capacity can be increased",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.REGIONAL).
            WithCapacityGb(1024).
            WithValidFileShareName(),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1280)
        },
    )
    
    // Test disallowed update
    canNotChangeSkr(
        "GcpNfsVolume BASIC_SSD tier instance capacity can not be reduced",
        newTestGcpNfsVolumeBuilder().
            WithTier(cloudresourcesv1beta1.BASIC_SSD).
            WithCapacityGb(2561).
            WithValidFileShareName(),
        func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
            b.(*testGcpNfsVolumeBuilder).WithCapacityGb(2560)
        },
        "BASIC_SSD tier capacityGb cannot be reduced",
    )
})
```

#### When to Write API Validation Tests

Write API validation tests when:
- Adding new CRD fields with validation rules
- Implementing field constraints (min/max values, patterns, enums)
- Adding immutability rules (fields that can't be changed after creation)
- Implementing cross-field validation (one field depends on another)
- Adding business logic in webhooks

**What to Test**:
1. **Valid values**: Ensure all valid combinations are accepted
2. **Invalid values**: Ensure invalid values are rejected with clear error messages
3. **Boundary conditions**: Test min/max values, edge cases
4. **Immutability**: Ensure immutable fields can't be changed
5. **Conditional logic**: Test validation that depends on other fields

#### Running API Validation Tests

```bash
# Run all API validation tests
go test ./internal/api-tests -v

# Run specific test file
go test ./internal/api-tests -run TestAPIs -ginkgo.focus="GcpNfsVolume"

# Run with verbose output
go test ./internal/api-tests -v -ginkgo.v
```

### E2E Tests
- Located in `e2e/` directory
- Use Cucumber/Gherkin syntax
- Test against real Kyma clusters (in CI)
- Feature files in `e2e/features/`

## Quick Reference

### Action Return Values

| Return | Effect |
|--------|--------|
| `nil, ctx` | Continue to next action |
| `error, ctx` | Stop with error, requeue |
| `composed.StopAndForget, nil` | Stop, don't requeue |
| `composed.StopWithRequeue, nil` | Stop, requeue immediately |
| `composed.StopWithRequeueDelay(duration), nil` | Stop, requeue after delay |

### Common Predicates

```go
composed.MarkedForDeletionPredicate       // Has deletion timestamp
composed.Not(predicate)                   // Negate predicate
statewithscope.GcpProviderPredicate      // Is GCP provider
statewithscope.AwsProviderPredicate      // Is AWS provider
statewithscope.AzureProviderPredicate    // Is Azure provider
```

### State Method Patterns

```go
state.Obj()                               // Get K8s object
state.K8sClient()                         // Get K8s client
state.UpdateObj(ctx)                      // Update object
state.UpdateObjStatus(ctx)                // Update status
state.Scope()                             // Get scope (focal state)
state.ObjAsRedisInstance()                // Type-specific getter
```

### Test Helpers

```go
Eventually(action).WithArguments(...).Should(Succeed())
LoadAndCheck(ctx, client, obj, actions, assertions...)
HavingConditionTrue(conditionType)
HavingState(stateValue)
IsDeleted(ctx, client, obj)
```

## Getting Help

1. **Architecture questions**: See `docs/contributor/architecture/`
2. **New reconciler pattern**: Look at `pkg/kcp/provider/gcp/subnet/` (recommended for new code)
3. **Legacy multi-provider pattern**: See `pkg/kcp/redisinstance/` and `pkg/kcp/provider/gcp/redisinstance/`
4. **Test examples**: See `internal/controller/cloud-control/redisinstance_gcp_test.go`
5. **API docs**: Check inline documentation in `api/` directory

## Final Tips for AI Agents

1. **Always follow the GcpSubnet pattern for new resources** - it's the recommended approach
2. **RedisInstance pattern is legacy** - only use it when maintaining existing multi-provider CRDs
3. **Read the state types carefully** - understanding State/Focal/Composed layering is critical
4. **Actions are composable and sequential** - think in terms of action chains
5. **Test with mocks** - use testinfra framework, don't mock manually
6. **One action per file** - keep code organized and maintainable
7. **Provider-specific CRDs for new resources** - don't add to multi-provider CRDs
8. **Feature flags matter** - always check if feature is enabled
9. **Status updates are explicit** - never assume they happen automatically
10. **State extends focal.State directly in new pattern** - no intermediate shared layer

When in doubt, grep for existing implementations of similar resources and follow the same patterns. Consistency is key to maintainability.

### Quick Pattern Recognition

**Is this new code?** → Use **GcpSubnet pattern**
- Provider-specific CRD name (e.g., `GcpSubnet`, `AzureRedisEnterprise`)
- State directly extends `focal.State`
- Reconciler in provider package (`pkg/kcp/provider/gcp/subnet/reconcile.go`)
- Single reconciler, no provider switching

**Is this legacy code?** → It uses **RedisInstance pattern**
- Multi-provider CRD (e.g., `RedisInstance`, `NfsInstance`)
- Shared state layer in `pkg/kcp/<resource>/types/`
- Provider switching in shared reconciler
- Three-layer state hierarchy
