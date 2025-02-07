# VPC Peering

The Cloud Manager module provides managed Virtual Private Cloud (VPC) peering functionality that allows you to peer the Kyma VPC network with a remote VPC network. Virtual network peering is possible only between networks of the same cloud providers. VPC peering in Kyma is fully automated. It means that Cloud Manager configures the peering on both Kyma's and cloud provider's side.

## Cloud Providers

When you configure VPC peering in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC Peering feature of the following cloud providers:

* Amazon Web Services [VPC peering](https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html)
* Google Cloud [VPC Network Peering](https://cloud.google.com/vpc/docs/vpc-peering)
* Microsoft Azure [Virtual network peering](https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-peering-overview)

You can configure Cloud Manager's VPC peering using a dedicated custom resource (CR) corresponding with the cloud provider for your Kyma cluster, namely AwsVpcPeering CR, GcpVpcPeering CR, or AzureVpcPeering CR.

## Prerequisites

Before you initiate VPC peering from a Kyma cluster, you must perform the following actions:

* Authorize Cloud Manager in the remote cloud provider landscape. For more information, see [Authorizing Cloud Manager in the Remote Cloud Provider](00-50-vpc-peering-authorization.md).
* Due to security reasons, to be able to use the Cloud Manager's VPC peering, you must allow SAP BTP, Kyma runtime to peer with your network. To allow the peering, tag your remote VPC network with the Kyma shoot name.
  
  > [!TIP] For more information see the Allow Kyma to Peer with Your Network sections in the VPC peering tutorials:
  > * [Create Virtual Private Cloud Peering in Amazon Web Services](./tutorials/01-30-10-aws-vpc-peering.md#allow-sap-btp-kyma-runtime-to-peer-with-your-network)
  > * [Create Virtual Private Cloud Peering in Google Cloud](./tutorials/01-30-20-gcp-vpc-peering.md#allow-sap-btp-kyma-runtime-to-peer-with-your-network)
  > * [Create Virtual Private Cloud Peering in Microsoft Azure](./tutorials/01-30-30-azure-vpc-peering.md#allow-sap-btp-kyma-runtime-to-peer-with-your-remote-network)

## Lifecycle

AwsVpcPeering CR, GcpVpcPeering CR, or AzureVpcPeering CR are cluster-level resources. Once one of the VPC peering resources is applied, the status of the VPC peering connection is reflected in that CR. The limit of the number of VPC Peering CRs per Kyma cluster depends on the quotas for each cloud provider individually.

## Related Information

* [Cloud Manager Resources: VPC Peering](./resources/README.md)
* [Tutorials](./tutorials/README.md)
