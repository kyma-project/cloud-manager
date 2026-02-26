# AccessibleFrom Field Refactoring Plan

## Overview

Refactor the `AccessibleFrom` field from `[]string` with magic `"all"` value to a discriminated union struct for improved type safety and developer experience.

## K8s API Precedent

This refactoring follows established Kubernetes API conventions where a `type` field discriminates behavior:

| K8s API | Type Field | Conditional Fields | Example |
|---------|------------|-------------------|---------|
| **Ingress** | `pathType: Exact\|Prefix\|ImplementationSpecific` | `path` interpretation varies | [networking/v1 Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types) |
| **Service** | `type: ClusterIP\|NodePort\|LoadBalancer\|ExternalName` | `nodePort`, `loadBalancerIP`, `externalName` | [core/v1 Service](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) |
| **PersistentVolume** | `accessModes: [ReadWriteOnce\|ReadOnlyMany\|ReadWriteMany]` | Mount behavior | [core/v1 PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) |
| **HorizontalPodAutoscaler** | `metrics[].type: Resource\|Pods\|Object\|External` | `resource`, `pods`, `object`, `external` | [autoscaling/v2 HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-metrics-apis) |
| **NetworkPolicy** | `policyTypes: [Ingress\|Egress]` | `ingress`, `egress` rules | [networking/v1 NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) |

**Pattern**: Enum type field + conditional struct fields = Discriminated Union

## Current Design (Before)

```go
// Both GcpNfsVolumeBackupSpec and GcpNfsBackupScheduleSpec
AccessibleFrom []string `json:"accessibleFrom,omitempty"`
```

**Usage:**
```yaml
accessibleFrom: ["all"]  # Magic string
# or
accessibleFrom: ["shoot-1", "subaccount-xyz"]
```

**Problems:**
- Magic string `"all"` lacks discoverability
- Complex CEL validation rule for mutual exclusivity
- No type safety
- Poor IDE support

## Target Design (After)

```go
// +kubebuilder:validation:Enum=All;Specific
type AccessibleFromType string

const (
    AccessibleFromTypeAll      AccessibleFromType = "All"
    AccessibleFromTypeSpecific AccessibleFromType = "Specific"
)

// +kubebuilder:validation:XValidation:rule="self.type != 'Specific' || size(self.targets) > 0", message="targets required when type is Specific"
// +kubebuilder:validation:XValidation:rule="self.type != 'All' || !has(self.targets) || size(self.targets) == 0", message="targets must be empty when type is All"
type AccessibleFrom struct {
    // Type specifies the access scope.
    // "All" allows access from all shoots in the same global account and GCP project.
    // "Specific" requires Targets to be defined.
    // +kubebuilder:validation:Required
    Type AccessibleFromType `json:"type"`
    
    // Targets is an array of shootNames or subaccountIds.
    // Required when Type is "Specific", must be empty when Type is "All".
    // +optional
    // +listType=set
    // +kubebuilder:validation:MaxItems=10
    Targets []string `json:"targets,omitempty"`
}
```

**Usage:**
```yaml
accessibleFrom:
  type: All
---
accessibleFrom:
  type: Specific
  targets:
    - shoot-prod-1
    - subaccount-xyz
```

---

## Files to Modify

### Phase 1: API Type Definitions

| # | File | Change |
|---|------|--------|
| 1.1 | [api/cloud-resources/v1beta1/gcpnfsvolumebackup_types.go](api/cloud-resources/v1beta1/gcpnfsvolumebackup_types.go) | Add `AccessibleFromType` enum, `AccessibleFrom` struct; update `GcpNfsVolumeBackupSpec.AccessibleFrom` field type |
| 1.2 | [api/cloud-resources/v1beta1/gcpnfsbackupschedule_types.go](api/cloud-resources/v1beta1/gcpnfsbackupschedule_types.go) | Update `GcpNfsBackupScheduleSpec.AccessibleFrom` field to use shared `AccessibleFrom` struct |
| 1.3 | Run `make manifests` | Regenerate CRD YAML files |

### Phase 2: SKR Reconciler - GcpNfsVolumeBackup

| # | File | Change |
|---|------|--------|
| 2.1 | [pkg/skr/gcpnfsvolumebackup/v1/state.go](pkg/skr/gcpnfsvolumebackup/v1/state.go) | Update `specCommaSeparatedAccessibleFrom()` to handle struct; update `HasProperLabels()`, `HasAllStatusLabels()`, `SetFilestoreLabels()` |
| 2.2 | [pkg/skr/gcpnfsvolumebackup/v2/state.go](pkg/skr/gcpnfsvolumebackup/v2/state.go) | Same changes as v1 |
| 2.3 | [pkg/skr/gcpnfsvolumebackup/v1/addLabelsToNfsBackup.go](pkg/skr/gcpnfsvolumebackup/v1/addLabelsToNfsBackup.go) | Update status AccessibleFrom handling if needed |
| 2.4 | [pkg/skr/gcpnfsvolumebackup/v2/addLabelsToNfsBackup.go](pkg/skr/gcpnfsvolumebackup/v2/addLabelsToNfsBackup.go) | Same as v1 |

### Phase 3: SKR Reconciler - GcpNfsVolumeRestore

| # | File | Change |
|---|------|--------|
| 3.1 | [pkg/skr/gcpnfsvolumerestore/state.go](pkg/skr/gcpnfsvolumerestore/state.go) | Update `IsAccessibleFromShoot()` - logic unchanged, just reading labels |

### Phase 4: SKR Reconciler - BackupSchedule

| # | File | Change |
|---|------|--------|
| 4.1 | [pkg/skr/backupschedule/backupImplGcpNfs.go](pkg/skr/backupschedule/backupImplGcpNfs.go) | Update `getBackupObject()` to pass struct instead of slice |

### Phase 5: Test Infrastructure

| # | File | Change |
|---|------|--------|
| 5.1 | [pkg/testinfra/dsl/gcpNfsVolumeBackup.go](pkg/testinfra/dsl/gcpNfsVolumeBackup.go) | Update `WithGcpNfsVolumeBackupAccessibleFrom()` signature; add helpers for `All` and `Specific` modes |
| 5.2 | [pkg/testinfra/dsl/nfsBackupSchedule.go](pkg/testinfra/dsl/nfsBackupSchedule.go) | Update `WithGcpNfsBackupScheduleAccessibleFrom()` |
| 5.3 | [pkg/testinfra/dsl/gcpNfsVolumeBackup.go](pkg/testinfra/dsl/gcpNfsVolumeBackup.go) | Update `HavingGcpNfsVolumeBackupAccessibleFromStatus()` assertion |

### Phase 6: Controller Tests

| # | File | Change |
|---|------|--------|
| 6.1 | [internal/controller/cloud-resources/gcpnfsvolumebackup_test.go](internal/controller/cloud-resources/gcpnfsvolumebackup_test.go) | Update test to use new struct |
| 6.2 | [internal/controller/cloud-resources/gcpnfsbackupschedule_test.go](internal/controller/cloud-resources/gcpnfsbackupschedule_test.go) | Update test to use new struct |
| 6.3 | [internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go](internal/controller/cloud-resources/gcpnfsvolumebackupdiscovery_test.go) | Verify label checks still work |

### Phase 7: API Validation Tests

| # | File | Change |
|---|------|--------|
| 7.1 | [internal/api-tests/skr_gcpnfsvolumebackup_test.go](internal/api-tests/skr_gcpnfsvolumebackup_test.go) | Add validation tests for `AccessibleFrom` struct |

### Phase 8: Configuration & Samples

| # | File | Change |
|---|------|--------|
| 8.1 | [config/samples/cloud-resources_v1beta1_gcpnfsvolumebackup.yaml](config/samples/cloud-resources_v1beta1_gcpnfsvolumebackup.yaml) | Update sample YAML |
| 8.2 | [config/samples/cloud-resources_v1beta1_gcpnfsbackupschedule.yaml](config/samples/cloud-resources_v1beta1_gcpnfsbackupschedule.yaml) | Update sample YAML |

### Phase 9: Generated Files (via `make manifests`)

| # | File | Change |
|---|------|--------|
| 9.1 | [config/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml](config/crd/bases/cloud-resources.kyma-project.io_gcpnfsvolumebackups.yaml) | Auto-generated |
| 9.2 | [config/crd/bases/cloud-resources.kyma-project.io_gcpnfsbackupschedules.yaml](config/crd/bases/cloud-resources.kyma-project.io_gcpnfsbackupschedules.yaml) | Auto-generated |
| 9.3 | `config/dist/**` | Auto-generated |

---

## Implementation Details

### Helper Function Changes

**Current `specCommaSeparatedAccessibleFrom()` in state.go:**
```go
func (s *State) specCommaSeparatedAccessibleFrom() string {
    backup := s.ObjAsGcpNfsVolumeBackup()
    sort.Strings(backup.Spec.AccessibleFrom)
    return strings.Join(backup.Spec.AccessibleFrom, ",")
}
```

**New implementation:**
```go
func (s *State) specCommaSeparatedAccessibleFrom() string {
    backup := s.ObjAsGcpNfsVolumeBackup()
    accessibleFrom := backup.Spec.AccessibleFrom
    
    if accessibleFrom == nil {
        return ""
    }
    
    if accessibleFrom.Type == cloudresourcesv1beta1.AccessibleFromTypeAll {
        return "all"
    }
    
    targets := make([]string, len(accessibleFrom.Targets))
    copy(targets, accessibleFrom.Targets)
    sort.Strings(targets)
    return strings.Join(targets, ",")
}
```

**Current label iteration pattern:**
```go
for _, shoot := range backup.Spec.AccessibleFrom {
    s.fileBackup.Labels[ConvertToAccessibleFromKey(shoot)] = util.GcpLabelBackupAccessibleFrom
}
```

**New implementation:**
```go
func (s *State) getAccessibleFromTargets() []string {
    backup := s.ObjAsGcpNfsVolumeBackup()
    accessibleFrom := backup.Spec.AccessibleFrom
    
    if accessibleFrom == nil {
        return nil
    }
    
    if accessibleFrom.Type == cloudresourcesv1beta1.AccessibleFromTypeAll {
        return []string{"all"}
    }
    
    return accessibleFrom.Targets
}

// Usage:
for _, target := range s.getAccessibleFromTargets() {
    s.fileBackup.Labels[ConvertToAccessibleFromKey(target)] = util.GcpLabelBackupAccessibleFrom
}
```

---

## Validation Rules

### CEL Validation on AccessibleFrom struct

```go
// +kubebuilder:validation:XValidation:rule="self.type != 'Specific' || size(self.targets) > 0", message="targets required when type is Specific"
// +kubebuilder:validation:XValidation:rule="self.type != 'All' || !has(self.targets) || size(self.targets) == 0", message="targets must be empty when type is All"
```

**Logic pattern:** `NOT(condition) OR requirement` = `condition IMPLIES requirement`

| Rule | Reads as |
|------|----------|
| `type != 'Specific' \|\| size(targets) > 0` | If Specific → must have targets |
| `type != 'All' \|\| size(targets) == 0` | If All → targets must be empty |

### Test Cases for API Validation

| Test | Input | Expected |
|------|-------|----------|
| Valid All | `type: All` | ✅ Pass |
| Valid Specific | `type: Specific, targets: [a]` | ✅ Pass |
| Invalid: Specific without targets | `type: Specific, targets: []` | ❌ Fail |
| Invalid: All with targets | `type: All, targets: [a]` | ❌ Fail |
| Invalid: missing type | `targets: [a]` | ❌ Fail |

---

## Execution Checklist

- [x] **Phase 1**: Update API types
  - [x] Add `AccessibleFromType` enum and `AccessibleFrom` struct
  - [x] Update `GcpNfsVolumeBackupSpec`
  - [x] Update `GcpNfsBackupScheduleSpec`
  - [x] Run `make manifests`
  - [x] Run `make generate`
  
- [x] **Phase 2**: Update GcpNfsVolumeBackup reconciler
  - [x] Update v1/state.go helper functions
  - [x] Update v2/state.go helper functions
  - [x] Update addLabelsToNfsBackup actions (both v1 and v2)
  
- [x] **Phase 3**: Update GcpNfsVolumeRestore reconciler
  - [x] Verify `IsAccessibleFromShoot()` works with label-based lookup
  
- [x] **Phase 4**: Update BackupSchedule reconciler
  - [x] Update `backupImplGcpNfs.go`
  
- [x] **Phase 5**: Update test infrastructure
  - [x] Update DSL helpers in testinfra
  - [x] Add new builder functions for All/Specific modes
  
- [x] **Phase 6**: Update controller tests
  - [x] Update gcpnfsvolumebackup_test.go
  - [x] Update gcpnfsbackupschedule_test.go
  - [x] Verify discovery tests
  
- [x] **Phase 7**: Add API validation tests
  - [x] Add tests for AccessibleFrom validation rules
  
- [x] **Phase 8**: Update samples
  - [x] Update YAML samples
  
- [x] **Phase 9**: Final verification
  - [x] Run `make test`
  - [x] Run `make build`
  - [x] Run API validation tests

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Label logic regression | Medium | Comprehensive controller tests with label verification |
| Backup schedule creating invalid backups | High | Test schedule → backup creation flow |
| Restore failing due to label mismatch | High | Test restore with new label format |

---

## Timeline Estimate

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: API Types | 30 min |
| Phase 2: Backup Reconciler | 1 hour |
| Phase 3: Restore Reconciler | 15 min |
| Phase 4: Schedule Reconciler | 15 min |
| Phase 5: Test Infrastructure | 30 min |
| Phase 6: Controller Tests | 45 min |
| Phase 7: API Validation Tests | 30 min |
| Phase 8: Samples | 10 min |
| Phase 9: Verification | 30 min |
| **Total** | **~4-5 hours** |
