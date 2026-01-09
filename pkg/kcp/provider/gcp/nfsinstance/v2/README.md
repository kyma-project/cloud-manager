# GCP NfsInstance V2 (Modern Implementation)

**Status**: ACTIVE - Modern streamlined implementation  
**Pattern**: OLD Reconciler Pattern (maintains compatibility with multi-provider CRD)  
**Feature Flag**: `gcpNfsInstanceV2` (default: false, uses v1)

## Overview

This directory contains the modernized GCP NfsInstance implementation that follows the OLD reconciler pattern while providing a cleaner, more maintainable codebase. It maintains compatibility with the multi-provider NfsInstance CRD but organizes code more logically.

## Architecture

- **Pattern**: OLD Reconciler Pattern (multi-provider CRD)
- **State Hierarchy**: `focal.State` → `types.State` (shared) → `v2.State` (GCP-specific)
- **Client**: `cloud.google.com/go/filestore/apiv1` (Modern protobuf API)
- **CRD**: `cloud-control.kyma-project.io/v1beta1.NfsInstance` (multi-provider)

## Key Improvements Over V1

### 1. Modern GCP Client
- Uses `cloud.google.com/go/filestore/apiv1` (modern protobuf-based API)
- Cleaner client interface with operation tracking
- Better error handling and type safety
- Follows NEW pattern for client initialization (GcpClientProvider)

### 2. Simplified Package Structure
- Flat file organization at package root
- Actions as individual files (not grouped in subdirectories)
- Clear action naming (createInstance, updateInstance, deleteInstance, etc.)
- Easier to navigate and understand

### 3. Streamlined State Management
- Cleaner state struct with modern protobuf types
- Simple state getters/setters
- No separate state machine package needed

### 4. Improved Action Flow
- Clear branching with `IfElse` composition
- Separate wait actions (waitInstanceReady, waitInstanceDeleted)
- Operation polling handled explicitly
- Better separation of concerns

### 5. Enhanced Testing
- Mock client for unit tests
- Better test organization
- Avoid trivial getter/setter tests

### 6. Consistent Error Handling
- Uniform error patterns across all actions
- Better error context
- Consistent requeue strategies

## Package Structure

```
v2/
├── README.md                      → This file (architecture overview)
├── reconcile.go                   → Entry point, wraps state factory
├── state.go                       → State definition with protobuf types
├── compose.go                     → Action composition with IfElse branching
├── util.go                        → Helper utilities
│
├── client/                        → GCP Filestore client abstraction
│   ├── filestoreClient.go        → Client interface & implementation (protobuf API)
│   └── mockFilestoreClient.go    → Mock client for tests
│
└── Actions (individual files at package root):
    ├── pollOperation.go          → Poll pending GCP operations
    ├── loadInstance.go           → Load instance from GCP
    ├── createInstance.go         → Create filestore instance
    ├── updateInstance.go         → Update filestore instance
    ├── deleteInstance.go         → Delete filestore instance
    ├── waitInstanceReady.go      → Wait for instance to become ready
    ├── waitInstanceDeleted.go    → Wait for deletion to complete
    └── updateStatus.go           → Update NfsInstance status
```

## Reconciliation Flow

The v2 reconciler follows this streamlined action sequence:

1. **AddCommonFinalizer** → Ensure finalizer for cleanup
2. **pollOperation** → Poll pending GCP operations (if any)
3. **loadInstance** → Fetch instance from GCP Filestore API
4. **Branch: Not Marked for Deletion** (`IfElse` composition)
   - **createInstance** → Create instance if it doesn't exist
   - **waitInstanceReady** → Wait for instance to become ready
   - **updateInstance** → Update instance if changes needed
   - **updateStatus** → Update NfsInstance status
   - **StopAndForget** → Complete reconciliation
5. **Branch: Marked for Deletion**
   - **deleteInstance** → Delete filestore instance if it exists
   - **waitInstanceDeleted** → Wait for deletion to complete
   - **RemoveCommonFinalizer** → Remove finalizer
   - **StopAndForget** → Complete reconciliation

## State Machine

The v2 implementation uses a simplified approach with wait actions instead of a complex state machine:

| Action | Condition | Behavior |
|--------|-----------|----------|
| **createInstance** | Instance doesn't exist in GCP | Create instance, store operation name |
| **waitInstanceReady** | Instance is being created or updated | Wait until instance state is READY |
| **updateInstance** | Instance exists but differs from spec | Update instance, store operation name |
| **deleteInstance** | Instance exists and marked for deletion | Delete instance, store operation name |
| **waitInstanceDeleted** | Instance is being deleted | Wait until instance is fully deleted |
| **pollOperation** | Operation name stored in status | Poll operation until completion |

The status conditions track the reconciliation state:
- `Ready`: Instance is ready and matches spec
- `Processing`: Operation in progress (create/update/delete)
- `Error`: Error occurred during reconciliation

## Client Interface

The client interface provides a clean abstraction over the GCP Filestore API using modern protobuf types:

```go
type FilestoreClient interface {
    // GetInstance retrieves a Filestore instance by name.
    GetInstance(ctx context.Context, projectId, location, instanceId string) (*filestorepb.Instance, error)

    // CreateInstance creates a new Filestore instance.
    // Returns the operation name for tracking.
    CreateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance) (string, error)

    // UpdateInstance updates an existing Filestore instance.
    // The updateMask specifies which fields should be updated.
    // Returns the operation name for tracking.
    UpdateInstance(ctx context.Context, projectId, location, instanceId string, instance *filestorepb.Instance, updateMask []string) (string, error)

    // DeleteInstance deletes a Filestore instance.
    // Returns the operation name for tracking.
    DeleteInstance(ctx context.Context, projectId, location, instanceId string) (string, error)

    // GetOperation retrieves the status of a long-running operation.
    // Returns true if the operation is done, false otherwise.
    GetOperation(ctx context.Context, operationName string) (bool, error)
}
```

The client uses `cloud.google.com/go/filestore/apiv1` (modern protobuf-based API) and follows the NEW pattern for client initialization via `GcpClientProvider`.

## State Structure

```go
type State struct {
    types.State                           // Shared NfsInstance state (OLD pattern)
    
    filestoreClient v2client.FilestoreClient  // GCP Filestore client
    instance        *filestorepb.Instance     // Cached GCP instance (protobuf type)
}
```

The state is simple and focuses on essential data:
- Embeds `types.State` for OLD pattern compatibility
- Uses modern protobuf types from `cloud.google.com/go/filestore/apiv1/filestorepb`
- Client initialization follows NEW pattern (GcpClientProvider)

## Key Differences from V1

| Aspect | V1 | V2 |
|--------|----|----|
| **GCP Client** | `google.golang.org/api/file/v1` (REST wrapper) | `cloud.google.com/go/filestore/apiv1` (protobuf) |
| **Client Pattern** | OLD pattern (ClientProvider) | NEW pattern (GcpClientProvider) |
| **File organization** | Mixed/unclear structure | Flat action files at package root |
| **Action composition** | Sequential with manual branching | `IfElse` composition with clear branching |
| **Operation tracking** | Stored in state with manual polling | Operation name in status, explicit poll/wait actions |
| **Wait logic** | Inline in actions | Separate wait actions (waitInstanceReady, waitInstanceDeleted) |
| **State struct** | Many intermediate fields | Simple with protobuf types |
| **Validation** | Mixed in actions | Inline in create/update actions |
| **Testing** | Many trivial tests | Focus on business logic |

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

### Current Testing Status
- **Unit Tests**: Limited to utility functions (util_test.go - tier conversion logic)
- **Controller Tests**: Integration tests in [internal/controller/cloud-control/nfsinstance_gcp_v2_test.go](/internal/controller/cloud-control/nfsinstance_gcp_v2_test.go)
  - Create and delete lifecycle
  - Create, update, and delete lifecycle
  - Uses testinfra framework with GCP mock server
  - Feature flag gated (skips when `gcpNfsInstanceV2` disabled)
- **Mocks**: Mock client available in client/mockFilestoreClient.go

### Test Coverage
- ✅ Complete reconciliation flows (create → ready → update → delete)
- ✅ GCP Filestore instance creation and tracking
- ✅ Status updates (Ready condition, host, path, capacity)
- ✅ Capacity updates (scale up)
- ✅ Deletion with finalizer cleanup

### Recommended Approach for Additional Tests
When adding more tests, focus on:
- Error handling and edge cases
- Error recovery scenarios
- Validation logic
- **Avoid** testing simple getters/setters
- **Avoid** testing framework code

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
