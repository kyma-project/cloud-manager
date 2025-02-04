# VPC Peering

The Cloud Manager module provides managed Virtual Private CLoud (VPC) peering functionality which allows you to peer the Kyma VPC network with a remote VPC network. Virtual network peering is possible only between networks of the same cloud providers. VPC peering in Kyma is fully automated. It means that Cloud Manager configures the peering on both Kyma's and cloud provider's side.

## Cloud Providers

When you configure VPC peering in Kyma, you depend on the cloud provider of your Kyma cluster. The cloud provider in use determines the exact implementation.

The Cloud Manager module supports the VPC Peering feature of three cloud providers:
