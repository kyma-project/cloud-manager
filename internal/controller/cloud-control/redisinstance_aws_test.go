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

var _ = Describe("Feature: KCP RedisInstance", func() {

	It("Scenario: KCP AWS RedisInstance is created and deleted", func() {

		name := "6e6ff0b2-3edb-4d6e-8ae5-fbd3d3644ce2"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "7017ef87-3814-4dc5-bcd1-966d2f44e285"
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsReadReplicas(int32(readReplicas)),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("Then AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisInstance.Status.Id)
		})

		By("And Then AWS Redis has defined custom parameters", func() {
			remoteParameters := awsMock.DescribeAwsElastiCacheParametersByName("cm-" + redisInstance.Name)

			Expect(remoteParameters["active-defrag-cycle-max"]).To(Equal(parameters["active-defrag-cycle-max"]))
		})

		By("When AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		By("And Then RedisInstance has .status.primaryEndpoint set", func() {
			Expect(len(redisInstance.Status.PrimaryEndpoint) > 0).To(Equal(true))
		})

		By("And Then RedisInstance has .status.readEndpoint set", func() {
			Expect(len(redisInstance.Status.ReadEndpoint) > 0).To(Equal(true))
		})

		By("And Then RedisInstance has .status.authString set", func() {
			Expect(len(redisInstance.Status.AuthString) > 0).To(Equal(true))
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisInstance is upgraded (6.x -> 7.0)", func() {

		name := "3c44ec47-b214-4ce9-9c1b-292b0508fb98"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "82eaa314-b9d1-4345-986b-9f1b0aae39e7"
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("And Given RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsReadReplicas(int32(readReplicas)),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisInstance.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Given RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		upgradVersionTarget := "7.0"
		By("When RedisInstance manifest has EngineVersion is upgraded", func() {
			Eventually(UpdateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithKcpAwsEngineVersion(upgradVersionTarget),
				).
				Should(Succeed(), "failed updating RedisInstance manifest")
		})

		By("Then RedisInstance has condition Updating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeUpdating),
				).
				Should(Succeed(), "expected RedisInstance to have Updating condition, but it didn't")
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

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisInstance is upgraded (7.0 -> 7.1)", func() {

		name := "23549459-1164-4684-980d-27a0a579018b"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "5ce2dcfd-87a2-4c51-a654-4795b0fa235f"
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

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "7.0"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := ptr.To("sun:23:00-mon:01:30")
		authEnabled := true
		readReplicas := 1

		parameters := map[string]string{
			"active-defrag-cycle-max": "85",
		}

		By("And Given RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAws(),
					WithKcpAwsCacheNodeType(cacheNodeType),
					WithKcpAwsEngineVersion(engineVersion),
					WithKcpAwsAutoMinorVersionUpgrade(autoMinorVersionUpgrade),
					WithKcpAwsAuthEnabled(authEnabled),
					WithKcpAwsPreferredMaintenanceWindow(preferredMaintenanceWindow),
					WithKcpAwsParameters(parameters),
					WithKcpAwsReadReplicas(int32(readReplicas)),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		awsMock := infra.AwsMock().MockConfigs(scope.Spec.Scope.Aws.AccountId, scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisInstance.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given when AWS Redis UserGroup is Active", func() {
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Given RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		upgradVersionTarget := "7.1"
		By("When RedisInstance manifest has EngineVersion is upgraded", func() {
			Eventually(UpdateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithKcpAwsEngineVersion(upgradVersionTarget),
				).
				Should(Succeed(), "failed updating RedisInstance manifest")
		})

		By("Then RedisInstance has condition Updating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeUpdating),
				).
				Should(Succeed(), "expected RedisInstance to have Updating condition, but it didn't")
		})

		By("When AWS ElastiCache Nodes are upgraded and Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
			awsMock.SetAwsElastiCacheEngineVersion(*awsElastiCacheClusterInstance.ReplicationGroupId, upgradVersionTarget)
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				).
				Should(Succeed(), "expected RedisInstance to has Ready state, but it didn't")
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("And When AWS Redis user group is deleted", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})
})
