# VPC Peering

The Cloud Manager module provides managed Virtual Private CLoud (VPC) peering functionality which allows you to peer the Kyma VPC network with a remote VPC network. Virtual network peering is possible only between networks of the same cloud providers. VPC Peering in Kyma is fully automated. It means that Cloud Manager configures the peering on both Kyma's and cloud provider's side.

## Cloud Providers

When you configure VPC peering in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC Peering feature of three cloud providers:

* [Amazon Web Services VPC peering](https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html)
* [Google Cloud VPC Network Peering](https://cloud.google.com/vpc/docs/vpc-peering)
* [Microsoft Azure Virtual network peering](https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-peering-overview)

You can configure Cloud Manager's VPC peering using a dedicated custom resource corresponding with the cloud provider for your Kyma cluster, namely AwsVpcPeering CR, GcpVpcPeering CR, or AzureVpcPeering CR.

## Prerequisites

Initiating VPC peering from a Kyma cluster requires:

* Appropriate authorization and permissions assigned to Cloud Manager in the remote cloud provider network. For more informatiom, see the relevant documents for:
  * Amazon Web Services: See [Authorization](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/resources/04-30-10-aws-vpc-peering.md#authorization).
  * Google Cloud: See [Required Permissions in the Remote Project](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/resources/04-30-20-gcp-vpc-peering.md).
  * Microsoft Azure: See [Authorization](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/resources/04-30-30-azure-vpc-peering.md).
* The remote network tagged with the Kyma shoot name. For more information, see the relevant tutorials for:
  * [Create Virtual Private Cloud Peering in Amazon Web Services](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/tutorials/01-30-10-aws-vpc-peering.mdn).
  * [Create Virtual Private Cloud Peering in Google Cloud](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/tutorials/01-30-20-gcp-vpc-peering.md).
  * [Create Virtual Private Cloud Peering in Microsoft Azure](https://github.com/kyma-project/cloud-manager/blob/main/docs/user/tutorials/01-30-30-azure-vpc-peering.md).

## Lifecycle

AwsVpcPeering CR, GcpVpcPeering CR, or AzureVpcPeering CR are cluster-level resources. Once one of the VPC peering resources is applied, the status of the VPC peering connection is reflected in that CR itself. The limit of the number of VPC Peering CRs per Kyma clusrer depends on the quotas for each cloud provider. More info about the quotas and limits could be found in the links provided above [TBD - I could not find the details.]

## Related Information

* [Cloud Manager Resources: VPC Peering](./resources#vpc-peering)
* [Tutorials](./tutorials/README.md)