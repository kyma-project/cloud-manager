---
name: write-tests
description: Write controller tests for Cloud Manager reconcilers using Ginkgo/Gomega. Use when adding tests, writing test scenarios, using Eventually() assertions, or testing with mocks.
---

# Write Controller Tests

Create BDD-style tests for Cloud Manager reconcilers using Ginkgo/Gomega with mocks.

## Quick Start

1. Create test file in `internal/controller/cloud-control/` or `cloud-resources/`
2. Use `testinfra` for test setup
3. Write scenarios with `Describe/It/By` structure
4. Use `Eventually()` for all reconciliation assertions
5. Mock cloud provider state transitions

## Rules

### MUST
- Use `Eventually()` for reconciliation assertions
- Use `LoadAndCheck()` helper pattern
- Mock cloud provider APIs (never call real APIs)
- Test both create and delete paths
- Test error conditions

### MUST NOT
- Use synchronous assertions for reconciled state
- Call real cloud provider APIs
- Skip error condition tests
- Use hardcoded timeouts without reason

## Test Structure

```go
var _ = Describe("Feature: GcpRedisCluster", func() {

    It("Scenario: Create and delete cluster", func() {
        scope := &cloudcontrolv1beta1.Scope{}
        cluster := &cloudcontrolv1beta1.GcpRedisCluster{}

        By("Given Scope exists", func() {
            Eventually(CreateScopeGcp).
                WithArguments(infra.Ctx(), infra, scope).
                Should(Succeed())
        })

        By("When GcpRedisCluster is created", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    WithScope(scope.Name),
                ).Should(Succeed())
        })

        By("Then cluster has ID assigned", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    NewObjActions(),
                    HavingStatusId(),
                ).Should(Succeed())
        })

        By("And cluster exists in mock", func() {
            Expect(infra.GcpMock().GetCluster(cluster.Status.Id)).NotTo(BeNil())
        })

        By("When GCP marks cluster ready", func() {
            infra.GcpMock().SetClusterState(cluster.Status.Id, "ACTIVE")
        })

        By("Then cluster has Ready condition", func() {
            Eventually(LoadAndCheck).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster,
                    NewObjActions(),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).WithTimeout(5 * time.Second).
                Should(Succeed())
        })

        // DELETE PATH

        By("When cluster is deleted", func() {
            Eventually(Delete).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster).
                Should(Succeed())
        })

        By("Then cluster is removed from GCP", func() {
            Eventually(func() bool {
                return infra.GcpMock().GetCluster(cluster.Status.Id) == nil
            }).Should(BeTrue())
        })

        By("And cluster resource is deleted", func() {
            Eventually(IsDeleted).
                WithArguments(infra.Ctx(), infra.KCP().Client(), cluster).
                Should(Succeed())
        })
    })
})
```

## Eventually Pattern

```go
// Basic pattern
Eventually(LoadAndCheck).
    WithArguments(ctx, client, obj, NewObjActions(), matcher).
    Should(Succeed())

// With timeout
Eventually(LoadAndCheck).
    WithArguments(ctx, client, obj, NewObjActions(), matcher).
    WithTimeout(10 * time.Second).
    WithPolling(200 * time.Millisecond).
    Should(Succeed())
```

## Common Matchers

| Matcher | Purpose |
|---------|---------|
| `HavingState(state)` | Check `.Status.State` |
| `HavingConditionTrue(type)` | Check condition is True |
| `HavingConditionFalse(type)` | Check condition is False |
| `HavingStatusId()` | Check `.Status.Id` is set |
| `HavingFinalizer(name)` | Check finalizer present |
| `HavingDeletionTimestamp()` | Check marked for deletion |

## Test Helpers

```go
// Create resource
Eventually(Create<Resource>).
    WithArguments(ctx, client, obj, WithOption(value)).
    Should(Succeed())

// Update resource
Eventually(Update).
    WithArguments(ctx, client, obj, func(o *Type) {
        o.Spec.Field = newValue
    }).Should(Succeed())

// Delete resource
Eventually(Delete).
    WithArguments(ctx, client, obj).
    Should(Succeed())

// Check deleted
Eventually(IsDeleted).
    WithArguments(ctx, client, obj).
    Should(Succeed())
```

## Mock State Transitions

```go
// Set mock state (triggers reconciler)
By("When GCP marks resource as READY", func() {
    infra.GcpMock().SetResourceState(id, "READY")
})

// Verify mock state
By("Then resource exists in mock", func() {
    resource := infra.GcpMock().GetResource(id)
    Expect(resource).NotTo(BeNil())
    Expect(resource.Status).To(Equal("READY"))
})

// Simulate error
By("When GCP returns error", func() {
    infra.GcpMock().SetError(id, errors.New("API error"))
})
```

## Test File Location

| Resource Type | Location |
|---------------|----------|
| KCP resources | `internal/controller/cloud-control/<resource>_test.go` |
| SKR resources | `internal/controller/cloud-resources/<resource>_test.go` |
| Provider-specific | `<resource>_gcp_test.go`, `<resource>_azure_test.go` |

## Running Tests

```bash
# All tests
make test

# Specific test file
go test ./internal/controller/cloud-control -v -ginkgo.focus="GcpRedisCluster"

# Single scenario
go test ./internal/controller/cloud-control -v \
    -ginkgo.focus="GcpRedisCluster" \
    -ginkgo.focus="Create and delete"

# Verbose
go test ./internal/controller/cloud-control -v -ginkgo.v
```

## Checklist

- [ ] Test file in correct location
- [ ] Uses testinfra framework
- [ ] Describe/It/By structure
- [ ] Eventually() for all reconciliation checks
- [ ] Create path tested
- [ ] Delete path tested
- [ ] Error conditions tested
- [ ] Mock state transitions used
- [ ] Tests pass with `make test`

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Test timeout | Increase Eventually timeout, check mock state |
| Resource stuck | Set mock state to trigger transition |
| Assertion fails immediately | Wrap in Eventually() |
| Context canceled | Increase test timeout |

## Related

- Full guide: [docs/agents/guides/CONTROLLER_TESTS.md](../../../docs/agents/guides/CONTROLLER_TESTS.md)
- Creating mocks: `/create-mocks`
- Quick reference: [docs/agents/reference/QUICK_REFERENCE.md](../../../docs/agents/reference/QUICK_REFERENCE.md)
