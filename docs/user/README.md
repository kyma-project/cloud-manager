
# Cloud Manager Module

Use the Cloud Manager module to manage infrastructure providers' resources from a Kyma cluster.

## What is Cloud Manager?

Cloud Manager manages access to infrastructure providers' resources and products in SAP BTP, Kyma runtime. Once you add the Cloud Manager module to your Kyma cluster, it brings the resources in a secure way.

## Features

The Cloud Manager module provides the following feature:

* Network File System (NFS) server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster.
* Virtual Private Cloud (VPC) peering between your project and the Kyma cluster.
* Google Cloud, Amazon Web Services, and Azure Redis offering.

## Architecture

Kyma Cloud Manager Operator runs in Kyma Control Plane (KCP) and does remote reconciliation on Kyma clusters that have the Cloud Manager module added. It brings various Custom Resource Definitions (CRDs) each representing a specific cloud resource from the underlying infrastructure provider subscription.

Cloud Manager defines two API groups:

* `cloud-resources.kyma-project.io`: user-facing API available in SAP BTP, Kyma runtime
* `cloud-control.kyma-project.io`: low-level API in KCP projected from the user-facing API in SAP BTP, Kyma runtime

![API and Reconcilers](../contributor/architecture/assets/api-and-reconcilers.drawio.svg "API and Reconcilers")

Cloud Manager has one central active component running in KCP that runs two sets of reconciliation loops - one for each API group.

Cloud Resources reconcilers remotely reconcile the `cloud-resources.kyma-project.io` API group from SAP BTP, Kyma runtime into the low-level cloud-control.kyma-project.io API group in KCP.

Cloud Control reconcilers locally reconcile the `cloud-control.kyma-project.io` API group into the resources of an infrastructure provider.

Both sets of reconcilers must also maintain the status of the resources they reconcile. This means that Cloud Control reconcilers observe the status of the Cloud Resources resource group and project it into the status of the Cloud Control resources in KCP. At the same time, Cloud Resources reconcilers observe the status of the Cloud Control resources and project it into the status of the Cloud Control resource group in SKR.

### Cloud Manager Operator

### Cloud Control Reconcilers

### Cloud Resources Reconcilers

### Controller Runtime Manager

### Custom Manager

## API / Custom Resources Definitions

The `cloud-resources.kyma-project.io` Custom Resource Definition (CRD) describes the kind and the format of data that Cloud Manager` uses to configure resources.

See the documentation related to the [Cloud Manager custom resources](./resources/README.md) (CRs).

## Related Information

* [Cloud Manager module tutorials](./tutorials/README.md) provide step-by-step instructions on creating, using and disposing cloud resources.
