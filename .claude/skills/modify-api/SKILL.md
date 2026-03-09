---
name: modify-api
description: Modify existing CRD APIs in Cloud Manager. Use when adding fields to CRDs, changing API types, updating validation rules, or running make manifests.
---

# Modify CRD API

Add or modify fields in existing Cloud Manager CRDs.

## Quick Start

1. Edit type definition in `api/<group>/v1beta1/<resource>_types.go`
2. Add Kubebuilder validation markers
3. Run `make manifests && make generate`
4. Update reconciler if needed
5. For SKR: Run patch and sync scripts
6. Add tests

## Adding a Field

**File**: `api/cloud-control/v1beta1/<resource>_types.go`

```go
type GcpResourceSpec struct {
    // Existing fields...

    // NEW FIELD
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:validation:Maximum=5
    ReplicaCount int32 `json:"replicaCount,omitempty"`
}
```

## Validation Markers

```go
// Required
// +kubebuilder:validation:Required

// Optional with default
// +kubebuilder:validation:Optional
// +kubebuilder:default=3

// Numeric
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100

// String
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=63
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

// Enum
// +kubebuilder:validation:Enum=STANDARD;PREMIUM;ENTERPRISE

// Immutable (CEL)
// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="field is immutable"
```

## Generate After Changes

### KCP Resources

```bash
make manifests
make generate
```

### SKR Resources

```bash
make manifests
make generate
./config/patchAfterMakeManifests.sh
./config/sync.sh
```

## Adding to Status

```go
type GcpResourceStatus struct {
    // Existing fields...

    // NEW STATUS FIELD
    // +optional
    Endpoint string `json:"endpoint,omitempty"`

    // +optional
    Port int32 `json:"port,omitempty"`
}
```

## Update Reconciler

If field affects behavior, update reconciler actions:

```go
func updateResource(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    obj := state.ObjAsGcpResource()

    // Check if field changed
    if state.remoteResource.ReplicaCount != obj.Spec.ReplicaCount {
        // Update cloud resource
        err := state.client.Update(ctx, UpdateRequest{
            ReplicaCount: obj.Spec.ReplicaCount,
        })
        if err != nil {
            return err, ctx
        }
    }

    return nil, ctx
}
```

## Version Annotation (SKR)

For SKR CRDs, update version in patch script:

**File**: `config/patchAfterMakeManifests.sh`

```bash
yq -i '.metadata.annotations."cloud-resources.kyma-project.io/version" = "v0.0.X"' \
    $SCRIPT_DIR/crd/bases/cloud-resources.kyma-project.io_gcpresources.yaml
```

## Breaking Changes

### FORBIDDEN Without Approval
- Remove existing field
- Change field type
- Make optional field required
- Change enum values

### ALLOWED
- Add optional field
- Add field with default
- Add new enum value
- Relax validation (increase max, decrease min)

## Checklist

- [ ] Type definition updated
- [ ] Validation markers added
- [ ] `make manifests` run
- [ ] `make generate` run
- [ ] For SKR: patch script updated
- [ ] For SKR: sync script run
- [ ] Reconciler updated if needed
- [ ] Tests added for new field
- [ ] No breaking changes (or user approved)

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Field not in CRD | Run `make manifests` |
| DeepCopy missing | Run `make generate` |
| SKR CRD outdated | Run sync.sh |
| Validation not working | Check marker syntax |

## Related

- Full workflow: [docs/agents/DEVELOPMENT_WORKFLOW.md](../../../docs/agents/DEVELOPMENT_WORKFLOW.md)
- Validation: `/api-validation`
