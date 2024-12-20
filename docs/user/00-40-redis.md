# Redis

## Overview

The Cloud Manager module provides managed Redis which allows you to use caching functionality directly from your cloud provider.

When you create a Redis instance in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

Cloud Manager module supports the Redis feature of three cloud providers:

* Amazon Web Services' [Amazon ElastiCashe for Redis OSS](https://aws.amazon.com/elasticache/redis/)
* Google Cloud's [Memorystore](https://cloud.google.com/memorystore?hl=en)
* Microsoft Azure's [Azure Cache for Redis](https://azure.microsoft.com/en-us/products/cache)

You can configure Cloud Manager's Redis instances using a dedicated Redis instance custom resource corresponding with the cloud provider for your Kyma cluster, namely AwsRedisInstance CR, GcpRedisInstance CR, or AzureRedisInstance CR.

## Prerequisites

To instantiate Redis, an IpRange CR must exist in the Kyma cluster. IpRange defines network address space reserved for your cloud provider's NFS resources. If you don't create the IpRange CR manually, Cloud Manager creates a default IpRange CR with the default address space and Classless Inter-Domain Routing (CIDR) selected. For more information, see [IpRange Custom Resoucre](./resources/04-10-iprange.md).

## Lifecycle

AwsRedisInstance, GcpRedisInstance, and AzureRedisInstance are namespace-level CRs. Once you create an AwsRedisInstance, GcpRedisInstance, or AzureRedisInstance resource, the following are also created automatically:

* IpRange CR
  * IpRange is a cluster-level CR.
  * Only one IpRange CR can exist per cluster.
  * If you don't want the default IpRange to be used, create one manually.
* Secret CR
  * The Secret is a namespace-level CR.
  * The Secret's name is the same as the name of the respective Redis instance CR.
  * The Secret holds values and information used to access the Redis instance.

## Related Information

* [Cloud Manager Resources: NFS](./resources/README.md#redis).
* [Tutorials](./tutorials/README.md).
* Pricing (link TBD)
