
# Cloud Manager Module

Use the Cloud Manager module to manage cloud providers' resources from a Kyma cluster.

## What is Cloud Manager?

Cloud Manager is strictly coupled with the cloud provider where your Kyma cluster is deployed. The module manages access to the chosen resources and products of that particular cloud provider. Once you add Cloud Manager to your Kyma cluster, the module brings the offered resources in a secure way.

## Scope

Cloud Manager supports Amazon Web Services (AWS), Google Cloud, and Microsoft Azure (Azure) as cloud providers for SAP BTP, Kyma runtime.

## Features

The Cloud Manager module provides the following features tailored for each of the cloud providers.

* [NFS](./00-20-nfs.md): Network File System (NFS) server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster.
* [VPC peering](./00-30-vpc-peering.md): Virtual Private Cloud (VPC) peering between your Kyma runtime and remote cloud provider's project, account, or subscription.
* [Redis](./00-40-redis.md): cloud provider-flavored cache that can be used in your Kyma cluster.

> [!NOTE]
> The NFS feature is offered for Amazon Web Service and Google Cloud only.

## Architecture

## API / Custom Resources Definitions

The `cloud-resources.kyma-project.io` Custom Resource Definition (CRD) describes the kind and the format of data that Cloud Manager` uses to configure resources. For more information, see [Cloud Manager Resources](./resources/README.md) (CRs).

## Related Information

* [Cloud Manager module tutorials](./tutorials/README.md) provide step-by-step instructions on creating, using and disposing cloud resources.
