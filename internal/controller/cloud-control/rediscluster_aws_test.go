package cloudcontrol

import (
	"errors"
	"strings"
	"time"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	awsrediscluster "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/rediscluster"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// isAuthChangingModify reports whether a recorded ModifyReplicationGroup
// invocation touches any auth-related field (AuthTokenSecretString, or the
// user-group attach/detach fields that AWS treats as an auth-flip via the
// implicit AuthTokenUpdateStrategy=Delete on the real client). Used by the
// detached-orphan backfill scenario to prove the reconciler issues no
// auth-changing modify while cleaning up a legacy detached user group.
func isAuthChangingModify(m awsclient.ModifyElastiCacheClusterOptions) bool {
	return m.AuthTokenSecretString != nil ||
		len(m.UserGroupIdsToAdd) > 0 ||
		len(m.UserGroupIdsToRemove) > 0
}

var _ = Describe("Feature: KCP RedisCluster", func() {

	It("Scenario: KCP AWS RedisCluster is created and deleted", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "79b1dd0e-06f2-4fa5-b4c1-b9eba45aabd4"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("Then AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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

		By("And Then RedisCluster has .status.nodeType set", func() {
			Expect(redisCluster.Status.NodeType).To(Equal(cacheNodeType))
		})

		By("And Then RedisCluster has .status.shardCount set", func() {
			Expect(redisCluster.Status.ShardCount).To(Equal(int32(shardCount)))
		})

		By("And Then RedisCluster has .status.replicasPerShard set", func() {
			Expect(redisCluster.Status.ReplicasPerShard).To(Equal(int32(readReplicas)))
		})

		By("And Then no transient user group was created (auth-enabled create path)", func() {
			Expect(awsMock.CreateUserGroupCallCount()).To(Equal(0),
				"expected zero CreateUserGroup calls on the auth-enabled create path, got %d",
				awsMock.CreateUserGroupCallCount())
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

	It("Scenario: KCP AWS RedisCluster downgrades auth (true -> false) via transient user group", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "15f363d9-d22d-408d-bc51-872d2f925133"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "86041363-3fda-4643-8cd5-75a871ba27e4"
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

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		By("And Given RedisCluster is created with authEnabled=true", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws-downgrade"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAws(),
					WithKcpAwsCacheNodeType("cache.m5.large"),
					WithKcpAwsEngineVersion("7.0"),
					WithKcpAwsAutoMinorVersionUpgrade(true),
					WithKcpAwsAuthEnabled(true),
					WithKcpAwsShardCount(int32(3)),
					WithKcpAwsReadReplicas(int32(1)),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected RedisCluster to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisCluster.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given RedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready")).
				Should(Succeed(), "expected RedisCluster to reach Ready before downgrade")
		})

		By("And Given no transient user group has been created yet", func() {
			Expect(awsMock.CreateUserGroupCallCount()).To(Equal(0))
		})

		By("When RedisCluster spec.authEnabled flips to false", func() {
			Eventually(UpdateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithKcpAwsAuthEnabled(false)).
				Should(Succeed(), "failed flipping authEnabled to false")
		})

		expectedUGName := awsrediscluster.GetAwsElastiCacheUserGroupName(redisCluster.Name)

		By("Then the transient user group is created and eventually reaches Active", func() {
			Eventually(func() int {
				return awsMock.CreateUserGroupCallCount()
			}).Should(Equal(1), "expected exactly one CreateUserGroup call")
			// Lock in at-most-once across subsequent reconcile passes.
			Consistently(func() int {
				return awsMock.CreateUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(1), "CreateUserGroup must be at-most-once")
			awsMock.SetAwsElastiCacheUserGroupLifeCycleState(expectedUGName, awsmeta.ElastiCache_UserGroup_ACTIVE)
		})

		By("And Then the reconciler drives auth to disabled and the UG is detached", func() {
			// AWS-side "modifying" needs to settle back to Available for the
			// reconciler to make progress across passes.
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
			// Lock in at-most-once for DeleteUserGroup too.
			Consistently(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(1), "DeleteUserGroup must be at-most-once")
			// Reconciler waits for DELETING → gone; simulate final tombstone removal.
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
		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})
		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})
		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisCluster with a detached-orphan user group backfills to zero", func() {
		// Pre-fix legacy state: an auth-disabled Ready cluster with an
		// existing detached cm-<name> user group. On the next reconcile the
		// delete-side predicate must fire and drive the account to zero
		// cm-<name> user groups.

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "6014090c-99e1-48e2-819b-0d07cb293109"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "86b69e36-16e1-47fa-b0f7-4ee6534e49f5"
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

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		By("And Given RedisCluster is created with authEnabled=false", func() {
			Eventually(CreateRedisCluster).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					WithName(name),
					WithRemoteRef("skr-redis-example-aws-orphan"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisClusterAws(),
					WithKcpAwsCacheNodeType("cache.m5.large"),
					WithKcpAwsEngineVersion("7.0"),
					WithKcpAwsAutoMinorVersionUpgrade(true),
					WithKcpAwsAuthEnabled(false),
					WithKcpAwsShardCount(int32(3)),
					WithKcpAwsReadReplicas(int32(1)),
				).
				Should(Succeed(), "failed creating RedisCluster")
		})

		awsMock := awsAccount.Region(scope.Spec.Region)
		expectedUGName := awsrediscluster.GetAwsElastiCacheUserGroupName(redisCluster.Name)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
				Should(Succeed(), "expected RedisCluster to get status.id")
			awsElastiCacheClusterInstance = awsMock.GetAwsElastiCacheByName(redisCluster.Status.Id)
		})

		By("And Given AWS Redis is Available", func() {
			awsMock.SetAwsElastiCacheLifeCycleState(*awsElastiCacheClusterInstance.ReplicationGroupId, awsmeta.ElastiCache_AVAILABLE)
		})

		By("And Given RedisCluster reaches Ready (no user group involved yet)", func() {
			// Establish the "auth-disabled Ready cluster" precondition BEFORE
			// seeding the legacy orphan. Otherwise the delete-side predicate
			// would fire during initial reconcile and block Ready until the
			// orphan is cleaned up — inverting the scenario's semantics.
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready")).
				Should(Succeed(), "expected RedisCluster to reach Ready before orphan is introduced")
		})

		var baselineDeleteCalls int
		By("And Given we seed a detached legacy cm-<name> user group in Active status", func() {
			// Simulate pre-fix state: the UG appears in the account (attached
			// to nothing) after the cluster reached Ready.
			awsMock.PreCreateUserGroup(expectedUGName, awsmeta.ElastiCache_UserGroup_ACTIVE)
			baselineDeleteCalls = awsMock.DeleteUserGroupCallCount()
		})

		By("And When an annotation change triggers a reconcile", func() {
			// The reconciler doesn't watch AWS-side state; a k8s-side change
			// on the CR is what fires the watch event that picks up the
			// newly-seeded orphan. Bumping an annotation is guaranteed to
			// change resourceVersion and re-queue the object.
			Eventually(func() error {
				fresh := &cloudcontrolv1beta1.RedisCluster{}
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisCluster), fresh); err != nil {
					return err
				}
				if fresh.Annotations == nil {
					fresh.Annotations = map[string]string{}
				}
				fresh.Annotations["cloud-manager.kyma-project.io/orphan-backfill-test"] = "1"
				return infra.KCP().Client().Update(infra.Ctx(), fresh)
			}).Should(Succeed(), "failed touching RedisCluster to trigger reconcile")
		})

		By("Then reconciler deletes the detached orphan on the next reconcile", func() {
			// Wait for the reconciler to observe the detached UG and issue Delete.
			Eventually(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}).Should(Equal(baselineDeleteCalls+1), "expected exactly one DeleteUserGroup call after orphan was introduced")
			Consistently(func() int {
				return awsMock.DeleteUserGroupCallCount()
			}, 500*time.Millisecond).Should(Equal(baselineDeleteCalls+1), "backfill DeleteUserGroup must be at-most-once")

			// No auth-changing modify was issued — this is pure backfill.
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
		By("When RedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "failed deleting RedisCluster")
		})
		By("And When AWS Redis state is deleted", func() {
			awsMock.DeleteAwsElastiCacheByName(*awsElastiCacheClusterInstance.ReplicationGroupId)
		})
		By("Then RedisCluster does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster).
				Should(Succeed(), "expected RedisCluster not to exist (be deleted), but it still exists")
		})
	})

	It("Scenario: KCP AWS RedisCluster is upgraded (6.x -> 7.0)", func() {

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "f6c81db7-05ea-4a5d-b73f-11b1eb15935c"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "6.x"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		name := "0c71148c-d5aa-42bd-82c3-de17f1bb269e"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeAws).
				WithArguments(infra.Ctx(), infra, scope, awsAccount.AccountId(), WithName(name)).
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
				).
				Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisCluster := &cloudcontrolv1beta1.RedisCluster{}
		cacheNodeType := "cache.m5.large"
		engineVersion := "7.0"
		autoMinorVersionUpgrade := true
		preferredMaintenanceWindow := new("sun:23:00-mon:01:30")
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

		awsMock := awsAccount.Region(scope.Spec.Region)

		var awsElastiCacheClusterInstance *elasticachetypes.ReplicationGroup
		By("And Given AWS Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id")).
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
