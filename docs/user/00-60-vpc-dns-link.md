# VPC DNS Link

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

VPC DNS Link in the Cloud Manager module enables linking the Kyma Virtual Private Cloud (VPC) network to either a remote DNS private zone or a DNS private resolver.

VPC DNS Link is possible only between networks and the remote private DNS zones or DNS private resolvers of the same cloud providers. VPC DNS Link in SAP BTP, Kyma runtime is fully automated. It means that Cloud Manager configures the link in the specified remote private DNS zone or DNS private resolver.

## Cloud Providers

When you configure VPC DNS Link in SAP BTP, Kyma runtime, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC DNS Link feature of the following cloud providers:

* Microsoft Azure [virtual network links](https://learn.microsoft.com/en-us/azure/dns/private-dns-virtual-network-links) and [ruleset links](https://learn.microsoft.com/en-us/azure/dns/private-resolver-endpoints-rulesets#ruleset-links) <!-- VPC DNS Link for Microsoft Azure is not part of external Help Portal docs-->

You can configure Cloud Manager's VPC DNS Link using a dedicated custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely:

* AzureVpcDnsLink CR <!-- VPC DNS Link for Microsoft Azure is not part of external Help Portal docs-->

For more information, see [VPC DNS Link Resources](./resources/README.md#vpc-dns-link-resources).

## Prerequisites

Before you create VPC DNS Link from a Kyma cluster, you must perform the following actions:

* Authorize Cloud Manager in the remote cloud provider landscape. For more information, see [Authorizing Cloud Manager in the Remote Cloud Provider](00-31-vpc-peering-authorization.md).
* Tag the remote network with the Kyma shoot name. For more information, see the following tutorials:
    * [Allow SAP BTP, Kyma Runtime to Link to Your Private DNS Zone](./tutorials/01-60-10-azure-dns-zone-vpc-link.md#allow-sap-btp-kyma-runtime-to-link-to-your-private-dns-zone) in Linking Your Kyma Network to Microsoft Azure Private DNS Zone.
    * [Allow SAP BTP, Kyma Runtime to Link to Your DNS Private Resolver](./tutorials/01-60-20-azure-dns-resolver-vpc-link.md#allow-sap-btp-kyma-runtime-to-link-to-your-dns-private-resolver) in Linking Your Kyma Network to Microsoft Azure DNS Private Resolver.


## Lifecycle

VPC DNS Link CRs are cluster-level resources. Once a VPC DNS Link resource is applied, the status of the VPC DNS Link is reflected in that CR. 

When you delete a VPC DNS Link CR, the VPC link in the remote cloud provider landscape is deleted automatically.

### Limitations

The limit on the number of VPC DNS Link CRs per Kyma cluster depends on the quotas for each cloud provider.

## Related Information

* [Cloud Manager Resources: VPC DNS Link](./resources/README.md#vpc-dns-link-resources)
