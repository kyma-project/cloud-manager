# Guide: Writing Controller Tests

**Target Audience**: LLM coding agents  
**Prerequisites**: [ADD_KCP_RECONCILER.md](ADD_KCP_RECONCILER.md), [ADD_SKR_RECONCILER.md](ADD_SKR_RECONCILER.md)  
**Purpose**: Step-by-step guide for writing Ginkgo/Gomega tests with mocked cloud providers  
**Context**: Cloud Manager uses testinfra framework for BDD testing with GCP/Azure/AWS mocks

## Authority: Controller Testing Requirements

### MUST

- **MUST use Ginkgo/Gomega**: BDD testing framework REQUIRED, no other test frameworks allowed
- **MUST use testinfra**: Use `pkg/testinfra` for test infrastructure and mocks
- **MUST test create AND delete**: Both create and delete scenarios REQUIRED for all tests
- **MUST use Eventually**: Use `Eventually()` for async operations, NEVER `Expect()` directly on changing state
- **MUST mock cloud providers**: Use GcpMock/AzureMock/AwsMock, NEVER call real cloud APIs
- **MUST use LoadAndCheck**: Use `Eventually(LoadAndCheck).WithArguments(...)` pattern for assertions
- **MUST test error conditions**: Test API failures, not found errors, conflicts
- **MUST use descriptive names**: Test names use "Feature:" and "Scenario:" format

### MUST NOT

- **NEVER call real cloud APIs**: All cloud provider calls MUST be mocked
- **NEVER use time.Sleep for waiting**: Use `Eventually()` with appropriate timeout, not hardcoded sleeps
- **NEVER test implementation details**: Test observable behavior (status, conditions), not internal state
- **NEVER hardcode namespaces**: Use `infra.Namespace()` or similar helpers
- **NEVER skip cleanup**: Use `DeferCleanup()` or BeforeEach/AfterEach for resource cleanup

### ALWAYS

- **ALWAYS use By() blocks**: Structure tests with `By("description", func() { ... })` for readability
- **ALWAYS use Eventually timeout**: Specify timeout like `Eventually(action, "10s", "1s")` for clarity
- **ALWAYS check errors**: Use `Should(Succeed())` on function returns, check error handling
- **ALWAYS test status sync**: Verify status fields updated correctly from cloud provider state

### NEVER

- **NEVER rely on test execution order**: Each test MUST be independent
- **NEVER mutate shared state**: Use fresh resources per test or BeforeEach setup
- **NEVER ignore mock setup**: Configure mock responses BEFORE testing behavior

## Test Structure: Ginkgo BDD Pattern

### ❌ WRONG: No Structure or Async Handling

```go
// NEVER: No Eventually, hardcoded waits, no By() blocks
func TestRedis(t *testing.T) {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    redis.Name = "test"
    client.Create(context.Background(), redis)  // No error check!
    
    time.Sleep(5 * time.Second)  // WRONG: Hardcoded wait
    
    client.Get(context.Background(), name, redis)  // No Eventually!
    if redis.Status.State != "Ready" {  // WRONG: Race condition
        t.Fatal("not ready")
    }
}
```

### ✅ CORRECT: Ginkgo BDD with Eventually

```go
// ALWAYS: Use Ginkgo Describe/It, By() blocks, Eventually
var _ = Describe("Feature: KCP GcpRedisCluster", func() {

    It("Scenario: Creates cluster successfully", func() {
        redis := &cloudcontrolv1beta1.GcpRedisCluster{}
        
        By("When GcpRedisCluster is created", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redis,
                    WithName("test-cluster"),
                    WithShardCount(3),
                ).
                Should(Succeed())  // Check creation success
        })
        
        By("Then resource becomes Ready", func() {
            Eventually(LoadAndCheck, "10s", "1s").  // Timeout + polling
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redis,
                    NewObjActions(),
                    HavingState(cloudcontrolv1beta1.StateReady),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                ).
                Should(Succeed())
        })
    })
})
```

## Test Suite Setup

### ❌ WRONG: No Envtest or Mock Setup

```go
// NEVER: No test infrastructure setup
func TestMain(m *testing.M) {
    os.Exit(m.Run())  // WRONG: No envtest, no mocks
}
```

### ✅ CORRECT: Complete Suite Setup with Mocks

**Location**: `internal/controller/cloud-control/suite_test.go`

```go
// ALWAYS: Setup envtest + mocks in suite
package cloudcontrol

import (
    "testing"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
    "github.com/kyma-project/cloud-manager/pkg/testinfra"
)

func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "KCP Controller Suite")
}

var _ = BeforeSuite(func() {
    // Setup envtest environment
    By("bootstrapping test environment")
    
    // Initialize test infrastructure
    infra = testinfra.Start()
    
    // Register CRDs
    infra.AddCRDs(
        cloudcontrolv1beta1.SchemeBuilder.SchemeBuilder,
    )
    
    // Setup mock providers
    infra.SetupGcpMock()
    infra.SetupAzureMock()
    infra.SetupAwsMock()
    
    // Start reconcilers with mocks
    infra.StartReconcilers(
        SetupGcpRedisClusterReconciler,
        SetupGcpSubnetReconciler,
        // ... other reconcilers
    )
})

var _ = AfterSuite(func() {
    By("tearing down test environment")
    infra.Stop()
})
```

## Testing KCP Reconcilers

### ❌ WRONG: Not Using Mocks

```go
// NEVER: No mock setup before testing
It("creates GCP Redis", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    Eventually(CreateGcpRedisCluster).Should(Succeed())
    
    // WRONG: No mock configured, reconciler will fail or timeout
    Eventually(LoadAndCheck).
        WithArguments(infra.Ctx(), infra.KCP().Client(), redis, 
            NewObjActions(), HavingState("Ready")).
        Should(Succeed())
})
```

### ✅ CORRECT: Configure Mock Then Test

```go
// ALWAYS: Setup mock BEFORE testing behavior
var _ = Describe("Feature: KCP GcpRedisCluster", func() {

    It("Scenario: Creates Redis cluster in GCP", func() {
        scope := &cloudcontrolv1beta1.Scope{}
        redis := &cloudcontrolv1beta1.GcpRedisCluster{}
        
        By("Given Scope exists", func() {
            Eventually(CreateScopeGcp).
                WithArguments(infra.Ctx(), infra, scope).
                Should(Succeed())
        })
        
        By("When GcpRedisCluster is created", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redis,
                    WithName("test-cluster"),
                    WithScope(scope.Name),
                    WithShardCount(3),
                    WithReplicaCount(2),
                ).
                Should(Succeed())
        })
        
        By("Then GCP mock receives cluster creation", func() {
            // MUST: Verify mock called
            Eventually(func() bool {
                cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
                return cluster != nil && cluster.ShardCount == 3
            }, "5s").Should(BeTrue())
        })
        
        By("When GCP completes provisioning", func() {
            // MUST: Configure mock response
            infra.GcpMock().SetRedisClusterState(
                redis.Status.Id,
                "ACTIVE",
                "redis.googleapis.com/endpoint:6379",
            )
        })
        
        By("Then resource becomes Ready", func() {
            Eventually(LoadAndCheck, "10s", "1s").
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    redis,
                    NewObjActions(),
                    HavingState(cloudcontrolv1beta1.StateReady),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                    func(obj *cloudcontrolv1beta1.GcpRedisCluster) {
                        Expect(obj.Status.PrimaryEndpoint).To(ContainSubstring("redis.googleapis.com"))
                    },
                ).
                Should(Succeed())
        })
    })
})
```

## Testing SKR Reconcilers

### ❌ WRONG: Not Verifying KCP Resource Creation

```go
// NEVER: Only testing SKR resource, ignoring KCP
It("creates SKR resource", func() {
    redisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
    Eventually(CreateGcpRedisCluster).Should(Succeed())
    
    // WRONG: Not checking if KCP resource created!
    Eventually(LoadAndCheck).
        WithArguments(infra.Ctx(), infra.SKR().Client(), redisCluster,
            NewObjActions(), HavingState("Ready")).
        Should(Succeed())
})
```

### ✅ CORRECT: Test SKR → KCP Projection

```go
// ALWAYS: Verify SKR creates KCP resource with correct mapping
var _ = Describe("Feature: SKR GcpRedisCluster", func() {

    It("Scenario: SKR resource creates and syncs with KCP", func() {
        redisCluster := &cloudresourcesv1beta1.GcpRedisCluster{}
        var kcpRedisCluster *cloudcontrolv1beta1.GcpRedisCluster
        
        By("When GcpRedisCluster is created in SKR", func() {
            Eventually(CreateGcpRedisCluster).
                WithArguments(
                    infra.Ctx(),
                    infra.SKR().Client(),
                    redisCluster,
                    WithName("test-cluster"),
                    WithShardCount(3),
                ).
                Should(Succeed())
        })
        
        By("Then ID is generated in SKR status", func() {
            Eventually(LoadAndCheck, "5s", "500ms").
                WithArguments(
                    infra.Ctx(),
                    infra.SKR().Client(),
                    redisCluster,
                    NewObjActions(),
                    func(obj *cloudresourcesv1beta1.GcpRedisCluster) {
                        Expect(obj.Status.Id).NotTo(BeEmpty())
                    },
                ).
                Should(Succeed())
        })
        
        By("Then KCP GcpRedisCluster is created", func() {
            Eventually(LoadAndCheck, "10s", "1s").
                WithArguments(
                    infra.Ctx(),
                    infra.SKR().Client(),
                    redisCluster,
                    NewObjActions(
                        func() {
                            // MUST: Load corresponding KCP resource
                            kcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{}
                            err := infra.KCP().Client().Get(
                                infra.Ctx(),
                                types.NamespacedName{
                                    Namespace: infra.KCP().Namespace(),
                                    Name:      redisCluster.Status.Id,
                                },
                                kcpRedisCluster,
                            )
                            Expect(err).NotTo(HaveOccurred())
                            
                            // MUST: Verify spec mapping
                            Expect(kcpRedisCluster.Spec.ShardCount).To(Equal(int32(3)))
                            
                            // MUST: Verify annotations
                            Expect(kcpRedisCluster.Annotations).To(HaveKey(cloudcontrolv1beta1.LabelRemoteName))
                            Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(redisCluster.Name))
                        },
                    ),
                ).
                Should(Succeed())
        })
        
        By("When KCP resource becomes Ready", func() {
            Eventually(UpdateStatus).
                WithArguments(
                    infra.Ctx(),
                    infra.KCP().Client(),
                    kcpRedisCluster,
                    WithState(cloudcontrolv1beta1.StateReady),
                    WithCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionTrue),
                    WithPrimaryEndpoint("redis.googleapis.com:6379"),
                ).
                Should(Succeed())
        })
        
        By("Then SKR status is synced from KCP", func() {
            Eventually(LoadAndCheck, "10s", "1s").
                WithArguments(
                    infra.Ctx(),
                    infra.SKR().Client(),
                    redisCluster,
                    NewObjActions(),
                    HavingState(cloudcontrolv1beta1.StateReady),
                    HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
                    func(obj *cloudresourcesv1beta1.GcpRedisCluster) {
                        // MUST: Verify user-relevant fields synced
                        Expect(obj.Status.PrimaryEndpoint).To(Equal("redis.googleapis.com:6379"))
                    },
                ).
                Should(Succeed())
        })
    })
})
```

## Testing Deletion

### ❌ WRONG: Not Waiting for Cloud Resource Deletion

```go
// NEVER: Only deleting Kubernetes resource without checking cloud cleanup
It("deletes resource", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    Eventually(CreateGcpRedisCluster).Should(Succeed())
    
    client.Delete(infra.Ctx(), redis)  // WRONG: Not checking finalizer or cloud deletion
    
    Eventually(func() bool {
        err := client.Get(infra.Ctx(), name, redis)
        return errors.IsNotFound(err)  // WRONG: Might succeed before cloud cleanup!
    }).Should(BeTrue())
})
```

### ✅ CORRECT: Verify Finalizer and Cloud Cleanup

```go
// ALWAYS: Test finalizer, cloud deletion, then Kubernetes deletion
It("Scenario: Deletes cluster and cleans up cloud resources", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    
    By("Given Redis cluster exists and is Ready", func() {
        Eventually(CreateGcpRedisCluster).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis, WithName("test")).
            Should(Succeed())
        
        infra.GcpMock().SetRedisClusterState(redis.Status.Id, "ACTIVE", "endpoint")
        
        Eventually(LoadAndCheck).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis,
                NewObjActions(), HavingState(cloudcontrolv1beta1.StateReady)).
            Should(Succeed())
    })
    
    By("When Redis cluster is deleted", func() {
        Eventually(Delete).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis).
            Should(Succeed())
    })
    
    By("Then resource has deletion timestamp", func() {
        Eventually(LoadAndCheck).
            WithArguments(
                infra.Ctx(),
                infra.KCP().Client(),
                redis,
                NewObjActions(),
                func(obj *cloudcontrolv1beta1.GcpRedisCluster) {
                    Expect(obj.DeletionTimestamp).NotTo(BeNil())
                },
            ).
            Should(Succeed())
    })
    
    By("Then GCP receives delete request", func() {
        Eventually(func() bool {
            cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
            return cluster == nil || cluster.State == "DELETING"
        }, "5s").Should(BeTrue())
    })
    
    By("When GCP completes deletion", func() {
        infra.GcpMock().DeleteRedisCluster(redis.Status.Id)
    })
    
    By("Then Kubernetes resource is deleted", func() {
        Eventually(IsDeleted, "15s", "1s").
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis).
            Should(Succeed())
    })
    
    By("Then cloud resource no longer exists", func() {
        cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
        Expect(cluster).To(BeNil())
    })
})
```

## Testing Error Conditions

### ❌ WRONG: Only Testing Happy Path

```go
// NEVER: Only testing successful scenarios
It("creates resource", func() {
    Eventually(CreateGcpRedisCluster).Should(Succeed())
    Eventually(LoadAndCheck).
        WithArguments(infra.Ctx(), infra.KCP().Client(), redis,
            NewObjActions(), HavingState("Ready")).
        Should(Succeed())
    // WRONG: No error condition testing!
})
```

### ✅ CORRECT: Test Error Handling

```go
// ALWAYS: Test API failures, not found, conflicts
It("Scenario: Handles GCP provisioning failure", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    
    By("When GcpRedisCluster is created", func() {
        Eventually(CreateGcpRedisCluster).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis, WithName("test")).
            Should(Succeed())
    })
    
    By("When GCP provisioning fails", func() {
        infra.GcpMock().SetRedisClusterState(
            redis.Status.Id,
            "FAILED",
            "",
        )
        infra.GcpMock().SetRedisClusterError(
            redis.Status.Id,
            "RESOURCE_EXHAUSTED: No capacity available",
        )
    })
    
    By("Then resource has Error condition", func() {
        Eventually(LoadAndCheck, "10s", "1s").
            WithArguments(
                infra.Ctx(),
                infra.KCP().Client(),
                redis,
                NewObjActions(),
                HavingState(cloudcontrolv1beta1.StateError),
                HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError),
                func(obj *cloudcontrolv1beta1.GcpRedisCluster) {
                    // MUST: Verify error message propagated
                    cond := meta.FindStatusCondition(obj.Status.Conditions, cloudcontrolv1beta1.ConditionTypeError)
                    Expect(cond).NotTo(BeNil())
                    Expect(cond.Message).To(ContainSubstring("RESOURCE_EXHAUSTED"))
                },
            ).
            Should(Succeed())
    })
})

It("Scenario: Handles resource not found in cloud provider", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    
    By("Given Redis cluster created but not in GCP", func() {
        Eventually(CreateGcpRedisCluster).
            WithArguments(infra.Ctx(), infra.KCP().Client(), redis, WithName("test")).
            Should(Succeed())
        
        // Simulate cluster deleted externally
        infra.GcpMock().DeleteRedisCluster(redis.Status.Id)
    })
    
    By("Then reconciler recreates resource", func() {
        Eventually(func() bool {
            cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
            return cluster != nil
        }, "10s", "1s").Should(BeTrue())
    })
})
```

## Test Helpers and DSL

### testinfra DSL Functions

#### Resource Creation Helpers

```go
// KCP resource creators
CreateScopeGcp(ctx, infra, scope, ...opts)
CreateGcpRedisCluster(ctx, client, redis, ...opts)
CreateGcpSubnet(ctx, client, subnet, ...opts)

// SKR resource creators  
CreateGcpRedisCluster(ctx, client, redisCluster, ...opts)

// Option pattern for configuration
WithName("test-name")
WithScope("scope-name")
WithShardCount(3)
WithReplicaCount(2)
```

#### Assertion Helpers

```go
// LoadAndCheck pattern
Eventually(LoadAndCheck).
    WithArguments(
        ctx,
        client,
        resource,
        NewObjActions(),           // Empty actions list
        HavingState("Ready"),       // State matcher
        HavingConditionTrue("Ready"), // Condition matcher
        func(obj *Type) {          // Custom assertion
            Expect(obj.Spec.Field).To(Equal(value))
        },
    ).
    Should(Succeed())

// Status update helper
UpdateStatus(ctx, client, resource,
    WithState("Ready"),
    WithCondition("Ready", metav1.ConditionTrue),
    WithPrimaryEndpoint("endpoint:6379"),
)

// Deletion helper
Delete(ctx, client, resource)
IsDeleted(ctx, client, resource)
```

#### Mock Configuration

```go
// GCP Mock
infra.GcpMock().SetRedisClusterState(id, "ACTIVE", endpoint)
infra.GcpMock().SetRedisClusterError(id, errorMessage)
infra.GcpMock().DeleteRedisCluster(id)
cluster := infra.GcpMock().GetRedisClusterByName(id)

// Azure Mock
infra.AzureMock().SetRedisEnterpriseState(id, "Running")
infra.AzureMock().SetRedisEnterpriseError(id, errorMessage)

// AWS Mock
infra.AwsMock().SetVpcState(id, "available")
```

## Common Test Patterns

### Pattern: Wait for Condition

```go
Eventually(LoadAndCheck, "10s", "1s").
    WithArguments(
        infra.Ctx(),
        infra.KCP().Client(),
        resource,
        NewObjActions(),
        HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
    ).
    Should(Succeed())
```

### Pattern: Check Multiple Fields

```go
Eventually(LoadAndCheck).
    WithArguments(
        infra.Ctx(),
        infra.KCP().Client(),
        resource,
        NewObjActions(),
        HavingState(cloudcontrolv1beta1.StateReady),
        HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
        func(obj *cloudcontrolv1beta1.GcpRedisCluster) {
            Expect(obj.Spec.ShardCount).To(Equal(int32(3)))
            Expect(obj.Status.PrimaryEndpoint).NotTo(BeEmpty())
            Expect(obj.Status.Id).NotTo(BeEmpty())
        },
    ).
    Should(Succeed())
```

### Pattern: Verify Mock Called

```go
By("Then GCP receives create request", func() {
    Eventually(func() bool {
        cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
        return cluster != nil && cluster.ShardCount == 3
    }, "5s", "500ms").Should(BeTrue())
})
```

### Pattern: Setup Common Resources

```go
var scope *cloudcontrolv1beta1.Scope

BeforeEach(func() {
    scope = &cloudcontrolv1beta1.Scope{}
    Eventually(CreateScopeGcp).
        WithArguments(infra.Ctx(), infra, scope).
        Should(Succeed())
})

It("test case 1", func() {
    redis := &cloudcontrolv1beta1.GcpRedisCluster{}
    Eventually(CreateGcpRedisCluster).
        WithArguments(infra.Ctx(), infra.KCP().Client(), redis,
            WithScope(scope.Name)).  // Use common scope
        Should(Succeed())
})
```

## Common Pitfalls

### Pitfall 1: Using Expect() Instead of Eventually()

**Symptom**: Tests flaky or fail with "condition not met" on first check

**Cause**:
```go
// WRONG: Immediate expectation on async operation
Expect(redis.Status.State).To(Equal("Ready"))  // Race condition!
```

**Solution**: ALWAYS use Eventually for async state:
```go
// CORRECT: Eventually for async operations
Eventually(LoadAndCheck).
    WithArguments(ctx, client, redis, NewObjActions(), 
        HavingState("Ready")).
    Should(Succeed())
```

### Pitfall 2: Not Configuring Mocks

**Symptom**: Tests timeout waiting for Ready state

**Cause**:
```go
// WRONG: No mock configuration
Eventually(CreateGcpRedisCluster).Should(Succeed())
Eventually(LoadAndCheck).  // WRONG: Mock never set to ACTIVE
    WithArguments(ctx, client, redis, NewObjActions(), HavingState("Ready")).
    Should(Succeed())  // Will timeout!
```

**Solution**: Configure mock state transitions:
```go
// CORRECT: Configure mock after creation
Eventually(CreateGcpRedisCluster).Should(Succeed())

// Set mock state
infra.GcpMock().SetRedisClusterState(redis.Status.Id, "ACTIVE", endpoint)

// Then check status
Eventually(LoadAndCheck).
    WithArguments(ctx, client, redis, NewObjActions(), HavingState("Ready")).
    Should(Succeed())
```

### Pitfall 3: Tests Not Independent

**Symptom**: Tests pass individually but fail when run together

**Cause**:
```go
// WRONG: Reusing resources across tests
var redis *cloudcontrolv1beta1.GcpRedisCluster

BeforeEach(func() {
    // Creates once, all tests share same resource
    Eventually(CreateGcpRedisCluster).Should(Succeed())
})

It("test 1", func() {
    // Modifies shared resource
})

It("test 2", func() {
    // Affected by test 1's modifications!
})
```

**Solution**: Create fresh resources per test:
```go
// CORRECT: Fresh resource per test
It("test 1", func() {
    redis1 := &cloudcontrolv1beta1.GcpRedisCluster{}
    Eventually(CreateGcpRedisCluster).
        WithArguments(ctx, client, redis1, WithName("redis1")).
        Should(Succeed())
})

It("test 2", func() {
    redis2 := &cloudcontrolv1beta1.GcpRedisCluster{}  // Independent resource
    Eventually(CreateGcpRedisCluster).
        WithArguments(ctx, client, redis2, WithName("redis2")).
        Should(Succeed())
})
```

### Pitfall 4: Forgetting Status.Id for SKR Tests

**Symptom**: SKR test fails with "KCP resource not found"

**Cause**:
```go
// WRONG: Trying to load KCP without waiting for ID generation
kcpRedisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}
err := infra.KCP().Client().Get(ctx,
    types.NamespacedName{
        Namespace: infra.KCP().Namespace(),
        Name:      redisCluster.Status.Id,  // Empty! Not generated yet
    },
    kcpRedisCluster)
```

**Solution**: Wait for ID generation first:
```go
// CORRECT: Wait for ID before loading KCP
By("Then ID is generated", func() {
    Eventually(LoadAndCheck).
        WithArguments(ctx, infra.SKR().Client(), redisCluster,
            NewObjActions(),
            func(obj *cloudresourcesv1beta1.GcpRedisCluster) {
                Expect(obj.Status.Id).NotTo(BeEmpty())
            }).
        Should(Succeed())
})

By("Then KCP resource created", func() {
    Eventually(LoadAndCheck).
        WithArguments(ctx, infra.SKR().Client(), redisCluster,
            NewObjActions(func() {
                kcpRedisCluster = &cloudcontrolv1beta1.GcpRedisCluster{}
                err := infra.KCP().Client().Get(ctx,
                    types.NamespacedName{
                        Namespace: infra.KCP().Namespace(),
                        Name:      redisCluster.Status.Id,  // Now has value
                    },
                    kcpRedisCluster)
                Expect(err).NotTo(HaveOccurred())
            })).
        Should(Succeed())
})
```

## Running Tests

### Run All Tests

```bash
make test  # Runs all controller tests
```

### Run Specific Suite

```bash
go test ./internal/controller/cloud-control -v        # KCP tests
go test ./internal/controller/cloud-resources -v      # SKR tests
```

### Run Specific Test

```bash
# By test name
go test ./internal/controller/cloud-control -run TestControllers

# Focus with Ginkgo
go test ./internal/controller/cloud-control -ginkgo.focus="GcpRedisCluster"
```

### Verbose Output

```bash
# Verbose Ginkgo
go test ./internal/controller/cloud-control -v -ginkgo.v

# Show all steps
go test ./internal/controller/cloud-control -v -ginkgo.v -ginkgo.trace
```

## Debugging Tests

### Print Debug Info

```go
By("Debug: Checking resource state", func() {
    fmt.Fprintf(GinkgoWriter, "Resource: %+v\n", redis)
    fmt.Fprintf(GinkgoWriter, "Status: %+v\n", redis.Status)
    fmt.Fprintf(GinkgoWriter, "Conditions: %+v\n", redis.Status.Conditions)
})
```

### Check Mock State

```go
By("Debug: Checking mock state", func() {
    cluster := infra.GcpMock().GetRedisClusterByName(redis.Status.Id)
    fmt.Fprintf(GinkgoWriter, "GCP cluster: %+v\n", cluster)
})
```

### Increase Timeouts

```go
// Longer timeout for debugging
Eventually(LoadAndCheck, "30s", "1s").  // 30s timeout instead of default
    WithArguments(...).
    Should(Succeed())
```

## Validation Checklist

### Suite Setup
- [ ] BeforeSuite initializes testinfra
- [ ] CRDs registered with infra.AddCRDs
- [ ] Mock providers setup (GcpMock/AzureMock/AwsMock)
- [ ] Reconcilers started with infra.StartReconcilers
- [ ] AfterSuite tears down environment

### Test Structure
- [ ] Uses Describe("Feature: ...", func())
- [ ] Uses It("Scenario: ...", func())
- [ ] Uses By("...", func()) blocks for steps
- [ ] Test names descriptive and follow convention

### Async Operations
- [ ] Uses Eventually() not Expect() for changing state
- [ ] Specifies timeout like Eventually(action, "10s", "1s")
- [ ] Uses Should(Succeed()) on function returns

### Mock Configuration
- [ ] Mocks configured BEFORE checking behavior
- [ ] Mock state transitions explicit (e.g., SetRedisClusterState)
- [ ] Verifies mock called with Eventually()

### Resource Testing
- [ ] Tests both create AND delete scenarios
- [ ] Waits for deletion timestamp on delete
- [ ] Verifies cloud resource cleanup
- [ ] Checks finalizer behavior

### SKR Testing
- [ ] Waits for Status.Id generation
- [ ] Verifies KCP resource created
- [ ] Checks spec mapping SKR → KCP
- [ ] Verifies KCP annotations set
- [ ] Tests status sync KCP → SKR

### Error Testing
- [ ] Tests API failures
- [ ] Tests resource not found scenarios
- [ ] Verifies error conditions propagated
- [ ] Tests recovery from errors

### Independence
- [ ] Each test creates own resources
- [ ] No shared mutable state between tests
- [ ] Tests pass individually and together

## Summary: Key Rules

1. **Ginkgo/Gomega ONLY**: BDD testing with testinfra framework REQUIRED
2. **Eventually for Async**: Use `Eventually()` for all changing state, never `Expect()` directly
3. **Mock Before Test**: Configure mock responses BEFORE testing behavior
4. **Test Create + Delete**: Both scenarios REQUIRED for complete coverage
5. **LoadAndCheck Pattern**: Use `Eventually(LoadAndCheck).WithArguments(...)` for assertions
6. **Independent Tests**: Each test creates fresh resources, no shared state
7. **By() Blocks**: Structure tests with descriptive `By("...", func())` blocks
8. **Verify Cloud Cleanup**: Test deletion includes cloud resource cleanup verification

## Next Steps

- [Create Mocks for New Providers](CREATING_MOCKS.md)
- [Add API Validation Tests](API_VALIDATION_TESTS.md)
- [Configure Feature Flags](FEATURE_FLAGS.md)
