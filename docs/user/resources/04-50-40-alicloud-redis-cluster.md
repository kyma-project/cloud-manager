# AlicloudRedisCluster Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `alicloudrediscluster.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Alibaba Cloud [ApsaraDB for Redis](https://www.alibabacloud.com/help/en/redis/) instance running in a cloud-native cluster (CE) architecture.
Once the cluster is provisioned, a Kubernetes Secret with endpoint and credential details is created in the same namespace. By default, the Secret has the same name as the AlicloudRedisCluster.

The cluster requires an IP range, allocated from an [IpRange CR](./04-10-iprange.md). If you don't reference one, the default IpRange is used. If the default IpRange doesn't exist, it is created. Create a non-default IpRange only when you need to control network segments to avoid range conflicts.

When creating AlicloudRedisCluster, the **redisTier** and **shardCount** fields are mandatory. **redisTier** sets the per-shard capacity, and **shardCount** sets the number of data shards and can be changed after creation. Optionally, you can set **engineVersion** and **authSecret**.

The AliCloud instance class is derived from the **redisTier** field. You can change **redisTier** after creation to resize per-shard capacity. **shardCount** can also be changed after creation.

## In-transit Encryption

In-transit encryption is always enabled. Communication with the Redis cluster requires a certificate. You can find it in the Secret at the **.data.CaCert.pem** path.

Authentication is always enabled. A generated password is provided in the Secret at the **.data.authString** path.

## Persistence

Persistence is not supported. Data is not written to durable storage (for example, data at rest).

## Redis Tiers

Each tier sets the memory capacity of a single shard. The total cluster capacity equals the per-shard capacity multiplied by `shardCount`.

| RedisTier | Capacity per shard (GiB) |
| --------- | ------------------------ |
| C3        | 1                        |
| C4        | 2                        |
| C5        | 4                        |
| C6        | 8                        |
| C7        | 16                       |

## Specification

This table lists the parameters of AlicloudRedisCluster, together with their descriptions:

| Parameter                  | Type   | Required | Description                                                                                                                                                                                                        |
| -------------------------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                | object | No       | IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created. Immutable.                                                                                  |
| **ipRange.name**           | string | Yes      | Name of the existing IpRange to use.                                                                                                                                                                              |
| **redisTier**              | string | Yes      | The per-shard capacity tier. Supported values are `C3`, `C4`, `C5`, `C6`, `C7`.                                                                                                                                   |
| **shardCount**             | int    | Yes      | Number of data shards. Minimum value is `1`, maximum value is `32`. You can change it after creation. Scaling triggers an online resharding operation, which may cause a temporary increase in latency.            |
| **replicasPerShard**       | int    | No       | Number of read-only replicas per shard. Fixed at `0` and cannot be changed. Independent read-replica configuration is not yet enabled. Defaults to `0`.                                                           |
| **engineVersion**          | string | No       | The version of the Redis engine. Supported values are `5.0`, `6.0`, and `7.0`. Defaults to `7.0`. Immutable; to change it, delete and recreate the cluster.                                                       |
| **authSecret**             | object | No       | Auth Secret options.                                                                                                                                                                                              |
| **authSecret.name**        | string | No       | Auth Secret name.                                                                                                                                                                                                 |
| **authSecret.labels**      | object | No       | Auth Secret labels. Keys and values must be strings.                                                                                                                                                              |
| **authSecret.annotations** | object | No       | Auth Secret annotations. Keys and values must be strings.                                                                                                                                                         |
| **authSecret.extraData**   | object | No       | Allows you to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). Keys and values must be strings. |

## Auth Secret Details

The following table lists the parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                  |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It shares the name with the AlicloudRedisCluster unless **authSecret.name** is set.   |
| **.metadata.labels**        | object | Specified custom labels (if any).                                                                           |
| **.metadata.annotations**   | object | Specified custom annotations (if any).                                                                      |
| **.data.host**              | string | Cluster discovery host.                                                                                     |
| **.data.port**              | string | Cluster discovery port.                                                                                     |
| **.data.discoveryEndpoint** | string | Cluster discovery endpoint. Provided in `<host>:<port>` format.                                             |
| **.data.authString**        | string | Auth string used to authenticate with the cluster.                                                          |
| **.data.CaCert.pem**        | string | CA Certificate that must be used for TLS.                                                                   |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudRedisCluster
metadata:
  name: alicloudrediscluster-sample
spec:
  redisTier: C3
  shardCount: 3
```

