package cloudcontrol

import (
	"time"

	"github.com/alibabacloud-go/tea/tea"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP AliCloud RedisInstance", func() {

	It("Scenario: KCP AliCloud RedisInstance is created and deleted", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "b2c3d4e5-f6a7-8901-bcde-f12345678901"
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
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-01"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-01",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}
		instanceClass := "tair.rdb.1g"
		engineVersion := "7.0"

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-alicloud-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAlicloud(),
					WithKcpAlicloudRedisInstanceClass(instanceClass),
					WithKcpAlicloudRedisEngineVersion(engineVersion),
				).Should(Succeed(), "failed creating RedisInstance")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("Then AliCloud Redis is created in Creating status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected RedisInstance to get status.id")
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldSet("status", "primaryEndpoint"),
					HavingFieldSet("status", "authString"),
				)
			}).Should(Succeed(), "expected RedisInstance to reach Ready state")
		})

		By("And Then SSL is enabled on the AliCloud instance", func() {
			entry := alicloudMock.GetRedisInstance(redisInstance.Status.Id)
			Expect(entry).NotTo(BeNil(), "expected mock entry to exist")
			Expect(entry.SslEnabled).To(BeTrue(), "expected SSL to be enabled")
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "failed deleting RedisInstance")
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance to be deleted")
		})
	})

	It("Scenario: KCP AliCloud RedisInstance load error is surfaced", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "c3d4e5f6-a7b8-9012-cdef-123456789012"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "d4e5f6a7-b8c9-0123-defg-234567890123"
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
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-02"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-02",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed(), "Expected KCP IpRange to become ready")
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-alicloud-redis-error"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAlicloud(),
					WithKcpAlicloudRedisInstanceClass("tair.rdb.1g"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
				).Should(Succeed(), "failed creating RedisInstance")
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisInstance gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("When AliCloud returns a permanent error on describe", func() {
			// Use a 403 SDK error so IsPermanentError() returns true and the
			// reconciler surfaces StateError rather than silently requeueing.
			alicloudMock.SetRedisInstanceError(redisInstance.Status.Id, &tea.SDKError{
				StatusCode: tea.Int(403),
				Code:       tea.String("Forbidden.RAM"),
				Message:    tea.String("simulated permanent AliCloud API failure"),
			})
		})

		By("Then RedisInstance has Error condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeError),
				).Should(Succeed(), "expected RedisInstance to surface error condition")
		})

		By("When error is cleared", func() {
			alicloudMock.SetRedisInstanceError(redisInstance.Status.Id, nil)
		})

		By("Then RedisInstance recovers to Ready", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				)
			}).Should(Succeed())
		})

		// cleanup
		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})
	})

	It("Scenario: KCP AliCloud RedisInstance instanceClass and readOnlyCount drift is reconciled", func() {

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()

		name := "d4e5f6a7-b8c9-0123-def0-456789012345"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, scope, alicloudAccount.Credentials().AccessKeyId, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "e5f6a7b8-c9d0-1234-ef01-567890123456"
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
					WithKcpIpRangeStatusVpcId("vpc-alicloud-test-09"),
					WithKcpIpRangeStatusSubnets(cloudcontrolv1beta1.IpRangeSubnet{
						Id:   "vsw-alicloud-test-09",
						Zone: "cn-hangzhou-a",
					}),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		redisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given RedisInstance is created with S tier (no read-only replica)", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-alicloud-redis-drift"),
					WithIpRange(kcpIpRangeName),
					WithScope(name),
					WithRedisInstanceAlicloud(),
					WithKcpAlicloudRedisInstanceClass("tair.rdb.1g"),
					WithKcpAlicloudRedisEngineVersion("7.0"),
					WithKcpAlicloudRedisReadOnlyCount(0),
				).Should(Succeed())
		})

		alicloudMock := alicloudAccount.Region(scope.Spec.Region)

		By("And Given RedisInstance gets its ID", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())
		})

		By("And Given AliCloud Redis transitions to Normal", func() {
			alicloudMock.TransitionAllToNormal()
		})

		By("And Given RedisInstance is Ready", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
				)
			}).Should(Succeed())
		})

		By("When instanceClass and readOnlyCount are changed (S→P tier upgrade)", func() {
			Eventually(func() error {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(redisInstance), redisInstance); err != nil {
					return err
				}
				redisInstance.Spec.Instance.Alicloud.InstanceClass = "tair.rdb.2g"
				redisInstance.Spec.Instance.Alicloud.ReadOnlyCount = 1
				return infra.KCP().Client().Update(infra.Ctx(), redisInstance)
			}).Should(Succeed())
		})

		By("Then AliCloud transitions to Changing and back to Normal with new class", func() {
			Eventually(func() error {
				alicloudMock.TransitionAllToNormal()
				return LoadAndCheck(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState("Ready"),
					HavingFieldValue("tair.rdb.2g", "status", "nodeType"),
				)
			}).Should(Succeed(), "expected RedisInstance to reach Ready with updated instanceClass")
		})

		// DELETE

		By("When RedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed())
		})
	})

})
