
# Cloud Manager Module

## What is Cloud Manager?

Cloud Manager is a central component that manages access to additional hyperscaler resources from the Kyma Runtime cluster. Its responsibility is to bring hyperscaler products/resources into the Kyma cluster in a secure way. Once Cloud Manager as a module is added in the Kyma cluster, Cloud Manager's features give you access to the respective products and resources of the hyperscaler providers.

## Features

Cloud Manager can provision the following cloud resources in the underlying cloud provider subscription:

* NFS server that can be used as a ReadWriteMany (RWX) volume in the Kyma cluster
* VPC Peering between your project and the Kyma cluster

## Architecture

Kyma Cloud Manager Operator runs in Kyma Control Plane and does remote reconciliation on Kyma clusters that
have the Cloud Manager module added. It brings various Custom Resource Definitions (CRDs) each representing some
a specific cloud resource from the underlying cloud provider subscription.

## API / Custom Resources Definitions

### IpRange CR

The `iprange.cloud-resources.kyma-project.io` CRD describes the VPC network
IP range used for IP address allocation for cloud resources that require an IP address.
To learn more, read the [IpRange Custom Resource](./resources/04-10-iprange.md) documentation.

### AwsNfsVolume CR

The `awsnfsvolume.cloud-resources.kyma-project.io` CRD describes the AWS EFS
instance that can be used as RWX volume in the cluster.
To learn more, read the [AwsNfsVolume Custom Resource](./resources/04-20-10-aws-nfs-volume.md) documentation.

### GcpNfsVolume CR

The `gcpnfsvolume.cloud-resources.kyma-project.io` CRD describes the GCP Filestore
instance that can be used as RWX volume in the cluster.
To learn more, read the [GcpNfsVolume Custom Resource](./resources/04-30-10-gcp-nfs-volume.md)  documentation.

### GcpVpcPeering CR
The `gcpvpcpeering.cloud-resources.kyma-project.io` CRD describes the VPC Peering
that you can use to peer the Kyma cluster with your Google Cloud project VPC.
To learn more, read the [GcpVpcPeering Custom Resource](./resources/04-50-gcp-vpc-peering.md) documentation.

## Related Information

To learn more about the Cloud Manager module, read the following:

* [Tutorials](./tutorials/README.md) that provide step-by-step instructions on creating, using and disposing cloud resources
