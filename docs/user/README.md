
# Cloud Manager Module

Use the Cloud Manager module to manage infrastructure providers' resources from a Kyma cluster.

## What is Cloud Manager?

Cloud Manager manages access to infrastructure providers' resources and products in SAP BTP, Kyma runtime. Once you add the Cloud Manager module to your Kyma cluster, it brings the resources in a secure way.

## Features

The Cloud Manager module provides the following feature:

* Network File System (NFS) server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster.
* Virtual Private Cloud (VPC) peering between your project and the Kyma cluster.
* Google Cloud and Amazon Web Services Redis offering.

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

The cloud-resources.kyma-project.io Custom Resource Definition (CRD) describes the kind and the format of data that Cloud Manager uses to configure resources.

See the documentation related to the [Cloud Manager custom resources](./resources/README.md) (CRs):

* IP Range
  * [IpRange](./resources/04-10-iprange.md)
* NFS
  * [AwsNfsVolume](./resources/04-20-10-aws-nfs-volume.md)
  * [AwsNfsVolumeBackup](./resources/04-20-11-aws-nfs-volume-backup.md)
  * [AwsNfsBackupSchedule](./resources/04-20-12-aws-nfs-backup-schedule.md)
  * [AwsNfsVolumeRestore](./resources/04-20-13-aws-nfs-volume-restore.md)
  * [GcpNfsVolume](./resources/04-20-20-gcp-nfs-volume.md)
  * [GcpNfsVolumeBackup](./resources/04-20-21-gcp-nfs-volume-backup.md)
  * [GcpNfsBackupSchedule](./resources/04-20-22-gcp-nfs-backup-schedule.md)
  * [GcpNfsVolumeRestore](./resources/04-20-23-gcp-nfs-volume-restore.md)
* VPC peering
  * [AzureVpcPeering](./resources/04-30-10-azure-vpc-peering.md)
  * [AwsVpcPeering](./resources/04-30-20-aws-vpc-peering.md)
  * [GcpVpcPeering](./resources/04-30-30-gcp-vpc-peering.md)
* Redis
  * [GcpRedisInstance](./resources/04-40-10-gcp-redis-instance.md)
  * [AwsRedisInstance](./resources/04-40-20-aws-redis-instance.md)
  * [AzureRedisInstance](./resources/04-40-30-azure-redis-instance.md)

## Resource Consumption

To learn more about the resources used by the Cloud Manager module, see [Kyma Modules' Sizing](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-modules-sizing?version=Cloud).

## Related Information

* [Cloud Manager module tutorials](./tutorials/README.md) provide step-by-step instructions on creating, using and disposing cloud resources.
