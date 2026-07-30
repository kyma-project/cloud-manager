package cloudcontrol

import (
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP AliCloud RedisCluster", func() {

	It("Scenario: KCP AliCloud RedisCluster is created and deleted", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "e5f6a7b8-c9d0-1234-efab-345678901234"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "f6a7b8c9-d0e1-2345-fabc-456789012345"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-03"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-03",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		instanceClass := "redis.shard.large.ce"
		engineVersion := "7.0"
		shardCount := int32(3)

		By("When RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass(instanceClass),
					WithKcpAlicloudRedisEngineVersion(engineVersion),
					WithKcpAlicloudRedisClusterShardCount(shardCount),
				).Should(Succeed(), "failed creating RedisCluster")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("Then AliCloud Redis cluster is created and gets an ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected RedisCluster to get status.id")
		})

		By("When AliCloud Redis transitions to Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldSet("status", "discoveryEndpoint"),
					HavingFieldSet("status", "authString"),
				).Should(Succeed(), "expected RedisCluster to reach Ready state")
		})

		By("And Then SSL is enabled on the AliCloud cluster", func() {
			entry := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
			Expect(entry).NotTo(BeNil(), "expected mock entry to exist")
			Expect(entry.SslEnabled).To(BeTrue(), "expected SSL to be enabled")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster to be deleted")
		})
	})

	It("Scenario: KCP AliCloud RedisCluster shard count is scaled up", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "a7b8c9d0-e1f2-3456-abcd-567890123456"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "b8c9d0e1-f2a3-4567-bcde-678901234567"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-04"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-04",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-scale"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(3),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisCluster gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("And Given AliCloud Redis is Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("And Given RedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		By("When shard count is increased to 6", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.ShardCount = 6
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then AliCloud transitions to Changing and back to Normal", func() {
			// Drive TransitionAllToNormal inside Eventually so it fires after
			// AddShardingNode has been called and the mock entry is Changing.
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldValue(int32(6), "status", "shardCount"),
				)
			}).Should(Succeed(), "expected RedisCluster to reach Ready with shardCount=6")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: KCP AliCloud RedisCluster replicasPerShard drift is reconciled", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "c9d0e1f2-a3b4-5678-cdef-678901234567"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "d0e1f2a3-b4c5-6789-def0-789012345678"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-05"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-05",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given RedisCluster is created with replicasPerShard=1", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-replica-drift"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(2),
					WithKcpAlicloudRedisClusterReplicasPerShard(1),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisCluster gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("And Given AliCloud Redis is Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("And Given RedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		By("When replicasPerShard is changed to 0", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.ReplicasPerShard = 0
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then AliCloud transitions to Changing and back to Normal with updated replicasPerShard", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldValue(int32(0), "status", "replicasPerShard"),
				)
			}).Should(Succeed(), "expected RedisCluster to reach Ready with replicasPerShard=0")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: KCP AliCloud RedisCluster SSL is re-enabled after spec change resets it", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "d1e2f3a4-b5c6-7890-defa-890123456789"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "e2f3a4b5-c6d7-8901-efab-901234567890"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-06"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-06",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-ssl-restore"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(2),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisCluster is Ready with SSL enabled", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
			alicloudMock.TransitionAllToNormal()
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
			entry := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
			Expect(entry).NotTo(BeNil())
			Expect(entry.SslEnabled).To(BeTrue(), "SSL should be enabled after initial provisioning")
		})

		By("When AliCloud resets SSL during a spec change (simulated)", func() {
			entry := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
			Expect(entry).NotTo(BeNil())
			entry.SslEnabled = false
		})

		By("When a spec change triggers reconcile", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.ShardCount = 3
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then SSL is re-enabled and RedisCluster returns to Ready", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldValue(int32(3), "status", "shardCount"),
				)
			}).Should(Succeed(), "expected RedisCluster to reach Ready with shardCount=3")
			entry := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
			Expect(entry).NotTo(BeNil())
			Expect(entry.SslEnabled).To(BeTrue(), "SSL should be re-enabled after reconcile")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: KCP AliCloud RedisCluster parameters are applied and cleared", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "a8b9c0d1-e2f3-4567-bcde-789012345678"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "b9c0d1e2-f3a4-5678-cdef-890123456789"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists and is Ready", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-07"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-07",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("When RedisCluster is created with initial parameters", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-params"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(2),
					WithKcpAlicloudRedisParameters(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And When AliCloud Redis transitions to Normal", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
			alicloudMock.TransitionAllToNormal()
		})

		By("Then RedisCluster is Ready and parameters are applied", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed(), "expected RedisCluster to reach Ready with parameters")
		})

		By("And Given mock Config reflects the applied parameters", func() {
			entry := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
			Expect(entry).NotTo(BeNil())
			entry.Config = `{"maxmemory-policy":"allkeys-lru"}`
		})

		By("When parameters are cleared", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.Parameters = nil
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then mock Config is cleared to {}", func() {
			Eventually(func() string {
				alicloudMock.TransitionAllToNormal()
				e := alicloudMock.GetRedisCluster(redisCluster.Status.Id)
				if e == nil {
					return ""
				}
				return e.Config
			}).Should(Equal("{}"))
		})

		By("Then RedisCluster is still Ready after parameter clearing", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				)
			}).Should(Succeed(), "expected RedisCluster to remain Ready after clearing parameters")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: KCP AliCloud RedisCluster shard count is scaled down", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "c1d2e3f4-a5b6-7890-cdef-012345678901"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "d2e3f4a5-b6c7-8901-defa-123456789012"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)

		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-08"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-08",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("And Given RedisCluster is created with 4 shards", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-alicloud-cluster-scale-down"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAlicloud(),
					WithKcpAlicloudRedisClusterInstanceClass("redis.shard.large.ce"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisClusterShardCount(4),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisCluster gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("And Given AliCloud Redis is Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("And Given RedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).Should(Succeed())
		})

		By("When shard count is decreased to 2", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), redisCluster); err != nil {
					return err
				}
				redisCluster.Spec.Instance.Alicloud.ShardCount = 2
				return infra.KCP().Client().Update(infra.Ctx(), redisCluster)
			}).Should(Succeed())
		})

		By("Then AliCloud transitions to Changing and back to Normal with reduced shard count", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldValue(int32(2), "status", "shardCount"),
				)
			}).Should(Succeed(), "expected RedisCluster to reach Ready with shardCount=2")
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed())
		})
	})
})
