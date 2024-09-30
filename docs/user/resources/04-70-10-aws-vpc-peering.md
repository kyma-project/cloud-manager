## AWS Vpc Peering


The `awsvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between
Kyma and the remote AWS Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks
of the same cloud provider.

Once an AwsVpcPeering CR is created and reconciled, the Cloud Manager controller first creates a virtual network peering
connection in the Virtual Private Cloud (VPC) network of the Kyma cluster in the underlying cloud provider and accepts
VPC peering connection in the remote cloud provider account.

Cloud Manager initiates VPC peering connection from the Kyma underlying cloud provider. Cloud Manager must be authorized
to accept VPC peering connection in the remote cloud provider account. For cross-account access Cloud Manager uses
[AssumeRole](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/sts/assume-role.html). 

You must create CloudManagerPeeringRole with trust policy that allows Cloud Manager principal
arn:aws:iam::{194230256199}:user/cloud-manager-peering-ENV to assume that role. ENV corresponds to dev, stage or prod.
```json
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::194230256199:user/cloud-manager-peering-ENV"
			},
			"Action": "sts:AssumeRole"
		}
	]
}
```
Create new managed policy CloudManagerPeeringAccess with following permissions:
* ec2:AcceptVpcPeeringConnection
* ec2:DescribeVpcs
* ec2:DescribeVpcPeeringConnections
* ec2:DescribeRouteTables
* ec2:CreateRoute
* ec2:CreateTags

Attach CloudManagerPeeringAccess policy to CloudManagerPeeringRole.

Kyma underlying cloud provider VPC peering connection is deleted as a part of AwsVpcPeering deletion, but remote VPC 
peering connection is left hanging, and must be deleted manually.


## Specification <!-- {docsify-ignore} -->


This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter           | Type   | Description                                                                        |
|---------------------|--------|------------------------------------------------------------------------------------|
| **remoteAccountId** | string | Specifies the the Amazon Web Services account ID of the owner of the accepter VPC. |
| **remoteRegion**    | string | Specifies the Region code for the accepter VPC.                                    |
| **remoteVpcId**     | string | Specifies the ID of the VPC with which you are creating the VPC peering connection |

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

See an exemplary AwsVpcPeering custom resource:

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
