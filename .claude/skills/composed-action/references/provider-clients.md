# Provider Client Patterns Reference

This document covers how to create cloud provider clients for KCP reconcilers.

## Overview

When a KCP reconciler needs to interact with cloud provider APIs, create a client in:

```
pkg/kcp/provider/{provider}/{resource}/client/client.go
```

## Client Pattern Components

Each client.go contains four parts:

1. **Client interface** - Methods for cloud operations
2. **NewClientProvider()** - Factory function returning the provider-specific `ClientProvider` type
3. **client struct** - Embeds facade clients from the base provider package
4. **newClient()** - Internal constructor

## ClientProvider Type Reference

| Provider | Type | Signature |
|----------|------|-----------|
| AWS | `awsclient.SkrClientProvider[T]` | `func(ctx, account, region, key, secret, role string) (T, error)` |
| Azure | `azureclient.ClientProvider[T]` | `func(ctx, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (T, error)` |
| GCP | `gcpclient.GcpClientProvider[T]` | `func(string) T` |
| SAP | `sapclient.SapClientProvider[T]` | `func(ctx, pp ProviderParams) (T, error)` |

**Note**: GCP is different - `NewClientProvider` takes `*gcpclient.GcpClients` as parameter because GCP uses a shared client pool.

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
    return &client{
        Ec2Client: ec2Client,
    }
}

var _ Client = (*client)(nil)

type client struct {
    awsclient.Ec2Client
}
```

### AWS Facade Clients

Available facade clients in `pkg/kcp/provider/aws/client/`:

- `Ec2Client` - EC2 operations (VPCs, subnets, security groups)
- `ElastiCacheClient` - ElastiCache operations
- `EfsClient` - EFS file system operations
- `Route53Client` - DNS operations
- `StsClient` - Security Token Service

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
    return &client{
        NetworkClient: networkClient,
    }
}

type client struct {
    azureclient.NetworkClient
}
```

### Azure Facade Clients

Available facade clients in `pkg/kcp/provider/azure/client/`:

- `NetworkClient` - Virtual network operations
- `SubnetClient` - Subnet operations
- `PrivateEndpointClient` - Private endpoint operations
- `RedisCacheClient` - Redis cache operations

---

## GCP Client Template

```go
// pkg/kcp/provider/gcp/{resource}/client/client.go
package client

import gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

type Client interface {
    // Embed wrapped client interfaces or define custom methods
    gcpclient.VpcNetworkClient
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
    return func(_ string) Client {
        return &client{
            VpcNetworkClient: gcpClients.NetworkWrapped(),
        }
    }
}

type client struct {
    gcpclient.VpcNetworkClient
}
```

### GCP Notes

- GCP uses a shared client pool (`*gcpclient.GcpClients`)
- `NewClientProvider` receives the client pool as parameter
- The `string` parameter in `GcpClientProvider[T] func(string) T` is ignored in production
- Access wrapped clients via methods like `NetworkWrapped()`, `RoutersWrapped()`

### GCP Facade Clients

Available in `pkg/kcp/provider/gcp/client/`:

- `VpcNetworkClient` - VPC network operations
- `ServiceUsageClient` - Service usage API
- `FilestoreClient` - Filestore operations
- `ComputeClient` - Compute operations

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
    // Embed client interfaces from sapclient package
    sapclient.NetworkClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
    return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
        f := sapclient.NewClientFactory(pp)
        nc, err := f.NetworkClient(ctx)
        if err != nil {
            return nil, err
        }
        return &client{
            NetworkClient: nc,
        }, nil
    }
}

var _ Client = (*client)(nil)

type client struct {
    sapclient.NetworkClient
}
```

### SAP Facade Clients

Available via `sapclient.ClientFactory`:

- `NetworkClient` - OpenStack network operations
- `ShareClient` - Manila share operations
- `SubnetClient` - OpenStack subnet operations

---

## Using Clients in Provider State

### StateFactory Pattern

```go
// pkg/kcp/provider/{provider}/{resource}/state.go
type State struct {
    types.State
    client Client
}

type StateFactory interface {
    NewState(ctx context.Context, baseState types.State) (context.Context, *State, error)
}

type stateFactory struct {
    clientProvider {provider}client.{Provider}ClientProvider[Client]
}

func NewStateFactory(clientProvider {provider}client.{Provider}ClientProvider[Client]) StateFactory {
    return &stateFactory{clientProvider: clientProvider}
}

func (f *stateFactory) NewState(ctx context.Context, baseState types.State) (context.Context, *State, error) {
    // Extract credentials from subscription
    client, err := f.clientProvider(
        ctx,
        // ... credentials from baseState.Subscription().Status.SubscriptionInfo...
    )
    if err != nil {
        return ctx, nil, err
    }
    return ctx, &State{State: baseState, client: client}, nil
}
```

### Accessing Credentials

Credentials come from `Subscription.Status.SubscriptionInfo`:

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

## See Also

- `references/kcp-reconciler.md` - KCP reconciler patterns
- `references/conventions.md` - Coding conventions
