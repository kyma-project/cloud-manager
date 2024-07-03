# GcpNfsVolume Custom Resource

The `gcpnfsvolume.cloud-resources.kyma-project.io` custom resource describes the GCP Filestore
instance that can be used as RWX volume in the cluster. Once the GCP Filestore instance is provisioned
in the underlying cloud provider subscription, also the corresponding PersistentVolume and
PersistentVolumeClaim are created in RWX mode, so they can be used from multiple cluster workloads.
To use it as a volume in the cluster workload, specify the workload volume of the `persistentVolumeClaim` type.
A created GcpNfsVolume can be deleted only where there are no workloads that
are using it, and when PV and PVC are unbound.

The zonal GCP Filestore requires 8 and regional GCP Filestore requires 64 IP address - to learn
more read [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range). 
Those IP addresses are
allocated from the [IpRange](./04-10-iprange.md). If IpRange is not specified in the GcpNfsVolume
then the default IpRange will be used. If default IpRange does not exist, it will be automatically created.
Manually create a non-default IpRange with specified CIDR and use it only in advanced cases of network topology
when you want to be in control of the network segments in order to avoid range conflicts with other networks.

You must specofy the GcpNfsVolume capacity. Minimum capacity depends on the chosen GCP Filestore tier.
To learn more read [GCP Filestore service tiers](https://cloud.google.com/filestore/docs/service-tiers).

You can optionally specify the `Tier` GCP Filestore configuration options. It's default value is `BASIC_HDD`, 
and can be one of `STANDARD`, `PREMIUM`, `BASIC_HDD`, `BASIC_SSD`, `HIGH_SCALE_SSD`, `ENTERPRISE`, `ZONAL`, `REGIONAL`.

By default, the created PV and PVC will have the same name as the GcpNfsVolume resource, but you can optionally
specify their names, labels and annotations if needed. If PV or PVC already exist with name equal to the one
being created, the provisioned GCP Filestore will remain and the GcpNfsVolume will be put to the Error state.

## Specification <!-- {docsify-ignore} -->
This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                                                                    |
|-----------------------------|---------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                 | object              | Optional IpRange reference. If omitted, default IpRange will be used, if default IpRange does not exist, it will be created                                                    |
| **ipRange.name**            | string              | Name of the existing IpRange to use.                                                                                                                                           |
| **capacityGb**              | int                 | Capacity of the provisioned volume.                                                                                                                                            |
| **tier**                    | string              | The GCP Filestore tier configuration option. One of `generalPurpose`, `maxIO`. Defaults to `generalPurpose`.                                                                   |
| **volume**                  | object              | The PersistentVolume options. Optional.                                                                                                                                        |
| **volume.name**             | string              | The PersistentVolume name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                        |
| **volume.labels**           | map\[string\]string | The PersistentVolume labels. Optional. Defaults to nil.                                                                                                                        |
| **volume.annotations**      | map\[string\]string | The PersistentVolume annotations. Optional. Defaults to nil.                                                                                                                   |
| **volumeClaim**             | object              | The PersistentVolumeClaim options. Optional.                                                                                                                                   |
| **volumeClaim.name**        | string              | The PersistentVolumeClaim name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                   |
| **volumeClaim.labels**      | map\[string\]string | The PersistentVolumeClaim labels. Optional. Defaults to nil.                                                                                                                   |
| **volumeClaim.annotations** | map\[string\]string | The PersistentVolumeClaim annotations. Optional. Defaults to nil.                                                                                                              |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |


## Example <!-- {docsify-ignore} -->

Example YAML manifest for IpRange:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-vol
spec:
  capacityGb: 1024
  tier: ENTERPRISE
---
apiVersion: v1
kind: Pod
metadata:
  name: workload
spec:
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: my-vol
  containers:
    - name: workload
      image: nginx
      volumeMounts:
        - mountPath: "/mnt/data1"
          name: data
```
