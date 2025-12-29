# GCP NfsInstance V2 (Modern Implementation)

**Status**: ACTIVE - Modern streamlined implementation  
**Pattern**: OLD Reconciler Pattern (maintains compatibility with multi-provider CRD)  
**Feature Flag**: `gcpNfsInstanceV2` (default: false, uses v1)

## Overview

This directory contains the modernized GCP NfsInstance implementation that follows the OLD reconciler pattern while providing a cleaner, more maintainable codebase. It maintains compatibility with the multi-provider NfsInstance CRD but organizes code more logically.

## Architecture

- **Pattern**: OLD Reconciler Pattern (multi-provider CRD)
- **State Hierarchy**: `focal.State` → `types.State` (shared) → `v2.State` (GCP-specific)
- **Client**: `google.golang.org/api/file/v1` (REST API wrapper)
- **CRD**: `cloud-control.kyma-project.io/v1beta1.NfsInstance` (multi-provider)

## Key Improvements Over V1

### 1. Better Organization
- Actions grouped by purpose (operations, validation, state management)
- Clear separation of concerns
- Easier to navigate and understand

### 2. Simplified State Management
- Cleaner state struct with better field naming
- Consolidated state helper methods
- Improved state machine logic in dedicated package

### 3. Improved Validation
- Centralized validation logic in validation package
- Clearer validation error messages
- Better separation of pre-flight and post-creation validations

### 4. Enhanced Client Abstraction
- Cleaner client interface
- Better error handling
- Easier to mock for testing

### 5. Streamlined Testing
- Focus on business logic tests
- Avoid trivial getter/setter tests
- Better test organization matching code structure

### 6. Consistent Error Handling
- Uniform error patterns across all actions
- Better error context
- Consistent requeue strategies

## Package Structure

```
v2/
├── README.md                      → This file (architecture overview)
├── reconcile.go                   → Entry point, state factory exposure
├── state.go                       → Streamlined state definition
├── actions.go                     → Composed action flow
│
├── client/                        → GCP Filestore client abstraction
│   ├── filestore.go              → Client interface & implementation
│   ├── types.go                  → Client types & constants
│   └── mock.go                   → Mock client for tests
│
├── operations/                    → CRUD operations and operation polling
│   ├── create.go                 → Create filestore instance
│   ├── update.go                 → Update filestore instance
│   ├── delete.go                 → Delete filestore instance
│   ├── load.go                   → Load instance from GCP
│   └── operation.go              → Operation polling logic
│
├── validation/                    → Validation logic
│   ├── preflight.go              → Pre-flight validations
│   ├── postcreate.go             → Post-creation validations
│   └── helpers.go                → Validation utilities
│
└── state/                         → State management
    ├── machine.go                → State machine transitions
    ├── status.go                 → Status update helpers
    └── comparison.go             → Filestore comparison logic
```

## Reconciliation Flow

The v2 reconciler follows this streamlined action sequence:

1. **validatePreflight** → Pre-flight validation (capacity, tier, network, IpRange)
2. **AddCommonFinalizer** → Ensure finalizer for cleanup
3. **pollOperation** → Poll pending GCP operations (if any)
4. **loadInstance** → Fetch instance from GCP Filestore API
5. **validatePostCreate** → Post-creation validation (no scale-down for BASIC tiers)
6. **stateMachine** → State machine logic + determine operation type
7. **Branch: Not Marked for Deletion**
   - **syncInstance** → Create or update instance based on operation type
   - **pollOperation** → Wait for GCP operation to complete
   - **updateStatus** → Update NfsInstance status
   - **StopAndForget** → Complete reconciliation
8. **Branch: Marked for Deletion**
   - **deleteInstance** → Delete filestore instance
   - **pollOperation** → Wait for deletion to complete
   - **RemoveCommonFinalizer** → Remove finalizer
   - **StopAndForget** → Complete reconciliation

## State Machine

The state machine manages the lifecycle of the GCP Filestore instance:

| GCP State | CRD State | Operation Type | Next Action |
|-----------|-----------|----------------|-------------|
| Not Found | Any | ADD | Create instance |
| CREATING | Processing | NONE | Poll operation |
| READY | Processing | MODIFY (if mismatch) | Update instance |
| READY | Processing | NONE (if match) | Update status to Ready |
| READY | Ready | MODIFY (if mismatch) | Update instance |
| READY | Ready | NONE (if match) | No-op, stop |
| ERROR | Error | NONE | Report error |
| DELETING | Deleting | NONE | Poll operation |
| Deleted | Any (marked for deletion) | DELETE | Already deleted, remove finalizer |

## Client Interface

The client interface provides a clean abstraction over the GCP Filestore API:

```go
type FilestoreClient interface {
    GetInstance(ctx context.Context, projectId, location, instanceId string) (*file.Instance, error)
    CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *file.Instance) (*file.Operation, error)
    UpdateInstance(ctx context.Context, projectId, location, instanceId, updateMask string, instance *file.Instance) (*file.Operation, error)
    DeleteInstance(ctx context.Context, projectId, location, instanceId string) (*file.Operation, error)
    GetOperation(ctx context.Context, projectId, operationName string) (*file.Operation, error)
}
```

## State Structure

```go
type State struct {
    types.State                           // Shared NfsInstance state (OLD pattern)
    
    client          FilestoreClient       // GCP Filestore client
    gcpInstance     *file.Instance        // Cached GCP instance
    pendingOp       *file.Operation       // In-progress operation
    opType          OperationType         // ADD, MODIFY, DELETE, NONE
    
    // State machine
    currentState    v1beta1.StatusState   // Current lifecycle state
    
    // Update tracking
    updateMask      []string              // Fields requiring update
}
```

## Operation Types

```go
type OperationType int

const (
    OpNone   OperationType = iota  // No operation needed
    OpAdd                          // Create new instance
    OpModify                       // Update existing instance
    OpDelete                       // Delete instance
)
```

## Key Differences from V1

| Aspect | V1 | V2 |
|--------|----|----|
| **File organization** | Flat structure | Organized in packages |
| **State helpers** | Mixed in state.go | Separated by concern (state/, operations/, validation/) |
| **Validation** | Split across files | Centralized in validation/ |
| **Client** | Directly in client/ | Better abstraction in client/ |
| **Testing** | Many trivial tests | Focus on business logic |
| **Action naming** | check*, sync* | Clearer purpose-based names |
| **Error handling** | Varied patterns | Consistent approach |
| **State machine** | Embedded in checkNUpdateState | Dedicated state/machine.go |

## Feature Flag Usage

The v2 implementation is enabled via the `gcpNfsInstanceV2` feature flag:

```yaml
# config/featureToggles/featureToggles.local.yaml
gcpNfsInstanceV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled  # Default to v1 for stability
```

In code:
```go
if feature.GcpNfsInstanceV2.Value(ctx) {
    return gcpnfsinstancev2.New(r.gcpStateFactoryV2)(ctx, state)
} else {
    return gcpnfsinstancev1.New(r.gcpStateFactoryV1)(ctx, state)
}
```

## Testing Strategy

### Unit Tests
- Focus on business logic and decision points
- Test error handling and edge cases
- Test state transitions and API interactions
- **Avoid** testing simple getters/setters
- **Avoid** testing framework code
- **Avoid** redundant tests

### Integration Tests
- Test complete reconciliation flows
- Test create → ready → update → delete lifecycle
- Test error recovery scenarios
- Use testinfra framework with mocked clients

## Migration Path

1. V1 remains default (feature flag = false)
2. V2 can be tested in non-production environments (feature flag = true)
3. Once V2 is validated, it becomes default
4. V1 is deprecated and eventually removed

## Maintenance Notes

- This implementation maintains the OLD reconciler pattern for compatibility
- The multi-provider CRD (`NfsInstance`) is shared across all providers
- Changes to CRD schema require coordination with other providers (AWS, Azure, SAP)
- State hierarchy must be maintained: `focal.State` → `types.State` → `v2.State`

## References

- [AGENTS.md](/AGENTS.md) - Agent instructions and rules
- [V1 Implementation](../v1/README.md) - Legacy implementation
- [Architecture: OLD Pattern](/docs/agents/architecture/RECONCILER_OLD_PATTERN.md)
- [Architecture: State Pattern](/docs/agents/architecture/STATE_PATTERN.md)
- [Architecture: Action Composition](/docs/agents/architecture/ACTION_COMPOSITION.md)
