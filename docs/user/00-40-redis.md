# Redis

The Cloud Manager module allows you to provision a cloud provider-managed Redis instance within your cluster network.

## Cloud Providers

When you create a Redis instance in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the Redis feature of the following cloud providers:

* Amazon Web Services [Amazon ElastiCashe for Redis OSS](https://aws.amazon.com/elasticache/redis)
* Google Cloud [Memorystore](https://cloud.google.com/memorystore?hl=en)
* Microsoft Azure [Azure Cache for Redis](https://azure.microsoft.com/en-us/products/cache)

You can configure Cloud Manager's Redis instances using a dedicated Redis instance custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely AwsRedisInstance CR, GcpRedisInstance CR, or AzureRedisInstance CR.

### Tiers

When you provision a Redis instance, you can choose between Standard or Premium Tier.

* Standard Tier offers one instance.
* Premium Tier offers high availability with automatic failover by provisioning an additional read replica of your instance.

## Prerequisites

To instantiate Redis, an IpRange CR must exist in the Kyma cluster. IpRange defines network address space reserved for your cloud provider's NFS resources. If you don't create the IpRange CR manually, Cloud Manager creates a default IpRange CR with the default address space and Classless Inter-Domain Routing (CIDR) selected. For more information, see [IpRange Custom Resoucre](./resources/04-10-iprange.md).

## Lifecycle

AwsRedisInstance, GcpRedisInstance, and AzureRedisInstance are namespace-level CRs. Once you create any of the Redis resources, the following are also created automatically:

* IpRange CR
  * IpRange is a cluster-level CR.
  * Only one IpRange CR can exist per cluster.
  * If you don't want the default IpRange to be used, create one manually.
* Secret CR
  * The Secret is a namespace-level CR.
  * The Secret's name is the same as the name of the respective Redis instance CR.
  * The Secret holds values and information used to access the Redis instance.

## Pricing

Using the Cloud Manager module and enabling Redis, introduces additional costs. For more information, see [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#calculation-with-the-cloud-manager-module).

## Related Information

* [Cloud Manager Resources: Redis](./resources/README.md#redis)
* [Tutorials](./tutorials/README.md)
* [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#calculation-with-the-cloud-manager-module)
