# Project Structure Guide

**Authority**: REQUIRED for navigating Cloud Manager codebase. All paths are deterministic.

**Target**: LLM coding agents needing to locate files, understand organization, and navigate dependencies.

**Related**: [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | [COMMON_PITFALLS.md](COMMON_PITFALLS.md) | [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md)

---

## Authority: Navigation Rules

### MUST Use For
- Locating reconciler implementations
- Finding CRD definitions
- Discovering test files
- Understanding package dependencies

### MUST NOT
- Guess file locations - consult this guide
- Modify files outside pattern boundaries
- Create files in wrong directories
- Import packages violating dependency rules

---

## Root Directory Structure

```
cloud-manager/
├── api/                        # CRD type definitions (NEVER edit zz_generated files)
├── cmd/                        # Application entry point (main.go)
├── config/                     # Kubernetes manifests and configuration
├── docs/agents/                # Agent documentation (this file)
├── e2e/                        # End-to-end tests (Gherkin features)
├── internal/                   # Internal packages (tests, private code)
├── pkg/                        # Main application code
└── tools/                      # Development tools
```

---

## API Definitions (`api/`)

### Structure

| Path | Purpose | Files |
|------|---------|-------|
| `api/cloud-control/v1beta1/` | KCP (control plane) CRDs | `*_types.go`, `common.go`, `scope_types.go` |
| `api/cloud-resources/v1beta1/` | SKR (user-facing) CRDs | `*_types.go` |

### Key Files

| File | Purpose | When to Modify |
|------|---------|----------------|
| `cloud-control/v1beta1/common.go` | Shared KCP types (conditions, status) | Adding common KCP status fields |
| `cloud-control/v1beta1/scope_types.go` | Cloud provider credentials | Never (managed by separate system) |
| `cloud-control/v1beta1/gcpsubnet_types.go` | NEW pattern KCP CRD | Adding NEW pattern GCP resource |
| `cloud-control/v1beta1/redisinstance_types.go` | OLD pattern multi-provider CRD | Maintaining OLD pattern resources |
| `cloud-resources/v1beta1/gcpnfsvolume_types.go` | SKR user-facing CRD | Adding user-facing resource |
| `**/zz_generated.deepcopy.go` | Generated DeepCopy methods | NEVER (auto-generated) |

### CRD Naming Rules

- **KCP NEW Pattern**: `Gcp<Resource>`, `Azure<Resource>`, `Aws<Resource>` (provider-specific)
- **KCP OLD Pattern**: `<Resource>Instance` (multi-provider with `.spec.instance.gcp/aws/azure`)
- **SKR Pattern**: Always `Gcp<Resource>`, `Azure<Resource>`, `Aws<Resource>` (provider-specific)

### After Modifying API

```bash
make manifests                              # Generate CRDs
./config/patchAfterMakeManifests.sh        # Add version annotations
./config/sync.sh                            # Sync to dist/ directories
```

---

## Application Code (`pkg/`)

### Top-Level Packages

| Package | Purpose | Import Rules |
|---------|---------|--------------|
| `pkg/composed/` | Core action composition framework | NO internal dependencies |
| `pkg/common/` | Shared utilities and actions | Imports `pkg/composed/` only |
| `pkg/kcp/` | KCP reconcilers (control plane) | Imports `pkg/composed/`, `pkg/common/` |
| `pkg/skr/` | SKR reconcilers (user-facing) | Imports `pkg/composed/`, `pkg/common/` |
| `pkg/testinfra/` | Test infrastructure and mocks | Imports `pkg/composed/`, `pkg/common/` |
| `pkg/feature/` | Feature flag system | Standalone |
| `pkg/config/` | Configuration management | Standalone |
| `pkg/util/` | General utilities | Standalone |

---

## Core Framework (`pkg/composed/`)

**Purpose**: Foundation for all reconciliation logic.

| File | Contains | When to Read |
|------|----------|--------------|
| `action.go` | Action type, ComposeActions, IfElse | Understanding action composition |
| `errors.go` | StopAndForget, StopWithRequeue | Understanding return values |
| `handling.go` | Result() function | Understanding reconciliation flow |
| `predicate.go` | Common predicates | Writing conditionals |
| `state.go` | Base State interface | Understanding state hierarchy |
| `updateStatus.go` | Status update helpers | Persisting status |

**Import Path**: `github.com/kyma-project/cloud-manager/pkg/composed`

---

## Common Actions (`pkg/common/actions/`)

**Purpose**: Reusable actions across all reconcilers.

### Key Subdirectories

| Directory | Contains | When to Use |
|-----------|----------|-------------|
| `focal/` | Scope management (loadScope, focal.State) | ALL KCP reconcilers |
| `addFinalizer.go` | Add finalizer action | Resources requiring cleanup |
| `removeFinalizer.go` | Remove finalizer action | After cleanup complete |

### focal Package

| File | Purpose | Usage |
|------|---------|-------|
| `state.go` | Focal state interface (adds Scope, Cluster) | Extending in provider state |
| `stateFactory.go` | Creates focal state | KCP reconciler initialization |
| `loadScope.go` | Loads Scope resource | First action in KCP reconcilers |

**Import Path**: `github.com/kyma-project/cloud-manager/pkg/common/actions/focal`

---

## KCP Reconcilers (`pkg/kcp/`)

### Directory Structure

```
pkg/kcp/
├── provider/                   # Provider-specific implementations
│   ├── gcp/                    # GCP resources
│   │   ├── client/             # GCP client management
│   │   │   ├── gcpClients.go   # NEW: Centralized client initialization
│   │   │   └── provider.go     # OLD: Legacy client provider
│   │   │
│   │   ├── subnet/             # NEW Pattern: Provider-specific CRD
│   │   │   ├── reconcile.go
│   │   │   ├── state.go
│   │   │   ├── client/         # Subnet-specific client
│   │   │   ├── load*.go
│   │   │   ├── create*.go
│   │   │   ├── delete*.go
│   │   │   ├── wait*.go
│   │   │   └── updateStatus.go
│   │   │
│   │   ├── redisinstance/      # Provider-specific logic (OLD pattern)
│   │   └── mock/               # GCP API mocks for testing
│   │
│   ├── aws/                    # AWS resources (similar structure)
│   └── azure/                  # Azure resources (similar structure)
│
└── redisinstance/              # OLD Pattern: Shared multi-provider logic
    ├── reconciler.go           # Provider switching reconciler
    ├── state.go                # Shared state
    └── types/                  # Shared state interface
```

### Finding Reconciler Implementation

**Decision Tree**:
```
What CRD are you working with?
├─ Provider-specific name (GcpSubnet)?
│  └─ Location: pkg/kcp/provider/<provider>/<resource>/
│     └─ Pattern: NEW
│
└─ Multi-provider name (RedisInstance)?
   ├─ Shared logic: pkg/kcp/<resource>/
   └─ Provider logic: pkg/kcp/provider/<provider>/<resource>/
   └─ Pattern: OLD
```

### NEW Pattern File Organization

| File | Purpose | MUST Contain |
|------|---------|--------------|
| `reconcile.go` | Reconciler setup | SetupWithManager, newFlow, newAction |
| `state.go` | State struct and factory | State extends focal.State, StateFactory |
| `client/<resource>Client.go` | Typed client interface | Resource-specific operations |
| `load<Dependency>.go` | Load dependency action | Network, VPC, parent resources |
| `load<Resource>.go` | Load target resource | Check if exists in cloud |
| `create<Resource>.go` | Create resource action | Create in cloud provider |
| `delete<Resource>.go` | Delete resource action | Delete from cloud provider |
| `wait<Operation>Done.go` | Wait for async operation | Poll until operation complete |
| `updateStatus.go` | Update status action | Sync status from cloud state |

### OLD Pattern File Organization

| File | Purpose | Location |
|------|---------|----------|
| `reconciler.go` | Provider switching | `pkg/kcp/<resource>/reconciler.go` |
| `state.go` | Shared state interface | `pkg/kcp/<resource>/types/state.go` |
| `state.go` | Shared state impl | `pkg/kcp/<resource>/state.go` |
| `new.go` | Provider actions | `pkg/kcp/provider/<provider>/<resource>/new.go` |
| `state.go` | Provider state | `pkg/kcp/provider/<provider>/<resource>/state.go` |
| `load*.go` | Actions | `pkg/kcp/provider/<provider>/<resource>/` |

---

## SKR Reconcilers (`pkg/skr/`)

**Purpose**: User-facing reconcilers that project to KCP.

### Structure

```
pkg/skr/
├── runtime/                    # SKR runtime management
│   ├── reconciler.go
│   └── skr.go
│
├── gcpnfsvolume/               # NEW: Provider-specific SKR
│   ├── reconciler.go
│   ├── state.go
│   ├── createKcpNfsVolume.go   # Create corresponding KCP resource
│   ├── loadKcpNfsVolume.go     # Load KCP resource
│   ├── deleteKcpNfsVolume.go   # Delete KCP resource
│   └── updateStatus.go         # Sync status from KCP
│
└── gcpredisinstance/           # OLD: Creates multi-provider KCP
    ├── reconciler.go
    ├── createKcpRedisInstance.go
    └── ...
```

### SKR File Organization

| File | Purpose | Key Logic |
|------|---------|-----------|
| `reconciler.go` | Reconciler setup | SetupWithManager, feature flag loading |
| `state.go` | State with SKR+KCP access | SKR client, KCP client |
| `createKcpNfsVolume.go` | Create KCP resource | UUID generation, KCP annotation |
| `loadKcpNfsVolume.go` | Load KCP resource | Find by KCP annotation |
| `deleteKcpNfsVolume.go` | Delete KCP resource | Remove KCP resource |
| `updateStatus.go` | Sync status | Copy KCP status to SKR |

---

## Testing Infrastructure (`pkg/testinfra/`)

**Purpose**: Mock infrastructure for controller tests.

| Directory | Contains | Usage |
|-----------|----------|-------|
| `infra.go` | Main test infrastructure | Start/stop test environment |
| `dsl/` | Test DSL helpers | LoadAndCheck, matchers, builders |
| `gardener/` | Gardener/Shoot utilities | Shoot testing |

### Mock Implementations

```
pkg/kcp/provider/gcp/mock/      # GCP API mocks
├── type.go                     # Server interface (aggregates all mocks)
├── server.go                   # Server implementation
├── memoryStoreClientFake.go    # Redis instance mock
├── computeClientFake.go        # Compute resources mock
└── ...

pkg/kcp/provider/aws/mock/      # AWS API mocks
pkg/kcp/provider/azure/mock/    # Azure API mocks
```

**Mock Pattern**: Each mock implements:
- **Client interface**: Methods reconcilers call
- **Utils interface**: Methods tests call to manipulate state

---

## Test Suites (`internal/`)

### Controller Tests

| Directory | Contains | Purpose |
|-----------|----------|---------|
| `internal/controller/cloud-control/` | KCP reconciler tests | Test KCP reconciliation logic |
| `internal/controller/cloud-resources/` | SKR reconciler tests | Test SKR reconciliation logic |
| `suite_test.go` | Test suite setup | Initialize testinfra |
| `*_gcp_test.go` | Provider-specific tests | GCP resource tests |
| `*_aws_test.go` | Provider-specific tests | AWS resource tests |

### API Validation Tests

| Directory | Contains | Purpose |
|-----------|----------|---------|
| `internal/api-tests/` | CRD validation tests | Test CRD validation rules |
| `builder_test.go` | Test helpers | canCreateSkr, canNotCreateSkr |
| `skr_*_test.go` | SKR CRD validation | Test SKR resource validation |
| `kcp_*_test.go` | KCP CRD validation | Test KCP resource validation |

---

## Configuration (`config/`)

| Directory | Purpose | When to Modify |
|-----------|---------|----------------|
| `crd/bases/` | Generated CRDs | Never (run `make manifests`) |
| `crd/patches/` | CRD patches | Customizing CRD manifests |
| `rbac/` | RBAC resources | Adding permissions |
| `manager/` | Manager deployment | Deployment config changes |
| `samples/` | Example resources | Adding examples |
| `featureToggles/` | Feature flag overrides | Local development |

### Post-Generation Scripts

| Script | Purpose | When to Run |
|--------|---------|-------------|
| `patchAfterMakeManifests.sh` | Add version annotations | After `make manifests` |
| `sync.sh` | Sync CRDs to dist/ | After patching |

---

## Command Entry Point (`cmd/main.go`)

**Key Responsibilities**:
1. Initialize manager
2. Load configuration
3. Setup cloud provider clients (NEW: `NewGcpClients()`)
4. Register reconcilers
5. Start manager

**Relevant Code Sections**:
```go
// Initialize GCP clients (NEW pattern)
gcpClients, err := gcpclient.NewGcpClients(ctx, config.GcpConfig.CredentialsFile, ...)

// Create state factories
subnetStateFactory := gcpsubnet.NewStateFactory(
    gcpsubnetclient.NewComputeClientProvider(gcpClients),
    env,
)

// Register reconcilers
if err = gcpsubnet.NewGcpSubnetReconciler(
    composedStateFactory,
    focalStateFactory,
    subnetStateFactory,
).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "GcpSubnet")
    os.Exit(1)
}
```

---

## Feature Flags (`pkg/feature/`)

| File | Purpose | When to Modify |
|------|---------|----------------|
| `flags.go` | Flag definitions | Adding new flags |
| `ff_ga.yaml` | GA features | Stable features |
| `ff_edge.yaml` | Edge features | Experimental features |
| `schema.json` | JSON schema | Validation rules |
| `provider.go` | Flag evaluation | Flag query logic |

---

## Navigation Guide

### Finding Reconciler by Resource Name

**Step 1**: Determine resource type (SKR or KCP)
- SKR: User-facing, in user clusters
- KCP: Control plane, manages cloud resources

**Step 2**: Check CRD name pattern
- Provider-specific (`GcpSubnet`) → NEW pattern
- Multi-provider (`RedisInstance`) → OLD pattern

**Step 3**: Navigate to directory

| Resource Type | Pattern | Directory |
|---------------|---------|-----------|
| KCP | NEW | `pkg/kcp/provider/<provider>/<resource>/` |
| KCP | OLD | `pkg/kcp/<resource>/` + `pkg/kcp/provider/<provider>/<resource>/` |
| SKR | Any | `pkg/skr/<resource>/` |

**Step 4**: Find key files
- `reconcile.go` - Entry point
- `state.go` - State definition
- Action files - Individual logic

### Finding Tests for Reconciler

**Controller Tests**: `internal/controller/cloud-control/<resource>_<provider>_test.go`

**API Validation**: `internal/api-tests/kcp_<resource>_test.go`

### Finding Mocks for Provider

**GCP**: `pkg/kcp/provider/gcp/mock/`
**AWS**: `pkg/kcp/provider/aws/mock/`
**Azure**: `pkg/kcp/provider/azure/mock/`

Look for:
- `type.go` - Server interface
- `*ClientFake.go` - Mock implementations

---

## Package Dependencies

### Dependency Rules

```
cmd/main.go
  ↓
pkg/kcp/provider/<provider>/<resource>/reconcile.go
  ↓
pkg/kcp/provider/<provider>/<resource>/state.go
  ↓
pkg/common/actions/focal/
  ↓
pkg/composed/
```

**MUST Follow**:
- `pkg/composed/` imports NO internal packages
- `pkg/common/` imports only `pkg/composed/`
- Reconcilers import `pkg/composed/` and `pkg/common/`
- Tests import `pkg/testinfra/`

**MUST NOT**:
- Circular imports
- Reconcilers importing other reconcilers
- `pkg/composed/` importing application packages

---

## File Naming Patterns

| Pattern | Example | Purpose |
|---------|---------|---------|
| `<resource>_types.go` | `gcpsubnet_types.go` | CRD type definition |
| `reconcile.go` | `reconcile.go` | Reconciler setup |
| `state.go` | `state.go` | State struct and factory |
| `load<Resource>.go` | `loadSubnet.go` | Load remote resource |
| `create<Resource>.go` | `createSubnet.go` | Create resource action |
| `update<Resource>.go` | `updateSubnet.go` | Update resource action |
| `delete<Resource>.go` | `deleteSubnet.go` | Delete resource action |
| `wait<Resource><State>.go` | `waitSubnetReady.go` | Wait for state transition |
| `updateStatus.go` | `updateStatus.go` | Status update action |
| `<resource>_test.go` | `subnet_test.go` | Controller tests |
| `<resource>_gcp_test.go` | `redisinstance_gcp_test.go` | Provider-specific tests |
| `*ClientFake.go` | `computeClientFake.go` | Mock client implementation |

---

## Quick Checklists

### Before Creating New KCP Reconciler

- [ ] Decide pattern: NEW (provider-specific CRD) or OLD (multi-provider)
- [ ] Check if CRD exists in `api/cloud-control/v1beta1/`
- [ ] Verify client pattern: NEW (GcpClients) or OLD (ClientProvider)
- [ ] Create directory: `pkg/kcp/provider/<provider>/<resource>/`
- [ ] Follow file naming conventions

### Before Creating New SKR Reconciler

- [ ] Verify corresponding KCP reconciler exists
- [ ] Check if CRD exists in `api/cloud-resources/v1beta1/`
- [ ] Create directory: `pkg/skr/<resource>/`
- [ ] Plan ID generation strategy (UUID recommended)

### Before Modifying API

- [ ] Understand breaking change implications
- [ ] Update CRD type file in `api/`
- [ ] Run `make manifests`
- [ ] Run `./config/patchAfterMakeManifests.sh`
- [ ] Run `./config/sync.sh`
- [ ] Update documentation

---

## Getting Help

**Can't find file?** → Use this guide's navigation trees

**Need quick lookup?** → [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

**Understanding patterns?** → [RECONCILER_PATTERN_COMPARISON.md](../architecture/RECONCILER_PATTERN_COMPARISON.md)

**Creating reconciler?** → [ADD_KCP_RECONCILER.md](../guides/ADD_KCP_RECONCILER.md)

**Writing tests?** → [CONTROLLER_TESTS.md](../guides/CONTROLLER_TESTS.md)
