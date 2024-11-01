# AzureRedisInstance Custom Resource
The `AzureRedisInstance.cloud-resources.kyma-project.io` is a namespace-scoped custom resource (CR).
It describes the Azure Cache for Redis instance.
Once the instance is provisioned, a Kubernetes Secret with endpoint and credential details is provided in the same namespace.
By default, the created auth Secret has the same name as AzureRedisInstance.

The current implementation supports the Premium tier, which is explained in detail on the [Azure Cache for Redis overview page](https://azure.microsoft.com/en-us/products/cache).

> [!TIP] _Only for advanced cases of network topology_
> 
> Redis requires 2 IP addresses per shard. The Premium tear supports [up to 3 shards.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-high-availability)
IP addresses can be configured using the IpRange CR.
To learn more, read [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range).
Those IP addresses are allocated from the [IpRange CR](./04-10-iprange.md).
If an IpRange CR is not specified in the AzureRedisInstance, then the default IpRange is used.
If the default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

When creating AzureRedisInstance, one field is mandatory: `sku.capacity`.

Optionally, you can specify the `redisConfiguration`, `redisVersion`, `shardCount` and `redisConfiguration` fields.

> [!Note] 
> Non SSL port is disabled.

# Specification

This table lists the parameters of AzureRedisInstance, together with their descriptions:

| Parameter                                              | Type   | Description                                                                                                                                                                                                                             |
|--------------------------------------------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                                            | object | Optional. IpRange reference. If omitted, the default IpRange is used. If the default IpRange does not exist, it will be created.                                                                                                        |
| **ipRange.name**                                       | string | Required. Name of the existing IpRange to use.                                                                                                                                                                                          | 
| **sku.capacity**                                       | int    | Required. The service capacity of the instance. Supported values are 1, 2, 3, and 4. Note that tear is 'P' - Premium.                                                                                                                   |
| **redisVersion**                                       | int    | Optional. The version of Redis software. Defaults to `6.0`.                                                                                                                                                                             |
| **shardCount**                                         | int    | Optional. Number of shards.                                                                                                                                                                                                             |
| **replicasPerPrimary**                                 | int    | Optional. The number of replicas to be created per primary.                                                                                                                                                                             |
| **redisConfiguration**                                 | object | Optional. Object containing Redis configuration options.                                                                                                                                                                                |
| **redisConfiguration.maxclients**                      | int    | Optional. Max number of Redis clients. Limited to [7,500 to 40,000.](https://azure.microsoft.com/en-us/pricing/details/cache/)                                                                                                          |
| **redisConfiguration.maxmemory-reserved**              | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting) |
| **redisConfiguration.maxmemory-delta**                 | int    | Optional. Gets or sets value in megabytes reserved for non-cache usage per shard e.g. failover.                                                                                                                                         | 
| **redisConfiguration.maxmemory-policy**                | int    | Optional. The setting for how Redis will select what to remove when maxmemory (the size of the cache offering you selected when you created the cache) is reached. Defaults to `volatile-lru`.                                          | 
| **redisConfiguration.maxfragmentationmemory-reserved** | int    | Optional. [Configure your maxmemory-reserved setting to improve system responsiveness.](https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices-memory-management#configure-your-maxmemory-reserved-setting) |

# Auth Secret Details

The following table list the meaningful parameters of the auth Secret:

| Parameter                 | Type   | Description                                                                                     |
|---------------------------|--------|-------------------------------------------------------------------------------------------------|
| **.metadata.name**        | string | Name of the auth Secret. It shares the name with AzureRedisInstance unless specified otherwise. |
| **.metadata.labels**      | object | Specified custom labels (if any)                                                                |
| **.metadata.annotations** | object | Specified custom annotations (if any)                                                           |
| **.data.host**            | string | Primary connection host. Base64 encoded.                                                        |
| **.data.port**            | string | Primary connection port. Base64 encoded.                                                        |
| **.data.primaryEndpoint** | string | Primary connection endpoint. Provided in <host>:<port> format. Base64 encoded.                  |
| **.data.authString**      | string | Auth string. Base64 encoded.                                                                    |


# Custom Resource Sample

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisInstance
metadata:
  name: azureRedisInstanceExample
spec:
  redisConfiguration:
    maxclients: "8"
  redisVersion: "6.0"
  sku:
    capacity: 1
```
