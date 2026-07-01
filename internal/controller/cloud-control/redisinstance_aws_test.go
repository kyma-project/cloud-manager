package cloudcontrol

import (
	"errors"
	"strings"
	"time"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsredisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP RedisInstance", func() {

	It("Scenario: KCP AWS RedisInstance is created and deleted", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "6e6ff0b2-3edb-4d6e-8ae5-fbd3d3644ce2"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("Then AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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

		By("And Then RedisInstance has .status.nodeType set", func() {
			Expect(redisInstance.Status.NodeType).To(Equal(cacheNodeType))
		})

		By("And Then RedisInstance has .status.replicaCount set", func() {
			Expect(redisInstance.Status.ReplicaCount).To(Equal(int32(readReplicas)))
		})

		By("And Then no transient user group was created (auth-enabled create path)", func() {
			Expect(awsMock.CreateUserGroupCallCount()).To(Equal(0),
				"expected zero CreateUserGroup calls on the auth-enabled create path, got %d",
				awsMock.CreateUserGroupCallCount())
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

	It("Scenario: KCP AWS RedisInstance downgrades auth (true -> false) via transient user group", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "f96812fd-7059-4a44-a63f-ce466e37b926"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "ac2a41f5-874b-456c-bf50-0b04d30b56f4"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName), WithScope(scope.Name)).
				Should(Succeed())
		})
		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		By("And Given RedisInstance is created with authEnabled=true", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws-downgrade"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAws(),
					WithKcpAwsCacheNodeType("cache.m5.large"),
					WithKcpAwsEngineVersion("7.0"),
					WithKcpAwsAutoMinorVersionUpgrade(true),
					WithKcpAwsAuthEnabled(true),
					WithKcpAwsReadReplicas(int32(1)),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected RedisInstance to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisInstance.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready")).
				Should(Succeed(), "expected RedisInstance to reach Ready before downgrade")
		})

		By("And Given no transient user group has been created yet", func() {
			Expect(awsMock.CreateUserGroupCallCount()).To(Equal(0))
		})

		By("When RedisInstance spec.authEnabled flips to false", func() {
			Eventually(UpdateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithKcpAwsAuthEnabled(false)).
				Should(Succeed(), "failed flipping authEnabled to false")
		})

		expectedUGName := awsredisinstance.GetAwsElastiCacheUserGroupName(redisInstance.Name)

		By("Then the transient user group is created and eventually reaches Active", func() {
			Eventually(func() int {
				return awsMock.CreateUserGroupCallCount()
			}).Should(Equal(1), "expected exactly one CreateUserGroup call")
			Consistently(func() int {
				return awsMock.CreateUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(1), "CreateUserGroup must be at-most-once")
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(expectedUGName, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Then the reconciler drives auth to disabled and the UG is detached", func() {
			Eventually(func() error {
				awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
				rg := awsMock.GetAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
				if ptr.Deref(rg.AuthTokenEnabled, true) {
					return errors.New("AuthTokenEnabled still true")
				}
				if len(rg.UserGroupIds) != 0 {
					return errors.New("UserGroupIds not empty yet")
				}
				return nil
			}).Should(Succeed(), "expected AuthTokenEnabled=false and UserGroupIds=[]")
		})

		By("And Then the transient user group is deleted", func() {
			Eventually(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}).Should(Equal(1), "expected exactly one DeleteUserGroup call")
			Consistently(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(1), "DeleteUserGroup must be at-most-once")
			awsMock.DeleteAwsElastiCacheUserGroupByName(expectedUGName)
		})

		By("Then AWS call sequence is: Create, ModifyAdd, ModifyRemove, Delete — each exactly once", func() {
			Expect(awsMock.CreateUserGroupCalls()).To(Equal([]string{expectedUGName}))
			Expect(awsMock.DeleteUserGroupCalls()).To(Equal([]string{expectedUGName}))
			modifies := awsMock.ModifyReplicationGroupCalls()
			var addAt, removeAt int = -1, -1
			for i, m := range modifies {
				if len(m.UserGroupIdsToAdd) > 0 && addAt < 0 {
					addAt = i
					Expect(m.UserGroupIdsToAdd).To(Equal([]string{expectedUGName}))
				}
				if len(m.UserGroupIdsToRemove) > 0 && removeAt < 0 {
					removeAt = i
					Expect(m.UserGroupIdsToRemove).To(Equal([]string{expectedUGName}))
				}
			}
			Expect(addAt).To(BeNumerically(">=", 0), "expected a ModifyReplicationGroup call adding the transient UG")
			Expect(removeAt).To(BeNumerically(">", addAt), "Remove modify must occur strictly after Add modify")
		})

		By("And Then DescribeUserGroup(cm-<name>) returns not-found", func() {
			Eventually(func() *elasticachetypes.UserGroup {
				return awsMock.GetAwsElastiCacheUserGroupByName(expectedUGName)
			}).Should(BeNil(), "user group must be gone from AWS inventory")
		})

		// Cleanup
		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})
		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})
		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisInstance with a detached-orphan user group backfills to zero", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "99dfd834-54c5-4e96-b6ac-82cc6a3b7c5f"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "5c5438d7-50d0-428b-a5b1-34382bc7b30d"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}
		kcpiprange.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName), WithScope(scope.Name)).
				Should(Succeed())
		})
		By("And Given KCP IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithKcpIpRangeStatusCidr(kcpIpRange.Spec.Cidr),
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		By("And Given RedisInstance is created with authEnabled=false", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws-orphan"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAws(),
					WithKcpAwsCacheNodeType("cache.m5.large"),
					WithKcpAwsEngineVersion("7.0"),
					WithKcpAwsAutoMinorVersionUpgrade(true),
					WithKcpAwsAuthEnabled(false),
					WithKcpAwsReadReplicas(int32(1)),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		awsMock := awsAccount.Region(scope.Spec.Region)
		expectedUGName := awsredisinstance.GetAwsElastiCacheUserGroupName(redisInstance.Name)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected RedisInstance to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisInstance.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given RedisInstance reaches Ready (no user group involved yet)", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready")).
				Should(Succeed(), "expected RedisInstance to reach Ready before orphan is introduced")
		})

		var baselineDeleteCalls int
		By("And Given we seed a detached legacy cm-<name> user group in Active status", func() {
			awsMock.PreCreateUserGroup(expectedUGName, awsmeta.ElastiCache_UserGroup_ACTIVE)
			baselineDeleteCalls = awsMock.DeleteUserGroupCallCount()
		})

		By("And When an annotation change triggers a reconcile", func() {
			// The reconciler doesn't watch AWS-side state; a k8s-side change
			// on the CR is what fires the watch event that picks up the
			// newly-seeded orphan.
			Eventually(func() error {
				fresh := &cloudcontrolv1beta1.RedisInstance{}
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisInstance), fresh); err != nil {
					return err
				}
				if fresh.Annotations == nil {
					fresh.Annotations = map[string]string{}
				}
				fresh.Annotations["cloud-manager.kyma-project.io/orphan-backfill-test"] = "1"
				return infra.KCP().Client().Update(infra.Ctx(), fresh)
			}).Should(Succeed(), "failed touching RedisInstance to trigger reconcile")
		})

		By("Then reconciler deletes the detached orphan on the next reconcile", func() {
			Eventually(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}).Should(Equal(baselineDeleteCalls+1), "expected exactly one DeleteUserGroup call after orphan was introduced")
			Consistently(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(baselineDeleteCalls+1), "backfill DeleteUserGroup must be at-most-once")

			for _, m := range awsMock.ModifyReplicationGroupCalls() {
				Expect(isAuthChangingModify(m)).To(BeFalse(),
					"backfill must not issue an auth-changing modify, got %+v", m)
			}
		})

		By("And Then final tombstone removal yields not-found", func() {
			awsMock.DeleteAwsElastiCacheUserGroupByName(expectedUGName)
			Eventually(func() *elasticachetypes.UserGroup {
				return awsMock.GetAwsElastiCacheUserGroupByName(expectedUGName)
			}).Should(BeNil())
		})

		// Cleanup
		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})
		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})
		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisInstance is upgraded (6.x -> 7.0)", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "3c44ec47-b214-4ce9-9c1b-292b0508fb98"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "23549459-1164-4684-980d-27a0a579018b"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "7.0"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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
