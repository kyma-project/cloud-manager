package cloudcontrol

import (
	"time"

	redispb "cloud.google.com/go/redis/apiv1/redispb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	iprangePkg "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	scopePkg "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP RedisInstance", func() {

	It("Scenario: KCP GCP RedisInstance is created and deleted", func() {

		name := "924a92cf-9e72-408d-a1e8-017a2fd8d42d"
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			scopePkg.Ignore.AddName(name)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

		kcpIpRangeName := "ffc7ebcc-114e-4d68-948c-241405fd01b5"
		kcpIpRange := &cloudcontrolv1beta1.IpRange{}

		// Tell IpRange reconciler to ignore this kymaName
		iprangePkg.Ignore.AddName(kcpIpRangeName)
		By("And Given KCP IPRange exists", func() {
			Eventually(CreateKcpIpRange).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), kcpIpRange,
					WithName(kcpIpRangeName),
					WithKcpIpRangeSpecScope(scope.Name),
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

		By("When RedisInstance is created", func() {
			Eventually(CreateRedisInstance).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					WithName(name),
					WithRemoteRef("skr-redis-example"),
					WithIpRange(kcpIpRangeName),
					WithInstanceScope(name),
					WithRedisInstanceGcp(),
					WithKcpGcpRedisInstanceTier("BASIC"),
					WithKcpGcpRedisInstanceMemorySizeGb(5),
					WithKcpGcpRedisInstanceRedisVersion("REDIS_7_0"),
					WithKcpGcpRedisInstanceTransitEncryption(&cloudcontrolv1beta1.TransitEncryptionGcp{
						ServerAuthentication: true,
					}),
					WithKcpGcpRedisInstanceConfigs(map[string]string{
						"maxmemory-policy": "allkeys-lru",
					}),
					WithKcpGcpRedisInstanceMaintenancePolicy(&cloudcontrolv1beta1.MaintenancePolicyGcp{
						DayOfWeek: &cloudcontrolv1beta1.DayOfWeekPolicyGcp{
							Day: "MONDAY",
							StartTime: cloudcontrolv1beta1.TimeOfDayGcp{
								Hours:   14,
								Minutes: 45,
							},
						},
					}),
				).
				Should(Succeed(), "failed creating RedisInstance")
		})

		var memorystoreRedisInstance *redispb.Instance
		By("Then GCP Redis is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingRedisInstanceStatusId()).
				Should(Succeed(), "expected RedisInstance to get status.id")
			memorystoreRedisInstance = infra.GcpMock().GetMemoryStoreRedisByName(redisInstance.Status.Id)
		})

		By("When GCP Redis is Available", func() {
			infra.GcpMock().SetMemoryStoreRedisLifeCycleState(memorystoreRedisInstance.Name, redispb.Instance_READY)
		})

		By("Then RedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
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

		By("And When GCP Redis state is deleted", func() {
			infra.GcpMock().DeleteMemorStoreRedisByName(memorystoreRedisInstance.Name)
		})

		By("Then RedisInstance does not exist", func() {
			Eventually(IsDeleted, 5*time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), redisInstance).
				Should(Succeed(), "expected RedisInstance not to exist (be deleted), but it still exists")
		})
	})

})
