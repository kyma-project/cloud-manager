# Cloud Manager - AI Agent Instructions

**Target**: LLM coding agents (Claude, Cursor, Copilot, etc.)

## Project Identity

Cloud Manager is a Kubernetes controller manager (Kubebuilder) that provisions cloud resources (AWS, Azure, GCP) for SAP BTP Kyma runtime.

## Core Architecture: Dual Reconciliation

**Data Flow**: SKR resource → KCP resource → Cloud Provider API → Status propagates back

| Layer | API Group | CRD Location | Reconciler Location | Example |
|-------|-----------|--------------|---------------------|---------|
| SKR (user-facing) | `cloud-resources.kyma-project.io` | `api/cloud-resources/` | `pkg/skr/` | GcpNfsVolume |
| KCP (control plane) | `cloud-control.kyma-project.io` | `api/cloud-control/` | `pkg/kcp/provider/` | GcpNfsInstance |

**Flow**: User creates SKR resource → SKR reconciler creates KCP resource → KCP reconciler provisions cloud infrastructure → Status flows back: Cloud → KCP → SKR

## Mandatory Rules

### Rule 1: Pattern Selection

| Scenario | Pattern | Example |
|----------|---------|---------|
| New resource (post-2024) | NEW - Provider-specific CRD | `GcpSubnet`, `AzureVNetLink` |
| Existing multi-provider | OLD - Maintain only | `RedisInstance`, `IpRange` |

**NEW pattern**: `pkg/kcp/provider/gcp/subnet/` - Direct state hierarchy
**OLD pattern**: `pkg/kcp/redisinstance/` - Provider switching with `BuildSwitchAction`

### Rule 2: Action Composition

```go
// MUST end flows with StopAndForget
composed.ComposeActions("main",
    loadDependencies,
    loadResource,
    createOrUpdate,
    waitReady,
    updateStatus,
    composed.StopAndForgetAction,  // REQUIRED
)
```

| Return | Effect | Use Case |
|--------|--------|----------|
| `nil, ctx` | Continue | Normal flow |
| `error, ctx` | Stop + requeue | Failure |
| `composed.StopAndForget, nil` | Stop, no requeue | Success |
| `composed.StopWithRequeueDelay(d), nil` | Stop + delayed requeue | Wait for async |

### Rule 3: Status Updates

```go
// ALWAYS check errors
err := state.UpdateObjStatus(ctx)
if err != nil {
    return err, ctx
}
```

### Rule 4: Testing

- Use `testinfra` mocks, NEVER real cloud APIs
- Wrap assertions in `Eventually()` for async reconciliation
- Test both create and delete paths

### Rule 5: Feature Flags

Cloud Manager uses feature flags to enable/disable specific CRD APIs per environment.

**General `apiDisabled` flag**: When set, the reconciler won't run at all - no manual check needed in your code.

**Custom flags**: For feature-specific behavior, check flags in reconciler actions:

```go
if feature.MyCustomFlag.Value(ctx) {
    // Feature-specific logic
}
```

**Flag locations**:
- `pkg/feature/ff_ga.yaml` - Stable features
- `pkg/feature/ff_edge.yaml` - Experimental features

## Forbidden Actions

- **NEVER commit or push** - User handles all git commits
- Modify CRD field types without approval
- Mix NEW and OLD patterns
- Skip `make manifests` after API changes
- Use real cloud APIs in tests

## Quick Commands

| Task | Command |
|------|---------|
| Generate CRDs | `make manifests && ./config/patchAfterMakeManifests.sh && ./config/sync.sh` |
| Run all tests | `make test` |
| Run tests for path | `make test TEST_PATH=pkg/kcp/provider/gcp/subnet` |
| Build | `make build` |

## Key Directories

| Path | Purpose |
|------|---------|
| `api/cloud-control/v1beta1/` | KCP CRDs |
| `api/cloud-resources/v1beta1/` | SKR CRDs |
| `pkg/kcp/provider/<provider>/<resource>/` | KCP reconcilers (NEW) |
| `pkg/skr/<resource>/` | SKR reconcilers |
| `pkg/composed/` | Action framework |
| `internal/controller/` | Tests |

## Detailed Documentation

| Topic | Location |
|-------|----------|
| Project structure | [docs/agents/reference/PROJECT_STRUCTURE.md](docs/agents/reference/PROJECT_STRUCTURE.md) |
| Quick reference | [docs/agents/reference/QUICK_REFERENCE.md](docs/agents/reference/QUICK_REFERENCE.md) |
| Common pitfalls | [docs/agents/reference/COMMON_PITFALLS.md](docs/agents/reference/COMMON_PITFALLS.md) |
| Development workflow | [docs/agents/DEVELOPMENT_WORKFLOW.md](docs/agents/DEVELOPMENT_WORKFLOW.md) |
| State pattern | [docs/agents/architecture/STATE_PATTERN.md](docs/agents/architecture/STATE_PATTERN.md) |
| Action composition | [docs/agents/architecture/ACTION_COMPOSITION.md](docs/agents/architecture/ACTION_COMPOSITION.md) |
| NEW vs OLD patterns | [docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md](docs/agents/architecture/RECONCILER_PATTERN_COMPARISON.md) |

## Task Guides

| Task | Guide |
|------|-------|
| Add KCP reconciler | [docs/agents/guides/ADD_KCP_RECONCILER.md](docs/agents/guides/ADD_KCP_RECONCILER.md) |
| Add SKR reconciler | [docs/agents/guides/ADD_SKR_RECONCILER.md](docs/agents/guides/ADD_SKR_RECONCILER.md) |
| Write tests | [docs/agents/guides/CONTROLLER_TESTS.md](docs/agents/guides/CONTROLLER_TESTS.md) |
| Create mocks | [docs/agents/guides/CREATING_MOCKS.md](docs/agents/guides/CREATING_MOCKS.md) |
| Feature flags | [docs/agents/guides/FEATURE_FLAGS.md](docs/agents/guides/FEATURE_FLAGS.md) |

## Decision Trees

### Which Pattern?
```
New code? ─YES─▶ NEW Pattern (GcpSubnet style)
    │
    NO
    │
    ▼
Maintaining existing? ─▶ Use pattern IT already uses
```

### Where to Add Code?
```
KCP reconciler ─▶ pkg/kcp/provider/<provider>/<resource>/
SKR reconciler ─▶ pkg/skr/<resource>/
Shared action  ─▶ pkg/common/actions/
Mock           ─▶ pkg/kcp/provider/<provider>/mock/
Test           ─▶ internal/controller/cloud-control/ or cloud-resources/
```

## Authority

1. This file (AGENTS.md) - highest
2. docs/agents/architecture/ - patterns
3. docs/agents/guides/ - procedures
4. Existing code - may be legacy, verify pattern first

---
*Last Updated: 2025-03-05*
