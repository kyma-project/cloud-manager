# Feature flags

This feature is based on the https://gofeatureflag.org/ library. When initialized with [`Initialize()`](./init.go)
it loads single toggle configuration yaml and stores an instance of the `*ffclient.GoFeatureFlag` 
in the [`provider`](./init.go) singleton variable. 

Each feature flag implemented in Cloud Manager should provide an anti corruption layer implementing 
the [`Feature`](./types.go) interface. In its implementation the actual flag value should be
obtained using the singleton instance in the [`provider`](./init.go) variable.

## Common evaluation context

Common evaluation context keys are defined as instances of the [`Key`](./types.go) type.

| Variable        | Description                                                                                                                                                         |
|-----------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| landscape       | Possible values: dev, stage, or prod. Determines the current landscape. Always defined.                                                                             |
| feature         | General feature name.                                                                                                                                               |
| plane           | Possible values: kcp, skr. Determines the plane of the feature/operation.                                                                                           |
| provider        | Cloud provider name as specified in the Shoot resource (aws, gcp, azure).                                                                                           |
| brokerPlan      | The KEB broker plan name, as specified in the KCP Kyma.                                                                                                             |
| globalAccount   | The BTP Global Account for the SKR. For non-SKR features not defined.                                                                                               |
| subAccount      | The BTP Subaccount for the SKR. For non-SKR features not defined.                                                                                                   |
| kyma            | The KCP Kyma name for the SKR. For non-SKR features not defined.                                                                                                    |
| shoot           | The Shoot name for the SKR. For non-SKR features not defined.                                                                                                       |
| region          | The region of the SKR as defined in the Shoot.                                                                                                                      |
| objKindGroup    | In format `lower([kind].[group])`. Defined only for features related to specific Object Kind.                                                                       |   
| crdKindGroup    | In format `lower([kind].[group])`. Defined in case CRD is handled and contains CRD's `spec.names.kind` and `spec.group`.                                            |
| busolaKindGroup | In format `lower([kind].[group])`. Defined in case of Busola extension ConfigMap and contains `general.resource.kind` and `general.resource.group`.                 |
| allKindGroups   | In format of `[objKindGroup],[crdKindGroup],[busolaKindGroup]`. Some of the elements may be empty. Use with `co`/`contains` operator to easily check any kindGroup. |

The evaluation context is stored in the golang context and is built using the [`ContextBuilder`](./context.go).

## General feature names

| Feature   | Description                             |
|-----------|-----------------------------------------|
| nfs       | All NFS Volume related features.        |
| nfsBackup | All NFS Volume Backup related features. |
| peering   | All VPC Peering related features.       |

