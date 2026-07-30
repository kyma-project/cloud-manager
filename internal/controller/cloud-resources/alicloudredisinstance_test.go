package cloudresources

import (
	"github.com/google/uuid"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR AlicloudRedisInstance", func() {

	It("Scenario: SKR AlicloudRedisInstance is created and deleted", func() {

		skrIpRangeName := uuid.NewString()
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := uuid.NewString()

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(skrIpRangeName)
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).Should(Succeed())
		})

		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		alicloudRedisInstanceName := uuid.NewString()
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: alicloudRedisInstanceName}))
		alicloudRedisInstance := &cloudresourcesv1beta1.AlicloudRedisInstance{}

		const authSecretName = "alicloud-redis-auth-secret"
		authSecretLabels := map[string]string{"env": "test"}
		authSecretAnnotations := map[string]string{"note": "alicloud"}
		extraData := map[string]string{
			"endpoint": "{{.host}}:{{.port}}",
		}

		By("When AlicloudRedisInstance is created", func() {
			Eventually(CreateAlicloudRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					WithName(alicloudRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisInstanceRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
					WithAlicloudRedisInstanceEngineVersion("5.0"),
					WithAlicloudRedisInstanceAuthSecretName(authSecretName),
					WithAlicloudRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAlicloudRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
					WithAlicloudRedisInstanceAuthSecretExtraData(extraData),
				).Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected SKR AlicloudRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(WithName(alicloudRedisInstance.Status.Id)),
				).Should(Succeed())

			By("And it has KCP labels")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(alicloudRedisInstance.Name))
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(alicloudRedisInstance.Namespace))

			By("And spec.instance.alicloud matches SKR tier")
			Expect(kcpRedisInstance.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.master.small.default"))
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.ReadOnlyCount).To(Equal(int32(0)))
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.EngineVersion).To(Equal("5.0"))
		})

		kcpPrimaryEndpoint := "r-test123.redis.rds.aliyuncs.com:6379"
		kcpAuthString := uuid.NewString()

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpPrimaryEndpoint),
					WithRedisInstanceAuthString(kcpAuthString),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Then SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret,
					NewObjActions(
						WithName(authSecretName),
						WithNamespace(alicloudRedisInstance.Namespace),
					),
					HavingLabelKeys(
						util.WellKnownK8sLabelComponent,
						util.WellKnownK8sLabelPartOf,
						util.WellKnownK8sLabelManagedBy,
					),
					HavingLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, alicloudRedisInstance.Status.Id),
					HavingLabels(authSecretLabels),
					HavingAnnotations(authSecretAnnotations),
				).Should(Succeed())

			Expect(authSecret.Data).To(HaveKey("primaryEndpoint"))
			Expect(authSecret.Data).To(HaveKey("host"))
			Expect(authSecret.Data).To(HaveKey("port"))
			Expect(authSecret.Data).To(HaveKey("authString"))
			Expect(authSecret.Data).To(HaveKeyWithValue("endpoint", []byte(kcpPrimaryEndpoint)))
		})

		// DELETE

		By("When AlicloudRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisInstance redisTier is changed (S→P upgrade)", func() {

		skrIpRangeName := uuid.NewString()
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := uuid.NewString()

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(skrIpRangeName)
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).Should(Succeed())
		})

		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		alicloudRedisInstanceName := uuid.NewString()
		alicloudRedisInstance := &cloudresourcesv1beta1.AlicloudRedisInstance{}

		By("When AlicloudRedisInstance is created with tier S1", func() {
			Eventually(CreateAlicloudRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					WithName(alicloudRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisInstanceRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
					WithAlicloudRedisInstanceEngineVersion("5.0"),
				).Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created with S1 spec", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(WithName(alicloudRedisInstance.Status.Id)),
				).Should(Succeed())

			Expect(kcpRedisInstance.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.master.small.default"))
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.ReadOnlyCount).To(Equal(int32(0)))
		})

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint("r-modify-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("And Given SKR AlicloudRedisInstance is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		By("When redisTier is changed to P2", func() {
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(alicloudRedisInstance), alicloudRedisInstance); err != nil {
					return err
				}
				alicloudRedisInstance.Spec.RedisTier = cloudresourcesv1beta1.AlicloudRedisTierP2
				return infra.SKR().Client().Update(infra.Ctx(), alicloudRedisInstance)
			}).Should(Succeed())
		})

		By("Then KCP RedisInstance spec is updated with P2 instanceClass and readOnlyCount=1", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(),
					HavingFieldValue("redis.amber.master.mid.multithread", "spec", "instance", "alicloud", "instanceClass"),
					HavingFieldValue(int32(1), "spec", "instance", "alicloud", "readOnlyCount"),
				).Should(Succeed())
		})

		// DELETE

		By("When AlicloudRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisInstance parameters are propagated and cleared", func() {

		skrIpRangeName := uuid.NewString()
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := uuid.NewString()

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(skrIpRangeName)
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).Should(Succeed())
		})

		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		alicloudRedisInstanceName := uuid.NewString()
		alicloudRedisInstance := &cloudresourcesv1beta1.AlicloudRedisInstance{}
		initialParams := map[string]string{"maxmemory-policy": "allkeys-lru"}

		By("When AlicloudRedisInstance is created with parameters", func() {
			Eventually(CreateAlicloudRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					WithName(alicloudRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisInstanceRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
					WithAlicloudRedisInstanceEngineVersion("5.0"),
					WithAlicloudRedisInstanceParameters(initialParams),
				).Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created and parameters are propagated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(WithName(alicloudRedisInstance.Status.Id)),
				).Should(Succeed())

			Expect(kcpRedisInstance.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(kcpRedisInstance.Spec.Instance.Alicloud.Parameters).To(HaveKeyWithValue("maxmemory-policy", "allkeys-lru"))
		})

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint("r-params-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("And Given SKR AlicloudRedisInstance is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		By("When parameters are cleared", func() {
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(alicloudRedisInstance), alicloudRedisInstance); err != nil {
					return err
				}
				alicloudRedisInstance.Spec.Parameters = nil
				return infra.SKR().Client().Update(infra.Ctx(), alicloudRedisInstance)
			}).Should(Succeed())
		})

		By("Then KCP RedisInstance parameters are cleared", func() {
			Eventually(func() bool {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(kcpRedisInstance), kcpRedisInstance); err != nil {
					return false
				}
				return len(kcpRedisInstance.Spec.Instance.Alicloud.Parameters) == 0
			}).Should(BeTrue(), "expected KCP RedisInstance parameters to be cleared")
		})

		// DELETE

		By("When AlicloudRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisInstance reflects Updating condition from KCP", func() {

		skrIpRangeName := uuid.NewString()
		skrIpRange := &cloudresourcesv1beta1.IpRange{}
		skrIpRangeId := uuid.NewString()

		By("And Given SKR IpRange exists", func() {
			skriprange.Ignore.AddName(skrIpRangeName)
			Eventually(CreateSkrIpRange).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithName(skrIpRangeName),
				).Should(Succeed())
		})

		By("And Given SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusCidr(skrIpRange.Spec.Cidr),
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		alicloudRedisInstanceName := uuid.NewString()
		alicloudRedisInstance := &cloudresourcesv1beta1.AlicloudRedisInstance{}

		By("When AlicloudRedisInstance is created", func() {
			Eventually(CreateAlicloudRedisInstance).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					WithName(alicloudRedisInstanceName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisInstanceRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
					WithAlicloudRedisInstanceEngineVersion("5.0"),
				).Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					NewObjActions(WithName(alicloudRedisInstance.Status.Id)),
				).Should(Succeed())
		})

		By("When KCP RedisInstance has Updating condition (alongside Ready)", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithConditions(
						KcpReadyCondition(),
						metav1.Condition{
							Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
							Status:  metav1.ConditionTrue,
							Reason:  cloudcontrolv1beta1.ConditionTypeUpdating,
							Message: "Instance is updating",
						},
					),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance reflects StateUpdating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeUpdating),
					HavingFieldValue(cloudresourcesv1beta1.StateUpdating, "status", "state"),
				).Should(Succeed())
		})

		By("When KCP RedisInstance transitions to Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint("r-updating-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance transitions to Ready and Updating condition is removed", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed())
		})

		// DELETE

		By("When AlicloudRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisInstance does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisInstance).
				Should(Succeed())
		})
	})
})
