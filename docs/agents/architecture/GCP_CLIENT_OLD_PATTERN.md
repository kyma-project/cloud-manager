# GCP Client OLD Pattern (ClientProvider)

**Authority**: Reference only (legacy pattern)  
**Prerequisite For**: Maintaining legacy GCP resources  
**Must Read Before**: Modifying NfsInstance, NfsBackup, or NfsRestore

**Prerequisites**:
- MUST have read: [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md)
- MUST understand: Why NEW pattern is preferred

**Skip This File If**:
- You are creating new GCP resources (use [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md))
- You are not working on NfsInstance, NfsBackup, or NfsRestore

## Pattern Status

**Status**: ⚠️ **LEGACY - DO NOT USE FOR NEW CODE**  
**Location**: [pkg/kcp/provider/gcp/client/provider.go](../../../../pkg/kcp/provider/gcp/client/provider.go)  
**Used By**: NfsInstance, NfsBackup, NfsRestore (legacy only)

## Rules: OLD Pattern

### ONLY FOR

1. ONLY maintain existing NfsInstance, NfsBackup, NfsRestore
2. ONLY understand when fixing bugs in legacy code
3. ONLY reference when comparing with NEW pattern

### NEVER

1. NEVER use for new GCP resources
2. NEVER create new on-demand client providers
3. NEVER use legacy REST APIs when modern library exists
4. NEVER replicate this pattern

## Pattern Characteristics

| Aspect | OLD Pattern |
|--------|-------------|
| Client Creation | On-demand, then cached |
| HTTP Client | Single shared `*http.Client` |
| API Style | Google API Discovery (REST) |
| Package | `google.golang.org/api/*` |
| Token Refresh | Periodic HTTP client renewal (6 hours) |
| Status | Legacy only |

## Architecture

**Generic Provider** (`provider.go`):
```go
type ClientProvider[T any] func(ctx context.Context, credentialsFile string) (T, error)

func NewCachedClientProvider[T comparable](p ClientProvider[T]) ClientProvider[T] {
    // Caches result of first call
    // Subsequent calls return cached client
}

func GetCachedGcpClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
    // Returns cached *http.Client
    // Periodically renewed (every 6 hours)
}
```

**Key Points**:
- Single HTTP client shared across services
- Clients created on first use
- Manual token refresh every 6 hours

## FilestoreClient Example

```go
func NewFilestoreClientProvider() client.ClientProvider[FilestoreClient] {
    return client.NewCachedClientProvider(
        func(ctx context.Context, credentialsFile string) (FilestoreClient, error) {
            httpClient, err := client.GetCachedGcpClient(ctx, credentialsFile)
            if err != nil {
                return nil, err
            }

            fsClient, err := file.NewService(ctx, option.WithHTTPClient(httpClient))
            if err != nil {
                return nil, fmt.Errorf("error obtaining GCP File Client: [%w]", err)
            }
            return NewFilestoreClient(fsClient), nil
        },
    )
}
```

## Usage in State Factory

```go
type stateFactory struct {
    filestoreClientProvider gcpclient.ClientProvider[client.FilestoreClient]
}

func (f *stateFactory) NewState(ctx context.Context, nfsState types.State) (*State, error) {
    credentialsFile := getCredentialsFile(nfsState)
    
    // May create client on-demand
    filestoreClient, err := f.filestoreClientProvider(ctx, credentialsFile)
    if err != nil {
        return nil, err  // Fails during reconciliation
    }
    
    return &State{
        State:           nfsState,
        filestoreClient: filestoreClient,
    }, nil
}
```

## Why We Moved Away

| Problem | NEW Pattern Solution |
|---------|---------------------|
| Legacy REST libraries | Modern gRPC Cloud Client Libraries |
| On-demand creation | Create once at startup |
| Runtime failures | Fail fast at startup |
| Hard to test | Easy to mock GcpClients |
| Shared HTTP client | Per-service token providers |
| Manual token refresh | Automatic token refresh |

## Common Pitfalls

### Pitfall 1: Using OLD Pattern for New Code

**Frequency**: Rare  
**Impact**: Technical debt, code review rejection  
**Detection**: ClientProvider used in new resource

❌ **WRONG**:
```go
// Creating new resource with OLD pattern
func NewMyNewClientProvider() client.ClientProvider[MyNewClient] {
    return client.NewCachedClientProvider(...)
}
```

✅ **CORRECT**:
```go
// Use NEW pattern
type GcpClients struct {
    MyNewService *mynewservice.MyServiceClient
}

func NewMyNewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[MyNewClient] {
    return func() MyNewClient {
        return NewMyNewClient(gcpClients)
    }
}
```

**Why It Fails**: OLD pattern is legacy, not for new code  
**How to Fix**: Use NEW pattern with GcpClients  
**Prevention**: Read [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md) before coding

### Pitfall 2: Client Creation Failures During Reconciliation

**Frequency**: Occasional  
**Impact**: Reconciliation fails unexpectedly  
**Detection**: Errors in reconciliation logs from client creation

❌ **WRONG** (OLD Pattern):
```go
// Client creation fails during reconciliation
filestoreClient, err := f.filestoreClientProvider(ctx, credentialsFile)
if err != nil {
    return nil, err  // User sees reconciliation failure
}
```

✅ **CORRECT** (NEW Pattern):
```go
// Client created at startup in main.go
gcpClients, err := gcpclient.NewGcpClients(ctx, ...)
if err != nil {
    setupLog.Error(err, "unable to create GCP clients")
    os.Exit(1)  // Fail fast before reconciliation starts
}
```

**Why It Fails**: OLD pattern defers errors to runtime  
**How to Fix**: Use NEW pattern, create clients at startup  
**Prevention**: Always fail fast at startup

## Summary Checklist

When maintaining OLD pattern code:
- [ ] Verify you are working on NfsInstance, NfsBackup, or NfsRestore
- [ ] Understand on-demand client creation
- [ ] Handle client creation errors properly
- [ ] Do not add new resources using this pattern

When creating new code:
- [ ] STOP - do not use OLD pattern
- [ ] Use [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md) instead
- [ ] Add client to GcpClients struct
- [ ] Create at startup in main.go

## Related Documentation

**MUST READ NEXT**:
- [NEW Pattern](GCP_CLIENT_NEW_PATTERN.md) - Required for all new code
- [Hybrid Pattern](GCP_CLIENT_HYBRID.md) - When mixing required

**REFERENCE**:
- [NEW Reconciler Pattern](RECONCILER_NEW_PATTERN.md) - How clients fit into reconcilers
