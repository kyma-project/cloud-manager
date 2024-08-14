# GcpVpcPeering Custom Resource

The `gcpvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) describes the VPC peering
connection that you can create to allow communication between Kyma and a remote VPC in Google Cloud Platform (GCP).
It allows you to consume services available on the remote VPC from the Kyma cluster.

### Required Permissions on the Remote Project
In order to create a VPC peering connection, the following permissions are required and must be granted to the Kyma service account provided:

| Parameter                          | Description                                                                 |
|------------------------------------|-----------------------------------------------------------------------------|
| compute.networks.addPeering        | Required to create the peering request in the remote project and VPC.       |
| compute.networks.get               | Required to fetch the list of existing VPC peerings from the remote VPC.    |
| compute.networks.ListEffectiveTags | Required to check if the remote VPC is tagged with the Kyma shoot name tag. |


## Specification <!-- {docsify-ignore} -->
This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter              | Type   | Description                                                                                                                                                        |
|------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **importCustomRoutes** | bool   | If set to `true`, custom routes are exported from the remote VPC and imported into Kyma.                                                                           |
| **remotePeeringName**  | string | The VPC Peering name in the remote project. To find it, select **Google Cloud project under VPC > {VPC Name} > VPC Network Peering** in your Google Cloud Project. |
| **remoteProject**      | string | The Google Cloud project to be peered with Kyma. The remote VPC is located in this project.                                                                        |
| **remoteVpcName**      | string | The name of the remote VPC to be peered with Kyma.                                                                                                                 |

**Status:**

| Parameter                         | Type       | Description                                                                                                                         |
|-----------------------------------|------------|-------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Represents the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                                |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                               |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                            |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                 |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                          |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                      |


## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary GcpVpcPeering custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpVpcPeering
metadata:
  name: "peering-with-kyma-dev"
spec:
  remotePeeringName: "peering-dev-vpc-to-kyma-dev"
  remoteProject: "my-remote-project"
  remoteVpc: "default"
  importCustomRoutes: false
```