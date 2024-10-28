# AwsRedisInstance Custom Resource
The `awsredisinstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the AWS ElastiCache Redis instance.
After the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as the AwsRedisInstance, unless specified otherwise.

The current implementation creates a single node replication group with cluster mode disabled.

The AwsRedisInstance requires an `/28` IpRange. Those IP addresses are allocated from the [IpRange](./04-10-iprange.md).
If the IpRange is not specified in the AwsRedisInstance, the default IpRange is used.
If a default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified CIDR and use it only in advanced cases of network topology when you want to be in control of the network segments to avoid range conflicts with other networks.

When creating AwsRedisInstance, there is only one mandatory field: `cacheNodeType`.
It specifies the underlying machine that will be used for the cache.

As in-transit encryption is always enabled, communication with the Redis instance requires a trusted Certificate Authority (CA). You must install it on the container (e.g., using `apt-get install -y ca-certificates && update-ca-certificate`).

Optionally, you can specify the `engineVersion`, `authEnabled`, `parameters`, and `preferredMaintenanceWindow` fields.

# Specification

This table lists the parameters of AwsRedisInstance, together with their descriptions:

| Parameter                                         | Type   | Description                                                                                                                                                                                                 |
| --------------------------------------------------| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **ipRange**                                       | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                            |
| **ipRange.name**                                  | string | Required. Name of the existing IpRange to use.                                                                                                                                                              |
| **cacheNodeType**                                 | string | Required. A node is the smallest building block of an Amazon ElastiCache deployment. It is a fixed-size chunk of secure, network-attached RAM. For supported node tyes, check [Amazon's Supported node types page](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html) |
| **engineVersion**                                 | string | Optional. The version number of the cache engine to be used for the clusters in this replication group. To see all supported versions, check [Amazon's Supported ElastiCache (Redis OSS) versions page](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/supported-engine-versions.html). Defaults to `"7.0"`. |
| **authEnabled**                                   | bool   | Optional. Enables using an AuthToken (password) when issuing Redis OSS commands. Defaults to `false`.                                                                                                       |
| **readReplicas**                                  | number | Optional. Number of read replicas. If greater than zero, automatic failover is enabled to ensure the high availability of the instance. Supported values are `0` and `1`. Defaults to `0`.                       |
| **parameters**                                    | object | Optional. Provided values are passed to the Redis configuration. Supported values can be read on [Amazons's Redis OSS-specific parameters page](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/ParameterGroups.Redis.html). If left empty, defaults to an empty object. |
| **preferredMaintenanceWindow**                    | string | Optional. Defines a desired window during which updates can be applied. If not provided, maintenance events can be performed at any time during the default time window. To learn more about maintenance window limitations and requirements, see [Managing maintenance](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/maintenance-window.html). |
| **authSecret**                                    | object | Optional. Auth Secret options.                                                                                                                                                                              |
| **authSecret.name**                               | string | Optional. Auth Secret name.                                                                                                                                                                                 |
| **authSecret.labels**                             | object | Optional. Auth Secret labels. Keys and values must be a string.                                                                                                                                             |
| **authSecret.annotations**                        | object | Optional. Auth Secret annotations. Keys and values must be a string.                                                                                                                                        |

# Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                   | Type   | Description                                                                                                 |
| --------------------------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| **.metadata.name**          | string | Name of the auth Secret. It will share the name with the AwsRedisInstance unless specified otherwise        |
| **.metadata.labels**        | object | Specified custom labels (if any)                                                                            |
| **.metadata.annotations**   | object | Specified custom annotations (if any)                                                                       |
| **.data.host**              | string | Primary connection host.                                                                                    |
| **.data.port**              | string | Primary connection port.                                                                                    |
| **.data.primaryEndpoint**   | string | Primary connection endpoint. Provided in <host>:<port> format.                                              |
| **.data.authString**        | string | Auth string. Provided if authEnabled is set to true.                                                        |


# Custom Resource Sample

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsRedisInstance
metadata:
  name: awsredisinstance-sample
spec:
  cacheNodeType: cache.t2.micro
  engineVersion: "7.0"
  autoMinorVersionUpgrade: true
  authEnabled: true
  readReplicas: 1
  parameters:
    maxmemory-policy: volatile-lru
    activedefrag: "yes"
  preferredMaintenanceWindow: sun:23:00-mon:01:30
```
