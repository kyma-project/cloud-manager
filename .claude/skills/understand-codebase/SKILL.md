---
name: understand-codebase
description: Navigate and understand Cloud Manager codebase structure. Use when finding files, exploring project layout, understanding package organization, or locating specific code.
disable-model-invocation: true
---

# Understand Codebase

Navigate and understand Cloud Manager project structure.

## Repository Overview

Cloud Manager is a Kubernetes controller manager (Kubebuilder) that provisions cloud resources (AWS, Azure, GCP) for SAP BTP Kyma runtime.

## Key Directories

```
cloud-manager/
├── api/                          # CRD definitions
│   ├── cloud-control/v1beta1/    # KCP (control plane) resources
│   └── cloud-resources/v1beta1/  # SKR (user-facing) resources
│
├── pkg/
│   ├── composed/                 # Action composition framework
│   │   ├── action.go             # ComposeActions, IfElse
│   │   └── state.go              # Base State interface
│   │
│   ├── common/
│   │   └── actions/
│   │       └── focal/            # Scope management for KCP
│   │
│   ├── kcp/                      # KCP reconcilers
│   │   └── provider/
│   │       ├── gcp/<resource>/   # GCP resources (NEW pattern)
│   │       ├── azure/<resource>/ # Azure resources
│   │       └── aws/<resource>/   # AWS resources
│   │
│   ├── skr/                      # SKR reconcilers
│   │   └── <resource>/           # User-facing resources
│   │
│   └── testinfra/                # Test infrastructure
│
├── internal/controller/          # Controller tests
│   ├── cloud-control/            # KCP resource tests
│   └── cloud-resources/          # SKR resource tests
│
├── config/
│   ├── crd/bases/                # Generated CRDs
│   ├── patchAfterMakeManifests.sh
│   └── sync.sh
│
└── cmd/main.go                   # Controller registration
```

## Finding Files

### By Resource Type

| Looking for | Location |
|-------------|----------|
| KCP CRD definition | `api/cloud-control/v1beta1/<resource>_types.go` |
| SKR CRD definition | `api/cloud-resources/v1beta1/<resource>_types.go` |
| KCP reconciler | `pkg/kcp/provider/<provider>/<resource>/` |
| SKR reconciler | `pkg/skr/<resource>/` |
| Controller test | `internal/controller/cloud-control/<resource>_test.go` |
| Mock | `pkg/kcp/provider/<provider>/mock/` |
| Generated CRD | `config/crd/bases/` |

### By Provider

| Provider | KCP Location | Example Resources |
|----------|--------------|-------------------|
| GCP | `pkg/kcp/provider/gcp/` | subnet, nfsinstance, redisinstance |
| Azure | `pkg/kcp/provider/azure/` | vnetpeering, redisinstance |
| AWS | `pkg/kcp/provider/aws/` | nfsinstance, vpcpeering |

### By File Purpose

| File Name | Purpose |
|-----------|---------|
| `reconcile.go` | Reconciler entry point, SetupWithManager |
| `state.go` | State struct and factory |
| `load*.go` | Load remote resource action |
| `create*.go` | Create resource action |
| `update*.go` | Update resource action |
| `delete*.go` | Delete resource action |
| `wait*.go` | Wait for async operation |
| `updateStatus.go` | Status update action |
| `client/client.go` | Cloud provider client interface |

## API Groups

| Group | Version | Purpose |
|-------|---------|---------|
| `cloud-control.kyma-project.io` | v1beta1 | KCP resources (internal) |
| `cloud-resources.kyma-project.io` | v1beta1 | SKR resources (user-facing) |

## Pattern Recognition

### NEW Pattern (Use for new code)
- CRD: Provider-specific name (`GcpSubnet`, `AzureVNetLink`)
- Location: `pkg/kcp/provider/<provider>/<resource>/`
- State: Extends `focal.State` directly

### OLD Pattern (Maintenance only)
- CRD: Multi-provider name (`RedisInstance`, `NfsInstance`)
- Location: `pkg/kcp/<resource>/` + `pkg/kcp/provider/<provider>/<resource>/`
- State: 3-layer hierarchy

## Core Framework Files

| File | Purpose |
|------|---------|
| `pkg/composed/action.go` | Action composition (ComposeActions, IfElse) |
| `pkg/composed/state.go` | Base State interface |
| `pkg/common/actions/focal/state.go` | Focal state with Scope |
| `pkg/feature/feature.go` | Feature flag definitions |

## Test Infrastructure

| Location | Purpose |
|----------|---------|
| `pkg/testinfra/infra.go` | Test setup and teardown |
| `pkg/testinfra/dsl/` | Test DSL helpers |
| `pkg/kcp/provider/<provider>/mock/` | Provider mocks |

## Quick Commands

```bash
# Find file by pattern
find . -name "*subnet*" -type f

# Find code by content
grep -r "GcpSubnet" --include="*.go"

# List all resources in API
ls api/cloud-control/v1beta1/*_types.go

# List all GCP reconcilers
ls pkg/kcp/provider/gcp/
```

## Related

- Full structure: [docs/agents/reference/PROJECT_STRUCTURE.md](../../../docs/agents/reference/PROJECT_STRUCTURE.md)
- Architecture: [docs/agents/architecture/](../../../docs/agents/architecture/)
- Quick reference: [docs/agents/reference/QUICK_REFERENCE.md](../../../docs/agents/reference/QUICK_REFERENCE.md)
