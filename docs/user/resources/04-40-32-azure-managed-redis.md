# AzureManagedRedis Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `azuremanagedredis.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes an Azure Managed Redis (Microsoft.Cache/redisEnterprise) instance provisioned in your Kyma runtime through Cloud Manager.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the AzureManagedRedis, unless specified otherwise.

A single SKR kind covers three workload classes — single-node dev, HA production, and HA + sharded clustered Redis — through the **redisTier** field.

> [!NOTE]
> Azure Managed Redis is a separate product from Azure Cache for Redis. It is **not yet available** in Azure China or Azure US Government regions; only commercial Azure regions are supported.

> [!TIP]
> Only for advanced cases of network topology: IP addresses are allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the AzureManagedRedis, then the default IpRange is used. If the default IpRange does not exist, it is automatically created. Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

## Specification

This table lists the parameters of **AzureManagedRedis.spec**:

| Parameter | Type | Required | Immutable | Description |
|-----------|------|----------|-----------|-------------|
| **redisTier** | string | Yes | No* | Kyma service tier. Encodes the underlying SKU, high-availability, and clustering policy. See the [Tier Reference](#tier-reference) below. Allowed values: `S1`, `S2`, `S3`, `S4`, `S5`, `P1`, `P2`, `P3`, `P4`, `P5`, `C3`, `C4`, `C5`, `C6`, `C7`. *Tier family (`S`, `P`, `C`) is immutable; see [Scaling](#scaling). |
| **authSecret** | object | No | No | Customizes the generated connection Secret. |
| **authSecret.name** | string | No | Yes | Name of the Secret. Defaults to the `AzureManagedRedis` resource name. |
| **authSecret.labels** | map[string]string | No | No | Labels applied to the Secret. |
| **authSecret.annotations** | map[string]string | No | No | Annotations applied to the Secret. |
| **authSecret.extraData** | map[string]string | No | No | Additional Secret keys, supporting Go templates that reference `host`, `port`, `primaryEndpoint`, `authString`. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |
| **ipRange** | object | No | No | IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created. |
| **ipRange.name** | string | Yes* | No | Name of the existing IpRange to use. *Required when **ipRange** is specified. |

### Tier Reference

The `redisTier` letter encodes three things in one field. Choose by workload class first, then by size.

| Tier | Underlying Azure SKU | Memory (GB) | High-Availability | Clustering Policy |
|------|----------------------|-------------|-------------------|-------------------|
| `S1` | `Balanced_B0`        | 1   | No  | EnterpriseCluster |
| `S2` | `Balanced_B3`        | 3   | No  | EnterpriseCluster |
| `S3` | `Balanced_B5`        | 6   | No  | EnterpriseCluster |
| `S4` | `Balanced_B10`       | 12  | No  | EnterpriseCluster |
| `S5` | `Balanced_B20`       | 24  | No  | EnterpriseCluster |
| `P1` | `ComputeOptimized_X5`   | 6   | Yes | EnterpriseCluster |
| `P2` | `ComputeOptimized_X10`  | 12  | Yes | EnterpriseCluster |
| `P3` | `ComputeOptimized_X20`  | 24  | Yes | EnterpriseCluster |
| `P4` | `ComputeOptimized_X50`  | 60  | Yes | EnterpriseCluster |
| `P5` | `ComputeOptimized_X100` | 120 | Yes | EnterpriseCluster |
| `C3` | `ComputeOptimized_X5`   | 6   | Yes | OSSCluster        |
| `C4` | `ComputeOptimized_X10`  | 12  | Yes | OSSCluster        |
| `C5` | `ComputeOptimized_X20`  | 24  | Yes | OSSCluster        |
| `C6` | `ComputeOptimized_X50`  | 60  | Yes | OSSCluster        |
| `C7` | `ComputeOptimized_X100` | 120 | Yes | OSSCluster        |

P-tiers and C-tiers map to the **same Azure SKU at the same price** when their underlying SKU matches (P1↔C3, P2↔C4, P3↔C5, P4↔C6, P5↔C7); they differ only in the database-level `ClusteringPolicy`. Choose `C*` if your client uses Redis Cluster commands and key hash-slot routing. Choose `P*` for a single primary endpoint.

> [!NOTE]
> S-tiers have no replica and no automatic failover. A node failure results in downtime and potential data loss. Use S-tiers for development and testing only.

### Scaling

You can scale an existing instance up or down by changing `spec.redisTier` to another tier **within the same family**:

- **S-family** (single-node): `S1` ↔ `S2` ↔ `S3` ↔ `S4` ↔ `S5`
- **P-family** (HA, EnterpriseCluster): `P1` ↔ `P2` ↔ `P3` ↔ `P4` ↔ `P5`
- **C-family** (HA, OSSCluster): `C3` ↔ `C4` ↔ `C5` ↔ `C6` ↔ `C7`

Switching between families (for example, `S3` → `P1` or `P2` → `C4`) is **not allowed** and is rejected by the API server. The reason is that a family change would alter high-availability or clustering policy, which Azure does not support for a running instance.

> [!NOTE]
> Azure scales an instance without data loss or client connection interruption when the target tier is a valid scale target. However, a brief performance impact may occur during the scaling operation.

## Status Fields

| Field | Type | Description |
|-------|------|-------------|
| **id** | string | Internal Cloud Manager identifier. |
| **state** | string | Lifecycle state: `Processing`, `Creating`, `Ready`, `Deleting`, `Error`. |
| **observedGeneration** | int64 | Most recent reconciled `metadata.generation`. |
| **conditions** | list | Standard Kubernetes conditions, including `Ready`, `Error`, and `Deleting`. |

## Auth Secret Details

Once **state** is `Ready`, Cloud Manager creates (or updates) a Kubernetes Secret in the same namespace.
The following table lists the meaningful parameters of the auth Secret:

| Parameter | Type | Description |
|-----------|------|-------------|
| **.metadata.name** | string | Name of the auth Secret. It shares the name with AzureManagedRedis unless specified otherwise. |
| **.metadata.labels** | object | Specifies custom labels (if any). |
| **.metadata.annotations** | object | Specifies custom annotations (if any). |
| **.data.primaryEndpoint** | string | Hostname of the AMR instance (hostname only, not `host:port`). Base64 encoded. |
| **.data.host** | string | Hostname of the AMR instance (hostname only, not `host:port`). The value is the same as **primaryEndpoint**. Base64 encoded. |
| **.data.port** | string | Redis client port. Always `10000` for Azure Managed Redis. Base64 encoded. |
| **.data.authString** | string | Access key for client authentication. Base64 encoded. |

Any keys you list under **spec.authSecret.extraData** are added on top, with Go template expansion against the four base keys above.

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureManagedRedis
metadata:
  name: my-managed-redis
  namespace: default
spec:
  redisTier: P2
```

A clustered example:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureManagedRedis
metadata:
  name: my-redis-cluster
  namespace: default
spec:
  redisTier: C5
  authSecret:
    name: redis-cluster-conn
    labels:
      app: my-app
    extraData:
      url: "redis://:{{.authString}}@{{.host}}:{{.port}}"
```

## Limitations

- **TLS only.** AMR enforces TLS 1.2 client connections. Plaintext is not supported.
- **Redis version.** Azure Managed Redis runs Redis 7.4. The version is managed by Azure; there is no **redisVersion** field.
- **Single database per instance.** Cloud Manager provisions exactly one database (`default`) per `AzureManagedRedis` resource.
- **Public network access is disabled.** Connections are only possible from within the Kyma SKR network through the auto-provisioned Private Endpoint.
- **No sovereign cloud support.** AMR is not yet available in Azure China or US Government clouds. Use `AzureRedisInstance` / `AzureRedisCluster` (legacy `armredis` SKUs) on those landscapes.
- **Auth Secret name conflict.** If a Kubernetes Secret with the same name already exists in the namespace and belongs to a different `AzureManagedRedis` resource, the instance enters the `Error` state. Use a unique **spec.authSecret.name** to avoid conflicts.
- **Family change requires recreation.** Switching families (for example, `P1` → `S3`) requires deleting and recreating the resource; data will be lost. Scaling within the same family does not require recreation.
