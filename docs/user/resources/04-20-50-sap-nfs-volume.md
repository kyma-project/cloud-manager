# SapNfsVolume Custom Resource

> [!WARNING]
> This is a beta feature available only per request for SAP-internal teams.

The `sapnfsvolume.cloud-resources.kyma-project.io` custom resource (CR) describes an NFS volume that can be provisioned and used as a ReadWriteMany (RWX) volume in OpenStack environments. Once the NFS volume is provisioned in the underlying OpenStack infrastructure, the corresponding PersistentVolume (PV) and PersistentVolumeClaim (PVC) are created in RWX mode, allowing them to be used from multiple cluster workloads simultaneously. To use the SapNfsVolume CR as a volume in the cluster workload, specify the workload volume of the `persistentVolumeClaim` type. A created SapNfsVolume can be deleted only when there are no workloads using it and when the PV and PVC are unbound.

The SapNfsVolume requires IP addresses allocated from an [IpRange](./04-10-iprange.md). If an IpRange is not specified in the SapNfsVolume, then the default IpRange is used. If a default IpRange does not exist, it is automatically created. Manually create a non-default IpRange with a specified CIDR and use it only in advanced cases of network topology when you want to control the network segments to avoid range conflicts with other networks.

You must specify the capacity of the SapNfsVolume using the `capacityGb` field, which defines the storage capacity in gigabytes.

By default, the created PV and PVC have the same name as the SapNfsVolume resource, but you can optionally specify their names, labels, and annotations if needed. If a PV or PVC already exists with the same name as the one being created, the provisioned NFS volume remains and the SapNfsVolume is put into the `Error` state.

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                                                                                                                         |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                 | object              | Optional IpRange reference. If omitted, default IpRange will be used, if default IpRange does not exist, it will be created                                                                                                         |
| **ipRange.name**            | string              | Name of the existing IpRange to use.                                                                                                                                                                                                |
| **capacityGb** (required)   | int                 | Capacity of the requested volume in GiB. Must be greater than zero.                                                                                                                                                                |
| **volume**                  | object              | The PersistentVolume options. Optional.                                                                                                                                                                                             |
| **volume.name**             | string              | The PersistentVolume name. Optional. Defaults to the SapNfsVolume status ID.                                                                                                                                                       |
| **volume.labels**           | map\[string\]string | The PersistentVolume labels. Optional. Defaults to nil.                                                                                                                                                                             |
| **volume.annotations**      | map\[string\]string | The PersistentVolume annotations. Optional. Defaults to nil.                                                                                                                                                                        |
| **volumeClaim**             | object              | The PersistentVolumeClaim options. Optional.                                                                                                                                                                                        |
| **volumeClaim.name**        | string              | The PersistentVolumeClaim name. Optional. Defaults to the name of the SapNfsVolume resource.                                                                                                                                        |
| **volumeClaim.labels**      | map\[string\]string | The PersistentVolumeClaim labels. Optional. Defaults to nil.                                                                                                                                                                        |
| **volumeClaim.annotations** | map\[string\]string | The PersistentVolumeClaim annotations. Optional. Defaults to nil.                                                                                                                                                                   |

**Status:**

| Parameter                         | Type       | Description                                                                                                                                                         |
|-----------------------------------|------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **id**                            | string     | The unique identifier of the provisioned NFS volume in the OpenStack environment.                                                                                  |
| **server**                        | string     | The server address of the provisioned NFS volume.                                                                                                                  |
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`.                                  |
| **capacity**                      | quantity   | The provisioned capacity of the volume. This matches the spec.capacityGb upon completion of provisioning.                                                          |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                                                                |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                                                               |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                                                            |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                                                 |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                                                          |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                                                      |

## Sample Custom Resource

See an exemplary SapNfsVolume custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: SapNfsVolume
metadata:
  name: my-sap-nfs-vol
  namespace: default
spec:
  capacityGb: 1000
```
