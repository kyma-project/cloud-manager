# IpRange Custom Resource

The `iprange.cloud-resources.kyma-project.io` custom resource specifies the VPC network
IP range that will be used for IP address allocation for cloud resources that require IP address. 

You are allowed to have one IpRange CR. If there are multiple IpRange resources in the cluster, the
oldest one will be reconciled and other will be ignored and put to the Error state.

Once IpRange is created and reconciled, the Cloud Manager controller reserves the specified IP range
in the VPC network of the cluster in the underlying cloud provider. IP address from that range will
be assigned to the provisioned resources of the cloud provider that require IP addresses. Once a 
cloud resource is assigned the local VPC network IP address it becomes functional and usable from the
cluster network and from the cluster workloads.

It is not required for the user to create an IpRange resource. Once needed it will be automatically created
and CIDR range automatically chosen adjacent and with the same size to the cluster nodes IP range. 
For most use cases this automatic allocation is sufficient. 

You might be interested in manually creating an IpRange resource with specific CIDR in advanced cases of
VPC network topology when cluster and cloud resources are not the only resources in the network, so you
can avoid IP range collisions. 

IpRange can be deleted and deprovisioned only if there are no cloud resources using it. In other words,
an IpRange and it's underlying VPC network address range can be purged only if there are no cloud resources
using an IP from that range.


## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter | Type   | Description                                                                          |
|-----------|--------|--------------------------------------------------------------------------------------|
| **cidr**  | string | Specifies the CIDR of the IP range that will be allocated. For example 10.250.4.0/22 |

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
kind: IpRange
metadata:
  name: my-range
spec:
  cidr: 10.250.4.0/22
```
