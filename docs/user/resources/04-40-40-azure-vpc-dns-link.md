# AzureVpcDnsLink Custom Resource

The `azurevpcdnslink.cloud-resources.kyma-project.io` custom resource (CR) specifies the link between Kyma network and the remote Azure private DNS zone.

Once an `AzureVpcDnsLink` CR is created and reconciled, the Cloud Manager controller creates an Azure Virtual Network Link in the private DNS zone of the remote underlying cloud provider landscape, pointing to the Kyma underlying cloud provider network.

## Specification

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                | Type    | Description                                                                                                                         |
|--------------------------|---------|-------------------------------------------------------------------------------------------------------------------------------------|
| **remoteLinkName**       | string  | Specifies the name of the Virtual Network Link in the remote subscription.                                                     |
| **remotePrivateDnsZone** | string  | Specifies the ID of the Private DNS zone in the remote subscription.                                                           |
| **remoteTenant**         | string  | Optional. Specifies the tenant ID of the remote subscription. Defaults to Kyma cluster underlying cloud provider subscription tenant. |

**Status:**

| Parameter                         | Type       | Description                                                                                 |
|-----------------------------------|------------|---------------------------------------------------------------------------------------------|
| **state**                         | string     | Signifies the current state of CustomObject.                                                |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                        |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                       |
| **conditions.message**            | string     | Provides more details about the condition status change.                                    |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                         |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.  |
| **conditions.type**               | string     | Provides a short description of the condition.                                              |

## Sample Custom Resource

See an exemplary AzureVpcDnsLink custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureVpcDnsLink
metadata:
  name: link-to-example-com
spec:
  remoteLinkName: link-to-my-kyma
  remotePrivateDnsZone: /subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/MyResourceGroup/providers/Microsoft.Network/privateDnsZones/example.com
  remoteTenant: ac3ddba3-536d-4b6f-aad7-03b942e46aca
```
