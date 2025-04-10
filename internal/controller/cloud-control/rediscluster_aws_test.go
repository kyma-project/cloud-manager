package cloudcontrol

import (
	"errors"
	"strings"
	"time"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: KCP RedisCluster", func() {

	It("Scenario: KCP AWS RedisCluster is created and deleted", func() {

		name := "79b1dd0e-06f2-4fa5-b4c1-b9eba45aabd4"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "d1a362d2-6a45-4908-9252-605b1970487a"
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
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1
		shardCount := 3

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("When RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws3"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsShardCount(int32(shardCount)),
					WithKcpAwsReadReplicas(int32(readReplicas)),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("Then AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingRedisClusterStatusId()).
				Should(Succeed(), "expected RedisCluster to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisCluster.Status.Id)
		})

		By("And Then AWS Redis has defined custom parameters", func() {
			remoteParameters := awsMock.DescribeAwsElastiCacheParametersByName("cm-" + redisCluster.Name)

			Expect(remoteParameters["active-defrag-cycle-max"]).To(Equal(parameters["active-defrag-cycle-max"]))
		})

		By("When AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
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

		By("And Then RedisCluster has .status.discoveryEndpoint set", func() {
			Expect(len(redisCluster.Status.DiscoveryEndpoint) > 0).To(Equal(true))
		})

		By("And Then RedisCluster has .status.authString set", func() {
			Expect(len(redisCluster.Status.AuthString) > 0).To(Equal(true))
		})

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisCluster is upgraded (6.x -> 7.0)", func() {

		name := "f6c81db7-05ea-4a5d-b73f-11b1eb15935c"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "20351a3e-7e28-446f-8d6a-421a90b0f65f"
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
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1
		shardCount := 3

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("And Given RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws4"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsReadReplicas(int32(readReplicas)),
					WithKcpAwsShardCount(int32(shardCount)),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingRedisClusterStatusId()).
				Should(Succeed(), "expected RedisCluster to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisCluster.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Given RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisCluster to has Ready state, but it didn't")
		})

		upgradVersionTarget := "7.0"
		By("When RedisCluster manifest has EngineVersion is upgraded", func() {
			Eventually(UpdateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithKcpAwsEngineVersion(upgradVersionTarget),
				).
				Should(Succeed(), "failed updating RedisCluster manifest")
		})

		By("Then RedisCluster has condition Updating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeUpdating),
				).
				Should(Succeed(), "expected RedisCluster to have Updating condition, but it didn't")
		})

		By("Then Aws ElastiCache is given temp ParamGroup for upgrade", func() {
			Eventually(func() error {
				node := awsMock.GetAWsElastiCacheNodeByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
				if strings.Contains(ptr.Deref(node.CacheParameterGroup.CacheParameterGroupName, ""), "temp") {
					return nil
				}

				return errors.New("not using temp paramgroup")
			}).Should(Succeed())
		})

		By("When AWS ElastiCache Nodes are upgraded and Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
			awsMock.SetAwsElastiCacheEngineVersion(*awsElastiCacheClusterInstance.ReplicationGroupId, upgradVersionTarget)
		})

		By("Then AWS ElastiCache is switched to main ParamGroup", func() {
			Eventually(func() error {
				node := awsMock.GetAWsElastiCacheNodeByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
				if strings.Contains(ptr.Deref(node.CacheParameterGroup.CacheParameterGroupName, ""), "temp") {
					return errors.New("not using main paramgroup")
				}

				return nil
			}).Should(Succeed())
		})

		By("When Switch to main param group is done", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
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

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisCluster is upgraded (7.0 -> 7.1)", func() {

		name := "0c71148c-d5aa-42bd-82c3-de17f1bb269e"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "32a0ef3a-50a6-4c2c-8994-0b119a3ba7ad"
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
		cacheNodeType := "cache.m5.large"
		engineVersion := "7.0"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1
		shardCount := 3

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("And Given RedisCluster is created", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws5"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsReadReplicas(int32(readReplicas)),
					WithKcpAwsShardCount(int32(shardCount)),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingRedisClusterStatusId()).
				Should(Succeed(), "expected RedisCluster to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisCluster.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Given RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisCluster to has Ready state, but it didn't")
		})

		upgradVersionTarget := "7.1"
		By("When RedisCluster manifest has EngineVersion is upgraded", func() {
			Eventually(UpdateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithKcpAwsEngineVersion(upgradVersionTarget),
				).
				Should(Succeed(), "failed updating RedisCluster manifest")
		})

		By("Then RedisCluster has condition Updating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeUpdating),
				).
				Should(Succeed(), "expected RedisCluster to have Updating condition, but it didn't")
		})

		By("When AWS ElastiCache Nodes are upgraded and Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
			awsMock.SetAwsElastiCacheEngineVersion(*awsElastiCacheClusterInstance.ReplicationGroupId, upgradVersionTarget)
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

		// DELETE

		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})
})
