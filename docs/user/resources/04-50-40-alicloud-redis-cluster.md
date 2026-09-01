# AlicloudRedisCluster Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `alicloudrediscluster.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Alibaba Cloud [ApsaraDB for Redis](https://www.alibabacloud.com/help/en/redis/) instance running in cluster (proxy-based sharded) architecture.
Once the cluster is provisioned, a Kubernetes Secret with endpoint and credential details is created in the same namespace. By default, the Secret has the same name as the AlicloudRedisCluster.

The cluster requires an IP range, allocated from an [IpRange CR](./04-10-iprange.md). If you do not reference one, the default IpRange is used and created if it does not exist. Create a non-default IpRange only when you need to control network segments to avoid range conflicts.

When creating AlicloudRedisCluster, the `redisTier` and `shardCount` fields are mandatory. `redisTier` sets the per-shard capacity; `shardCount` sets the number of data shards and can be changed after creation. Optionally, you can set `engineVersion` and `authSecret`.

The AliCloud instance class is derived from `redisTier` and `shardCount` at creation time. It cannot be changed afterwards, except by scaling `shardCount`.

## In-transit Encryption

In-transit encryption is always enabled. Communication with the Redis cluster requires a certificate. The certificate can be found in the Secret on the `.data.CaCert.pem` path.

Authentication is always enabled. A generated password is provided in the Secret on the `.data.authString` path.

## Persistence

Persistence is not supported. Data is not written to durable storage (i.e., data at rest).

## Redis Tiers

Each tier sets the memory capacity of a single shard. The total cluster capacity equals the per-shard capacity multiplied by `shardCount`.

| RedisTier | Capacity per shard (GiB) |
| --------- | ------------------------ |
| C3        | 4                        |
| C4        | 8                        |
| C5        | 16                       |
| C6        | 32                       |
| C7        | 64                       |

## Specification

This table lists the parameters of AlicloudRedisCluster, together with their descriptions:

| Parameter                  | Type   | Description                                                                                                                                                                                                        |
| -------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created. Immutable.                                                                        |
| **ipRange.name**           | string | Required. Name of the existing IpRange to use.                                                                                                                                                                    |
| **redisTier**              | string | Required. The per-shard capacity tier. Supported values are `C3`, `C4`, `C5`, `C6`, `C7`.                                                                                                                         |
| **shardCount**             | int    | Required. Number of data shards. Minimum value is 1, maximum value is 32. Can be changed after creation; scaling triggers an online resharding operation, which may cause a temporary increase in latency.        |
| **replicasPerShard**       | int    | Optional. Number of read-only replicas per shard. Fixed at `0` and cannot be changed. All supported `redisTier` values use proxy-based instance classes, which do not support read replicas. Defaults to `0`.     |
| **engineVersion**          | string | Optional. The version of the Redis engine. Supported values are `5.0`, `6.0`, and `7.0`. Defaults to `5.0`. Immutable; to change it, delete and recreate the cluster.                                              |
| **authSecret**             | object | Optional. Auth Secret options.                                                                                                                                                                                    |
| **authSecret.name**        | string | Optional. Auth Secret name.                                                                                                                                                                                       |
| **authSecret.labels**      | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                                   |
| **authSecret.annotations** | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                              |
| **authSecret.extraData**   | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table lists the meaningful parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                  |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It shares the name with the AlicloudRedisCluster unless `authSecret.name` is set.   |
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

