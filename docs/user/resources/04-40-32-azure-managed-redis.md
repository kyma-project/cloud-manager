# AzureManagedRedis Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `azuremanagedredis.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes an Azure Managed Redis (Microsoft.Cache/redisEnterprise) instance provisioned in your Kyma runtime through Cloud Manager.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the AzureManagedRedis, unless specified otherwise.

A single SKR kind covers three workload classes — single-node dev, HA production, and HA + sharded clustered Redis — through the `redisTier` field.

> Azure Managed Redis is a separate product from Azure Cache for Redis. It is **not yet available** in Azure China or Azure US Government regions; only commercial Azure regions are supported.

> [!TIP] _Only for advanced cases of network topology_
> IP addresses are allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the AzureManagedRedis, then the default IpRange is used. If the default IpRange does not exist, it is automatically created. Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating AzureManagedRedis, one field is mandatory: `redisTier`.

Optionally, you can specify the `authSecret` and `ipRange` fields.

## Specification

This table lists the parameters of `AzureManagedRedis.spec`.

| Parameter | Type | Description |
|-----------|------|-------------|
| **redisTier** (required) | string | Kyma service tier. Encodes the underlying SKU + high-availability + clustering policy. See the [Tier Reference](#tier-reference) below. Allowed values: `S1`, `S2`, `S3`, `S4`, `S5`, `P1`, `P2`, `P3`, `P4`, `P5`, `C3`, `C4`, `C5`, `C6`, `C7`. **Immutable.** |
| **authSecret** | object | Optional. Customises the generated connection secret. |
| **authSecret.name** | string | Optional. Name of the secret. Defaults to the `AzureManagedRedis` resource name. **Immutable once set.** |
| **authSecret.labels** | map[string]string | Optional. Labels applied to the secret. |
| **authSecret.annotations** | map[string]string | Optional. Annotations applied to the secret. |
| **authSecret.extraData** | map[string]string | Optional. Additional secret keys, supporting Go templates that reference `host`, `port`, `primaryEndpoint`, `authString`. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |
| **ipRange** | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created. |
| **ipRange.name** | string | Required when `ipRange` is specified. Name of the existing IpRange to use. |

### Tier Reference

The `redisTier` letter encodes three things in one field. Choose by workload class first, then by size.

| Tier | Underlying Azure SKU | Memory (GB) | High-Availability | Clustering Policy | Use case |
|------|----------------------|-------------|-------------------|-------------------|----------|
| `S1` | `Balanced_B0`        | 1   | No  | EnterpriseCluster | Dev / test (no replica, no HA)      |
| `S2` | `Balanced_B3`        | 3   | No  | EnterpriseCluster | Dev / test (no replica, no HA)      |
| `S3` | `Balanced_B5`        | 6   | No  | EnterpriseCluster | Dev / test (no replica, no HA)      |
| `S4` | `Balanced_B10`       | 12  | No  | EnterpriseCluster | Dev / test (no replica, no HA)      |
| `S5` | `Balanced_B20`       | 24  | No  | EnterpriseCluster | Dev / test (no replica, no HA)      |
| `P1` | `ComputeOptimized_X5`   | 6   | Yes | EnterpriseCluster | Production single-shard HA       |
| `P2` | `ComputeOptimized_X10`  | 12  | Yes | EnterpriseCluster | Production single-shard HA       |
| `P3` | `ComputeOptimized_X20`  | 24  | Yes | EnterpriseCluster | Production single-shard HA       |
| `P4` | `ComputeOptimized_X50`  | 60  | Yes | EnterpriseCluster | Production single-shard HA       |
| `P5` | `ComputeOptimized_X100` | 120 | Yes | EnterpriseCluster | Production single-shard HA       |
| `C3` | `ComputeOptimized_X5`   | 6   | Yes | OSSCluster        | Production sharded (Redis Cluster) |
| `C4` | `ComputeOptimized_X10`  | 12  | Yes | OSSCluster        | Production sharded (Redis Cluster) |
| `C5` | `ComputeOptimized_X20`  | 24  | Yes | OSSCluster        | Production sharded (Redis Cluster) |
| `C6` | `ComputeOptimized_X50`  | 60  | Yes | OSSCluster        | Production sharded (Redis Cluster) |
| `C7` | `ComputeOptimized_X100` | 120 | Yes | OSSCluster        | Production sharded (Redis Cluster) |

P-tiers and C-tiers map to the **same Azure SKU at the same price** when their underlying SKU matches (P1↔C3, P2↔C4, P3↔C5, P4↔C6, P5↔C7); they differ only in the database-level `ClusteringPolicy`. Choose `C*` if your client uses Redis Cluster commands and key hash-slot routing; choose `P*` for a single primary endpoint.

> [!NOTE]
> S-tiers have no replica and no automatic failover. A node failure results in downtime and potential data loss. Use S-tiers for development and testing only.

## Status Fields

| Field | Type | Description |
|-------|------|-------------|
| **id** | string | Internal Cloud Manager identifier. |
| **state** | string | Lifecycle state: `Processing`, `Creating`, `Ready`, `Deleting`, `Error`. |
| **observedGeneration** | int64 | Most recent reconciled `metadata.generation`. |
| **conditions** | list | Standard Kubernetes conditions, including `Ready`, `Error`, and `Deleting`. |

## Auth Secret Details

Once `state` is `Ready`, Cloud Manager creates (or updates) a Kubernetes `Secret` in the same namespace.
The following table lists the meaningful parameters of the auth Secret:

| Parameter | Type | Description |
|-----------|------|-------------|
| **.metadata.name** | string | Name of the auth Secret. It shares the name with AzureManagedRedis unless specified otherwise. |
| **.metadata.labels** | object | Specified custom labels (if any). |
| **.metadata.annotations** | object | Specified custom annotations (if any). |
| **.data.primaryEndpoint** | string | Hostname of the AMR instance (hostname only, not `host:port`). Base64 encoded. |
| **.data.host** | string | Same as `primaryEndpoint`. Base64 encoded. |
| **.data.port** | string | Redis client port. Always `10000` for Azure Managed Redis. Base64 encoded. |
| **.data.authString** | string | Access key for client authentication. Base64 encoded. |

Any keys you list under `spec.authSecret.extraData` are added on top, with Go template expansion against the four base keys above.

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

## Notes and Limitations

- **Immutable fields.** `spec.redisTier` and `spec.authSecret.name` cannot be changed after creation. To resize or change the tier, delete and recreate the resource (data will be lost). All other fields (`authSecret.labels`, `authSecret.annotations`, `authSecret.extraData`, `ipRange`) are mutable and take effect on the next reconciliation.
- **Public network access is disabled.** Connections are only possible from within the Kyma SKR network through the auto-provisioned Private Endpoint.
- **TLS only.** AMR enforces TLS 1.2 client connections; plaintext is not supported.
- **Redis version.** Azure Managed Redis runs Redis 7.4. There is no `redisVersion` field; the version is managed by Azure.
- **Single database per instance.** Cloud Manager provisions exactly one database (`default`) per `AzureManagedRedis` resource.
- **No sovereign cloud support.** AMR is not yet available in Azure China or US Government clouds. Use `AzureRedisInstance` / `AzureRedisCluster` (legacy `armredis` SKUs) on those landscapes.
- **Auth secret name conflict.** If a Kubernetes Secret with the same name already exists in the namespace and belongs to a different `AzureManagedRedis` resource, the instance will enter `Error` state. Use a unique `spec.authSecret.name` to avoid conflicts.
