# AwsVpcPeering Custom Resource

The `awsvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between
Kyma and the remote AWS Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks
of the same cloud provider.

Once an `AwsVpcPeering` CR is created and reconciled, the Cloud Manager controller creates a VPC peering connection in 
the Kyma cluster underlying cloud provider account and accepts VPC peering connection in the remote cloud provider account.

## Required Actions in the Remote Project

Before creating the VPC peering, please tag your AWS account VPC with the Kyma shoot name tag.  
For more information, check the [Create Virtual Private Cloud Peering in Amazon Web Services](../tutorials/01-30-10-aws-vpc-peering.md) tutorial.

## Deleting `AwsVpcPeering`

Kyma's underlying cloud provider VPC peering connection is deleted as a part of AwsVpcPeering deletion. The remote VPC 
peering connection is left hanging, and must be deleted manually.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter           | Type   | Description                                                                                  |
|---------------------|--------|----------------------------------------------------------------------------------------------|
| **remoteAccountId** | string | Required. Specifies the the Amazon Web Services account ID of the owner of the accepter VPC. |
| **remoteRegion**    | string | Required. Specifies the Region code for the accepter VPC.                                    |
| **remoteVpcId**     | string | Required. Specifies the ID of the VPC with which you are creating the VPC peering connection |

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

See an exemplary `AwsVpcPeering` custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsVpcPeering
metadata:
  name: peering-to-vpc-11122233
spec:
  remoteVpcId: vpc-11122233
  remoteRegion: us-west-2
  remoteAccountId: 123456789012
```
