
# Cloud Manager Module

## What is Cloud Manager?

Cloud Manager is a central component that manages access to additional hyperscaler resources from the Kyma Runtime cluster. Its responsibility is to bring hyperscaler products/resources into the Kyma cluster in a secure way. Once Cloud Manager as a module is added to the Kyma cluster, Cloud Manager's features give you access to the respective products and resources of the infrastructure providers.

## Features

Cloud Manager can provision the following cloud resources in the underlying cloud provider subscription:

* Network File System (NFS) server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster
* Virtual Private Cloud (VPC) peering between your project and the Kyma cluster

## Architecture

Kyma Cloud Manager Operator runs in Kyma Control Plane and does remote reconciliation on Kyma clusters that
have the Cloud Manager module added. It brings various Custom Resource Definitions (CRDs) each representing
a specific cloud resource from the underlying cloud provider subscription.

## API / Custom Resources Definitions

For more information on Cloud Manager's API and custom resources (CRs), see the [`/resources`](./resources/README.md) directory.

## Related Information

To learn more about the Cloud Manager module, read the following:

* [Tutorials](./tutorials/README.md) that provide step-by-step instructions on creating, using and disposing cloud resources
