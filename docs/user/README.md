# Kyma Cloud Manager Module

Cloud Manager is a central component that manages access to additional Hyperscaler resources from the Kyma Runtime 
cluster. Its responsibility is to bring Hyperscaler products/resources onto the Kyma cluster in a secure way. 
Once Cloud Manager as a module is enabled in the Kyma cluster, Cloud Manager's features will give you access to the 
respective products and resources of the Hyperscaler providers.

## Features

### NFS Storage

Cloud Manager provides a feature to provision NFS (Network File System) instance in the underlying cloud provider.
Provisioned NFS instance can then be used as a K8s Persistent Volume. 
Many use cases require the capability to mount the volume as read-write by multiple Nodes (`ReadWriteMany`).
Depending on the cloud provider, a different custom resource is provided to support this use case, 
with different cloud provider based limitations. 
More details (separated per cloud providers) in the sections below. 

#### GCP

Custom resource for provisioning the GCP NFS Volume is `GcpNfsVolume`.
During creation of `GcpNfsVolume`, `PersistentVolume` and `PersistentVolumeClaim` will be created automatically.
`PersistentVolumeClaim` can then be used in your Pods to access the underlying NFS.

By default, `PersistentVolume` and `PersistentVolumeClaim` will have the same name as the `GcpNfsVolume`, 
unless specified otherwise (see the example below).

List of `NfsVolume` parameters:

| Param                        | Required | Description                                                                                           |
|------------------------------|----------|-------------------------------------------------------------------------------------------------------|
| spec.capacityGb              | Yes      | Size of the cloud volume. The minimum size is 1,024 GiB (1 TiB).                                      |
| spec.ipRange                 | No       | Name of existing IP Range. If not specified, default will be used.                                    |
| spec.location                | No       | [Filestore instances live in zones within regions.]( https://cloud.google.com/filestore/docs/regions) |
| spec.tier                    | No       | [Service tier.](https://cloud.google.com/filestore/docs/performance)                                  |
| spec.volume.name             | No       | Name of the PersistentVolume that will be created.                                                    |
| spec.volume.labels           | No       |                                                                                                       |
| spec.volume.annotations      | No       |                                                                                                       |
| spec.volumeClaim.name        | No       | Name of the PersistentVolumeClaim that will be created.                                               |
| spec.volumeClaim.labels      | No       |                                                                                                       |
| spec.volumeClaim.annotations | No       |                                                                                                       |

> **WARNING**
> Be careful when naming the `volume` and `volumeClaim` resources. 
> Referencing the ones that are already in use will cause the error.
> Additionally, be aware that `PersistentVolume` is a cluster wide resource, while NFS and `PersitentVolumeClaim` are namespaced.

Example 
```
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: nfs-volume-demo
  namespace: demo
spec:
  capacityGb: 1024
  ipRange:
    name: demo-iprange
    namespace: demo
  location: us-central1-a
  tier: BASIC_HDD
  volume:
    name: persistent-volume-demo
  volumeClaim:
    name: persistent-claim-volume-demo
```

More info (including the limitations) can be found at official GCP [doc page](https://cloud.google.com/architecture/partners/netapp-cloud-volumes/creating-nfs-volumes).

#### AWS

Custom resource for provisioning the GCP NFS Volume is `AwsNfsVolume`.
During creation of `AwsNfsVolume`, `PersistentVolume` and `PersistentVolumeClaim` will be created automatically.
`PersistentVolumeClaim` can then be used in your Pods to access the underlying NFS.

By default, `PersistentVolume` and `PersistentVolumeClaim` will have the same name as the `AwsNfsVolume`,
unless specified otherwise (see the example below).

List of `NfsVolume` parameters:

| Param                        | Required | Description                                                                       |
|------------------------------|----------|-----------------------------------------------------------------------------------|
| spec.capacity                | Yes      | Size of the cloud volume.                                                         |
| spec.ipRange                 | No       | Name of existing IP Range. If not specified, default will be used.                |
| spec.performanceMode         | No       | [PerformanceMode.]( https://docs.aws.amazon.com/efs/latest/ug/performance.html)   |
| spec.throughput              | No       | [ThroughputMode.](https://docs.aws.amazon.com/efs/latest/ug/performance.html)     |
| spec.volume.name             | No       | Name of the PersistentVolume that will be created.                                |
| spec.volume.labels           | No       |                                                                                   |
| spec.volume.annotations      | No       |                                                                                   |
| spec.volumeClaim.name        | No       | Name of the PersistentVolumeClaim that will be created.                           |
| spec.volumeClaim.labels      | No       |                                                                                   |
| spec.volumeClaim.annotations | No       |                                                                                   |

> **WARNING**
> Be careful when naming the `volume` and `volumeClaim` resources.
> Referencing the ones that are already in use will cause the error.
> Additionally, be aware that `PersistentVolume` is a cluster wide resource, while NFS and `PersitentVolumeClaim` are namespaced.

Example
```
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolume
metadata:
  name: aws-nfs-volume-demo
  namespace: demo 
spec:
  capacity: 10G
  volume:
    labels:
      foo: bar
    annotations:
      baz: qux
    name: volume-demo
  volumeClaim:
    name: volume-claim-demo
    labels:
      foo: bar
    annotations:
      baz: qux
```

More info (including the limitations) can be found at official AWS [doc page](https://docs.aws.amazon.com/filegateway/latest/files3/CreatingAnNFSFileShare.html).


#### Azure

Azure is not supported as this Hyperscaler has the NFS `ReadWriteMany` feature already available.