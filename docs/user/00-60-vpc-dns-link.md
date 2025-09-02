# VPC DNS Link

VPC DNS Link in the Cloud Manager module enables linking Kyma Virtual Private Cloud (VPC) network to remote private DNS zone.

VPC DNS Link is possible only between networks and remote private DNS of the same cloud providers. VPC DNS Link in Kyma is fully automated. It means that Cloud Manager configures the link in specified remote private DNS zone.


## Cloud Providers

When you configure VPC DNS Link in SAP BTP, Kyma runtime, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC DNS Link feature of the following cloud providers:

* Microsoft Azure [Virtual Network Link](https://learn.microsoft.com/en-us/azure/dns/private-dns-virtual-network-links) <!-- VPC DNS Link for Microsoft Azure is not part of external Help Portal docs-->

You can configure Cloud Manager's VPC DNS Link using a dedicated custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely:

* AzureVpcDnsLink CR <!-- VPC DNS Link for Microsoft Azure is not part of external Help Portal docs-->

For more information, see [VPC DNS Link Resources](./resources/README.md#vpc-dns-link-resources).

## Prerequisites

Before you create VPC DNS Link from a Kyma cluster, you must perform the following actions:

* Authorize Cloud Manager in the remote cloud provider landscape. For more information, see [Authorizing Cloud Manager in the Remote Cloud Provider](00-31-vpc-peering-authorization.md).
* Tag the remote network with the Kyma shoot name. For more information, see the following tutorials:
    * [Allow SAP BTP, Kyma Runtime to link with Your Private DNS zone](./tutorials/01-40-10-azure-vpc-dns-link.md#allow-sap-btp-kyma-runtime-to-link-with-your-private-dns-zone) in Create VPC DNS Link in Microsoft Azure.


## Lifecycle

VPC DNS Link CRs are cluster-level resources. Once a VPC DNS Link resource is applied, the status of the VPC DNS Link is reflected in that CR. 

When you delete a VPC DNS Link CR, the VPC link in remote cloud provider landscape is deleted automatically.

### Limitations

The limit on the number of VPC DNS Link CRs per Kyma cluster depends on the quotas for each cloud provider.

## Related Information

* [Cloud Manager Resources: VPC DNS Link](./resources/README.md#vpc-dns-link-resources)
