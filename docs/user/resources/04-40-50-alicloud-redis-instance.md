# AlicloudRedisInstance Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `alicloudredisinstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Alibaba Cloud [ApsaraDB for Redis (Tair)](https://www.alibabacloud.com/help/en/redis/) instance.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is created in the same namespace. By default, the Secret has the same name as the AlicloudRedisInstance.

The instance requires an IP range, allocated from an [IpRange CR](./04-10-iprange.md). If you do not reference one, the default IpRange is used and created if it does not exist. Create a non-default IpRange only when you need to control network segments to avoid range conflicts.

When creating AlicloudRedisInstance, only the `redisTier` field is mandatory. It selects both the service tier (**Standard** or **Premium**) and the memory capacity. Optionally, you can set `engineVersion` and `authSecret`.

## In-transit Encryption

In-transit encryption is always enabled. Communication with the Redis instance requires a certificate. The certificate can be found in the Secret on the `.data.CaCert.pem` path.

Authentication is always enabled. A generated password is provided in the Secret on the `.data.authString` path.

## Persistence

Persistence is not supported. Data is not written to durable storage (i.e., data at rest).

## Redis Tiers

### Standard Tier

In the **Standard** service tier, the instance runs as a master with a replica for automatic failover. It does not have a read-only replica.

| RedisTier | Capacity (GiB) | AliCloud Instance Class |
| --------- | -------------- | ----------------------- |
| S1        | 1              | tair.rdb.1g             |
| S2        | 2              | tair.rdb.2g             |
| S3        | 4              | tair.rdb.4g             |
| S4        | 8              | tair.rdb.8g             |
| S5        | 16             | tair.rdb.16g            |

### Premium Tier

In the **Premium** service tier, the instance comes with a read-only replica in addition to the master and its failover replica. Thus, it can serve read traffic from the replica.

| RedisTier | Capacity (GiB) | AliCloud Instance Class |
| --------- | -------------- | ----------------------- |
| P1        | 4              | tair.rdb.4g             |
| P2        | 8              | tair.rdb.8g             |
| P3        | 16             | tair.rdb.16g            |
| P4        | 32             | tair.rdb.32g            |
| P5        | 64             | tair.rdb.64g            |

## Specification

This table lists the parameters of AlicloudRedisInstance, together with their descriptions:

| Parameter                  | Type   | Description                                                                                                                                                                                                        |
| -------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created. Immutable.                                                                        |
| **ipRange.name**           | string | Required. Name of the existing IpRange to use.                                                                                                                                                                    |
| **redisTier**              | string | Required. The Redis tier of the instance. Supported values are `S1`, `S2`, `S3`, `S4`, `S5` for the **Standard** offering, and `P1`, `P2`, `P3`, `P4`, `P5` for the **Premium** offering. The service tier (`S` or `P`) is immutable; only the capacity within a service tier can be changed. |
| **engineVersion**          | string | Optional. The version of the Redis engine. Supported values are `5.0`, `6.0`, and `7.0`. Defaults to `7.0`. Immutable; to change it, delete and recreate the instance.                                             |
| **authSecret**             | object | Optional. Auth Secret options.                                                                                                                                                                                    |
| **authSecret.name**        | string | Optional. Auth Secret name.                                                                                                                                                                                       |
| **authSecret.labels**      | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                                   |
| **authSecret.annotations** | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                              |
| **authSecret.extraData**   | object | Optional. Additional Secret Data entries. Keys and values must be a string. Allows users to define additional data fields that will be present in the Secret. The well-known data fields can be used as templates. The templating follows the [Golang templating syntax](https://pkg.go.dev/text/template). |

## Auth Secret Details

The following table lists the meaningful parameters of the auth Secret:

| Parameter                 | Type   | Description                                                                                                |
| ------------------------- | ------ | --------------------------------------------------------------------------------------------------------- |
| **.metadata.name**        | string | Name of the auth Secret. It shares the name with the AlicloudRedisInstance unless `authSecret.name` is set. |
| **.metadata.labels**      | object | Specified custom labels (if any).                                                                         |
| **.metadata.annotations** | object | Specified custom annotations (if any).                                                                    |
| **.data.host**            | string | Primary connection host.                                                                                  |
| **.data.port**            | string | Primary connection port.                                                                                  |
| **.data.primaryEndpoint** | string | Primary connection endpoint. Provided in `<host>:<port>` format.                                          |
| **.data.authString**      | string | Auth string used to authenticate with the instance.                                                       |
| **.data.CaCert.pem**      | string | CA Certificate that must be used for TLS.                                                                 |

## Sample Custom Resource

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AlicloudRedisInstance
metadata:
  name: alicloudredisinstance-sample
spec:
  redisTier: "P1"
  engineVersion: "7.0"
```

