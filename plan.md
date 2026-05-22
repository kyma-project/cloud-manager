# GCP Security Reconciler — Implementation Plan

## Context

Implementing PCI DSS compliance for GCP by enabling 5 SCC security services at the GCP **project** level.

Wireframe already exists in `pkg/kcp/provider/gcp/security/` (state.go, new.go are stubs).
Reference implementation: `pkg/kcp/provider/azure/security/`.

---

## Deep Dive: GCP Security Command Center (SCC)

### SCC API Families

There are two related but separate GCP SCC APIs:

| Package | Purpose |
|---------|---------|
| `cloud.google.com/go/securitycenter/apiv1` | Findings, sources, assets (read/query) |
| `cloud.google.com/go/securitycentermanagement/apiv1` | **Enable/disable security services** ← we need this |

Source: [pkg.go.dev/cloud.google.com/go/securitycentermanagement/apiv1](https://pkg.go.dev/cloud.google.com/go/securitycentermanagement/apiv1)

### SecurityCenterService Resource

The resource managed by this API is `SecurityCenterService`, representing a security scanner or detection service at a given resource scope (org / folder / project).

Resource name pattern (project-scope):
```
projects/{project}/locations/global/securityCenterServices/{service_id}
```

Key Go type — `securitycentermanagementpb.SecurityCenterService`:
```go
type SecurityCenterService struct {
    Name                     string
    IntendedEnablementState  SecurityCenterService_EnablementState  // what we SET
    EffectiveEnablementState SecurityCenterService_EnablementState  // what GCP applies
    Modules                  map[string]*SecurityCenterService_ModuleSettings
    UpdateTime               *timestamppb.Timestamp
}
```

### Enablement State Enum

From [github.com/googleapis/google-cloud-go](https://github.com/googleapis/google-cloud-go/blob/main/securitycentermanagement/apiv1/securitycentermanagementpb/security_center_management.pb.go):

| Const | Value | Meaning |
|-------|-------|---------|
| `ENABLEMENT_STATE_UNSPECIFIED` | 0 | Default, unused |
| `INHERITED` | 1 | Inherit from parent (org/folder) |
| `ENABLED` | 2 | Explicitly enabled |
| `DISABLED` | 3 | Explicitly disabled |
| `INGEST_ONLY` | 4 | Ingest findings but don't enable |

We use `ENABLED` and `DISABLED`.

### Key Client Methods

```go
// Get current state for a service
func (c *Client) GetSecurityCenterService(
    ctx context.Context,
    req *securitycentermanagementpb.GetSecurityCenterServiceRequest,
    opts ...gax.CallOption,
) (*securitycentermanagementpb.SecurityCenterService, error)

// Update (enable/disable) a service
func (c *Client) UpdateSecurityCenterService(
    ctx context.Context,
    req *securitycentermanagementpb.UpdateSecurityCenterServiceRequest,
    opts ...gax.CallOption,
) (*securitycentermanagementpb.SecurityCenterService, error)
```

`UpdateSecurityCenterServiceRequest` fields:
- `SecurityCenterService` — the updated resource (only changed fields)
- `UpdateMask *fieldmaskpb.FieldMask` — specify `["intended_enablement_state"]`
- `ValidateOnly bool` — dry-run flag

### Service Name Identifiers

From the Go proto file and GCP documentation:

| Feature | Service ID in resource name |
|---------|----------------------------|
| Security Health Analytics | `security-health-analytics` |
| Web Security Scanner | `web-security-scanner` |
| Event Threat Detection | `event-threat-detection` |
| VM Threat Detection | `vm-threat-detection` |

Note: "Vulnerability Assessment" (Rapid Vulnerability Detection) is **not exposed** via the `securityCenterServices` endpoint. It is managed through a separate GCP API (VM Manager / OS Config) and is out of scope for this reconciler.

---

## Multi-SKR / Multi-Runtime Per Project Logic

SCC services are **project-scoped**. Multiple SKR Runtimes can exist inside the same GCP project. The existing framework already handles this:

In `pkg/kcp/runtime/securityEnabledDetermine.go`:
- Lists all Runtimes sharing the same `SecretBindingName` (= same GCP project credentials)
- Sets `securityServiceEnabledOnSubscription = true` if **any** of them has the security label

Exposed via `runtimetypes.State`:
- `SecurityServiceEnabledOnSubscription() bool` → `true` if any runtime in project has security on

**Enable rule**: enable SCC services when `SecurityServiceEnabledOnSubscription() == true`
**Disable rule**: disable SCC services only when `SecurityServiceEnabledOnSubscription() == false` (= all runtimes in project are security-off)

---

## New Dependency — REQUIRES EXPLICIT APPROVAL

`cloud.google.com/go/securitycentermanagement/apiv1` is NOT in go.mod.

This is forbidden without explicit user approval per AGENTS.md. Before coding Phase 1b and 1c below, **confirm with the user** that this dependency may be added via `go get cloud.google.com/go/securitycentermanagement`.

---

## File Structure

```
pkg/kcp/provider/gcp/
├── client/
│   ├── clientSecurityCenterMgmt.go  NEW — SecurityCenterManagementClient iface + impl
│   └── gcpClients.go                MODIFY — add SCC mgmt client field + init + wrapped
└── security/
    ├── client/
    │   └── client.go                REWRITE — Client iface + NewClientProvider
    ├── state.go                     REWRITE — StateFactory with GCP client, loaded state fields
    ├── new.go                       REWRITE — action pipeline
    ├── sccServicesLoad.go           NEW
    ├── sccServicesEnable.go         NEW
    └── sccServicesDisable.go        NEW
```

---

## Implementation Steps

### Phase 0 — Dependency ✅ DONE

```bash
go get cloud.google.com/go/securitycentermanagement@latest
go mod tidy
```

### Phase 1 — New SecurityCenterManagementClient ✅ DONE

**`pkg/kcp/provider/gcp/client/clientSecurityCenterMgmt.go`** (new file)

```go
package client

import (
    securitycentermanagement "cloud.google.com/go/securitycentermanagement/apiv1"
    "cloud.google.com/go/securitycentermanagement/apiv1/securitycentermanagementpb"
    "google.golang.org/protobuf/types/known/fieldmaskpb"
)

type SecurityCenterManagementClient interface {
    GetSecurityCenterService(ctx context.Context, name string) (*securitycentermanagementpb.SecurityCenterService, error)
    UpdateSecurityCenterService(
        ctx context.Context,
        svc *securitycentermanagementpb.SecurityCenterService,
        mask *fieldmaskpb.FieldMask,
    ) (*securitycentermanagementpb.SecurityCenterService, error)
}

// sccMgmtClient is the concrete implementation
type sccMgmtClient struct {
    inner *securitycentermanagement.Client
}
```

### Phase 1c — Add to GcpClients

**`pkg/kcp/provider/gcp/client/gcpClients.go`**

1. Add field:
   ```go
   SecurityCenterManagement *securitycentermanagement.Client
   ```
2. In `NewGcpClients`: create SCC management client using GRPC (same pattern as redisCluster):
   ```go
   sccTokenProvider, err := b.WithScopes(securitycentermanagement.DefaultAuthScopes()).BuildTokenProvider()
   sccTokenSource := oauth2adapt.TokenSourceFromTokenProvider(sccTokenProvider)
   sccClient, err := securitycentermanagement.NewClient(ctx,
       option.WithTokenSource(sccTokenSource),
       option.WithGRPCDialOption(grpc.WithUnaryInterceptor(metrics.UnaryClientInterceptor())),
   )
   ```
3. Add `SecurityCenterManagementWrapped()` method returning `SecurityCenterManagementClient`.

### Phase 2 — Security Package Client ✅ DONE

**`pkg/kcp/provider/gcp/security/client/client.go`**

```go
package client

import gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"

type Client interface {
    gcpclient.SecurityCenterManagementClient     // Get + Update SCC services
}

func NewClientProvider(gcpClients *gcpclient.GcpClients) gcpclient.GcpClientProvider[Client] {
    return func(_ string) Client {
        return &client{
            SecurityCenterManagementClient: gcpClients.SecurityCenterManagementWrapped(),
        }
    }
}
```

### Phase 3 — State ✅ DONE

**`pkg/kcp/provider/gcp/security/state.go`**

```go
type State struct {
    runtimetypes.State
    gcpClient securityclient.Client

    // sccServices: keyed by service ID (e.g. "security-health-analytics")
    // nil = not yet loaded
    sccServices map[string]*securitycentermanagementpb.SecurityCenterService
}

type StateFactory interface {
    NewState(ctx context.Context, runtimeState runtimetypes.State) (context.Context, composed.State, error)
}

type stateFactory struct {
    gcpClientProvider gcpclient.GcpClientProvider[securityclient.Client]
}
```

`NewState` implementation:
- Extracts project from `runtimeState.Subscription().Status.SubscriptionInfo.Gcp.Project`
- Gets client: `f.gcpClientProvider(project)`
- Logs `gcpProjectId`
- Returns `(ctx, newState(...), nil)`

Return type matches Azure pattern: `(context.Context, composed.State, error)`.

### Phase 4 — Action Files ✅ DONE

#### `sccServicesLoad.go`

```go
var sccServiceIDs = []string{
    "security-health-analytics",
    "web-security-scanner",
    "event-threat-detection",
    "vm-threat-detection",
    "rapid-vulnerability-detection",  // [INFERRED] = Vulnerability Assessment
}

func sccServicesLoad(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    project := state.Subscription().Status.SubscriptionInfo.Gcp.Project

    state.sccServices = make(map[string]*securitycentermanagementpb.SecurityCenterService)
    for _, svcID := range sccServiceIDs {
        name := fmt.Sprintf("projects/%s/locations/global/securityCenterServices/%s", project, svcID)
        svc, err := state.gcpClient.GetSecurityCenterService(ctx, name)
        if err != nil {
            return composed.LogErrorAndReturn(err, "Error loading SCC service "+svcID, composed.StopWithRequeue, ctx)
        }
        state.sccServices[svcID] = svc
    }
    return nil, ctx
}
```

#### `sccServicesEnable.go`

Iterates `state.sccServices`. For each service where `IntendedEnablementState != ENABLED`:
- Calls `UpdateSecurityCenterService` setting `IntendedEnablementState = ENABLED`
- Update mask: `["intended_enablement_state"]`
- Returns `StopWithRequeue` if any update was made (so state reloads next reconcile)

#### `sccServicesDisable.go`

Same pattern as enable, setting `IntendedEnablementState = DISABLED`.

### Phase 5 — Action Pipeline ✅ DONE

**`pkg/kcp/provider/gcp/security/new.go`**

```go
func New(sf StateFactory) composed.Action {
    return func(ctx context.Context, st composed.State) (error, context.Context) {
        runtimeState := st.(runtimetypes.State)
        cctx, state, err := sf.NewState(ctx, runtimeState)
        if cctx != nil {
            ctx = cctx
        }
        if err != nil {
            return err, ctx
        }

        return composed.ComposeActionsNoName(
            // SCC services: project-scoped, controlled by subscription-level flag
            sccServicesLoad,
            composed.IfElse(
                runtimeState.SecurityServiceEnabledOnSubscriptionPredicate,
                sccServicesEnable,
                sccServicesDisable,
            ),
        )(ctx, state)
    }
}
```

---

## Key Design Decisions

### 1. Multi-SKR project logic via subscription predicate
`SecurityServiceEnabledOnSubscriptionPredicate` gates SCC operations — ANY runtime in the project triggers enable, ALL must be off to disable.

### 2. Idempotency via load-then-check
`sccServicesEnable/Disable` load current state first and only call the GCP API when an actual change is needed. Returns `StopWithRequeue` if a change was made so the next reconcile verifies success.

### 3. Vulnerability Assessment is out of scope
GCP does not expose Vulnerability Assessment / Rapid Vulnerability Detection via the `securityCenterServices` management API. It is managed through VM Manager / OS Config — a separate API surface not covered by this reconciler.

### 4. SCC services at `locations/global`
The location for project-level SCC service management is always `global`, regardless of the GCP region of the runtime. Resource path: `projects/{project}/locations/global/securityCenterServices/{service}`.
Source: [pkg.go.dev/cloud.google.com/go/securitycentermanagement/apiv1](https://pkg.go.dev/cloud.google.com/go/securitycentermanagement/apiv1)

---

---

## Phase 6 — GCP IAM Permissions

**File**: `docs/contributor/permissions/gcp/gcp_default.yaml`

### Permissions needed

#### SCC Security Center Management API

`GetSecurityCenterService` and `UpdateSecurityCenterService` require:

| Permission | Used by |
|-----------|---------|
| `securitycentermanagement.securityCenterServices.get` | `sccServicesLoad` |
| `securitycentermanagement.securityCenterServices.update` | `sccServicesEnable`, `sccServicesDisable` |

Normally granted via `roles/securitycenter.settingsEditor` or `roles/securitycentermanagement.editor`, but we add the granular permissions directly.

### Diff to apply

```yaml
# Add new securitycentermanagement block (alphabetical order, after servicenetworking.*, before serviceusage.*):
  - securitycentermanagement.securityCenterServices.get
  - securitycentermanagement.securityCenterServices.update
```

---

## What Is NOT Covered (out of scope per task description)

- SCC findings export configuration (log export comes later)
- Creating BigQuery/Pub/Sub SCC export sinks
- SCC notification configs
- Tests (separate skill: `/testing-cloud-manager-code`)
