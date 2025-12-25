# GCP Client Hybrid Pattern (IpRange Case Study)

**Authority**: Reference (special case pattern)  
**Prerequisite For**: Understanding IpRange implementation  
**Must Read Before**: Working with resources that mix GCP API types

**Prerequisites**:
- MUST have read: [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md)
- MUST have read: [OLD Pattern](GCP_CLIENT_OLD_PATTERN.md)

**Skip This File If**:
- You have modern Cloud Client Library available for all APIs
- You are not working on IpRange or similar mixed-API resources

## Pattern Status

**Status**: ⚠️ **PRAGMATIC - USE ONLY WHEN NECESSARY**  
**Example**: IpRange (Compute + Service Networking APIs)  
**Reason**: Some GCP APIs lack modern Cloud Client Libraries

## Rules: Hybrid Pattern

### ONLY IF

1. ONLY IF no modern Cloud Client Library exists for required API
2. ONLY IF mixing modern (gRPC) and legacy (REST) APIs
3. ONLY IF documented why hybrid approach is needed

### MUST

1. MUST use NEW pattern when modern library exists
2. MUST use OLD pattern only when no alternative
3. MUST document why hybrid approach is required
4. MUST create clean interfaces for both client types
5. MUST migrate to NEW pattern when modern library becomes available

### NEVER

1. NEVER use legacy API when modern library exists
2. NEVER use hybrid as default approach

## Why Hybrid?

Some GCP APIs lack modern Cloud Client Libraries:

| API | Modern Library | Must Use |
|-----|----------------|----------|
| Compute API | ✅ `cloud.google.com/go/compute` | NEW Pattern (gRPC) |
| Service Networking API | ❌ No modern library | OLD Pattern (REST) |

**Decision**: Use best available client for each API

## IpRange Client Structure

**Location**: [pkg/kcp/provider/gcp/iprange/client/](../../../../pkg/kcp/provider/gcp/iprange/client/)

### Compute Client (NEW Pattern)

**computeClient.go**:
```go
type ComputeClient interface {
    CreatePscIpRange(ctx, project, vpc, name, desc, ipAddress string, prefix int64) (string, error)
    GetIpRange(ctx, project, name string) (*computepb.Address, error)
    DeleteGlobalAddress(ctx, project, name string) (string, error)
    GetComputeOperation(ctx, project, operation string) (*computepb.Operation, error)
}

// Uses GcpClients (NEW pattern)
func NewComputeClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[ComputeClient] {
    return func() ComputeClient {
        return NewComputeClient(gcpClients)
    }
}

type computeClient struct {
    addressesClient      *compute.AddressesClient
    globalOperations     *compute.GlobalOperationsClient
}

func NewComputeClient(gcpClients *gcpclient.GcpClients) ComputeClient {
    return &computeClient{
        addressesClient:  gcpClients.ComputeAddresses,
        globalOperations: gcpClients.GlobalOperations,
    }
}
```

**Characteristics**:
- Modern gRPC library
- Pre-created in GcpClients
- Fast, fail-fast at startup

### Service Networking Client (OLD Pattern)

**serviceNetworkingClient.go**:
```go
type ServiceNetworkingClient interface {
    CreateServiceConnection(ctx, project, vpc string, reservedIpRanges []string) (*servicenetworking.Operation, error)
    ListServiceConnections(ctx, project, vpc string) ([]*servicenetworking.Connection, error)
    PatchServiceConnection(ctx, project, vpc string, reservedIpRanges []string, force bool) (*servicenetworking.Operation, error)
    DeleteServiceConnection(ctx, project, vpc string) (*servicenetworking.Operation, error)
    GetServiceNetworkingOperation(ctx, operation string) (*servicenetworking.Operation, error)
}

// Uses ClientProvider (OLD pattern) - no modern library
func NewServiceNetworkingClientProvider() gcpclient.GcpClientProvider[ServiceNetworkingClient] {
    return func() ServiceNetworkingClient {
        return NewServiceNetworkingClient()
    }
}

type serviceNetworkingClient struct {
    serviceNetworkingClientProvider gcpclient.ClientProvider[*servicenetworking.APIService]
}

func NewServiceNetworkingClient() ServiceNetworkingClient {
    return &serviceNetworkingClient{
        serviceNetworkingClientProvider: NewServiceNetworkingServiceProvider(),
    }
}
```

**Characteristics**:
- Legacy REST API (`google.golang.org/api/servicenetworking/v1`)
- Created on-demand (no modern library available)
- No alternative available

## IpRange State

**State uses both client types**:
```go
type State struct {
    types.State  // Multi-provider pattern
    
    computeClient            client.ComputeClient           // NEW pattern (gRPC)
    serviceNetworkingClient  client.ServiceNetworkingClient // OLD pattern (REST)
    
    address       *computepb.Address               // From Compute API
    psaConnection *servicenetworking.Connection    // From Service Networking API
}

type stateFactory struct {
    computeClientProvider            gcpclient.GcpClientProvider[client.ComputeClient]
    serviceNetworkingClientProvider  gcpclient.GcpClientProvider[client.ServiceNetworkingClient]
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState types.State) (*State, error) {
    return &State{
        State:                   ipRangeState,
        computeClient:           f.computeClientProvider(),           // NEW pattern
        serviceNetworkingClient: f.serviceNetworkingClientProvider(), // OLD pattern
    }, nil
}
```

## When Hybrid is Acceptable

✅ **Use Hybrid when**:
- Cloud provider doesn't offer modern SDK for required API
- Different APIs in same resource need different approaches
- Pragmatic mixing is more maintainable than waiting
- Both clients have clean, testable interfaces

❌ **Do NOT use Hybrid when**:
- Modern library exists for all APIs (use NEW pattern only)
- Only legacy APIs needed (use OLD pattern, plan migration)

## Best Practices for Hybrid

### Practice 1: Clean Interfaces

```go
// Both clients implement testable interfaces
type ComputeClient interface {
    CreatePscIpRange(...) (string, error)
    // ...
}

type ServiceNetworkingClient interface {
    CreateServiceConnection(...) (*servicenetworking.Operation, error)
    // ...
}
```

### Practice 2: Document Why

```go
// serviceNetworkingClient.go

// NOTE: Service Networking API does not have modern Cloud Client Library
// (cloud.google.com/go/servicenetworking does not exist).
// We must use legacy REST API: google.golang.org/api/servicenetworking/v1
// TODO: Migrate to modern library when available
```

### Practice 3: Monitor for Modern Libraries

```bash
# Periodically check if modern library available
go get cloud.google.com/go/servicenetworking@latest
```

## Migration Path

**IF** modern library becomes available:

1. Check for `cloud.google.com/go/<service>`
2. Add client to GcpClients struct
3. Update service client implementation
4. Remove OLD pattern client provider
5. Test thoroughly
6. Update documentation

## Common Pitfalls

### Pitfall 1: Using Hybrid When Not Needed

**Frequency**: Rare  
**Impact**: Unnecessary complexity  
**Detection**: Modern library exists but not used

❌ **WRONG**:
```go
// Using legacy API when modern exists
import "google.golang.org/api/compute/v1"  // Legacy
```

✅ **CORRECT**:
```go
// Use modern library when available
import compute "cloud.google.com/go/compute/apiv1"  // Modern gRPC
```

**Why It Fails**: Adds unnecessary complexity  
**How to Fix**: Use modern library if it exists  
**Prevention**: Always check for modern library first

### Pitfall 2: Not Documenting Why Hybrid

**Frequency**: Occasional  
**Impact**: Confusion for future maintainers  
**Detection**: No comment explaining legacy API usage

❌ **WRONG**:
```go
// No explanation why legacy API used
import "google.golang.org/api/servicenetworking/v1"
```

✅ **CORRECT**:
```go
// NOTE: Service Networking API lacks modern Cloud Client Library.
// Must use legacy REST API. Migrate when modern library available.
import "google.golang.org/api/servicenetworking/v1"
```

**Why It Fails**: Future maintainers don't understand why  
**How to Fix**: Add comment explaining necessity  
**Prevention**: Always document why mixing patterns

## Summary Checklist

Before using hybrid pattern:
- [ ] Verify no modern library exists for required API
- [ ] Check `cloud.google.com/go/<service>`
- [ ] Document why hybrid is necessary
- [ ] Plan migration path when modern library available

When implementing hybrid:
- [ ] Use NEW pattern for APIs with modern libraries
- [ ] Use OLD pattern only when no alternative
- [ ] Create clean interfaces for both types
- [ ] Add documentation explaining why
- [ ] Add TODO for migration

After implementation:
- [ ] Test both client types thoroughly
- [ ] Document which APIs use which pattern
- [ ] Periodically check for modern library availability

## Related Documentation

**MUST READ NEXT**:
- [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md) - Preferred approach
- [OLD Pattern](GCP_CLIENT_OLD_PATTERN.md) - When modern library unavailable

**REFERENCE**:
- [NEW Reconciler Pattern](RECONCILER_NEW_PATTERN.md) - How clients fit into reconcilers
