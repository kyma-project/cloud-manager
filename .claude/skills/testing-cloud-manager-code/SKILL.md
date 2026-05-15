---
name: testing-cloud-manager-code
description: Use when writing tests in Cloud Manager — reconciler behavior (controller tests), CRD field validation and immutability (API tests), complex pure functions (unit tests), or full feature scenarios on a real cluster (e2e tests). Also use when creating mock implementations for cloud provider APIs.
---

# Testing Cloud Manager Code

## Test Type Decision

| What are you testing? | Test type | Location |
|---|---|---|
| Reconciler create / update / delete / error handling | Controller test | `internal/controller/cloud-control/` or `cloud-resources/` |
| CRD field validation, allowed values, immutability | API validation test | `internal/api-tests/` |
| Complex pure function (enum conversion, schedule calculation, validation logic) | Unit test | Same package as code |
| Action with complex branching/fallback logic (multiple code paths, hard to cover via controller test) | Unit test with stubs | Same package as action |
| Full feature scenario on a real cluster | E2E test | `e2e/` |

When uncertain: **controller test** is the right default for reconciler behavior.

## Controller Tests

Framework: Ginkgo/Gomega BDD. Infrastructure: `pkg/testinfra`. Cloud APIs: always mocked, never real.

**Structure**:
```go
var _ = Describe("Feature: KCP GcpRedisCluster", func() {
    It("Scenario: Creates cluster successfully", func() {
        redis := &cloudcontrolv1beta1.GcpRedisCluster{}

        By("When resource is created", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(infra.Ctx(), infra.KCP().Client(), redis, WithName("test")).
                Should(Succeed())
        })

        By("When GCP reports ACTIVE", func() {
            infra.GcpMock().SetRedisClusterState(redis.Status.Id, "ACTIVE", "endpoint:6379")
        })

        By("Then resource becomes Ready", func() {
            Eventually(LoadAndCheck, "10s", "1s").
                WithArguments(infra.Ctx(), infra.KCP().Client(), redis,
                    NewObjActions(),
                    HavingState(cloudcontrolv1beta1.StateReady),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).Should(Succeed())
        })
    })
})
```

**Rules**:
- Always use `Eventually(LoadAndCheck)` — never `Expect()` on async state
- Configure mock BEFORE the assertion that depends on it
- Test create AND delete in the **same** `It` block — do not split into separate scenarios
- Delete all K8s resources at end of the scenario (cleanup)
- Each `It` block uses its own fresh resource and mock subscription/account — no shared state
- Never use `time.Sleep`

See `references/controller-test-patterns.md` for suite setup, deletion testing, error conditions, and feature flag skip guards.

## API Validation Tests

Location: `internal/api-tests/`. Naming: `kcp_<resource>_test.go` or `skr_<resource>_test.go`.

Two builder cases:
- If the API type already has a `<Type>Builder` in its package, use it directly (no local builder needed)
- Otherwise, define a local `test<Type>Builder` in the test file

Use the 7 helpers — pass a builder, never a raw object:

```go
// Valid creation
canCreateSkr("REGIONAL tier with valid capacity: 1024", builder.WithTier(REGIONAL).WithCapacityGb(1024))

// Invalid creation — expectedErrorMsg is required
canNotCreateSkr("capacity below minimum", builder.WithCapacityGb(100), "capacityGb must be at least 1024")

// Mutable field
canChangeSkr("capacity can be increased", builder, func(b Builder) { b.WithCapacityGb(2048) })

// Immutable field — expectedErrorMsg is required
canNotChangeSkr("tier is immutable", builder, func(b Builder) { b.WithTier(BASIC_SSD) }, "tier is immutable")

// KCP variants: canCreateKcp / canNotCreateKcp / canNotChangeKcp
```

For each validated field: test a valid value, an invalid value, boundary conditions (min/max), and immutability.

See `references/api-validation-patterns.md` for the builder pattern and complete examples.

## Unit Tests

Write for two cases:

**1. Complex pure functions** — enum/tier conversion, schedule calculation, capacity validation logic.

**2. Actions with complex branching or fallback logic** — when the action has multiple code paths that are impractical to exercise through controller tests (e.g., a fallback naming strategy, VPC ownership validation, conditional error classification). Use hand-written stubs — no testinfra, no Ginkgo.

**Do NOT write for**: simple pass-through actions, state factories, client wrappers, anything where setting up a stub is more work than just using a controller test.

Symptom you chose the wrong type: you need a real reconciler loop or testinfra to observe the behavior you care about → use a controller test instead.

**Pattern for action unit tests** (see `pkg/kcp/provider/gcp/iprange/v3/loadAddress_test.go`):
- Use `testing.T` + `testify/assert`, not Ginkgo/Gomega
- Write a minimal stub implementing only the client interface methods the action calls; `panic("unimplemented")` for unused methods
- Build state manually: `focal.NewStateFactory().NewState(composed.NewStateFactory(cluster).NewState(...))`
- Use `fake.NewClientBuilder().WithScheme(...).Build()` for the k8s client
- Group related cases under `t.Run("loadAddress", ...)`, with a shared `setupTest()` that resets all variables

Location: same package as code under test (`*_test.go` alongside the file).

## E2E Tests

Location: `e2e/`. Gherkin feature files. Runs against a real cluster. Write only when a full end-to-end scenario must be validated beyond what controller tests cover.

## Fake Clock (time-dependent reconcilers)

When a reconciler injects `clock.Clock`, tests use a fake clock to advance time without real waits:

```go
// Suite setup
var testFakeClock *clock.FakeClock

// BeforeSuite:
testFakeClock = clock.NewFakeClock(time.Now())
Expect(SetupGcpNfsBackupScheduleReconciler(infra.Registry(), env, testFakeClock)).NotTo(HaveOccurred())

// In test — advance time, then assert
By("When time advances past scheduled run", func() {
    testFakeClock.Step(2 * time.Minute)
})
```

`util.SetSpeedyTimingForTests()` shrinks requeue delays to milliseconds so tests don't wait for real intervals.

## Creating Provider Mocks

Each provider has a different mock architecture — they are **not** interchangeable:

- **GCP mock2** (`pkg/kcp/provider/gcp/mock2/`) — use for all **new** GCP resources. Subscription-based: each test scenario calls `infra.GcpMock2().NewSubscription("prefix")` and `defer gcpMock.Delete()`. State lives in the `Store` per subscription.
- **GCP mock** (`pkg/kcp/provider/gcp/mock/`) — old GCP resources only. Flat global state, direct mutation via `infra.GcpMock().*`.
- **AWS mock** (`pkg/kcp/provider/aws/mock/`) — Account model: `infra.AwsMock().NewAccount()` / `defer awsAccount.Delete()`.
- **SAP/OpenStack mock** (`pkg/kcp/provider/sap/mock/`) — Project model: `infra.SapMock().NewProject()`. No `Delete()` — random credentials provide isolation. Config methods (`AddNetwork`, `AddRouter`, `SetShareStatus`) seed pre-existing OpenStack resources.
- **Azure mock** (`pkg/kcp/provider/azure/mock/`) — flat Server with per-resource store interfaces.

See `references/mock-creation.md` for implementation details for each provider.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| `Expect(obj.Status.State)` on reconciled state | Use `Eventually(LoadAndCheck)` |
| Mock not configured → test timeouts waiting for Ready | Set mock state BEFORE assertion |
| Shared resource across `It` blocks | Create fresh resource per `It` |
| `canNotCreate*` call missing expected error message | Always provide expected error string |
| Unit test for an action that has no branching logic | Use controller test; stub-based unit tests are only worth it when the action has multiple non-trivial paths |
| No deletion test | Every reconciler scenario must delete and verify `IsDeleted` |
| Create and delete in separate `It` blocks | Put create, update, and delete in one `It` block |
| Forgetting `defer gcpMock.Delete()` / `defer awsAccount.Delete()` | Add defer immediately after creating the subscription/account |

## Reference Files

- `references/controller-test-patterns.md` — Suite setup, deletion testing, error conditions, SKR→KCP projection, feature flag skip guards, fake clock details
- `references/api-validation-patterns.md` — Builder pattern, all 7 helpers with complete examples, cross-field and boundary testing
- `references/mock-creation.md` — Dual-interface pattern, thread safety, state transitions, async operations mock
