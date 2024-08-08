# AzureRedisInstance Custom Resource

The `azureredisinstance.cloud-resources.kyma-project.io` custom resource (CR) describes the Azure cache for Redis 
(https://azure.microsoft.com/en-us/products/cache)
instance that can be used as a cache in the cluster. Note that current implementation does not support backup, 
therefore using AzureRedisInstance might not be the best option for the permanent storage. Once the
Azure cache for Redis instance is provisioned in the underlying cloud provider subscription, also the corresponding
ResourceGroup is created as Azure cloud provider mandates it. The corresponding end point and secret token are provided
in the Secret resource which is created inside Kyma cluster. Secret will be named the same as the AzureRedisInstance.

A created AzureRedisInstance can be deleted only where there are no workloads that are using it.

The Azure cache for Redis `geo-location` value will be set to the same value as the Kyma cluster where you applied the manifest.

The underlying Redis tier is Premium as it supports VirtualNetwork.

## Specification <!-- {docsify-ignore} -->
This table lists the parameters of the given resource together with their descriptions. Note that this table is subject
to change, as the underlying Azure API could change with the new versions. Therefore, use it is basic info and for any
in-depth spec, please refer to MS Azure latest documentation. 

Also note that not all the parameters supported on the Azure side can be used from Kyma 
(eg `properties.redisConfiguration.aof-backup-enabled`).

**Spec:**

| Parameter                                              | Type     | Description                                                                                       |
|--------------------------------------------------------|----------|---------------------------------------------------------------------------------------------------|
| **enableNonSslPort**                                   | bool     | Specifies whether the non-ssl Redis server port (6379) is enabled. Defaults to `false`. Optional. |
| **redisVersion**                                       | string   | Redis version. Defaults to `6.0`. Optional.                                                       |
| **shardCount**                                         | quantity | The number of shards to be created on a Premium Cluster Cache. Optional.                          |
| **replicasPerPrimary**                                 | quantity | The number of replicas to be created per primary. Optional.                                       |                              
| **sku**                                                | object   | The SKU of the Redis cache to deploy. Required.                                                   |
| **sku.capacity**                                       | quantity | The size of the Redis cache to deploy. Valid values: 1, 2, 3, 4. Required.                        |
| **redisConfiguration**                                 | object   | Redis Settings. Not all settings supported on the Azure side are supported in Kyma. Optional.     |
| **redisConfiguration.maxclients**                      | string   | The max clients config. Optional.                                                                 |
| **redisConfiguration.maxfragmentationmemory-reserved** | string   | Value in megabytes reserved for fragmentation per shard. Optional.                                |
| **redisConfiguration.maxmemory-delta**                 | string   | Value in megabytes reserved for non-cache usage per shard e.g. failover. Optional.                |
| **redisConfiguration.maxmemory-policy**                | string   | The eviction strategy used when your data won't fit within its memory limit. Optional.            |
| **redisConfiguration.maxmemory-reserved**              | string   | Value in megabytes reserved for non-cache usage per shard e.g. failover. Optional.                |
| **redisConfiguration.notify-keyspace-events**          | string   | The keyspace events which should be monitored. Optional.                                          |
| **redisConfiguration.zonal-configuration**             | string   | Zonal Configuration. Optional.                                                                    |

An exemplary AzureRedisInstance custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureRedisInstance
metadata:
  name: azure-redis-example
spec:
  enableNonSslPort: false
  redisConfiguration:
    maxclients: "5"
  redisVersion: "6.0"
  sku:
    capacity: 1
```



