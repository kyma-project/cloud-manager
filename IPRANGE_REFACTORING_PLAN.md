# KCP IpRange Refactoring Plan

## Overview
Refactor KCP IpRange implementation from OLD client pattern to NEW client pattern, following the GcpSubnet structure for cleaner, more maintainable code.

## Current State Analysis

### Problems with Current Implementation
1. **OLD Client Pattern**: Uses `ClientProvider[T]` with cached HTTP client (legacy REST APIs)
2. **Messy Structure**: Mixed v2 wrapper layer, unclear action organization
3. **Inconsistent with Modern Code**: Doesn't follow GcpSubnet pattern (NEW pattern)
4. **Complex State Management**: Unnecessary `v2/` subdirectory and wrapper functions
5. **Legacy REST APIs**: Uses `google.golang.org/api/servicenetworking/v1` instead of modern Cloud Client Libraries

### Current Directory Structure
```
pkg/kcp/iprange/
├── allocate/
├── allocateIpRange.go
├── reconciler.go (multi-provider with switch)
├── state.go (shared state)
└── types/state.go

pkg/kcp/provider/gcp/iprange/
├── client/
│   ├── serviceNetworkingClient.go (OLD pattern)
│   └── computeClient.go (OLD pattern)
├── new.go (wrapper to v2)
├── state.go (wrapper factory)
└── v2/ (actual implementation)
    ├── new.go
    ├── state.go
    ├── loadAddress.go
    ├── loadPsaConnection.go
    ├── syncAddress.go
    ├── syncPsaConnection.go
    └── ... many other actions
```

### Target Structure (Following RedisInstance Multi-Provider Pattern)
```
pkg/kcp/iprange/
├── reconciler.go (multi-provider with switch)
├── state.go (shared state implementation)
├── types/
│   └── state.go (shared state interface - extends focal.State)
├── kcpNetworkCreate.go (shared actions)
├── kcpNetworkDelete.go
├── kcpNetworkLoad.go
├── kymaNetworkLoad.go
├── kymaPeeringCreate.go
├── kymaPeeringDelete.go
├── allocateIpRange.go (shared allocation logic)
└── ... other shared actions

pkg/kcp/provider/gcp/iprange/
├── new.go (provider action composition - replaces v2 wrapper)
├── state.go (extends iprange/types.State)
├── client/
│   ├── computeClient.go (NEW pattern)
│   └── serviceNetworkingClient.go (NEW pattern)
├── loadAddress.go
├── loadPsaConnection.go
├── createAddress.go
├── deleteAddress.go
├── createPsaConnection.go
├── deletePsaConnection.go
├── updatePsaConnection.go
├── waitOperationDone.go
├── updateStatus.go
├── updateStatusId.go
├── preventCidrEdit.go
├── copyCidrToStatus.go
├── validateCidr.go
└── allocateIpRange.go
```

---

## Refactoring Plan

### Phase 0: Client Capability Investigation (CRITICAL - DO FIRST!)
**Status**: ⬜ TODO

#### Task 0.1: Investigate Service Networking API Support
- [ ] Check if modern Cloud Client Library exists for Service Networking
  - Look for `cloud.google.com/go/servicenetworking`
  - Check Google Cloud Go SDK documentation
  - Review release notes for recent additions
- [ ] Compare OLD vs NEW client capabilities:
  - OLD: `google.golang.org/api/servicenetworking/v1`
  - NEW: Potential `cloud.google.com/go` equivalent
- [ ] **Critical Methods Needed**:
  - `ListServiceConnections(ctx, projectId, vpcId) ([]*Connection, error)`
  - `CreateServiceConnection(ctx, projectId, vpcId, reservedIpRanges) (*Operation, error)`
  - `PatchServiceConnection(ctx, projectId, vpcId, reservedIpRanges) (*Operation, error)` (force update)
  - `DeleteServiceConnection(ctx, projectId, vpcId) (*Operation, error)`
  - `GetOperation(ctx, operationName) (*Operation, error)`
- [ ] **Document findings**:
  - If NEW client exists with all methods → Proceed with refactoring
  - If NEW client missing methods → Document gap and consider hybrid approach
  - If no NEW client exists → Keep OLD pattern for Service Networking only

#### Task 0.2: Verify Compute API Coverage in GcpClients
- [ ] Check if `GcpClients.ComputeAddresses` supports all needed operations:
  - `Get(projectId, region, addressName)` - Load global address
  - `Insert(projectId, region, address)` - Create global address  
  - `Delete(projectId, region, addressName)` - Delete global address
- [ ] Check if `GcpClients` has global operations client:
  - Current: `RegionOperations` exists
  - Need: `GlobalOperations` for global address operations
  - Add if missing
- [ ] Verify operation polling capabilities

#### Task 0.3: Create Decision Matrix
- [ ] Document which clients can use NEW pattern:
  ```
  | Client              | NEW Pattern Available? | Action                    |
  |---------------------|------------------------|---------------------------|
  | Compute Addresses   | ✅ Yes (existing)      | Use GcpClients            |
  | Service Networking  | ❓ TBD                 | Investigate               |
  | Global Operations   | ❓ Check if exists     | Add if missing            |
  ```
- [ ] **Decision Point**: If Service Networking has no NEW client:
  - **Option A**: Keep OLD pattern for Service Networking, use NEW for Compute (hybrid)
  - **Option B**: Wait for Google to release NEW client
  - **Option C**: Wrapper approach - use OLD client inside NEW-style interface

#### Task 0.4: Update Plan Based on Findings
- [ ] If gaps found, update Phase 1 and Phase 2 tasks accordingly
- [ ] Add notes about which parts use NEW vs OLD pattern
- [ ] Adjust timeline estimates if hybrid approach needed

**Expected Outcome**: Clear understanding of what's possible with current GCP client libraries. **DO NOT PROCEED** with Phase 1 until this investigation is complete.

**Notes**:
- This investigation was last done ~1 year ago
- Google regularly adds new services to Cloud Client Libraries
- Check `cloud.google.com/go` repository releases for recent updates
- Service Networking might now be available or under different package name

---

### Phase 1: Add New GCP Client Libraries to GcpClients
**Status**: ⬜ TODO

#### Task 1.1: Add ServiceNetworking Client to GcpClients (Based on Phase 0 Findings)
- [ ] **IF NEW client exists**: Add to `GcpClients` struct
  ```go
  ServiceNetworking *servicenetworking.Client  // NEW pattern
  ```
- [ ] **IF NO NEW client**: Keep OLD pattern with note
  ```go
  // ServiceNetworking uses OLD pattern (no Cloud Client Library available yet)
  // Accessed via ClientProvider in iprange-specific code
  ```
- [ ] Initialize based on Phase 0 decision
- [ ] Document why OLD pattern is kept if necessary

#### Task 1.2: Ensure Compute Client is Available in GcpClients
- [ ] Verify `ComputeAddresses` is already in `GcpClients` ✅ (confirmed in current code)
- [ ] Check for `GlobalOperations` client for global address operations
- [ ] Add if missing: `ComputeGlobalOperations *compute.GlobalOperationsClient`
- [ ] Note: IpRange uses **global** addresses, not regional

#### Task 1.3: Update cmd/main.go GcpClients Initialization
- [ ] Verify `NewGcpClients()` call includes all needed credentials
- [ ] Add any missing initialization for IpRange-specific clients

**Expected Outcome**: GcpClients struct contains all necessary clients for IpRange operations using NEW pattern.

---

### Phase 2: Create NEW Pattern Client Interfaces
**Status**: ⬜ TODO

#### Task 2.1: Refactor pkg/kcp/provider/gcp/iprange/client/computeClient.go
- [ ] Create typed `ComputeClient` interface for IpRange operations
  ```go
  type ComputeClient interface {
      GetAddress(ctx context.Context, projectId, addressName string) (*computepb.Address, error)
      CreateAddress(ctx context.Context, projectId string, address *computepb.Address) (*computepb.Operation, error)
      DeleteAddress(ctx context.Context, projectId, addressName string) (*computepb.Operation, error)
      GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error)
  }
  ```
- [ ] Create `NewComputeClientProvider(gcpClients *gcpclient.GcpClients)` function
- [ ] Implement wrapper client using `gcpClients.ComputeAddresses`
- [ ] Remove OLD pattern `ClientProvider[T]` usage

#### Task 2.2: Refactor pkg/kcp/provider/gcp/iprange/client/serviceNetworkingClient.go (Conditional)
- [ ] **IF NEW client available (from Phase 0)**:
  - Create `NewServiceNetworkingClientProvider(gcpClients *gcpclient.GcpClients)`
  - Implement wrapper using `gcpClients.ServiceNetworking`
  - Remove OLD pattern implementation
- [ ] **IF NO NEW client available**:
  - Keep interface but document OLD pattern usage
  - Consider creating wrapper that accepts OLD-style client
  - Update provider to use `ClientProvider[ServiceNetworkingClient]` with cached HTTP client
  - Add TODO comment linking to tracking issue for NEW client migration
- [ ] **Interface remains the same** (business operations):
  ```go
  type ServiceNetworkingClient interface {
      ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error)
      CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
      PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error)
      DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error)
      GetOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error)
  }
  ```

**Expected Outcome**: Clean client interfaces following NEW pattern, wrapping GcpClients.

---

### Phase 3: Restructure State (Multi-Provider Pattern)
**Status**: ⬜ TODO

#### Task 3.1: Keep/Update Shared State in pkg/kcp/iprange/types/state.go
- [ ] Keep `types.State` interface extending `focal.State`
  ```go
  type State interface {
      focal.State
      ObjAsIpRange() *cloudcontrolv1beta1.IpRange
      
      // Shared methods needed by all providers
      Network() *cloudcontrolv1beta1.Network
      ExistingCidrRanges() []string
      SetExistingCidrRanges([]string)
  }
  ```
- [ ] This is the shared interface ALL providers extend

#### Task 3.2: Keep Shared State Implementation in pkg/kcp/iprange/state.go
- [ ] Keep basic implementation that embeds `focal.State`
  ```go
  type state struct {
      focal.State
      
      existingCidrRanges    []string
      networkKey            client.ObjectKey
      isCloudManagerNetwork bool
      isKymaNetwork         bool
      network               *cloudcontrolv1beta1.Network
      kymaNetwork           *cloudcontrolv1beta1.Network
      kymaPeering           *cloudcontrolv1beta1.VpcPeering
  }
  
  func newState(focalState focal.State) types.State {
      return &state{State: focalState}
  }
  ```

#### Task 3.3: Create GCP-Specific State in pkg/kcp/provider/gcp/iprange/state.go
- [ ] Remove `v2/` wrapper completely
- [ ] Create `State` struct that extends `types.State` (shared IpRange state)
  ```go
  type State struct {
      types.State  // Extends shared iprange state (which extends focal.State)
      
      computeClient            client.ComputeClient
      serviceNetworkingClient  client.ServiceNetworkingClient
      env                      abstractions.Environment
      
      // GCP-specific remote resources
      address        *computepb.Address
      psaConnection  *servicenetworking.Connection
      operation      interface{}  // can be compute or servicenetworking operation
      
      // GCP-specific state
      peeringIpRanges []string
  }
  ```
- [ ] Create `StateFactory` interface
  ```go
  type StateFactory interface {
      NewState(ctx context.Context, ipRangeState types.State) (*State, error)
  }
  ```
- [ ] Implement `stateFactory` struct with NEW pattern client providers
- [ ] Implement `NewStateFactory()` constructor
- [ ] Remove old `generalStateFactory` wrapper

**Expected Outcome**: Three-layer state hierarchy (composed → focal → shared iprange → GCP-specific) following RedisInstance pattern.

---

### Phase 4: Flatten and Refactor Actions (Remove v2/ Directory)
**Status**: ⬜ TODO

#### Task 4.1: Move and Refactor Core Actions
Move all actions from `v2/` to main `iprange/` directory and refactor:

##### loadAddress.go
- [ ] Move `v2/loadAddress.go` → `loadAddress.go`
- [ ] Update to use NEW state
- [ ] Simplify logic, remove unnecessary abstractions
- [ ] Update tests: `loadAddress_test.go`

##### createAddress.go (new, extracted from syncAddress)
- [ ] Create `createAddress.go`
- [ ] Extract address creation logic from `v2/syncAddress.go`
- [ ] Call `state.computeClient.CreateAddress()`
- [ ] Set operation in state for tracking

##### deleteAddress.go (new, extracted from syncAddress)
- [ ] Create `deleteAddress.go`
- [ ] Extract address deletion logic from `v2/syncAddress.go`
- [ ] Call `state.computeClient.DeleteAddress()`
- [ ] Set operation in state for tracking

##### loadPsaConnection.go
- [ ] Move `v2/loadPsaConnection.go` → `loadPsaConnection.go`
- [ ] Update to use NEW state
- [ ] Simplify logic
- [ ] Update tests: `loadPsaConnection_test.go`

##### createPsaConnection.go (new, extracted from syncPsaConnection)
- [ ] Create `createPsaConnection.go`
- [ ] Extract PSA connection creation from `v2/syncPsaConnection.go`
- [ ] Call `state.serviceNetworkingClient.CreateServiceConnection()`

##### updatePsaConnection.go (new, extracted from syncPsaConnection)
- [ ] Create `updatePsaConnection.go`
- [ ] Extract PSA connection update from `v2/syncPsaConnection.go`
- [ ] Call `state.serviceNetworkingClient.UpdateServiceConnection()`

##### deletePsaConnection.go (new, extracted from syncPsaConnection)
- [ ] Create `deletePsaConnection.go`
- [ ] Extract PSA connection deletion from `v2/syncPsaConnection.go`
- [ ] Call `state.serviceNetworkingClient.DeleteServiceConnection()`

#### Task 4.2: Refactor Supporting Actions
- [ ] Move `v2/preventCidrEdit.go` → `preventCidrEdit.go` and update
- [ ] Move `v2/copyCidrToStatus.go` → `copyCidrToStatus.go` and update
- [ ] Move `v2/validateCidr.go` → `validateCidr.go` and update
- [ ] Move `v2/updateStatusId.go` → `updateStatusId.go` and update
- [ ] Move `v2/identifyPeeringIpRanges.go` → `identifyPeeringIpRanges.go` and update
- [ ] Create `waitOperationDone.go` (similar to GcpSubnet's `waitCreationOperationDone.go`)
- [ ] Create `updateStatus.go` for status updates

#### Task 4.3: Refactor State Management Actions
- [ ] Analyze `v2/compareStates.go` - determine if needed or can be simplified
- [ ] Analyze `v2/updateState.go` - simplify or remove if logic can be inline
- [ ] Remove `v2/checkGcpOperation.go` - replace with cleaner `waitOperationDone`

#### Task 4.4: Move Allocation Logic
- [ ] Move `v2/allocateIpRange.go` → `allocateIpRange.go`
- [ ] Keep compatible with shared `pkg/kcp/iprange/allocateIpRange.go` if needed
- [ ] Update to use NEW state

**Expected Outcome**: All actions are flat in `pkg/kcp/provider/gcp/iprange/`, following one-action-per-file pattern like GcpSubnet.

---

### Phase 5: Update Provider Action Composition (Multi-Provider Pattern)
**Status**: ⬜ TODO

#### Task 5.1: Refactor new.go (keep, but simplify)
- [ ] Keep `pkg/kcp/provider/gcp/iprange/new.go` (this is the provider entry point)
- [ ] Implement action composition following RedisInstance pattern
  ```go
  func New(stateFactory StateFactory) composed.Action {
      return func(ctx context.Context, st composed.State) (error, context.Context) {
          // Convert shared iprange state to GCP-specific state
          state, err := stateFactory.NewState(ctx, st.(types.State))
          if err != nil {
              ipRange := st.Obj().(*v1beta1.IpRange)
              return composed.PatchStatus(ipRange).
                  SetExclusiveConditions(metav1.Condition{
                      Type:    v1beta1.ConditionTypeError,
                      Status:  metav1.ConditionTrue,
                      Reason:  v1beta1.ReasonGcpError,
                      Message: err.Error(),
                  }).
                  SuccessError(composed.StopAndForget).
                  Run(ctx, st)
          }
          
          return composed.ComposeActions(
              "gcpIpRange",
              actions.AddCommonFinalizer(),
              preventCidrEdit,
              copyCidrToStatus,
              validateCidr,
              loadAddress,
              loadPsaConnection,
              composed.IfElse(
                  composed.Not(composed.MarkedForDeletionPredicate),
                  composed.ComposeActions(
                      "create-update",
                      createAddress,
                      waitOperationDone,
                      updateStatusId,
                      identifyPeeringIpRanges,
                      createOrUpdatePsaConnection,
                      updateStatus,
                  ),
                  composed.ComposeActions(
                      "delete",
                      deletePsaConnection,
                      deleteAddress,
                      waitOperationDone,
                      actions.RemoveCommonFinalizer(),
                      composed.StopAndForgetAction,
                  ),
              ),
              composed.StopAndForgetAction,
          )(ctx, state)  // Pass GCP-specific state
      }
  }
  ```
- [ ] Remove complex state machine logic from `v2/new.go`
- [ ] Simplify flow to be more declarative

#### Task 5.2: Update NewAllocateIpRangeAction
- [ ] Keep `NewAllocateIpRangeAction` for allocation phase (called from shared reconciler)
  ```go
  func NewAllocateIpRangeAction(stateFactory StateFactory) composed.Action {
      return func(ctx context.Context, st composed.State) (error, context.Context) {
          state, err := stateFactory.NewState(ctx, st.(types.State))
          if err != nil {
              // Handle error
          }
          return allocateIpRange(ctx, state)
      }
  }
  ```
- [ ] Simplify wrapper, remove v2 indirection

#### Task 5.3: Delete Old v2/ Directory
- [ ] Delete entire `pkg/kcp/provider/gcp/iprange/v2/` directory
- [ ] Update imports in `new.go`
- [ ] Keep the multi-provider reconciler in `pkg/kcp/iprange/reconciler.go` (with provider switch)

**Expected Outcome**: Clean provider action composition following RedisInstance multi-provider pattern, no v2 wrapper.

---

### Phase 6: Update main.go Wiring
**Status**: ⬜ TODO

#### Task 6.1: Update GCP IpRange State Factory Initialization
- [ ] Find GCP IpRange state factory creation in `cmd/main.go`
- [ ] Replace OLD pattern client providers with NEW pattern:
  ```go
  gcpIpRangeStateFactory := gcpiprange.NewStateFactory(
      gcpiprangeclient.NewComputeClientProvider(gcpClients),
      gcpiprangeclient.NewServiceNetworkingClientProvider(gcpClients),
      env,
  )
  ```
- [ ] Remove old `gcpiprangeclient.NewServiceNetworkingClient()` call
- [ ] Remove old `gcpiprangeclient.NewComputeClient()` call if exists

#### Task 6.2: Verify Reconciler Registration
- [ ] Ensure `iprange.NewIPRangeReconciler()` call in main.go still works
- [ ] Verify all state factories are correctly passed

**Expected Outcome**: main.go uses NEW pattern client providers from GcpClients.

---

### Phase 7: Update Tests
**Status**: ⬜ TODO

#### Task 7.1: Update Unit Tests
- [ ] Update tests in `pkg/kcp/provider/gcp/iprange/v2/*_test.go`
- [ ] Move to main `pkg/kcp/provider/gcp/iprange/` directory
- [ ] Update mocks to use NEW pattern clients
- [ ] Ensure all tests pass with new structure

#### Task 7.2: Update Controller Tests
- [ ] Find IpRange controller tests in `internal/controller/cloud-control/`
- [ ] Update test setup to use NEW pattern
- [ ] Verify all test scenarios pass

#### Task 7.3: Add Missing Tests
- [ ] Add tests for new actions if coverage is missing
- [ ] Test error handling paths
- [ ] Test deletion flows

**Expected Outcome**: All tests passing with NEW pattern.

---

### Phase 8: Cleanup and Documentation
**Status**: ⬜ TODO

#### Task 8.1: Code Cleanup
- [ ] Remove any unused imports
- [ ] Remove commented-out code
- [ ] Ensure consistent error handling
- [ ] Add proper logging where needed

#### Task 8.2: Verify Feature Flags
- [ ] Ensure feature flag checks are in place
- [ ] Test with different feature flag configurations

#### Task 8.3: Update AGENTS.md
- [ ] Update IpRange example in AGENTS.md to reflect NEW pattern
- [ ] Document any IpRange-specific patterns

#### Task 8.4: Final Verification
- [ ] Run `make test` - all tests pass
- [ ] Run `make lint` - no linting errors
- [ ] Run `make manifests` and `make generate` - no changes
- [ ] Manual smoke test if possible

**Expected Outcome**: Clean, well-documented code following NEW pattern.

---

## Success Criteria

### Must Have
- ✅ All GCP IpRange clients use NEW pattern (GcpClients)
- ✅ State structure follows RedisInstance multi-provider pattern (three-layer hierarchy)
- ✅ Shared state in `pkg/kcp/iprange/types/` for all providers
- ✅ GCP-specific state extends shared state
- ✅ No `v2/` subdirectory
- ✅ One action per file
- ✅ All tests passing
- ✅ No linting errors

### Should Have
- ✅ Simplified reconciliation logic (less complex than current state machine)
- ✅ Clear separation of concerns (create/update/delete actions)
- ✅ Consistent with other multi-provider resources (RedisInstance, NfsInstance)

### Nice to Have
- ✅ Improved error messages
- ✅ Better logging
- ✅ Additional test coverage

---

## Risk Assessment

### High Risk
- **Breaking PSA Connections**: Private Service Access is critical for Redis/NFS. Thoroughly test deletion/recreation flows.
- **CIDR Allocation Logic**: Complex logic for finding free CIDR ranges. Must preserve correctness.

### Medium Risk
- **Operation Tracking**: GCP operations are async. Ensure proper wait/polling logic.
- **Multi-provider Compatibility**: Don't break AWS/Azure/SAP IpRange implementations.

### Low Risk
- **Client Migration**: NEW pattern is well-established and working in other resources.
- **State Structure**: Direct focal.State extension is simpler and clearer.

---

## Testing Strategy

### Unit Tests
- Test each action independently with mocked clients
- Test CIDR validation and allocation logic
- Test state transitions

### Integration Tests (Controller Tests)
- Test full IpRange lifecycle (create → ready → delete)
- Test PSA connection creation/update/deletion
- Test address allocation and deallocation
- Test operation polling and error handling
- Test with different Scope configurations

### Manual Testing (if possible)
- Deploy to dev environment
- Create IpRange
- Create dependent resources (Redis, NFS)
- Delete dependent resources
- Delete IpRange
- Verify no resource leaks

---

## Rollback Plan

If refactoring causes issues:
1. Revert all changes to `pkg/kcp/provider/gcp/iprange/`
2. Revert changes to `cmd/main.go`
3. Keep OLD pattern clients temporarily
4. Re-test thoroughly before second attempt

---

## Timeline Estimate

- **Phase 0**: 2-4 hours (Investigation - CRITICAL, do first!)
- **Phase 1**: 2-3 hours (GcpClients integration) - depends on Phase 0 findings
- **Phase 2**: 2-4 hours (Client interfaces) - may need hybrid approach
- **Phase 3**: 2-3 hours (State refactoring)
- **Phase 4**: 6-8 hours (Action refactoring - most complex)
- **Phase 5**: 3-4 hours (Reconciler)
- **Phase 6**: 1-2 hours (main.go wiring) - depends on Phase 0 findings
- **Phase 7**: 4-6 hours (Tests)
- **Phase 8**: 2-3 hours (Cleanup/docs)

**Total Estimate**: 24-37 hours (increased due to potential hybrid approach)

---

## Progress Tracking

### Legend
- ⬜ TODO - Not started
- 🔄 IN PROGRESS - Currently working on
- ✅ DONE - Completed
- ⚠️ BLOCKED - Blocked by dependency or issue
- ❌ SKIPPED - Decided not to implement

### Overall Progress
- **Phase 0: ⬜ TODO** ⚠️ **MUST COMPLETE FIRST - BLOCKS ALL OTHER PHASES**
- Phase 1: ⬜ TODO (blocked by Phase 0)
- Phase 2: ⬜ TODO (blocked by Phase 0)
- Phase 3: ⬜ TODO
- Phase 4: ⬜ TODO
- Phase 5: ⬜ TODO
- Phase 6: ⬜ TODO (may need adjustment based on Phase 0)
- Phase 7: ⬜ TODO
- Phase 8: ⬜ TODO

---

## Notes

### Key Differences from Current Implementation
1. **No v2 wrapper**: Direct implementation in main package
2. **NEW clients**: Use GcpClients singleton instead of ClientProvider
3. **Simpler state machine**: Replace complex StatePredicate switching with clear IfElse composition
4. **One action per file**: Better organization and maintainability
5. **Three-layer state**: composed → focal → shared iprange → GCP-specific (multi-provider pattern)
6. **Shared reconciler**: Keeps provider switching in `pkg/kcp/iprange/reconciler.go`

### Alignment with AGENTS.md Guidance
- ✅ Follows OLD/Legacy Pattern (RedisInstance) - multi-provider CRD
- ✅ Shared state layer for all providers
- ✅ GCP-specific state extends shared state
- ✅ NEW client pattern (GcpClients)
- ✅ One action per file
- ✅ Clean state hierarchy with proper extension
- ✅ Composed action pattern
- ✅ Provider switching in shared reconciler

---

## Questions to Resolve

1. **Service Networking API**: ⚠️ **CRITICAL - Resolve in Phase 0**
   - **Last checked**: ~1 year ago (no NEW client available)
   - **Action**: Re-investigate if `cloud.google.com/go/servicenetworking` now exists
   - **Fallback**: If still no NEW client, document hybrid approach (NEW for Compute, OLD for Service Networking)
   - **Impact**: May require keeping OLD pattern for Service Networking while using NEW for other clients

2. **Allocation Logic**: Should CIDR allocation stay in shared `pkg/kcp/iprange/` or move to GCP-specific?
   - Current: Shared across all providers
   - Recommendation: Keep shared for now to avoid breaking other providers

3. **State Machine Complexity**: Can we simplify the current state machine in `v2/new.go`?
   - Current: Uses `StatePredicate` switching based on `curState`
   - Target: Use clear IfElse conditions based on resource existence

4. **Shared vs Provider-Specific**: Which actions belong in shared `pkg/kcp/iprange/` vs GCP-specific?
   - Shared: Network/Peering operations (used by all providers)
   - GCP-specific: Address and PSA connection operations

---

**Document Version**: 1.0  
**Created**: 2024-12-09  
**Last Updated**: 2024-12-09
