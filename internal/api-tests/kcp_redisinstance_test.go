package api_tests

import (
	"time"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: KCP RedisInstance .status patch", Ordered, func() {
	name := "8c3b91a9-2265-4e92-810d-df61cf9bc16a"
	scope := &cloudcontrolv1beta1.Scope{}
	kcpIpRangeName := "3cec2a0b-a98b-4dac-9b9a-ba4888a00934"
	kcpIpRange := &cloudcontrolv1beta1.IpRange{}

	BeforeAll(func() {
		By("Given Scope exists", func() {
			kcpscope.Ignore.AddName(name)

			Eventually(CreateScopeGcp).
				WithArguments(infra.Ctx(), infra, scope, WithName(name)).
				Should(Succeed())
		})

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
	})

	It("Scenario: KCP RedisInstance supports patch status", func() {
		obj := &cloudcontrolv1beta1.RedisInstance{}
		name := "eac4ab45-6c9e-4aff-a457-9d6da06d93af"

		By("When RedisInstance is created", func() {

			Expect(CreateRedisInstance(
				infra.Ctx(), infra.KCP().Client(), obj,
				WithName(name),
				WithRemoteRef("skr-redis-example"),
				WithIpRange(kcpIpRangeName),
				WithScope(name),
				WithRedisInstanceGcp(),
				WithKcpGcpRedisInstanceTier("BASIC"),
				WithKcpGcpRedisInstanceMemorySizeGb(5),
				WithKcpGcpRedisInstanceRedisVersion("REDIS_7_0"),
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
			)).To(Succeed())
		})

		By("Then RedisInstance has no conditions", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(obj.Status.Conditions).To(HaveLen(0))
		})

		By("When RedisInstance is patched with Ready condition", func() {
			meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: cloudcontrolv1beta1.ConditionTypeReady,
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.KCP().Client())).
				To(Succeed())
		})

		By("And When RedisInstance is patched with Error condition", func() {
			meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: cloudcontrolv1beta1.ConditionTypeError,
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.KCP().Client())).
				To(Succeed())
		})

		By("Then RedisInstance has two conditions", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(obj.Status.Conditions).To(HaveLen(2))
		})

		By("When RedisInstance Ready condition is removed", func() {
			meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)

			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.KCP().Client())).
				To(Succeed())
		})

		By("Then RedisInstance has one conditions", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(obj.Status.Conditions).To(HaveLen(1))
		})

		By("When RedisInstance Error condition is removed", func() {
			meta.RemoveStatusCondition(obj.Conditions(), cloudcontrolv1beta1.ConditionTypeError)

			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.KCP().Client())).
				To(Succeed())
		})

		By("Then RedisInstance has no conditions", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(obj.Status.Conditions).To(HaveLen(0))
		})

	})

})

var _ = Describe("Feature: KCP RedisInstance", func() {

	// GCP ===============================================

	It("Scenario: KCP RedisInstance GCP with IpRange can be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{
						MemorySizeGb: int32(16),
						RedisVersion: "REDIS_7_0",
						AuthEnabled:  true,
						Tier:         "BASIC",
						ReplicaCount: 0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP RedisInstance GCP without IpRange can not be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Gcp: &cloudcontrolv1beta1.RedisInstanceGcp{
						MemorySizeGb: int32(16),
						RedisVersion: "REDIS_7_0",
						AuthEnabled:  true,
						Tier:         "BASIC",
						ReplicaCount: 0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange name must not be empty"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	// AWS ===============================================

	It("Scenario: KCP RedisInstance AWS with IpRange can be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Aws: &cloudcontrolv1beta1.RedisInstanceAws{
						CacheNodeType:           "cache.m5.large",
						EngineVersion:           "7.0",
						AutoMinorVersionUpgrade: false,
						AuthEnabled:             true,
						ReadReplicas:            0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP RedisInstance AWS without IpRange can not be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Aws: &cloudcontrolv1beta1.RedisInstanceAws{
						CacheNodeType:           "cache.m5.large",
						EngineVersion:           "7.0",
						AutoMinorVersionUpgrade: false,
						AuthEnabled:             true,
						ReadReplicas:            0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange name must not be empty"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	// Azure ===============================================

	It("Scenario: KCP RedisInstance Azure with IpRange can be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Azure: &cloudcontrolv1beta1.RedisInstanceAzure{
						SKU: cloudcontrolv1beta1.AzureRedisSKU{
							Capacity: 1,
						},
						RedisVersion: "7.2",
						ShardCount:   0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP RedisInstance Azure without IpRange can not be created", func() {
		var err error
		obj := &cloudcontrolv1beta1.RedisInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.RedisInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.RedisInstanceInfo{
					Azure: &cloudcontrolv1beta1.RedisInstanceAzure{
						SKU: cloudcontrolv1beta1.AzureRedisSKU{
							Capacity: 1,
						},
						RedisVersion: "7.2",
						ShardCount:   0,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange name must not be empty"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})
})
