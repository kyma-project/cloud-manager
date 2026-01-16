# Cloud Manager - AI Agent Instructions

**Target Audience**: LLM coding agents  
**Last Updated**: 2025-01-24

## Repository Identity

Cloud Manager is a Kubernetes controller manager (built with Kubebuilder) that provisions cloud provider resources (AWS, Azure, GCP) for SAP BTP Kyma runtime.

## Core Architecture

### MUST UNDERSTAND: Dual Reconciliation Loops

**SKR Loop** (user-facing):
- RECONCILES: `cloud-resources.kyma-project.io/v1beta1` CRDs
- LOCATION: Remote Kyma clusters (user environments)
- CREATES: KCP resources in control plane
- EXAMPLES: `GcpNfsVolume`, `AzureRedisInstance`

**KCP Loop** (control plane):
- RECONCILES: `cloud-control.kyma-project.io/v1beta1` CRDs
- LOCATION: Control plane cluster
- PROVISIONS: Actual cloud provider resources via APIs
- EXAMPLES: `GcpSubnet`, `AzureVNetLink`

**Execution Flow** (ALWAYS in this order):
1. User creates SKR resource in their cluster
2. SKR reconciler creates KCP resource in control plane
3. KCP reconciler provisions cloud provider infrastructure
4. Status propagates back: cloud → KCP → SKR

## Agent Rules: MANDATORY CONSTRAINTS

### Rule 1: Pattern Selection (CRITICAL)

**NEW Pattern** (REQUIRED for all new code):
- MUST USE FOR: All new KCP/SKR reconcilers created after 2024
- CRD NAME PATTERN: Provider-specific (`GcpSubnet`, `AzureVNetLink`)
- STATE HIERARCHY: `composed.State` → `focal.State` → `ProviderState` (2 layers)
- REFERENCE IMPLEMENTATION: `pkg/kcp/provider/gcp/subnet/`
- NEVER ADD TO: Existing multi-provider CRDs

**OLD Pattern** (FORBIDDEN for new code):
- ONLY USE FOR: Maintaining `RedisInstance`, `NfsInstance`, `IpRange`
- CHARACTERISTIC: Multi-provider CRDs with provider sections
- STATE HIERARCHY: 3 layers (includes shared intermediate state)
- DO NOT REPLICATE: This pattern is deprecated

### Rule 2: Code Modification Boundaries

**ALLOWED**:
- Add new provider-specific CRDs (e.g., `GcpRedisCluster`)
- Add new actions in existing reconciler packages
- Fix bugs in existing reconcilers
- Add tests

**FORBIDDEN WITHOUT EXPLICIT USER APPROVAL**:
- Modify existing CRD API fields (breaking change)
- Change reconciler patterns (NEW ↔ OLD)
- Refactor cross-package state hierarchies
- Add new dependencies
- Modify feature flag behavior

### Rule 3: Action Composition Requirements

**MUST**:
- End all successful flows with `composed.StopAndForgetAction` OR `return composed.StopAndForget, nil`
- Return `nil, ctx` to continue to next action
- Return `error, ctx` for failures (triggers requeue)
- Order actions: load dependencies → load resource → create/update → wait → update status → stop

**MUST NOT**:
- Execute actions in parallel (always sequential)
- Create resources before loading dependencies
- Skip status updates
- Forget to add finalizers for resources requiring deletion cleanup
- Remove finalizers before completing deletion

### Rule 4: State Management

**ALWAYS**:
- Check state factory errors: `state, err := factory.NewState(ctx, focalState); if err != nil { handle }`
- Use type-specific getters: `state.ObjAsGcpSubnet()` not raw `state.Obj().(*Type)`
- Pass provider state to provider-specific actions
- Keep state immutable within single action

**NEVER**:
- Mutate state across action boundaries without explicit state update
- Assert to wrong state level (provider action receiving focal.State)
- Ignore state initialization errors

### Rule 5: Testing Requirements

**MUST INCLUDE**:
- Controller tests in `internal/controller/` using `pkg/testinfra` framework
- Test both create and delete paths
- Test error conditions (API failures, not found, conflicts)
- Use `Eventually()` with `LoadAndCheck()` for assertions

**UNIT TESTS** (optional, only for business logic):
- Only write unit tests for complex business logic functions (e.g., validation, conversion, comparison)
- DO NOT write unit tests for:
  - Mock/fake implementations
  - Simple getters/setters
  - Client wrappers
  - Actions that only orchestrate calls to other actions
- Example: `util_test.go` testing `convertTier()` enum conversion is appropriate
- Location: Same directory as the code being tested (`*_test.go`)

**TEST FILE LOCATION**:
- Controller tests: `internal/controller/cloud-control/*_test.go` or `internal/controller/cloud-resources/*_test.go`
- Unit tests: Same package as code under test (e.g., `pkg/kcp/provider/gcp/nfsinstance/v2/*_test.go`)
- Mocks: `pkg/kcp/provider/<provider>/mock/` or `pkg/testinfra/`

### Rule 6: Feature Flag Handling

**CHECK BEFORE EXECUTION**:
```go
if feature.ApiDisabled.Value(ctx) {
    return composed.StopAndForget, nil
}
```

**NEVER**:
- Skip feature flag checks in SKR reconcilers
- Modify feature flag definitions without approval

### Rule 7: Documentation Updates (MANDATORY)

**WHEN MODIFYING CODE**, MUST UPDATE:
- THIS FILE (`AGENTS.md`) if changing global patterns
- `docs/agents/architecture/*.md` if changing reconciler structure
- `docs/agents/guides/*.md` if changing development procedures
- `docs/agents/reference/COMMON_PITFALLS.md` if discovering new issues
- `docs/agents/reference/QUICK_REFERENCE.md` if adding new helpers

**UPDATE FORMAT**: Keep deterministic, agent-first language

## Quick Start for Agents

### Task: Add New KCP Reconciler

1. READ: [docs/agents/guides/ADD_KCP_RECONCILER.md](docs/agents/guides/ADD_KCP_RECONCILER.md)
2. VERIFY: Using NEW pattern (provider-specific CRD)
3. EXECUTE: Follow step-by-step guide
4. TEST: Create controller tests with mocks
5. UPDATE: Documentation if patterns change

### Task: Add New SKR Reconciler

1. READ: [docs/agents/guides/ADD_SKR_RECONCILER.md](docs/agents/guides/ADD_SKR_RECONCILER.md)
2. VERIFY: Corresponding KCP reconciler exists
3. EXECUTE: Create SKR → KCP bridge
4. TEST: Include feature flag tests
5. UPDATE: Documentation if patterns change

### Task: Fix Bug

1. READ: [docs/agents/reference/COMMON_PITFALLS.md](docs/agents/reference/COMMON_PITFALLS.md)
2. IDENTIFY: Which pattern (NEW or OLD)
3. LOCATE: Reconciler using [docs/agents/reference/PROJECT_STRUCTURE.md](docs/agents/reference/PROJECT_STRUCTURE.md)
4. FIX: Within pattern boundaries
5. TEST: Add regression test
6. UPDATE: Add to COMMON_PITFALLS.md if novel issue

### Task: Understand Codebase

1. READ IN ORDER:
   - [docs/agents/architecture/STATE_PATTERN.md](docs/agents/architecture/STATE_PATTERN.md) - Foundation
   - [docs/agents/architecture/ACTION_COMPOSITION.md](docs/agents/architecture/ACTION_COMPOSITION.md) - Logic flow
   - [docs/agents/architecture/RECONCILER_NEW_PATTERN.md](docs/agents/architecture/RECONCILER_NEW_PATTERN.md) - Current pattern
   - [docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md](docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md) - NEW vs OLD
2. REFERENCE: [docs/agents/reference/QUICK_REFERENCE.md](docs/agents/reference/QUICK_REFERENCE.md) during coding

## Critical Files Index

### Must-Read Architecture

| File | Purpose | When to Read |
|------|---------|--------------|
| [docs/agents/architecture/STATE_PATTERN.md](docs/agents/architecture/STATE_PATTERN.md) | Three-state hierarchy explained | Before any reconciler work |
| [docs/agents/architecture/ACTION_COMPOSITION.md](docs/agents/architecture/ACTION_COMPOSITION.md) | Action flow control | Before writing actions |
| [docs/agents/architecture/RECONCILER_NEW_PATTERN.md](docs/agents/architecture/RECONCILER_NEW_PATTERN.md) | Required pattern for new code | Before creating reconcilers |
| [docs/agents/architecture/RECONCILER_OLD_PATTERN.md](docs/agents/architecture/RECONCILER_OLD_PATTERN.md) | Legacy pattern explanation | When maintaining old reconcilers |
| [docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md](docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md) | NEW vs OLD side-by-side | When uncertain which pattern to use |

### Step-by-Step Guides

| File | Task | Prerequisites |
|------|------|---------------|
| [docs/agents/guides/ADD_KCP_RECONCILER.md](docs/agents/guides/ADD_KCP_RECONCILER.md) | Create cloud provider reconciler | Understanding STATE_PATTERN, ACTION_COMPOSITION |
| [docs/agents/guides/ADD_SKR_RECONCILER.md](docs/agents/guides/ADD_SKR_RECONCILER.md) | Create user-facing reconciler | KCP reconciler exists |
| [docs/agents/guides/CONTROLLER_TESTS.md](docs/agents/guides/CONTROLLER_TESTS.md) | Write unit tests | Reconciler code written |
| [docs/agents/guides/CREATING_MOCKS.md](docs/agents/guides/CREATING_MOCKS.md) | Mock cloud provider APIs | Understanding provider client patterns |
| [docs/agents/guides/API_VALIDATION_TESTS.md](docs/agents/guides/API_VALIDATION_TESTS.md) | Test CRD validation | CRD defined |

### Quick Reference

| File | Purpose | Use When |
|------|---------|----------|
| [docs/agents/reference/QUICK_REFERENCE.md](docs/agents/reference/QUICK_REFERENCE.md) | Cheat sheet for common patterns | During coding |
| [docs/agents/reference/COMMON_PITFALLS.md](docs/agents/reference/COMMON_PITFALLS.md) | Known issues and solutions | Debugging or preventing bugs |
| [docs/agents/reference/PROJECT_STRUCTURE.md](docs/agents/reference/PROJECT_STRUCTURE.md) | Codebase navigation | Finding files |

### Core Implementation Files

| Path | Contains | Do Not Modify Without |
|------|----------|----------------------|
| `pkg/composed/action.go` | Action composition core | Understanding impact on all reconcilers |
| `pkg/composed/state.go` | Base state implementation | Verifying no state breakage |
| `pkg/common/actions/focal/` | Scope management | Testing all KCP reconcilers |
| `api/cloud-control/v1beta1/` | KCP API definitions | User approval (breaking changes) |
| `api/cloud-resources/v1beta1/` | SKR API definitions | User approval (breaking changes) |

## Decision Trees

### Which Pattern Should I Use?

```
Is this NEW code (created after 2024)?
├─ YES: MUST use NEW Pattern (GcpSubnet style)
│  ├─ CRD: Provider-specific name
│  ├─ State: Directly extends focal.State
│  ├─ Package: Single provider package
│  └─ Example: pkg/kcp/provider/gcp/subnet/
│
└─ NO: Are you maintaining existing code?
   └─ YES: Understand the pattern IT uses
      ├─ Multi-provider CRD? → OLD Pattern (RedisInstance)
      └─ Provider-specific CRD? → NEW Pattern
```

### Which GCP Client Pattern?

```
Are you adding NEW GCP resource?
├─ YES: MUST use NEW Pattern (GcpClients)
│  ├─ Client init: Via GcpClientFactory.GetGcpClients()
│  ├─ Location: pkg/kcp/provider/gcp/client/gcpClients.go
│  └─ Example: pkg/kcp/provider/gcp/subnet/
│
└─ NO: Maintaining existing resource?
   ├─ Uses ClientProvider? → OLD Pattern (legacy)
   └─ Uses GcpClients? → NEW Pattern
```

### Where Do I Add Code?

```
What are you adding?
├─ New cloud provider resource
│  └─ CREATE: pkg/kcp/provider/<provider>/<resource>/
│     ├─ reconcile.go
│     ├─ state.go
│     ├─ client/ (if needed)
│     ├─ load*.go
│     ├─ create*.go
│     ├─ delete*.go
│     └─ updateStatus.go
│
├─ New user-facing resource
│  └─ CREATE: pkg/skr/<resource>/
│     ├─ reconciler.go
│     ├─ createKcpResource.go
│     ├─ updateStatus.go
│     └─ *_test.go
│
├─ New action (reusable)
│  └─ ADD TO: pkg/common/actions/
│
├─ New test mock
│  └─ ADD TO: pkg/kcp/provider/<provider>/mock/
│     OR pkg/testinfra/
│
└─ Bug fix
   └─ MODIFY: Existing file in pattern-consistent way
```

## API Groups Reference

| API Group | Version | Location | Reconciler Location | Purpose |
|-----------|---------|----------|-------------------|---------|
| `cloud-resources.kyma-project.io` | v1beta1 | `api/cloud-resources/v1beta1/` | `pkg/skr/` | User-facing SKR resources |
| `cloud-control.kyma-project.io` | v1beta1 | `api/cloud-control/v1beta1/` | `pkg/kcp/` | Control plane KCP resources |

## Action Return Value Semantics

| Return | Effect | Requeue | Use Case |
|--------|--------|---------|----------|
| `nil, ctx` | Continue to next action | N/A | Normal flow progression |
| `error, ctx` | Stop execution, log error | Yes (exponential backoff) | Unexpected failure |
| `composed.StopAndForget, nil` | Stop execution, success | No | Reconciliation complete |
| `composed.StopWithRequeue, nil` | Stop execution, success | Yes (immediate) | Check status again soon |
| `composed.StopWithRequeueDelay(duration), nil` | Stop execution, success | Yes (after duration) | Wait for async operation |

## Agent Execution Rules

### Before Making Changes

1. MUST READ relevant architecture documentation
2. MUST IDENTIFY which pattern (NEW or OLD) applies
3. MUST CHECK if modification requires user approval
4. MUST VERIFY tests exist for code being modified

### While Making Changes

1. MUST FOLLOW pattern boundaries (no mixing NEW and OLD)
2. MUST MAINTAIN existing action composition structure
3. MUST ADD tests for new code
4. MUST UPDATE status conditions appropriately
5. MUST HANDLE errors explicitly (no silent failures)

### After Making Changes

1. MUST RUN tests: `make test`
2. MUST VERIFY CRD generation: `make manifests`
3. MUST UPDATE documentation
4. MUST COMMIT atomic changes (one logical change per commit)

### Forbidden Actions

**NEVER**:
- Modify CRD field types without user approval
- Remove existing CRD fields
- Change action return semantics
- Introduce new external dependencies without approval
- Skip error handling
- Ignore test failures
- Leave incomplete implementations
- Create resources without finalizers if deletion cleanup needed
- Modify feature flag default values
- Change state hierarchy
- Break backwards compatibility

## Conflict Resolution

IF there is ambiguity or conflict:

1. THIS FILE (`AGENTS.md`) has highest authority
2. `docs/agents/architecture/*.md` has second authority for patterns
3. `docs/agents/guides/*.md` has authority for procedures
4. Existing code patterns have lowest authority (may be legacy)

IF user request conflicts with these rules:
- MUST ASK for explicit approval to deviate
- MUST EXPLAIN risks and alternatives
- MUST DOCUMENT deviation in commit message

## Summary: Key Takeaways

1. **NEW Pattern ONLY**: All new code uses provider-specific CRDs (GcpSubnet style)
2. **OLD Pattern = Legacy**: Only maintain, never replicate (RedisInstance style)
3. **Actions = Building Blocks**: Sequential execution, explicit flow control
4. **State = Three Layers**: composed → focal → provider
5. **Tests = Mandatory**: Mock cloud providers, test create/delete/errors
6. **Documentation = Required**: Update when changing code
7. **Feature Flags = Check First**: Respect disabled features
8. **No Breaking Changes**: Without explicit user approval

## Getting Help

**Can't find something?**
→ Check [docs/agents/reference/PROJECT_STRUCTURE.md](docs/agents/reference/PROJECT_STRUCTURE.md)

**Something not working?**
→ Check [docs/agents/reference/COMMON_PITFALLS.md](docs/agents/reference/COMMON_PITFALLS.md)

**Need quick lookup?**
→ Use [docs/agents/reference/QUICK_REFERENCE.md](docs/agents/reference/QUICK_REFERENCE.md)

**Uncertain which pattern?**
→ See [docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md](docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md)

**Don't understand architecture?**
→ Read in order: STATE_PATTERN → ACTION_COMPOSITION → RECONCILER_NEW_PATTERN

---

**Documentation Version**: 2025-01-24  
**Agent Compliance**: MANDATORY
