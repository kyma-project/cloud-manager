# GCP NfsInstance V1 (Legacy Implementation)

**Status**: LEGACY - Maintained for backward compatibility only  
**Deprecated**: This implementation is preserved from the original codebase  
**Replacement**: V2 implementation (when available via feature flag)

## Overview

This directory contains the original GCP NfsInstance implementation that follows the OLD reconciler pattern. It is maintained in this v1 subdirectory for backward compatibility and stability.

## Architecture

- **Pattern**: OLD Reconciler Pattern (multi-provider CRD)
- **State Hierarchy**: `focal.State` → `types.State` (shared) → `v1.State` (GCP-specific)
- **Client**: `google.golang.org/api/file/v1` (REST API wrapper)
- **CRD**: `cloud-control.kyma-project.io/v1beta1.NfsInstance` (multi-provider)

## Package Structure

```
v1/
├── README.md                      → This file
├── new.go                         → Entry point, action composition
├── state.go                       → State definition, factory
├── checkGcpOperation.go           → GCP operation polling
├── checkNUpdateState.go           → State machine logic
├── checkUpdateMask.go             → Update field tracking
├── loadNfsInstance.go             → Load from GCP
├── syncNfsInstance.go             → Create/Update/Delete GCP resource
├── validateAlways.go              → Pre-flight validation
├── validatePostCreate.go          → Post-creation validation
├── validations.go                 → Shared validation logic
├── *_test.go                      → Unit tests
├── client/
│   └── filestoreClient.go         → GCP Filestore client wrapper
└── testdata/
    └── nfs41Enabled.yaml          → Test fixtures
```

## Reconciliation Flow

The v1 reconciler follows this action sequence:

1. **validateAlways** → Pre-flight validation (capacity, tier, network)
2. **AddCommonFinalizer** → Ensure finalizer for cleanup
3. **checkGcpOperation** → Poll pending GCP operations
4. **loadNfsInstance** → Fetch instance from GCP Filestore API
5. **validatePostCreate** → Post-creation validation (no scale-down for BASIC tiers)
6. **checkNUpdateState** → State machine logic + determine operation (ADD/MODIFY/DELETE/NONE)
7. **checkUpdateMask** → Build update mask for MODIFY operations
8. **syncNfsInstance** → Execute operation (Create/Patch/Delete)
9. **RemoveCommonFinalizer** (if deleted) → Clean up finalizer

## GCP Filestore API Calls

V1 uses the following GCP Filestore API operations via `google.golang.org/api/file/v1`:

- **Instances.Get** → Load instance
- **Instances.Create** → Create new instance
- **Instances.Patch** → Update existing instance
- **Instances.Delete** → Delete instance
- **Operations.Get** → Poll operation status

## State Management

The v1 State struct extends `types.State` (shared layer) and adds GCP-specific fields:

```go
type State struct {
    types.State                         // Shared NfsInstance state
    
    curState        v1beta1.StatusState // Current lifecycle state
    operation       OperationType       // ADD, MODIFY, DELETE, NONE
    updateMask      []string            // Fields to update
    validations     []string            // Validation errors
    fsInstance      *file.Instance      // Cached GCP instance
    filestoreClient FilestoreClient     // GCP Filestore client
}
```

## Configuration Dependencies

V1 depends on:

1. **GCP Config** (`pkg/kcp/provider/gcp/config`):
   - `GcpRetryWaitTime` - Retry delay for operations
   - `GcpOperationWaitTime` - Wait time between operation polls

2. **Feature Flags** (`pkg/feature`):
   - `Nfs41Gcp` - Enable NFSv4.1 protocol for ZONAL/REGIONAL tiers

3. **IpRange Dependency**:
   - NfsInstance requires IpRange for network configuration
   - Loaded via shared reconciler layer

4. **Scope Dependency**:
   - GCP project ID and credentials from Scope CRD
   - Location (region/zone) from NfsInstance spec

## Usage

The v1 implementation is used by default when the `gcpNfsInstanceV2` feature flag is disabled (default behavior).

```go
// In pkg/kcp/nfsinstance/reconciler.go
if feature.GcpNfsInstanceV2.Value(ctx) {
    return v2.New(r.gcpStateFactoryV2)(ctx, state)
} else {
    return v1.New(r.gcpStateFactoryV1)(ctx, state)  // Default
}
```

## Migration to V2

A v2 implementation is planned with the following improvements:

- Modern GCP client library (`cloud.google.com/go/filestore`)
- Streamlined action organization
- Simplified state management
- Better separation of concerns
- Improved testing patterns

Users will be able to enable v2 via the `gcpNfsInstanceV2` feature flag once v2 is available.

## Maintenance Policy

**DO**:
- Fix critical bugs in v1
- Maintain test coverage
- Preserve backward compatibility
- Keep documentation updated

**DO NOT**:
- Add new features to v1 (add to v2 instead)
- Refactor v1 structure (it's frozen)
- Change reconciliation behavior
- Modify CRD schema

## References

- **Architecture**: See `docs/agents/architecture/RECONCILER_OLD_PATTERN.md`
- **Agent Instructions**: See `AGENTS.md`
- **V2 Plan**: See `GCP_NFSINSTANCE_REFACTOR_PLAN.md`

---

**Last Updated**: 2025-12-26  
**Version**: 1.0 (Legacy)  
**Deprecated**: Use V2 when available via feature flag
