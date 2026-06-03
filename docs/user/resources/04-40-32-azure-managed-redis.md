# AzureManagedRedis Custom Resource

The `azuremanagedredis.cloud-resources.kyma-project.io` Custom Resource Definition (CRD) describes a single Azure Managed Redis (Microsoft.Cache/redisEnterprise) cluster provisioned in your Kyma runtime through Cloud Manager. A single SKR kind covers three workload classes — single-node dev, HA production, and HA + sharded clustered Redis — through the `redisTier` field.

> Azure Managed Redis is a separate product from Azure Cache for Redis. It is **not yet available** in Azure China or Azure US Government regions; only commercial Azure regions are supported.

## Specification

This table lists the parameters of `AzureManagedRedis.spec`.

| Parameter | Type | Description |
|-----------|------|-------------|
| **redisTier** (required) | string | Kyma service tier. Encodes the underlying SKU + high-availability + clustering policy. See the [Tier Reference](#tier-reference) below. Allowed values: `S1`, `S2`, `S3`, `S4`, `S5`, `P1`, `P2`, `P3`, `P4`, `P5`, `C3`, `C4`, `C5`, `C6`, `C7`. **Immutable.** |
| **authSecret** | object | Customises the generated connection secret. |
| **authSecret.name** | string | Name of the secret. Defaults to the `AzureManagedRedis` resource name. |
| **authSecret.labels** | map[string]string | Labels applied to the secret. |
| **authSecret.annotations** | map[string]string | Annotations applied to the secret. |
| **authSecret.extraData** | map[string]string | Additional secret keys, supporting Go templates that reference `host`, `port`, `primaryEndpoint`, `authString`. |
| **ipRange** | object | Reference to an `IpRange` for private connectivity. If omitted, the default `IpRange` is used. |
| **ipRange.name** | string | Name of the referenced `IpRange`. |

### Tier Reference

The `redisTier` letter encodes three things in one field. Choose by workload class first, then by size.

| Tier | Underlying Azure SKU | Memory (GB) | High-Availability | Clustering Policy | Use case |
|------|----------------------|-------------|-------------------|-------------------|----------|
| `S1` | `Balanced_B0`        | 1   | No  | EnterpriseCluster | Dev / test                          |
| `S2` | `Balanced_B3`        | 3   | No  | EnterpriseCluster | Dev / test                          |
| `S3` | `Balanced_B5`        | 6   | No  | EnterpriseCluster | Dev / test                          |
| `S4` | `Balanced_B10`       | 12  | No  | EnterpriseCluster | Dev / test                          |
| `S5` | `Balanced_B20`       | 24  | No  | EnterpriseCluster | Dev / test                          |
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

P-tiers and C-tiers of equal numeric rank (e.g. `P3` and `C5`) provision the **same Azure SKU at the same price**; they differ only in the database-level `ClusteringPolicy`. Choose `C*` if your client uses Redis Cluster commands and key hash-slot routing; choose `P*` for a single primary endpoint.

## Status Fields

| Field | Type | Description |
|-------|------|-------------|
| **id** | string | Internal Cloud Manager identifier. |
| **state** | string | Lifecycle state: `Creating`, `Updating`, `Ready`, `Deleting`, `Error`. |
| **primaryEndpoint** | string | Hostname for client connections. |
| **port** | int32 | Redis client port (always `10000` for AMR). |
| **observedGeneration** | int64 | Most recent reconciled `metadata.generation`. |
| **conditions** | list | Standard Kubernetes conditions, including `Ready` and `Error`. |

## Connection Secret

Once `state` is `Ready`, Cloud Manager creates (or updates) a Kubernetes `Secret` in the same namespace with the following keys:

| Key | Description |
|-----|-------------|
| `primaryEndpoint` | Hostname of the AMR cluster. |
| `host` | Same as `primaryEndpoint` (kept for compatibility with other Redis SKR kinds). |
| `port` | Always `10000`. |
| `authString` | Access key for client authentication. |

Any keys you list under `spec.authSecret.extraData` are added on top, with Go template expansion against the four base keys above.

## Example

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

- **Immutable fields.** `spec.redisTier`, `spec.ipRange`, and `spec.authSecret.name` cannot be changed after creation. To resize, delete and recreate the resource (data will be lost).
- **Public network access is disabled.** Connections are only possible from within the Kyma SKR network through the auto-provisioned Private Endpoint.
- **TLS only.** AMR enforces TLS 1.2 client connections; plaintext is not supported.
- **Single database per cluster.** Cloud Manager provisions exactly one database (`default`) per `AzureManagedRedis` resource.
- **No sovereign cloud support.** AMR is not yet available in Azure China or US Government clouds. Use `AzureRedisInstance` / `AzureRedisCluster` (legacy `armredis` SKUs) on those landscapes.
