# RBAC Aggregation

Cloud Manager extends the default Kubernetes user-facing roles with permissions for Cloud Manager custom resources in SAP BTP, Kyma runtime (SKR).

This follows the Kyma-wide RBAC aggregation decision and uses native Kubernetes ClusterRole aggregation.

## Aggregated Roles

Cloud Manager provides these module roles:

- `kyma-cloud-manager-view`
- `kyma-cloud-manager-edit`
- `kyma-cloud-manager-admin`

These roles aggregate into the default Kubernetes roles through labels:

- `rbac.authorization.k8s.io/aggregate-to-view: "true"`
- `rbac.authorization.k8s.io/aggregate-to-edit: "true"`
- `rbac.authorization.k8s.io/aggregate-to-admin: "true"`

## Scope

Cloud Manager aggregation roles target the `cloud-resources.kyma-project.io` API group (SKR resources).

The roles do not grant permissions for `cloud-control.kyma-project.io`, because that API group belongs to Kyma Control Plane (KCP), not SKR.

## Permission Model

- **View role**: Read-only permissions (`get`, `list`, `watch`) for Cloud Manager resources.
- **Edit role**: Read and mutate permissions (`create`, `update`, `patch`, `delete`, `deletecollection`) for Cloud Manager resources.
- **Admin role**: Full access to Cloud Manager SKR resources.

## Bindings

To grant these permissions to workloads or users, create standard Kubernetes role bindings to the built-in `view`, `edit`, or `admin` ClusterRoles.

Because Cloud Manager roles are aggregated to these built-in roles, the effective permissions are expanded automatically.

For example, bind a service account to `view`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: app-view
subjects:
- kind: ServiceAccount
  name: app
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
```

After Cloud Manager is installed, this subject can read Cloud Manager resources without any additional custom role wiring.

## Notes

- Role aggregation is additive. If more module roles are installed, effective permissions for built-in roles can grow.
- Review role changes during upgrades as part of normal security governance.
