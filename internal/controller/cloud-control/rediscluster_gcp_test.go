package cloudcontrol

import (
	"time"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP RedisCluster", func() {

	It("Scenario: KCP GCP RedisCluster is created and deleted", func() {

		name := "485e202e-6d22-43c0-a936-47c107ffd574"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "577d95d3-1e62-4ac3-9883-581db961d84f"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithScope(scope.Name),
				).
				Should(Succeed())
		})

		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).WithTimeout(20*time.Second).WithPolling(200*time.Millisecond).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("When RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-rediscluster-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterGcp(),
					WithKcpGcpRedisClusterNodeType("REDIS_SHARED_CORE_NANO"),
					WithKcpGcpRedisClusterShardCount(3),
					WithKcpGcpRedisClusterReplicasPerShard(1),
					WithKcpGcpRedisClusterConfigs(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		var memorystoreRedisCluster *clusterpb.Cluster
		By("Then GCP Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingRedisClusterStatusId()).
				Should(Succeed(), "expected RedisCluster to get status.id")
			memorystoreRedisCluster = infra.GcpMock().GetMemoryStoreRedisClusterByName(redisCluster.Status.Id)
		})

		By("When GCP Redis is Available", func() {
			infra.GcpMock().SetMemoryStoreRedisClusterLifeCycleState(memorystoreRedisCluster.Name, clusterpb.Cluster_State(clusterpb.Cluster_ACTIVE))
		})

		By("Then RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisCluster to has Ready state, but it didn't")
		})

		By("And Then RedisCluster has .status.primaryEndpoint set", func() {
			Expect(len(redisCluster.Status.DiscoveryEndpoint) > 0).To(Equal(true))
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("And When GCP Redis state is deleted", func() {
			infra.GcpMock().DeleteMemorStoreRedisClusterByName(memorystoreRedisCluster.Name)
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})

})
