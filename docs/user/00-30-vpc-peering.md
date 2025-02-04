# VPC Peering

The Cloud Manager module provides managed Virtual Private CLoud (VPC) peering functionality which allows you to peer the Kyma VPC network with a remote VPC network. Virtual network peering is possible only between networks of the same cloud providers. VPC peering in Kyma is fully automated. It means that Cloud Manager configures the peering on both Kyma's and cloud provider's side.

## Cloud Providers

When you configure VPC peering in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC Peering feature of three cloud providers:

* Amazon Web Services [VPC peering](https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html)
* Google Cloud's [VPC Network Peering](https://cloud.google.com/vpc/docs/vpc-peering)
* Microsoft Azure's [Virtual network peering](https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-peering-overview)

## Prerequisites

To initiate VPC peering from a Kyma cluster, you must:

* Authorize Cloud Manager in the remote cloud provider landscape. For more information, see the relevant documents for:
  * Amazon Web Services: See [Authorization](./resources/04-30-10-aws-vpc-peering.md#authorization).
  * Google Cloud: See [Required Permissions in the Remote Project](./resources/04-30-20-gcp-vpc-peering.md).
  * Microsoft Azure: See [Authorization](./resources/04-30-30-azure-vpc-peering.md).
* Tag the remote network with the Kyma shoot name. For more information, see the relevant tutorials for:
  * [Create Virtual Private Cloud Peering in Amazon Web Services](./tutorials/01-30-10-aws-vpc-peering.md).
  * [Create Virtual Private Cloud Peering in Google Cloud](./tutorials/01-30-20-gcp-vpc-peering.md).
  * [Create Virtual Private Cloud Peering in Microsoft Azure](./tutorials/01-30-30-azure-vpc-peering.md).
