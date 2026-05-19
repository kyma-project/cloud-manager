# Provider Client Patterns Reference

Use this reference when creating a cloud provider client for a KCP reconciler.

## Client Location

```
pkg/kcp/provider/{provider}/{resource}/client/client.go
```

## Client Pattern Components

Each `client.go` contains four parts:

1. **`Client` interface** — methods for cloud operations
2. **`NewClientProvider()`** — factory returning the provider-specific `ClientProvider` type
3. **`client` struct** — embeds facade clients from the base provider package
4. **`newClient()`** — internal constructor

## ClientProvider Type Reference

| Provider | Type | Signature |
|----------|------|-----------|
| AWS | `awsclient.SkrClientProvider[T]` | `func(ctx, account, region, key, secret, role string) (T, error)` |
| Azure | `azureclient.ClientProvider[T]` | `func(ctx, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (T, error)` |
| GCP | `gcpclient.GcpClientProvider[T]` | `func(string) T` |
| SAP/OpenStack | `sapclient.SapClientProvider[T]` | `func(ctx, pp ProviderParams) (T, error)` |

**GCP is different**: `NewClientProvider` takes `*gcpclient.GcpClients` as parameter because GCP uses a shared client pool. The `string` parameter in the provider function is the GCP project ID and is **ignored in production** (the pool is already initialized). Pass `_` in the implementation.

---

## AWS Client Template

```go
// pkg/kcp/provider/aws/{resource}/client/client.go
package client

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/service/ec2"
    awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
)

type Client interface {
    // Define cloud operations here
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
    return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
        cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
        if err != nil {
            return nil, err
        }
        return newClient(awsclient.NewEc2Client(ec2.NewFromConfig(cfg))), nil
    }
}

func newClient(ec2Client awsclient.Ec2Client) Client {
    return &client{Ec2Client: ec2Client}
}

var _ Client = (*client)(nil)

type client struct {
    awsclient.Ec2Client
}
```

### AWS Facade Clients (`pkg/kcp/provider/aws/client/`)

| Facade | Operations |
|--------|-----------|
| `Ec2Client` | VPCs, subnets, security groups |
| `ElastiCacheClient` | ElastiCache |
| `EfsClient` | EFS file systems |
| `Route53Client` | DNS |
| `StsClient` | Security Token Service |

---

## Azure Client Template

```go
// pkg/kcp/provider/azure/{resource}/client/client.go
package client

import (
    "context"

    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
    azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
    // Define cloud operations here
}

func NewClientProvider() azureclient.ClientProvider[Client] {
    return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
        cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
        if err != nil {
            return nil, err
        }
        networkClientFactory, err := armnetwork.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())
        if err != nil {
            return nil, err
        }
        return newClient(
            azureclient.NewNetworkClient(networkClientFactory.NewVirtualNetworksClient()),
        ), nil
    }
}

func newClient(networkClient azureclient.NetworkClient) *client {
    return &client{NetworkClient: networkClient}
}

type client struct {
    azureclient.NetworkClient
}
```

### Azure Facade Clients (`pkg/kcp/provider/azure/client/`)

| Facade | Operations |
|--------|-----------|
| `NetworkClient` | Virtual networks |
| `SubnetClient` | Subnets |
| `PrivateEndpointClient` | Private endpoints |
| `RedisCacheClient` | Redis cache |

---

## GCP Client Template

```go
// pkg/kcp/provider/gcp/{resource}/client/client.go
package client

import gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

type Client interface {
    gcpclient.VpcNetworkClient
    // Add additional wrapped interfaces or custom methods
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
    return func(_ string) Client {   // project string is ignored — pool is pre-initialized
        return &client{
            VpcNetworkClient: gcpClients.NetworkWrapped(),
        }
    }
}

type client struct {
    gcpclient.VpcNetworkClient
}
```

### GCP Facade Clients (`pkg/kcp/provider/gcp/client/`)

| Facade | Access method | Operations |
|--------|--------------|-----------|
| `VpcNetworkClient` | `NetworkWrapped()` | VPC networks |
| `ServiceUsageClient` | — | Service usage API |
| `FilestoreClient` | — | Filestore |
| `ComputeClient` | — | Compute operations |
| `RoutersClient` | `RoutersWrapped()` | Cloud Routers |

---

## SAP/OpenStack Client Template

```go
// pkg/kcp/provider/sap/{resource}/client/client.go
package client

import (
    "context"

    sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
    sapclient.NetworkClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
    return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
        f := sapclient.NewClientFactory(pp)
        nc, err := f.NetworkClient(ctx)
        if err != nil {
            return nil, err
        }
        return &client{NetworkClient: nc}, nil
    }
}

var _ Client = (*client)(nil)

type client struct {
    sapclient.NetworkClient
}
```

### SAP Facade Clients (via `sapclient.ClientFactory`)

| Facade | Operations |
|--------|-----------|
| `NetworkClient` | OpenStack network operations |
| `ShareClient` | Manila share operations |
| `SubnetClient` | OpenStack subnets |

---

## Using Clients in Provider StateFactory

```go
// pkg/kcp/provider/{provider}/{resource}/state.go
type State struct {
    types.State        // or focal.State for single-provider
    client Client
}

type stateFactory struct {
    clientProvider {provider}client.{Provider}ClientProvider[Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState types.State) (context.Context, *State, error) {
    client, err := f.clientProvider(ctx, /* credentials from baseState below */)
    if err != nil {
        return ctx, nil, err
    }
    return ctx, &State{State: baseState, client: client}, nil
}
```

### Credential Access Paths

```go
// AWS
baseState.Subscription().Status.SubscriptionInfo.Aws.Account
baseState.Subscription().Status.SubscriptionInfo.Aws.Region

// Azure
baseState.Subscription().Status.SubscriptionInfo.Azure.TenantId
baseState.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId

// GCP
baseState.Subscription().Status.SubscriptionInfo.Gcp.ProjectId

// SAP/OpenStack
baseState.Subscription().Status.SubscriptionInfo.Sap.OpenStackProjectId
```

---

## Example Files in Codebase

| Provider | Example |
|----------|---------|
| AWS | `pkg/kcp/provider/aws/vpcnetwork/client/client.go` |
| Azure | `pkg/kcp/provider/azure/vpcnetwork/client/client.go` |
| GCP | `pkg/kcp/provider/gcp/vpcnetwork/client/client.go` |
| SAP | `pkg/kcp/provider/sap/vpcnetwork/client/client.go` |
| Multiple GCP clients | `pkg/kcp/provider/gcp/iprange/v3/state.go` |

## See Also

- `references/kcp-multi-provider.md` — Multi-provider reconciler using these clients
- `references/kcp-single-provider.md` — Single-provider reconciler using these clients
