package cloudcontrol

import (
	"time"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpsubnet "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP GcpRedisCluster", func() {

	It("Scenario: KCP GCP GcpRedisCluster is created and deleted", func() {

		name := "10a1ff0e-cb76-4eb2-ae70-2951bb6bc439"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpGcpSubnetName := "2599ac8d-b435-491e-8e17-30ede7b0b571"
		kcpGcpSubnet := &cloudcontrolv1beta1.GcpSubnet{}

		// Tell GcpSubnet reconciler to ignore this kymaName
		kcpsubnet.Ignore.AddName(kcpGcpSubnetName)
		By("And Given KCP Subnet exists", func() {
			Eventually(CreateKcpGcpSubnet).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet,
					WithName(kcpGcpSubnetName),
					WithScope(scope.Name),
					WithRemoteRef("foo-subnet"),
					WithKcpGcpSubnetSpecCidr("10.250.0.0/24"),
					WithKcpGcpSubnetPurposePrivate(),
				).
				Should(Succeed())
		})

		By("And Given KCP GcpSubnet has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpGcpSubnet,
					WithKcpGcpSubnetStatusCidr(kcpGcpSubnet.Spec.Cidr),
					WithConditions(KcpReadyCondition()),
				).WithTimeout(20*time.Second).WithPolling(200*time.Millisecond).
				Should(Succeed(), "Expected KCP GcpSubnet to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.GcpRedisCluster{}

		By("When GcpRedisCluster is created", func() {
			Eventually(CreateKcpGcpRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-rediscluster-example"),
					WithGcpSubnet(kcpGcpSubnetName),
					WithScope(scope.Name),
					WithKcpGcpRedisClusterNodeType("REDIS_SHARED_CORE_NANO"),
					WithKcpGcpRedisClusterShardCount(3),
					WithKcpGcpRedisClusterReplicasPerShard(1),
					WithKcpGcpRedisClusterConfigs(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
				).
				Should(Succeed(), "failed creating GcpRedisCluster")
		})

		var memorystoreGcpRedisCluster *clusterpb.Cluster
		By("Then GCP Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingKcpGcpRedisClusterStatusId()).
				Should(Succeed(), "expected GcpRedisCluster to get status.id")
			memorystoreGcpRedisCluster = infra.GcpMock().GetMemoryStoreRedisClusterByName(redisCluster.Status.Id)
		})

		By("When GCP Redis is Available", func() {
			infra.GcpMock().SetMemoryStoreRedisClusterLifeCycleState(memorystoreGcpRedisCluster.Name, clusterpb.Cluster_State(clusterpb.Cluster_ACTIVE))
		})

		By("Then GcpRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected GcpRedisCluster to has Ready state, but it didn't")
		})

		By("And Then GcpRedisCluster has .status.discoveryEndpoint set", func() {
			Expect(len(redisCluster.Status.DiscoveryEndpoint) > 0).To(Equal(true))
		})

		// DELETE

		By("When GcpRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting GcpRedisCluster")
		})

		By("And When GCP Redis state is deleted", func() {
			infra.GcpMock().DeleteMemorStoreRedisClusterByName(memorystoreGcpRedisCluster.Name)
		})

		By("Then GcpRedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected GcpRedisCluster not to exist (be deleted), but it still exists")
		})
	})

})
