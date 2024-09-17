# GcpRedisInstance Custom Resource
The `gcpredisinstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Google Memorystore Redis instance.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the GcpRedisInstance, unless specified otherwise.

The current implementation supports *Basic* and *Standard(without replicas)* tiers, which are explained in detail on the [Google's Memorystore for Redis overview page](https://cloud.google.com/memorystore/docs/redis/memorystore-for-redis-overview).

Redis requires a `/28` ip range.
To learn more, read [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range).
Those IP addresses are allocated from the [IpRange CR](./04-10-iprange.md).
If an IpRange CR is not specified in the GcpRedisInstance, then the default IpRange is used.
If the default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating GcpRedisInstance, two fields are mandatry: `memorySizeGb`, and `tier`.

Optionally, you can specify the `redisVersion`, `authEnabled`, `transitEncryption`, `redisConfigs`, and `maintenancePolicy` fields.

# Specification

This table lists the parameters of GcpRedisInstance, together with their descriptions:

| Parameter                                         | Type   | Description                                                                                                                                                                                                 |
| --------------------------------------------------| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                                       | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                            |
| **ipRange.name**                                  | string | Required. Name of the existing IpRange to use.                                                                                                                                                              |
| **tier**                                          | string | Required. The service tier of the instance. Supported values are `BASIC` and `STANDARD_HA`                                                                                                                  |
| **memorySizeGb**                                  | int    | Required. Redis memory size in GiB.                                                                                                                                                                         |
| **redisVersion**                                  | int    | Optional. The version of Redis software. Defaults to `REDIS_7_0`.                                                                                                                                           |
| **authEnabled**                                   | bool   | Optional. Indicates whether OSS Redis AUTH is enabled for the instance. If set to `true,` AUTH is enabled on the instance. Defaults to `true`                                                                |
| **transitEncryption**                             | object | Optional. Defines the way TLS is executed. Currently supports only one mode. If not provided, in-transit encryption is disabled. To learn more about in-transit encryption, see [About in-transit encryption](https://cloud.google.com/memorystore/docs/redis/about-in-transit-encryption). |
| **transitEncryption.serverAuthentication**        | bool   | Optional. If `true`, sets in-transit encryption mode to server authentication. If not provided, defaults to `false`.                                                                                            |
| **redisConfigs**                                  | object | Optional. Provided values are passed to the Redis configuration. Supported values can be read on [Google's Supported Redis configurations page](https://cloud.google.com/memorystore/docs/redis/supported-redis-configurations). If left empty, defaults to an empty object. |
| **maintenancePolicy**                             | object | Optional. Defines a desired maintenance policy. Only one policy can be active at a time.  If not provided, maintenance events can be performed at any time. To learn more about maintenance policy limitations and requirements, see [About maintenance on Memorystore for Redis](https://cloud.google.com/memorystore/docs/redis/about-maintenance). |
| **maintenancePolicy.dayOfWeek**                   | object | Optional. Defines maintenance policy to a specific day.                                                                                                                                                     |
| **maintenancePolicy.dayOfWeek.day**               | string | Required. The day of the week that maintenance updates occur. Supported values are `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`.                                            |
| **maintenancePolicy.dayOfWeek.startTime**         | object | Required. Defines the start time of the policy in UTC time.                                                                                                                                                 |
| **maintenancePolicy.dayOfWeek.startTime.hours**   | int    | Required. Hours of day in 24-hour format. Accepts values from 0 to 23                                                                                                                                       |
| **maintenancePolicy.dayOfWeek.startTime.minutes** | int    | Required. Minutes of an hour of the day. Accepts values from 0 to 59.                                                                                                                                       |
| **authSecret**                                    | object | Optional. Auth Secret options.                                                                                                                                                                              |
| **authSecret.name**                               | string | Optional. Auth Secret name.                                                                                                                                                                                 |
| **authSecret.labels**                             | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                             |
| **authSecret.annotations**                        | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                        |

# Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                 |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It will share the name with the GcpRedisInstance unless specified otherwise        |
| **.metadata.labels**        | object | Specified custom labels (if any)                                                                            |
| **.metadata.annotations**   | object | Specified custom annotations (if any)                                                                       |
| **.data.host**              | string | Primary connection host.                                                                                    |
| **.data.port**              | string | Primary connection port.                                                                                    |
| **.data.primaryEndpoint**   | string | Primary connection endpoint. Provided in <host>:<port> format.                                              |
| **.data.authString**        | string | Auth string. Provided if authEnabled is set to true.                                                        |
| **.data.CaCert.pem**        | string | CA Certificate that must be used for TLS. Provided if transit encryption is set to server authentication.   |


# Custom Resource Sample

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpRedisInstance
metadata:
  name: gcpredisinstance-sample
spec:
  memorySizeGb: 5
  tier: "STANDARD_HA"
  redisVersion: REDIS_7_0
  authEnabled: true
  transitEncryption:
    serverAuthentication: true
  redisConfigs:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  maintenancePolicy:
    dayOfWeek:
      day: "TUESDAY"
      startTime:
          hours: 15
          minutes: 45
```
