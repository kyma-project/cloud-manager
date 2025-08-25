# Redis Cluster

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

Use the Cloud Manager module to provision a Redis cluster. The Cloud Manager module allows you to provision a cloud provider-managed Redis cluster in cluster mode within your cluster network.

> [!NOTE]
> Using the Cloud Manager module and enabling Redis, introduces additional costs. For more information, see [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#loioc33bb114a86e474a95db29cfd53f15e6__section_cloud_manager).

## Cloud Providers

When you create a Redis cluster in SAP BTP, Kyma runtime, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the Redis cluster feature of the following cloud providers:

* Amazon Web Services [Amazon ElastiCashe for Redis OSS](https://aws.amazon.com/elasticache/redis)
* Google Cloud [Memorystore](https://cloud.google.com/memorystore?hl=en)
* Microsoft Azure [Azure Cache for Redis](https://azure.microsoft.com/en-us/products/cache)

You can configure Cloud Manager's Redis clusters using a dedicated Redis cluster custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely AwsRedisCluster CR, GcpRedisCluster CR, or AzureRedisCluster CR. For more information, see [Redis Resources](./resources/README.md#redis-cluster-resources).

### Tiers

The offered tires depends on the cloud provider. Cloud Manager doesn't offer all machine types available at the cloud provoders, but a dedicated subset.

## Prerequisites

To instantiate Redis cluster:

* For Amazon Web Service and Microsoft Azure, an IpRange CR must exist in the Kyma cluster. IpRange defines network address space reserved for your cloud provider's Redis resources. If you don't create the IpRange CR manually, Cloud Manager creates a default IpRange CR with the default address space and Classless Inter-Domain Routing(CIDR) selected. For more information, see [IpRange Custom Resoucre](./resources/04-10-iprange.md).
* For Google Cloud, a GcpSubnet CR must exist in the Kyma cluster. GcpSubnet specifies the VPC Network Subnet. Once needed, the resource is automatically created with the hardcoded CIDR `10.251.0.0/22`. For most use cases, this automatic allocation is sufficient. For more information, see [GcpSubnet Custom Resource](./resources/04-50-21-gcp-subnet.md).

## Lifecycle

AwsRedisCluster, GcpRedisCluster, and AzureRedisCluster are namespace-level CRs. Once you create any of the Redis cluster resources, the following resources are also created automatically:

* IpRange CR / GcpSubnet
  * IpRange and GcpSubnet are cluster-level CRs.
  * Only one IpRange CR or GcpSubnet CR can exist per cluster.
  * If you don't want the default IpRange or GcpSubnet to be used, create one manually.
* Secret CR
  * The Secret is a namespace-level CR.
  * The Secret's name is the same as the name of the respective Redis cluster CR.
  * The Secret holds values and information used to access the Redis cluster.

## Related Information

* [Using AwsRedisCluster Custom Resources](./tutorials/01-50-10-aws-redis-cluster.md)
* [Using GcpRedisCluster Custom Resources](./tutorials/01-50-20-gcp-redis-cluster.md)
* [Using AzureRedisCluster Custom Resources](./tutorials//01-50-30-azure-redis-cluster.md)
* [Cloud Manager Resources: Redis Cluster](./resources/README.md#redis-cluster-resources)
* [Calculation with the Cloud Manager Module](https://help.sap.com/docs/btp/sap-business-technology-platform-internal/commercial-information-sap-btp-kyma-runtime?state=DRAFT&version=Internal#calculation-with-the-cloud-manager-module)
