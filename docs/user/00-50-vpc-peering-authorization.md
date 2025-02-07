# Authorizing Cloud Manager in the Remote Cloud Provider

To create VPC peering in the Kyma environment, you must authorize the Cloud Manager module in the remote cloud provider to accept the connection.

## Amazon Web Services

For cross-account access in Amazon Web Services, Cloud Manager uses `AssumeRole`. `AssumeRole` requires specifying the trusted principle. For more information, see the [official Amazon Web Services documentation](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/sts/assume-role.html).

Use the following table to identify the Cloud Manager principal based on your Kyma landscape. Then perform the required actions:

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Principal                                      |
|------------------------------------|----------------------------------------|--------------------------------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | `arn:aws:iam::194230256199:user/cloud-manager-peering-stage` |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | `arn:aws:iam::194230256199:user/cloud-manager-peering-prod`  |

1. Create a new role named **CloudManagerPeeringRole** with a trust policy that allows the Cloud Manager principal to assume the role:

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

2. Create a new **CloudManagerPeeringAccess** managed policy with the following permissions:

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

3. Attach the **CloudManagerPeeringAccess** policy to the **CloudManagerPeeringRole**.

## Google Cloud

Grant the following permissions to the Kyma service account in your GCP project:

| Permission                           | Description                                                                 |
|--------------------------------------|-----------------------------------------------------------------------------|
| `compute.networks.addPeering`        | Required to create the peering request in the remote project and VPC.       |
| `compute.networks.get`               | Required to fetch the list of existing VPC peerings from the remote VPC.    |
| `compute.networks.ListEffectiveTags` | Required to check if the remote VPC is tagged with the Kyma shoot name tag. |

For more information on how to manage access to service accounts, see the [official Google Cloud documentation](https://cloud.google.com/iam/docs/manage-access-service-accounts).

### Service Account

For security reasons, each Kyma landscape has its own service account. Use the following table to identify the correct Cloud Manager service account for your Kyma landscape:

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Service Account                                          |
|------------------------------------|----------------------------------------|------------------------------------------------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | `cloud-manager-peering@sap-ti-dx-kyma-mps-stage.iam.gserviceaccount.com` |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | `cloud-manager-peering@sap-ti-dx-kyma-mps-prod.iam.gserviceaccount.com`  |

## Microsoft Azure

To authorize Cloud Manager in the remote subscription, Microsoft Azure requires specifying the service principal. Use the following table to identify the Cloud Manager service principal based on your Kyma landscape. Then perform the following actions:

| BTP Cockpit URL                    | Kyma Dashboard URL                     | Cloud Manager Service Principal  |
|------------------------------------|----------------------------------------|----------------------------------|
| https://canary.cockpit.btp.int.sap | https://dashboard.stage.kyma.cloud.sap | kyma-cloud-manager-peering-stage |
| https://emea.cockpit.btp.cloud.sap | https://dashboard.kyma.cloud.sap       | kyma-cloud-manager-peering-prod  |

For more information, see the official Microsoft Azure documentation on how to [Assign Azure roles using the Azure portal](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-portal) and how to [Manage service principals](https://learn.microsoft.com/en-us/azure/databricks/admin/users-groups/service-principals).
