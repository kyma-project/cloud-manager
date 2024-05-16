# Feature toggles

## Common evaluation context

| Variable      | Description                                                                             |
|---------------|-----------------------------------------------------------------------------------------|
| landscape     | Possible values: dev, stage, or prod. Determines the current landscape. Always defined. |
| plane         | Possible values: kcp, skr. Determines plane of the feature/operation.                   |
| feature       | General feature name.                                                                   |
| provider      | Cloud provider name as specified in the Shoot resource (aws, gcp, azure).               |
| brokerPlan    | The KEB broker plan name, as specified in the KCP Kyma.                                 |
| globalAccount | The BTP Global Account for the SKR. For non-SKR features not defined.                   |
| subAccount    | The BTP Subaccount for the SKR. For non-SKR features not defined.                       |
| kyma          | The KCP Kyma name for the SKR. For non-SKR features not defined.                        |
| shoot         | The Shoot name for the SKR. For non-SKR features not defined.                           |
| kindGroup     | In format [kind].[group]. Defined only for features related to specific Kind.           |   

## General feature names

| Feature   | Description                             |
|-----------|-----------------------------------------|
| nfs       | All NFS Volume related features.        |
| nfsBackup | All NFS Volume Backup related features. |
| peering   | All VPC Peering related features.       |

