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


## GcpNfsVolume CR

The `gcpnfsvolume.cloud-resources.kyma-project.io` CustomResourceDefinition (CRD) describes the GCP Filestore
instance that can be used as RWX volume in the cluster.
To learn more, read the [GcpNfsVolume custom resource documentation](./resources/04-30-10-gcp-nfs-volume.md).
