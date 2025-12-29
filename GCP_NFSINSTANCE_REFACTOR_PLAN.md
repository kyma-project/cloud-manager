# GCP NfsInstance Refactoring Plan

**Date**: 2025-12-25  
**Target**: `pkg/kcp/provider/gcp/nfsinstance`  
**Pattern**: OLD Reconciler Pattern (maintain compatibility)  
**Objective**: Streamline implementation, modernize client if available, preserve v1, create v2

---

## Executive Summary

The GCP NfsInstance implementation in `pkg/kcp/provider/gcp/nfsinstance` follows the OLD reconciler pattern (multi-provider CRD) and needs refactoring to:

1. **Preserve current implementation** in a `v1/` folder for stability
2. **Create streamlined v2** in a `v2/` folder with improved structure
3. **Evaluate modern GCP client** (cloud.google.com/go/filestore) availability
4. **Maintain OLD pattern** since NfsInstance is a legacy multi-provider CRD
5. **Improve code organization** without changing CRD schema or reconciliation logic

### Current Status

- **CRD**: `cloud-control.kyma-project.io/v1beta1.NfsInstance` (multi-provider)
- **Pattern**: OLD (3-layer state hierarchy)
- **Client**: `google.golang.org/api/file/v1` (REST API wrapper)
- **Location**: `pkg/kcp/provider/gcp/nfsinstance/`
- **Actions**: 10+ action files with tests
- **State Layers**: `focal.State` → `types.State` → GCP `State`

---

## Refactoring Checklist

### Phase 1: Assessment & Preparation

#### 1.1 Current Implementation Analysis
- [x] Catalog all existing files in `pkg/kcp/provider/gcp/nfsinstance/`
- [x] Document current actions and their purposes
- [x] Identify state management approach
- [x] Map client dependencies
- [x] Document current reconciliation flow
- [x] List all GCP Filestore API calls used
- [x] **Review v1 tests** - identify low-value tests to exclude from v2
- [x] **Categorize v1 tests**: business logic vs trivial vs redundant
- [x] Document configuration dependencies

**Files to analyze**:
```
pkg/kcp/provider/gcp/nfsinstance/
├── new.go                      → Entry point, action composition
├── state.go                    → State definition, factory
├── checkGcpOperation.go        → GCP operation polling
├── checkNUpdateState.go        → State machine logic
├── checkUpdateMask.go          → Update field tracking
├── loadNfsInstance.go          → Load from GCP
├── syncNfsInstance.go          → Create/Update/Delete GCP resource
├── validateAlways.go           → Pre-flight validation
├── validatePostCreate.go       → Post-creation validation
├── validations.go              → Shared validation logic
├── client/filestoreClient.go   → GCP Filestore client wrapper
└── testdata/                   → Test fixtures
```

#### 1.2 GCP Client Library Research
- [x] Check if `cloud.google.com/go/filestore` exists (Result: **YES** - https://pkg.go.dev/cloud.google.com/go/filestore)
- [x] Review `cloud.google.com/go/filestore` API documentation
- [x] Compare `cloud.google.com/go/filestore` with `google.golang.org/api/file/v1`
- [x] Verify all required operations are supported in modern client
- [x] Check for additional features in modern client
- [x] Compare with modern GCP client patterns (compute, redis examples)
- [x] Evaluate migration effort and benefits
- [x] Document API capabilities and limitations
- [x] Check for any deprecation notices
- [x] Make final client library decision

**Findings**:
- **Modern client EXISTS**: `cloud.google.com/go/filestore` at https://pkg.go.dev/cloud.google.com/go/filestore
- Current implementation uses `google.golang.org/api/file/v1` (REST API wrapper)
- Modern GCP services in codebase use `cloud.google.com/go/*` packages (compute, redis, etc.)
- **Decision**: Evaluate modern client thoroughly, likely migrate to `cloud.google.com/go/filestore` in v2

#### 1.3 Dependencies & Integration Points
- [x] Map all imports and dependencies
- [x] Identify shared code with other providers (AWS, Azure, SAP)
- [x] Document integration with `pkg/kcp/nfsinstance/` (shared layer)
- [x] List feature flags and configuration
- [x] Identify controller setup in `internal/controller/cloud-control/nfsinstance_controller.go`
- [x] Document IpRange dependency

---

### Phase 2: Feature Flag Setup

#### 2.1 Define Feature Flag
- [x] Add feature flag to feature flag schema
- [x] Name: `gcpNfsInstanceV2` (finalized)
- [x] Default value: `false` (use v1)
- [x] Add flag to `pkg/feature/ffGcpNfsInstanceV2.go`
- [x] Add flag to `pkg/feature/ff_ga.yaml` and `pkg/feature/ff_edge.yaml`
- [x] Document flag purpose and behavior

**Feature flag definition**:
```go
// pkg/feature/ffGcpNfsInstanceV2.go
package feature

import (
	"context"
)

const GcpNfsInstanceV2FlagName = "gcpNfsInstanceV2"

var GcpNfsInstanceV2 = &gcpNfsInstanceV2Info{}

type gcpNfsInstanceV2Info struct{}

func (k *gcpNfsInstanceV2Info) Value(ctx context.Context) bool {
	return provider.BoolVariation(ctx, GcpNfsInstanceV2FlagName, false)
}
```

**Configuration**:
```yaml
# pkg/feature/ff_ga.yaml
gcpNfsInstanceV2:
  variations:
    enabled: true
    disabled: false
  defaultRule:
    variation: disabled  # Default to v1
```

#### 2.2 Feature Flag Access Pattern
- [x] Add feature flag getter method
- [x] Document flag usage pattern
- [x] Plan flag check locations
- [x] Design fallback behavior

**Usage pattern**:
```go
if feature.GcpNfsInstanceV2.Value(ctx) {
    // Use v2 implementation
    return gcpnfsinstancev2.New(r.gcpStateFactoryV2)(ctx, state)
} else {
    // Use v1 implementation (default)
    return gcpnfsinstancev1.New(r.gcpStateFactoryV1)(ctx, state)
}
```

#### 2.3 Test Feature Flag
- [x] Unit tests for flag definition
- [x] Test default value (false)
- [x] Test flag override
- [x] Document testing approach

---

### Phase 3: V1 Preservation

#### 3.1 Create V1 Directory Structure
- [x] Create `pkg/kcp/provider/gcp/nfsinstance/v1/` directory
- [x] Move all current files to v1 folder
- [x] Update package declarations to `v1`
- [x] Create v1 README documenting legacy status
- [x] Add v1-specific godoc comments

**Commands**:
```bash
mkdir -p pkg/kcp/provider/gcp/nfsinstance/v1
mkdir -p pkg/kcp/provider/gcp/nfsinstance/v1/client
mkdir -p pkg/kcp/provider/gcp/nfsinstance/v1/testdata
```

#### 3.2 Move and Update V1 Files
- [x] Move `*.go` files to v1/
- [x] Move `*_test.go` files to v1/
- [x] Move `client/` directory to v1/client/
- [x] Move `testdata/` to v1/testdata/
- [x] Update package name in all files: `package nfsinstance` → `package v1`
- [x] Update imports in v1 files
- [x] Update client package name: `package client` → `package v1client`

**Files to move**:
```
checkGcpOperation.go → v1/checkGcpOperation.go
checkGcpOperation_test.go → v1/checkGcpOperation_test.go
checkNUpdateState.go → v1/checkNUpdateState.go
checkNUpdateState_test.go → v1/checkNUpdateState_test.go
checkUpdateMask.go → v1/checkUpdateMask.go
checkUpdateMask_test.go → v1/checkUpdateMask_test.go
loadNfsInstance.go → v1/loadNfsInstance.go
loadNfsInstance_test.go → v1/loadNfsInstance_test.go
new.go → v1/new.go
state.go → v1/state.go
state_test.go → v1/state_test.go
syncNfsInstance.go → v1/syncNfsInstance.go
syncNfsInstance_test.go → v1/syncNfsInstance_test.go
validateAlways.go → v1/validateAlways.go
validateAlways_test.go → v1/validateAlways_test.go
validatePostCreate.go → v1/validatePostCreate.go
validatePostCreate_test.go → v1/validatePostCreate_test.go
validations.go → v1/validations.go
validations_test.go → v1/validations_test.go
client/filestoreClient.go → v1/client/filestoreClient.go
testdata/* → v1/testdata/*
```

#### 3.3 Update V1 References
- [x] Update imports in `pkg/kcp/nfsinstance/reconciler.go`
- [x] Update imports in `internal/controller/cloud-control/nfsinstance_controller.go`
- [x] Add v1 import aliases where needed
- [x] Create v1 state factory wrapper if needed
- [x] Verify all tests still pass with v1 imports
- [x] Add deprecation notices in v1 godoc

**Import changes example**:
```go
// Before
import gcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance"

// After (temporary - will use v2 later)
import gcpnfsinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1"
```

---

### Phase 4: V2 Structure Design

#### 4.1 Define V2 Architecture
- [x] Design streamlined directory structure
- [x] Plan state management simplification
- [x] Define action organization strategy
- [x] Plan client interface improvements
- [x] Document differences from v1
- [x] Create v2 architecture document

**Proposed V2 structure**:
```
pkg/kcp/provider/gcp/nfsinstance/v2/
├── reconcile.go               → Entry point, state factory exposure
├── state.go                   → Streamlined state definition
├── actions.go                 → Composed action flow
├── client/
│   ├── filestore.go          → Client interface & implementation
│   ├── types.go              → Client types & constants
│   └── mock.go               → Mock client for tests
├── operations/
│   ├── create.go             → Create filestore instance
│   ├── update.go             → Update filestore instance
│   ├── delete.go             → Delete filestore instance
│   ├── load.go               → Load from GCP
│   └── operation.go          → Operation polling
├── validation/
│   ├── preflight.go          → Pre-flight validations
│   ├── postcreate.go         → Post-creation validations
│   └── helpers.go            → Validation utilities
├── state/
│   ├── machine.go            → State machine logic
│   ├── status.go             → Status update helpers
│   └── comparison.go         → Filestore comparison logic
└── README.md                  → V2 documentation
```

#### 4.2 Create V2 Base Files
- [x] Create v2 directory structure
- [x] Create v2 README with architecture overview
- [x] Create `reconcile.go` with state factory interface
- [x] Create `state.go` with simplified state struct
- [x] Create `actions.go` with action composition
- [x] Add comprehensive godoc comments

#### 4.3 Plan Streamlining Improvements
- [x] Consolidate validation logic
- [x] Simplify state machine transitions
- [x] Reduce code duplication
- [x] Improve error handling consistency
- [x] Better separation of concerns
- [x] More idiomatic Go patterns

**Key improvements**:
1. **Action organization**: Group related actions in packages
2. **State management**: Clearer state transitions
3. **Validation**: Centralized validation logic
4. **Client**: Better abstraction and mocking
5. **Testing**: Easier unit test setup
6. **Error handling**: Consistent patterns

---

### Phase 5: V2 Client Implementation

#### 5.1 Evaluate Modern Client Library
- [x] Study `cloud.google.com/go/filestore` API documentation
- [x] Review apiv1 package structure
- [x] Check CloudFilestoreManagerClient methods
- [x] Verify CRUD operations support
- [x] Check operation polling support
- [x] Compare with OLD client capabilities
- [x] List migration requirements
- [x] Assess compatibility with existing patterns

**Modern client package**: `cloud.google.com/go/filestore/apiv1`

**Key types to evaluate**:
- `CloudFilestoreManagerClient` - Main client
- `Instance` - Filestore instance representation
- `CreateInstanceRequest` - Create operation
- `UpdateInstanceRequest` - Update operation
- `DeleteInstanceRequest` - Delete operation
- Operations handling

#### 5.2 Design Client Interface
- [x] Define FilestoreClient interface
- [x] Document all required methods
- [x] Design mock client interface
- [x] Plan error handling strategy
- [x] Define retry logic (if needed)
- [x] Create client factory pattern

**Interface design**:
```go
type FilestoreClient interface {
    GetInstance(ctx context.Context, project, location, name string) (*Instance, error)
    CreateInstance(ctx context.Context, project, location, name string, instance *Instance) (*Operation, error)
    UpdateInstance(ctx context.Context, project, location, name string, mask []string, instance *Instance) (*Operation, error)
    DeleteInstance(ctx context.Context, project, location, name string) (*Operation, error)
    GetOperation(ctx context.Context, project, operationName string) (*Operation, error)
}
```

#### 5.3 Implement Client Wrapper
- [ ] Add `cloud.google.com/go/filestore` to go.mod if not present
- [ ] Create client implementation using `cloud.google.com/go/filestore/apiv1`
- [ ] Wrap GCP Filestore API calls
- [ ] Add logging and metrics
- [ ] Implement error handling
- [ ] Add retry logic if appropriate
- [ ] Document API rate limits

**Implementation file**: `v2/client/filestore.go`

#### 5.4 Create Mock Client
- [ ] Implement mock FilestoreClient
- [ ] Add configurable responses
- [ ] Support error injection
- [ ] Enable test scenarios
- [ ] Document mock usage

**Mock file**: `v2/client/mock.go`

#### 5.5 Client Testing
- [ ] Unit tests for client wrapper
- [ ] Test error scenarios
- [ ] Test retry logic
- [ ] Test metrics collection
- [ ] Mock client validation tests

#### Phase 5 Client Implementation Summary (COMPLETED 2025-12-29)

**Modern Client Adopted**: `cloud.google.com/go/filestore/apiv1`
- Added filestore module to go.mod
- Integrated CloudFilestoreManagerClient into GcpClients struct
- Created FilestoreClient interface with business operations
- Implemented client wrapper following NEW pattern (GcpClientProvider)
- Created comprehensive mock client with:
  - In-memory instance and operation storage
  - Auto-complete and manual operation modes
  - Error injection capabilities
  - State management (CREATING → READY, DELETING, etc.)
- Added unit tests for mock client (all CRUD operations)
- Used modern protobuf types (`filestorepb.Instance`, `filestorepb.OperationMetadata`)
- Consistent with existing GCP resources (compute, redis)

**Files Created**:
- `pkg/kcp/provider/gcp/nfsinstance/v2/client/filestoreClient.go` - Client implementation
- `pkg/kcp/provider/gcp/nfsinstance/v2/client/mockFilestoreClient.go` - Mock for testing

**Integration**:
- Added Filestore field to `GcpClients` struct
- Initialized in `NewGcpClients()` with proper token provider
- Client accessible via `gcpClients.Filestore`

---

### Phase 6: V2 State Management

#### 6.1 Implement State Structure
- [ ] Create streamlined State struct
- [ ] Embed `types.State` (maintain OLD pattern compatibility)
- [ ] Add client reference
- [ ] Add cached GCP resources
- [ ] Add operation tracking
- [ ] Document state fields

**State structure**:
```go
type State struct {
    types.State                           // Shared NfsInstance state (OLD pattern)
    
    client          FilestoreClient       // GCP Filestore client
    instance        *file.Instance        // Cached GCP instance
    operation       *file.Operation       // In-progress operation
    operationType   OperationType         // ADD, MODIFY, DELETE, NONE
    
    // State machine
    currentState    v1beta1.StatusState   // Current lifecycle state
    
    // Update tracking
    updateFields    []string              // Fields to update
    
    // Validation results
    validationErrs  []error               // Validation errors
}
```

#### 6.2 Implement State Factory
- [ ] Create StateFactory interface
- [ ] Implement state factory
- [ ] Handle client initialization
- [ ] Add error handling
- [ ] Support dependency injection
- [ ] Write factory tests

**Factory pattern** (OLD pattern compatible):
```go
type StateFactory interface {
    NewState(ctx context.Context, nfsInstanceState types.State) (*State, error)
}
```

#### 6.3 State Helper Methods
- [ ] Add state query methods
- [ ] Add state mutation methods
- [ ] Add comparison helpers
- [ ] Add conversion helpers
- [ ] Document method purposes
- [ ] Write helper tests

**Examples**:
- `doesFilestoreMatch()` → Check if GCP instance matches desired state
- `getGcpLocation()` → Resolve GCP region/zone
- `toGcpInstance()` → Convert CRD to GCP API model
- `needsUpdate()` → Determine if update needed

---

### Phase 7: V2 Action Implementation

#### 7.1 Validation Actions
- [ ] Implement pre-flight validations (`validation/preflight.go`)
  - [ ] Capacity validation
  - [ ] Tier validation
  - [ ] Network validation
  - [ ] IpRange validation
- [ ] Implement post-create validations (`validation/postcreate.go`)
  - [ ] No scale-down validation
  - [ ] Tier consistency validation
- [ ] Add validation helpers (`validation/helpers.go`)
- [ ] Write validation tests

**Action signature** (OLD pattern compatible):
```go
func validatePreflight(ctx context.Context, st composed.State) (error, context.Context)
```

#### 7.2 Load Actions
- [ ] Implement loadInstance (`operations/load.go`)
  - [ ] Call GCP Get API
  - [ ] Handle not found
  - [ ] Cache in state
  - [ ] Set state flags
- [ ] Write load action tests
- [ ] Test not-found scenarios
- [ ] Test API error handling

#### 7.3 Create Actions
- [ ] Implement createInstance (`operations/create.go`)
  - [ ] Build GCP instance spec
  - [ ] Handle protocol selection
  - [ ] Call Create API
  - [ ] Store operation
  - [ ] Set creating status
- [ ] Write create action tests
- [ ] Test validation integration
- [ ] Test error scenarios

#### 7.4 Update Actions
- [ ] Implement checkUpdateNeeded (`state/comparison.go`)
  - [ ] Compare capacity
  - [ ] Compare configuration
  - [ ] Build update mask
- [ ] Implement updateInstance (`operations/update.go`)
  - [ ] Apply update mask
  - [ ] Call Patch API
  - [ ] Store operation
  - [ ] Set updating status
- [ ] Write update action tests
- [ ] Test various update scenarios

#### 7.5 Delete Actions
- [ ] Implement deleteInstance (`operations/delete.go`)
  - [ ] Call Delete API
  - [ ] Store operation
  - [ ] Set deleting status
- [ ] Write delete action tests
- [ ] Test finalizer integration

#### 7.6 Operation Polling
- [ ] Implement pollOperation (`operations/operation.go`)
  - [ ] Get operation status
  - [ ] Handle completion
  - [ ] Handle errors
  - [ ] Calculate retry delay
- [ ] Write operation polling tests
- [ ] Test timeout scenarios

#### 7.7 State Machine
- [ ] Implement state transitions (`state/machine.go`)
  - [ ] Map GCP states to CRD states
  - [ ] Handle state changes
  - [ ] Trigger status updates
- [ ] Write state machine tests
- [ ] Document state flow

#### 7.8 Status Updates
- [ ] Implement updateStatus (`state/status.go`)
  - [ ] Update NfsInstance status
  - [ ] Set conditions
  - [ ] Set hosts/capacity
  - [ ] Handle state data
- [ ] Write status update tests
- [ ] Test condition management

---

### Phase 8: V2 Action Composition

#### 8.1 Compose Main Action Flow
- [ ] Create main action in `actions.go`
- [ ] Compose action pipeline
- [ ] Handle feature flags
- [ ] Add finalizer management
- [ ] Implement branching logic (create vs delete)
- [ ] Add stop conditions

**Action composition**:
```go
func New(stateFactory StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        state, err := stateFactory.NewState(ctx, st.(types.State))
        if err != nil {
            // Handle state creation error
        }
        
        return composed.ComposeActions(
            "gcpNfsInstance",
            validatePreflight,
            actions.AddCommonFinalizer(),
            pollOperation,
            loadInstance,
            validatePostCreate,
            stateMachine,
            checkUpdateNeeded,
            composed.IfElse(
                composed.Not(composed.MarkedForDeletionPredicate),
                createOrUpdateFlow,
                deleteFlow,
            ),
        )(ctx, state)
    }
}
```

#### 8.2 Create/Update Flow
- [ ] Compose create/update actions
- [ ] Order actions logically
- [ ] Handle conditional execution
- [ ] Add logging
- [ ] Document flow

```go
var createOrUpdateFlow = composed.ComposeActions(
    "create-update",
    syncInstance,           // Create or update
    pollOperation,          // Wait for completion
    updateStatus,           // Update CRD status
    composed.StopAndForgetAction,
)
```

#### 8.3 Delete Flow
- [ ] Compose delete actions
- [ ] Handle cleanup
- [ ] Remove finalizer
- [ ] Add logging
- [ ] Document flow

```go
var deleteFlow = composed.ComposeActions(
    "delete",
    deleteInstance,
    pollOperation,
    actions.RemoveCommonFinalizer(),
    composed.StopAndForgetAction,
)
```

---

### Phase 9: V2 Testing

#### 9.1 Unit Tests - Business Logic Focus
- [ ] **Review v1 tests** - identify useless tests to avoid replicating
- [ ] Test state factory (error cases, client initialization)
- [ ] Test state helpers (business logic only - comparison, conversion)
- [ ] Test validation actions (business rules - capacity limits, tier rules)
- [ ] Test load actions (error handling, not-found scenarios)
- [ ] Test create actions (GCP API request building, error handling)
- [ ] Test update actions (update mask calculation, business rules)
- [ ] Test delete actions (error handling, operation tracking)
- [ ] Test operation polling (state transitions, timeout logic)
- [ ] Test state machine (critical state transitions only)
- [ ] Test status updates (condition logic, status field mapping)
- [ ] **SKIP trivial tests**: getters, setters, simple type conversions
- [ ] **SKIP mock validation tests**: tests that only verify mock behavior
- [ ] **SKIP redundant tests**: duplicate coverage of same logic
- [ ] Achieve 70%+ code coverage (focused on business logic)

**Testing Principles for V2**:
- ✅ Test business logic and decision points
- ✅ Test error handling and edge cases
- ✅ Test state transitions and API interactions
- ❌ Don't test simple getters/setters
- ❌ Don't test framework code or library behavior
- ❌ Don't test mock implementations
- ❌ Don't duplicate tests without adding value

**Test structure**:
```
v2/
├── state_test.go
├── actions_test.go
├── client/
│   ├── filestore_test.go
│   └── mock_test.go
├── operations/
│   ├── create_test.go
│   ├── update_test.go
│   ├── delete_test.go
│   ├── load_test.go
│   └── operation_test.go
├── validation/
│   ├── preflight_test.go
│   └── postcreate_test.go
└── state/
    ├── machine_test.go
    ├── status_test.go
    └── comparison_test.go
```

#### 9.2 Integration Tests - Critical Paths Only
- [ ] Create test fixtures (minimal, reusable)
- [ ] Test complete reconciliation flow (happy path)
- [ ] Test create → ready flow (with GCP operations)
- [ ] Test update flow (capacity change)
- [ ] Test delete flow (with cleanup)
- [ ] Test error recovery (GCP API failures)
- [ ] Test operation timeout (long-running operations)
- [ ] Use testinfra framework
- [ ] **FOCUS**: End-to-end critical business scenarios
- [ ] **SKIP**: Testing every possible combination of states

#### 9.3 Mock Testing
- [ ] Verify mock client works
- [ ] Test error injection
- [ ] Test operation scenarios
- [ ] Validate test isolation

---

### Phase 10: V2 Integration with Feature Flag

#### 10.1 Update Shared Layer for Dual Support
- [ ] Update `pkg/kcp/nfsinstance/reconciler.go` to support both v1 and v2
- [ ] Accept both v1 and v2 state factories
- [ ] Implement feature flag check
- [ ] Route to v1 by default (feature flag = false)
- [ ] Route to v2 when feature flag = true
- [ ] Maintain backward compatibility
- [ ] Document migration path

**Reconciler update** (dual support):
```go
func NewNfsInstanceReconciler(
    composedStateFactory composed.StateFactory,
    focalStateFactory focal.StateFactory,
    awsStateFactory awsnfsinstance.StateFactory,
    azureStateFactory azurenfsinstance.StateFactory,
    gcpStateFactoryV1 gcpnfsinstancev1.StateFactory,  // V1
    gcpStateFactoryV2 gcpnfsinstancev2.StateFactory,  // V2
    sapStateFactory sapnfsinstance.StateFactory,
) NfsInstanceReconciler {
    return &nfsInstanceReconciler{
        // ...
        gcpStateFactoryV1: gcpStateFactoryV1,
        gcpStateFactoryV2: gcpStateFactoryV2,
    }
}
```

**Feature flag routing in action**:
```go
func (r *nfsInstanceReconciler) gcpAction(ctx context.Context, state composed.State) (error, context.Context) {
    if feature.GcpNfsInstanceUseV2.Value(ctx) {
        return gcpnfsinstancev2.New(r.gcpStateFactoryV2)(ctx, state)
    }
    return gcpnfsinstancev1.New(r.gcpStateFactoryV1)(ctx, state)
}
```

#### 10.2 Update Controller Setup for Dual Support
- [ ] Update `internal/controller/cloud-control/nfsinstance_controller.go`
- [ ] Add v1 and v2 imports
- [ ] Initialize both v1 and v2 state factories
- [ ] Pass both factories to reconciler
- [ ] Verify controller registration

**Controller setup** (dual support):
```go
import (
    gcpnfsinstancev1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1"
    gcpnfsinstancev2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2"
)

func SetupNfsInstanceReconciler(...) error {
    // Create both v1 and v2 state factories
    gcpStateFactoryV1 := gcpnfsinstancev1.NewStateFactory(filestoreClientProviderV1, env)
    gcpStateFactoryV2 := gcpnfsinstancev2.NewStateFactory(filestoreClientProviderV2, env)
    
    return NewNfsInstanceReconciler(
        nfsinstance.NewNfsInstanceReconciler(
            composedStateFactory,
            focalStateFactory,
            awsStateFactory,
            azureStateFactory,
            gcpStateFactoryV1,  // V1
            gcpStateFactoryV2,  // V2
            sapStateFactory,
        ),
    ).SetupWithManager(kcpManager)
}
```

#### 10.3 Feature Flag Testing
- [ ] Test with feature flag = false (v1 behavior)
- [ ] Test with feature flag = true (v2 behavior)
- [ ] Test flag toggling
- [ ] Verify no runtime errors
- [ ] Test both paths independently

#### 10.4 Update Imports
- [ ] Find all v1 imports
- [ ] Replace with v2 imports
- [ ] Verify compilation
- [ ] Update import aliases if needed

#### 10.5 Configuration
- [ ] Verify config compatibility
- [ ] Update feature flags if needed
- [ ] Document configuration changes
- [ ] Update config examples

---

### Phase 11: Validation & Testing

#### 11.1 Compilation
- [ ] Run `make build`
- [ ] Fix compilation errors
- [ ] Verify no warnings
- [ ] Check vendor dependencies

#### 11.2 Unit Tests
- [ ] Run `make test`
- [ ] Fix failing tests
- [ ] Verify v1 tests still pass
- [ ] Verify v2 tests pass
- [ ] Check test coverage

#### 11.3 Feature Flag Tests
- [ ] Test v1 path (default, flag = false)
- [ ] Test v2 path (flag = true)
- [ ] Test flag switching at runtime
- [ ] Verify isolation between v1 and v2

#### 11.4 Code Generation
- [ ] Run `make manifests`
- [ ] Verify CRDs unchanged
- [ ] Run `make generate`
- [ ] Commit generated code

#### 11.5 Linting
- [ ] Run `make lint` (if exists)
- [ ] Fix linting errors
- [ ] Check code style
- [ ] Verify godoc

---

### Phase 12: Documentation

#### 12.1 Code Documentation
- [ ] Update all godoc comments
- [ ] Document public interfaces
- [ ] Add package documentation
- [ ] Document state transitions
- [ ] Add usage examples

#### 12.2 Architecture Documentation
- [ ] Create v2/README.md
- [ ] Document v2 architecture
- [ ] Explain differences from v1
- [ ] Add flow diagrams
- [ ] Document design decisions

#### 12.3 Feature Flag Documentation
- [ ] Document feature flag usage
- [ ] Explain v1 vs v2 behavior
- [ ] Document default (v1)
- [ ] Explain how to enable v2
- [ ] Add troubleshooting guide

#### 12.4 Migration Guide
- [ ] Document v1 → v2 changes
- [ ] List breaking changes (if any)
- [ ] Provide upgrade instructions
- [ ] Document rollback procedure
- [ ] Add troubleshooting section

**Migration doc location**: `docs/contributor/gcp-nfsinstance-v2-migration.md`

#### 12.5 Agent Instructions
- [ ] Update AGENTS.md if needed
- [ ] Update architecture docs if patterns changed
- [ ] Add v2 to reference documentation
- [ ] Document v2 in guides

---

### Phase 13: Cleanup & Finalization

#### 13.1 Code Review
- [ ] Self-review all changes
- [ ] Check for code duplication
- [ ] Verify error handling
- [ ] Check logging consistency
- [ ] Verify test coverage

#### 13.2 V1 Deprecation (Future)
- [ ] Add deprecation notices to v1
- [ ] Document v1 sunset timeline
- [ ] Add warnings in v1 godoc
- [ ] Plan v1 removal (future)

**Deprecation notice**:
```go
// Package v1 provides the legacy GCP NfsInstance implementation.
//
// Deprecated: This package is maintained for backward compatibility only.
// New code should use v2 package. This package will be removed in a future release.
package v1
```

#### 13.3 Final Verification
- [ ] Run full test suite
- [ ] Verify e2e compatibility (if applicable)
- [ ] Check metrics still work
- [ ] Verify feature flags
- [ ] Test with real GCP account (manual)

#### 13.4 Commit Strategy
- [ ] Commit v1 move separately
- [ ] Commit v2 structure
- [ ] Commit v2 implementation in logical chunks
- [ ] Commit integration updates
- [ ] Commit documentation
- [ ] Squash or organize commits

**Suggested commits**:
1. `feat(feature): add gcpNfsInstanceUseV2 feature flag`
2. `refactor(gcp): move NfsInstance v1 to subdirectory`
3. `feat(gcp): add NfsInstance v2 structure and modern client`
4. `feat(gcp): add NfsInstance v2 state management`
5. `feat(gcp): add NfsInstance v2 actions`
6. `feat(gcp): integrate NfsInstance v2 with feature flag routing`
7. `docs(gcp): add NfsInstance v2 documentation and migration guide`
8. `test(gcp): add NfsInstance v2 tests and feature flag tests`

---

## Success Criteria

### Must Have
- [ ] Feature flag `gcpNfsInstanceUseV2` implemented with default = false
- [ ] V1 code preserved and functional in `v1/` directory
- [ ] V2 code implemented in `v2/` directory with streamlined structure
- [ ] V2 uses modern `cloud.google.com/go/filestore` client
- [ ] Both v1 and v2 integrated with feature flag routing
- [ ] All existing tests pass (v1 and v2)
- [ ] Feature flag tests verify both paths
- [ ] Code coverage maintained or improved
- [ ] No CRD API changes
- [ ] No breaking changes to reconciliation behavior
- [ ] OLD pattern maintained (3-layer state hierarchy)
- [ ] Documentation complete with feature flag usage

### Should Have
- [ ] Improved code organization and readability
- [ ] Reduced code duplication
- [ ] Better separation of concerns
- [ ] Modern GCP client patterns consistent with other resources
- [ ] Focused unit tests for v2 (business logic only, no trivial tests)
- [ ] Leaner test suite compared to v1
- [ ] Feature flag documentation
- [ ] Migration guide for developers
- [ ] V2 rollout plan

### Nice to Have
- [ ] Performance improvements
- [ ] Better error messages
- [ ] Enhanced logging
- [ ] Metrics improvements
- [ ] Integration test coverage

---

## Risk Assessment

### High Risk
- **Breaking reconciliation behavior**: Mitigation = preserve v1, extensive testing
- **State hierarchy changes**: Mitigation = maintain OLD pattern strictly
- **CRD incompatibility**: Mitigation = no CRD changes

### Medium Risk
- **Test coverage gaps**: Mitigation = comprehensive v2 testing
- **Integration issues**: Mitigation = gradual rollout, feature flag
- **Client API changes**: Mitigation = use stable google.golang.org/api

### Low Risk
- **Documentation gaps**: Mitigation = documentation phase
- **Code style inconsistency**: Mitigation = linting, review

---

## Timeline Estimate

**Total Estimated Effort**: 4-6 days

| Phase | Estimated Time | Priority |
|-------|---------------|----------|
| Phase 1: Assessment | 4-6 hours | High |
| Phase 2: Feature Flag | 2-3 hours | High |
| Phase 3: V1 Preservation | 2-3 hours | High |
| Phase 4: V2 Design | 3-4 hours | High |
| Phase 5: Client (modern) | 5-7 hours | High |
| Phase 6: State | 4-5 hours | High |
| Phase 7: Actions | 8-12 hours | High |
| Phase 8: Composition | 2-3 hours | High |
| Phase 9: Testing | 6-8 hours | High |
| Phase 10: Integration | 4-6 hours | High |
| Phase 11: Validation | 3-4 hours | High |
| Phase 12: Documentation | 4-5 hours | Medium |
| Phase 13: Cleanup | 2-3 hours | Medium |

---

## Notes

### Phase 1 Assessment Findings (Completed 2025-12-26)

#### Current Reconciliation Flow
The GCP NfsInstance reconciler follows this action sequence:
1. **validateAlways** → Pre-flight validation (capacity, tier, network)
2. **AddCommonFinalizer** → Ensure finalizer for cleanup
3. **checkGcpOperation** → Poll pending GCP operations
4. **loadNfsInstance** → Fetch instance from GCP Filestore API
5. **validatePostCreate** → Post-creation validation (no scale-down for BASIC tiers)
6. **checkNUpdateState** → State machine logic + determine operation (ADD/MODIFY/DELETE/NONE)
7. **checkUpdateMask** → Build update mask for MODIFY operations
8. **syncNfsInstance** → Execute operation (Create/Patch/Delete)
9. **RemoveCommonFinalizer** (if deleted) → Clean up finalizer

#### GCP Filestore API Calls Used (via `google.golang.org/api/file/v1`)
1. **Instances.Get** → `GetFilestoreInstance()` - Load instance
2. **Instances.Create** → `CreateFilestoreInstance()` - Create new instance
3. **Instances.Patch** → `PatchFilestoreInstance()` - Update existing instance
4. **Instances.Delete** → `DeleteFilestoreInstance()` - Delete instance
5. **Operations.Get** → `GetFilestoreOperation()` - Poll operation status

All operations supported in `cloud.google.com/go/filestore/apiv1` CloudFilestoreManagerClient.

#### V1 Test Analysis (Files reviewed)
**Business Logic Tests (KEEP for v2)**:
- `validations_test.go`: Tests capacity ranges for different tiers (BASIC_HDD, BASIC_SSD, ZONAL, REGIONAL) and scale-down rules
- `validateAlways_test.go`: Tests pre-flight validation logic
- `validatePostCreate_test.go`: Tests post-creation validation (no scale-down)
- `checkNUpdateState_test.go`: Tests state machine transitions (Creating → Ready → Updating → Error → Deleted)
- `checkUpdateMask_test.go`: Tests update mask calculation logic
- `syncNfsInstance_test.go`: Tests GCP API calls and operation handling
- `checkGcpOperation_test.go`: Tests operation polling logic
- `loadNfsInstance_test.go`: Tests loading instances from GCP

**Trivial Tests (SKIP for v2)**:
- Tests in `state_test.go` that only test simple state creation/getters
- Mock setup code without actual business logic validation
- Simple type conversion tests without decision logic

**Test Infrastructure Pattern**: Uses testinfra mocks with httptest servers - good pattern to maintain.

#### Configuration Dependencies
1. **GCP Config** (`pkg/kcp/provider/gcp/config`):
   - `GcpRetryWaitTime` - Retry delay for operations
   - `GcpOperationWaitTime` - Wait time between operation polls
2. **Feature Flags** (`pkg/feature`):
   - `Nfs41Gcp` - Enable NFSv4.1 protocol for ZONAL/REGIONAL tiers
3. **IpRange Dependency**:
   - NfsInstance requires IpRange for network configuration
   - Loaded via `loadIpRange` action in shared reconciler
4. **Scope Dependency**:
   - GCP project ID and credentials from Scope CRD
   - Location (region/zone) from NfsInstance spec

#### Integration Points
1. **Shared Reconciler** (`pkg/kcp/nfsinstance/reconciler.go`):
   - Multi-provider switch using `statewithscope` predicates
   - GCP action invoked via `gcpnfsinstance.New(r.gcpStateFactory)`
   - Shared actions: `loadIpRange`, `copyStatusHostsToHost`
2. **Controller Setup** (`internal/controller/cloud-control/nfsinstance_controller.go`):
   - `gcpnfsinstance.NewStateFactory(filestoreClientProvider, env)`
   - Uses `gcpclient.ClientProvider[FilestoreClient]` pattern
3. **State Hierarchy** (OLD Pattern):
   - `focal.State` → `types.State` (shared) → `nfsinstance.State` (GCP-specific)
   - GCP state adds: `filestoreClient`, `fsInstance`, `operation`, `updateMask`, `curState`

#### Modern Client Decision Rationale
**V2 WILL use `cloud.google.com/go/filestore/apiv1`**:
- ✅ Package exists and is well-maintained by Google
- ✅ Consistent with other GCP resources in codebase (compute, redis use `cloud.google.com/go/*`)
- ✅ Better type safety (generated protobuf types vs JSON REST wrappers)
- ✅ Native support for gRPC (better performance, streaming if needed)
- ✅ All required operations supported: Get, Create, Update, Delete, Operations
- ✅ Better error handling with grpc status codes
- ⚠️ Not currently in go.mod - will need to add dependency
- ⚠️ Requires client initialization pattern change (similar to compute/redis)

**Migration path**: v2 will use modern client, v1 remains unchanged for stability.

---

### GCP Client Decision
- **V1 Uses**: `google.golang.org/api/file/v1` (REST API wrapper)
- **V2 Uses**: `cloud.google.com/go/filestore/apiv1` (modern client library) ✅ **IMPLEMENTED**
- **Modern Client**: **EXISTS** at https://pkg.go.dev/cloud.google.com/go/filestore
- **Decision**: Migrated to modern client in v2 (Phase 5 completed)
- **Rationale**: 
  - Modern client available from Google
  - Consistent with other GCP resources (compute, redis, etc.)
  - Better type safety with protobuf types
  - Follows best practices for new GCP client usage
  - V1 remains unchanged for stability
  - Feature flag allows gradual rollout
- **Integration**: Added to `pkg/kcp/provider/gcp/client/gcpClients.go` as `Filestore` field

### OLD Pattern Compliance
- **MUST maintain 3-layer state hierarchy**: `focal.State` → `types.State` → GCP `State`
- **MUST NOT change to NEW pattern**: NfsInstance is legacy multi-provider CRD
- **MUST preserve provider switching**: Integration with shared reconciler layer
- **MUST maintain types.State interface**: Shared with other providers

### Streamlining Opportunities
1. **Better action organization**: Group by responsibility
2. **Simplified state transitions**: Clearer state machine
3. **Consolidated validations**: Reduce duplication
4. **Improved testing**: Focus on business logic, eliminate trivial tests
5. **Enhanced error handling**: Consistent patterns
6. **Leaner test suite**: Avoid v1's pattern of testing every trivial method

### Feature Flag Strategy
- **Default**: `gcpNfsInstanceV2 = false` (use v1, stable)
- **Feature Flag Name**: `gcpNfsInstanceV2` (finalized)
- **Implementation**: `pkg/feature/ffGcpNfsInstanceV2.go`
- **Configuration**: `pkg/feature/ff_ga.yaml` and `pkg/feature/ff_edge.yaml`
- **Tests**: `pkg/feature/ffGcpNfsInstanceV2_test.go`
- **Documentation**: `pkg/feature/README_GCP_NFSINSTANCE_V2.md`
- **Enable v2**: Set flag to `true` in feature toggles configuration
- **Rollout Plan**:
  1. Deploy with v2 implementation, flag = false (v1 active)
  2. Test v2 in dev/staging environments with flag = true
  3. Validate v2 behavior matches v1
  4. Gradually enable v2 in production
  5. Monitor metrics and error rates
  6. Full v2 rollout after validation period
  7. Deprecate v1 after v2 proven stable

### V1 Test Issues to Avoid in V2
The v1 implementation contains many low-value tests that should NOT be replicated:
- **Trivial getter/setter tests**: Testing simple property access adds no value
- **Mock validation tests**: Tests that only verify mock behavior, not business logic
- **Redundant coverage**: Multiple tests covering the same code path
- **Framework/library tests**: Testing external library behavior instead of our code
- **Simple type conversion tests**: Testing straightforward data transformations

**V2 Testing Philosophy**: "Test behavior, not implementation. Test decisions, not data flow."

### Future Considerations
- [ ] V1 removal timeline (after v2 fully validated)
- [ ] Evaluate performance metrics comparing v1 and v2
- [ ] Plan similar refactoring for other OLD pattern providers
- [ ] Consider feature flag removal after v1 sunset
- [ ] Apply lean testing approach to other provider refactorings

---

## References

- **AGENTS.md**: Main agent instructions
- **docs/agents/architecture/RECONCILER_OLD_PATTERN.md**: OLD pattern reference
- **docs/agents/architecture/STATE_PATTERN.md**: State hierarchy explanation
- **docs/agents/architecture/ACTION_COMPOSITION.md**: Action composition guide
- **pkg/kcp/provider/gcp/subnet/**: NEW pattern example (for comparison)
- **pkg/kcp/redisinstance/**: OLD pattern reference implementation

---

## Approval & Review

- [ ] Plan reviewed by team
- [ ] Technical approach approved
- [ ] Timeline acceptable
- [ ] Risk mitigation agreed
- [ ] Ready to proceed with implementation

**Plan Version**: 1.0  
**Last Updated**: 2025-12-25  
**Status**: DRAFT - READY FOR REVIEW
