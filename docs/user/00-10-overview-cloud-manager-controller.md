# Cloud Manager Controller


## Overview

Cloud Manager controller role is to manage cloud resources lifecycle from the underlying cloud provider.

## IpRange CR

The `iprange.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the VPC network 
IP range that will be used for IP address allocation for cloud resources that require IP address. 
To learn more, read the [IpRange custom resource documentation](./resources/04-10-iprange.md).

## AwsNfsVolume CR

The `awsnfsvolume.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the AWS EFS 
instance that can be used as RWX volume in the cluster. 
To learn more, read the [AwsNfsVolume custom resource documentation](./resources/04-20-10-aws-nfs-volume.md).

## AwsNfsVolumeBackup CR

The `awsnfsvolumebackup.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the AWS EFS
backup.
To learn more, read the [AwsNfsVolumeBackup custom resource documentation](./resources/04-20-20-aws-nfs-volume-backup.md).

## AwsNfsVolumeRestore CR

The `awsnfsvolumerestore.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the AWS EFS
restore.
To learn more, read the [AwsNfsVolumeRestore custom resource documentation](./resources/04-20-30-aws-nfs-volume-restore.md).


## GcpNfsVolume CR

The `gcpnfsvolume.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the GCP Filestore
instance that can be used as RWX volume in the cluster.
To learn more, read the [GcpNfsVolume custom resource documentation](./resources/04-30-10-gcp-nfs-volume.md).

## GcpNfsVolumeBackup CR

The `gcpnfsvolumebackup.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the GCP filestore
backup.
To learn more, read the [GcpNfsVolumeBackup custom resource documentation](./resources/04-30-20-gcp-nfs-volume-backup.md).

## GcpNfsVolumeRestore CR

The `gcpnfsvolumerestore.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the GCP filestore
restore.
To learn more, read the [GcpNfsVolumeRestore custom resource documentation](./resources/04-30-30-gcp-nfs-volume-restore.md).

