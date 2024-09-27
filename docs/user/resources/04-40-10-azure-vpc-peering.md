## Azure Vpc Peering


The `azurevpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between 
Kyma and the remote Azure Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks
of the same cloud provider.

Once an AzureVpcPeering CR is created and reconciled, the Cloud Manager controller first creates a virtual network peering 
connection in the Virtual Private Cloud (VPC) network of the Kyma cluster in the underlying cloud provider and accepts
VPC peering connection in the remote cloud provider subscription.

You must authorize Cloud Manager service principal `kyma-cloud-manager-ENV` in the remote cloud provider subscription to
accept VPC peering connection. Assign the following IAM roles to the Cloud Manager service principal in the remote subscription: 
* Classic Network Contributor
* Network Contributor

AzureVpcPeering can be deleted at any time but the VPC peering connection in the remote subscription must be deleted
manually.

## Specification <!-- {docsify-ignore} -->


This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter               | Type   | Description                                                                                                                                   |
|-------------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| **remotePeeringName**   | string | Specifies the name of the VNet peering in the remote subscription.                                                                            |
| **remoteVnet**          | string | Specifies the ID of the VNet in the remote subscription.                                                                                      |

**Status:**

| Parameter                         | Type       | Description                                                                                 |
|-----------------------------------|------------|---------------------------------------------------------------------------------------------|
| **id**                            | string     | Represents the VPC peering name on the Kyma cluster underlying cloud provider subscription. |
| **state**                         | string     | Signifies the current state of CustomObject.                                                |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                        |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                       |
| **conditions.message**            | string     | Provides more details about the condition status change.                                    |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                         |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.  |
| **conditions.type**               | string     | Provides a short description of the condition.                                              |


## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary AzureVpcPeering custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureVpcPeering
metadata:
  name: peering-to-my-vnet
spec:
  remotePeeringName: peering-to-my-kyma
  remoteVnet: /subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/virtualNetworks/MyVnet
```
