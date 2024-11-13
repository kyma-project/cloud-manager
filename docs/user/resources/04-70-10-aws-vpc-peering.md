# AwsVpcPeering Custom Resource


The `awsvpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between
Kyma and the remote AWS Virtual Private Cloud (VPC) network. Virtual network peering is only possible within the networks
of the same cloud provider.

Once an `AwsVpcPeering` CR is created and reconciled, the Cloud Manager controller creates a VPC peering connection in 
the Kyma cluster underlying cloud provider account and accepts VPC peering connection in the remote cloud provider account.

### Authorization

Cloud Manager must be authorized in the remote cloud provider account to accept VPC peering connection. For cross-account access,
Cloud Manager uses [`AssumeRole`](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/sts/assume-role.html).

Use the following table to identify Cloud Manager principal based on your Kyma landscape:

| BTP cockpit URL                    | Kyma dashboard URL                     | Cloud Manager principal                                    |
|------------------------------------|----------------------------------------|------------------------------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | `arn:aws:iam::194230256199:user/cloud-manager-peering-stage` |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | `arn:aws:iam::194230256199:user/cloud-manager-peering-prod`  |

1.  Create a new role named **CloudManagerPeeringRole** with a trust policy that allows Cloud Manager principal to assume the role:

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "AWS": "{CLOUD_MANAGER_PRINCIPAL}"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }

    ```

2.  Create a new managed policy **CloudManagerPeeringAccess** with the following permissions:
    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "Statement1",
                "Effect": "Allow",
                "Action": [
                    "ec2:AcceptVpcPeeringConnection",
                    "ec2:DescribeVpcs",
                    "ec2:DescribeVpcPeeringConnections",
                    "ec2:DescribeRouteTables",
                    "ec2:CreateRoute",
                    "ec2:CreateTags"
                ],
                "Resource": "*"
            }
        ]
    }
    ```

3.  Attach the **CloudManagerPeeringAccess** policy to the **CloudManagerPeeringRole**:

### Deleting `AwsVpcPeering`

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
