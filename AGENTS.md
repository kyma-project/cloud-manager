# Cloud Manager

Kubernetes controller manager (Kubebuilder) provisioning AWS, Azure, GCP resources for SAP BTP Kyma runtime.

## Two Reconciliation Loops

| Loop | API Group | Runs In | Purpose |
|------|-----------|---------|---------|
| SKR | `cloud-resources.kyma-project.io/v1beta1` | Remote Kyma clusters | User-facing resources |
| KCP | `cloud-control.kyma-project.io/v1beta1` | Control plane cluster | Cloud provider provisioning |

Execution flow: User creates SKR resource → SKR reconciler creates KCP resource → KCP reconciler calls cloud API → status propagates back.

## Before You Start Any Implementation

**Creating or modifying a reconciler?** Invoke `/composed-action` before writing code.  
**Writing tests?** Invoke `/testing-cloud-manager-code` before writing code.

## Skills

| Skill | Use When |
|-------|---------|
| `/composed-action` | Creating KCP/SKR reconcilers, adding providers, action pipelines, state factories |
| `/testing-cloud-manager-code` | Writing controller, API validation, unit, or e2e tests; creating provider mocks |

## Rules

**Sequential actions**: Actions execute one at a time. Never in parallel.

**Finalizer Law**: Every resource requiring deletion cleanup MUST have a finalizer. Remove the finalizer ONLY after deletion is fully confirmed. No exceptions: "simple resource", "no cloud API called" — if any cleanup is needed, a finalizer is required.

**Feature flag check**: Every SKR reconciler MUST check `composed.If(feature.ApiDisabledPredicate, composed.StopAndForgetAction)` early in the action pipeline.

**Injectable clock**: If a reconciler performs any time-based logic (scheduling, expiry, rate limiting), the state factory MUST accept `clock.Clock`. Use `clock.RealClock{}` in production, `clock.NewFakeClock()` in tests.

**New CRDs**: Provider-specific names only (`GcpRedisCluster`, `AzureVNetLink`). Never multi-provider CRDs with provider subsections.

**Legacy resources**: `RedisInstance`, `NfsInstance`, `IpRange` use the old multi-provider pattern. Maintain them as-is; do not replicate this pattern.

## Project Layout

```
api/cloud-control/v1beta1/    # KCP CRD type definitions
api/cloud-resources/v1beta1/  # SKR CRD type definitions
pkg/kcp/                      # KCP reconcilers (shared state + provider switch)
pkg/kcp/provider/<p>/<r>/     # Provider-specific implementations
pkg/skr/                      # SKR reconcilers
pkg/composed/                 # Action composition framework
pkg/common/actions/           # Shared reusable actions
internal/controller/          # Controller setup + controller tests
internal/api-tests/           # CRD API validation tests
e2e/                          # End-to-end tests
pkg/testinfra/                # Test infrastructure and provider mocks
pkg/kcp/provider/<p>/mock/    # Provider mock implementations
pkg/feature/                  # Feature flag definitions (ff_ga.yaml, ff_edge.yaml)
```

## Make Commands

| Command | When |
|---------|------|
| `make manifests` | After modifying API types |
| `make generate` | After modifying interfaces |
| `make test` | Before committing |
| `make test-ff` | After editing feature flag YAML |
| `make build` | Verify compilation |

After API changes run in sequence: `make manifests && ./config/patchAfterMakeManifests.sh && ./config/sync.sh`

## Forbidden Without Explicit User Approval

- Modify CRD field types
- Remove existing CRD fields
- Add external dependencies
- Change feature flag default values
- Modify state hierarchy
- Break backwards compatibility
