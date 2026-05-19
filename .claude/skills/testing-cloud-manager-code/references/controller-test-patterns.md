# Controller Test Patterns

## Suite Setup

Location: `internal/controller/cloud-control/suite_test.go` (or `cloud-resources/`)

```go
package cloudcontrol

import (
    "context"
    "testing"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/kyma-project/cloud-manager/pkg/testinfra"
)

var infra testinfra.Infra

func TestControllers(t *testing.T) {
    if len(os.Getenv("PROJECTROOT")) == 0 {
        t.Skip("Skipping — set PROJECTROOT to the Makefile directory. See `make test`.")
        return
    }
    RegisterFailHandler(Fail)
    RunSpecs(t, "KCP Controller Suite")
}

var _ = BeforeSuite(func() {
    var err error
    infra, err = testinfra.Start()
    Expect(err).NotTo(HaveOccurred(), "failed starting infra clusters")

    // Create namespaces needed for tests
    Expect(infra.KCP().GivenNamespaceExists(infra.KCP().Namespace())).NotTo(HaveOccurred())
    Expect(infra.SKR().GivenNamespaceExists(infra.SKR().Namespace())).NotTo(HaveOccurred())

    env := abstractions.NewMockedEnvironment(map[string]string{})

    // Register reconcilers — pass mock providers from infra.GcpMock2(), infra.AwsMock(), infra.AzureMock()
    Expect(SetupMyReconciler(
        infra.KcpManager(),
        infra.GcpMock2().NfsInstanceV2Provider(),
        env,
    )).NotTo(HaveOccurred())

    // Start all registered controllers
    infra.StartKcpControllers(context.Background())
})

var _ = AfterSuite(func() {
    err := testinfra.PrintMetrics()
    Expect(err).NotTo(HaveOccurred())
    Expect(infra.Stop()).NotTo(HaveOccurred())
})

var _ = ReportAfterSuite("gherkin report", testinfra.ReportAfterSuite)
```

## Scenario Structure Rule

**Create, update, and delete in one `It` block.** Do not split into separate scenarios. At the end of every scenario, delete all K8s resources created during the test — both the main resource and any prerequisites (Scope, IpRange, etc.) that the test created and owns.

This provides isolation: each scenario leaves no residue that can interfere with the next.

## GcpMock2 Subscription Pattern (new GCP resources)

`mock2` uses subscriptions for isolation. Each scenario creates its own subscription (= isolated GCP project) and defers deletion.

```go
It("Scenario: KCP GCP NfsInstance v2 is created, updated and deleted", func() {

    gcpMock := infra.GcpMock2().NewSubscription("nfs-instance-v2")
    defer gcpMock.Delete()  // cleans up all GCP-side resources for this scenario

    scope := &cloudcontrolv1beta1.Scope{}

    By("Given Scope exists", func() {
        kcpscope.Ignore.AddName(kymaName)
        Eventually(CreateScopeGcp2).
            WithArguments(infra.Ctx(), infra, scope, gcpMock.ProjectId(), WithName(kymaName)).
            Should(Succeed())
    })

    // ... create resource ...

    By("When NfsInstance is created", func() {
        Eventually(CreateObj).
            WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance, ...).
            Should(Succeed())
    })

    // For async GCP operations (long-running ops), resolve them manually:
    By("When GCP operation is resolved", func() {
        Eventually(func() error {
            it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
            for op, err := it.Next(); err == nil; op, err = it.Next() {
                if !op.Done && op.Name != "" {
                    return gcpMock.ResolveFilestoreOperation(infra.Ctx(), op.Name)
                }
            }
            return fmt.Errorf("no pending operation found yet")
        }).Should(Succeed())
    })

    By("Then NfsInstance has Ready condition", func() {
        Eventually(LoadAndCheck).
            WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
                NewObjActions(),
                HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                HavingState("Ready"),
            ).Should(Succeed())
    })

    // DELETE at end of scenario
    By("When NfsInstance is deleted", func() {
        Eventually(Delete).
            WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
            Should(Succeed())
    })

    By("And When delete operation is resolved", func() {
        Eventually(func() error {
            it := gcpMock.ListFilestoreOperations(infra.Ctx(), &longrunningpb.ListOperationsRequest{})
            for op, err := it.Next(); err == nil; op, err = it.Next() {
                if !op.Done {
                    return gcpMock.ResolveFilestoreOperation(infra.Ctx(), op.Name)
                }
            }
            return fmt.Errorf("no pending delete operation found yet")
        }).Should(Succeed())
    })

    By("Then NfsInstance is gone", func() {
        Eventually(IsDeleted, 5*time.Second).
            WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
            Should(Succeed())
    })
})
```

`gcpMock.ProjectId()` provides the isolated project ID for this subscription. Use it when creating GCP resources directly on the mock (network, address, etc.).

## AWS Account Pattern

AWS uses `Account` for isolation instead of subscriptions:

```go
It("Scenario: KCP AWS RedisInstance is created and deleted", func() {

    awsAccount := infra.AwsMock().NewAccount()
    defer awsAccount.Delete()  // cleans up all AWS-side resources

    scope := &cloudcontrolv1beta1.Scope{}
    By("Given Scope exists", func() {
        kcpscope.Ignore.AddName(name)
        Eventually(CreateScopeAws).
            WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
            Should(Succeed())
    })

    // ... create, assert, delete ...

    By("When RedisInstance is deleted", func() {
        Eventually(Delete).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
            Should(Succeed())
    })

    By("Then RedisInstance is gone", func() {
        Eventually(IsDeleted, 5*time.Second).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
            Should(Succeed())
    })
})
```

## GcpMock (old resources — mock/)

For resources that still use the old `mock/` (VpcPeering, old NfsInstance v1, etc.), state is manipulated directly — no subscription pattern:

```go
By("When GCP reports ACTIVE", func() {
    infra.GcpMock().SetRedisClusterState(redis.Status.Id, "ACTIVE", "endpoint:6379")
})
```

Check `suite_test.go` to see which mock is passed to which reconciler — that determines which to use in the test.

## SKR→KCP Projection Test Pattern

```go
It("Scenario: SKR resource creates and syncs with KCP", func() {
    skrRedis := &cloudresourcesv1beta1.GcpRedisCluster{}

    By("When SKR resource is created", func() {
        Eventually(CreateGcpRedisCluster).
            WithArguments(infra.Ctx(), infra.SKR().Client(), skrRedis, WithName("test")).
            Should(Succeed())
    })

    By("Then KCP resource is created with correct labels", func() {
        var kcpRedis *cloudcontrolv1beta1.GcpRedisCluster
        Eventually(LoadAndCheck, "10s", "1s").
            WithArguments(infra.Ctx(), infra.SKR().Client(), skrRedis,
                NewObjActions(func() {
                    kcpRedis = &cloudcontrolv1beta1.GcpRedisCluster{}
                    err := infra.KCP().Client().Get(infra.Ctx(),
                        types.NamespacedName{Namespace: infra.KCP().Namespace(), Name: skrRedis.Status.Id},
                        kcpRedis)
                    Expect(err).NotTo(HaveOccurred())
                    Expect(kcpRedis.Labels[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(skrRedis.Name))
                }),
            ).Should(Succeed())
    })

    // ... update KCP status, assert SKR syncs, then delete at end ...

    By("When SKR resource is deleted", func() {
        Eventually(Delete).
            WithArguments(infra.Ctx(), infra.SKR().Client(), skrRedis).
            Should(Succeed())
    })
    By("Then SKR resource is gone", func() {
        Eventually(IsDeleted, 5*time.Second).
            WithArguments(infra.Ctx(), infra.SKR().Client(), skrRedis).
            Should(Succeed())
    })
})
```

## Feature Flag v1/v2 Skip Guards

```go
// v1 test file
BeforeEach(func() {
    if feature.BackupScheduleV2.Value(context.Background()) {
        Skip("Skipping v1 tests: backupScheduleV2 flag is enabled")
    }
})

// v2 test file
BeforeEach(func() {
    if !feature.BackupScheduleV2.Value(context.Background()) {
        Skip("Skipping v2 tests: backupScheduleV2 flag is disabled")
    }
})
```

CI matrix tests both: flag=false (v1 runs, v2 skips) and flag=true (v2 runs, v1 skips).

## Common Assertion Helpers

```go
HavingState(cloudcontrolv1beta1.StateReady)
HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)
HavingDeletionTimestamp()
HavingFinalizer(cloudcontrolv1beta1.FinalizerName)
HavingFieldSet("status", "id")          // field is non-zero
HavingFieldValue(val, "status", "field") // field equals value

UpdateStatus(ctx, client, obj,
    WithConditions(KcpReadyCondition()),
    WithKcpIpRangeStatusCidr(cidr),
)

Delete(ctx, client, obj)
IsDeleted(ctx, client, obj)
```

## Fake Clock Pattern

```go
// suite_test.go
var testFakeClock *clock.FakeClock

// BeforeSuite:
testFakeClock = clock.NewFakeClock(time.Now())
Expect(SetupGcpNfsBackupScheduleReconciler(infra.Registry(), env, testFakeClock)).NotTo(HaveOccurred())
// util.SetSpeedyTimingForTests() shrinks requeue delays to ~1ms

// In test:
By("When time advances past scheduled run", func() {
    testFakeClock.Step(2 * time.Minute)
})
```

## Running Tests

```bash
make test
go test ./internal/controller/cloud-control -v
go test ./internal/controller/cloud-resources -v
go test ./internal/controller/cloud-control -ginkgo.focus="NfsInstance"
```
