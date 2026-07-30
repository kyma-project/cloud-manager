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

var _ = Describe("Feature: SKR AlicloudRedisCluster", func() {

	It("Scenario: SKR AlicloudRedisCluster is created and deleted", func() {

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

		alicloudRedisClusterName := uuid.NewString()
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: alicloudRedisClusterName}))
		alicloudRedisCluster := &cloudresourcesv1beta1.AlicloudRedisCluster{}

		const authSecretName = "alicloud-cluster-auth-secret"
		authSecretLabels := map[string]string{"env": "test"}
		authSecretAnnotations := map[string]string{"note": "alicloud-cluster"}
		extraData := map[string]string{
			"discovery": "{{.host}}:{{.port}}",
		}

		By("When AlicloudRedisCluster is created", func() {
			Eventually(CreateAlicloudRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					WithName(alicloudRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisClusterRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
					WithAlicloudRedisClusterShardCount(3),
					WithAlicloudRedisClusterEngineVersion("7.0"),
					WithAlicloudRedisClusterAuthSecretName(authSecretName),
					WithAlicloudRedisClusterAuthSecretLabels(authSecretLabels),
					WithAlicloudRedisClusterAuthSecretAnnotations(authSecretAnnotations),
					WithAlicloudRedisClusterAuthSecretExtraData(extraData),
				).Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed(), "expected SKR AlicloudRedisCluster to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(WithName(alicloudRedisCluster.Status.Id)),
				).Should(Succeed())

			By("And it has KCP labels")
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(skrKymaRef.Name))
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(alicloudRedisCluster.Name))
			Expect(kcpRedisCluster.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(alicloudRedisCluster.Namespace))

			By("And spec.instance.alicloud matches SKR tier")
			Expect(kcpRedisCluster.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.logic.sharding.4g.3db.0rodb.4proxy.default"))
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.ShardCount).To(Equal(int32(3)))
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.EngineVersion).To(Equal("7.0"))
		})

		kcpDiscoveryEndpoint := "r-cluster123.redis.rds.aliyuncs.com:6379"
		kcpAuthString := uuid.NewString()

		By("When KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint(kcpDiscoveryEndpoint),
					WithRedisInstanceAuthString(kcpAuthString),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
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
						WithNamespace(alicloudRedisCluster.Namespace),
					),
					HavingLabelKeys(
						util.WellKnownK8sLabelComponent,
						util.WellKnownK8sLabelPartOf,
						util.WellKnownK8sLabelManagedBy,
					),
					HavingLabel(cloudresourcesv1beta1.LabelRedisClusterStatusId, alicloudRedisCluster.Status.Id),
					HavingLabels(authSecretLabels),
					HavingAnnotations(authSecretAnnotations),
				).Should(Succeed())

			Expect(authSecret.Data).To(HaveKey("discoveryEndpoint"))
			Expect(authSecret.Data).To(HaveKey("host"))
			Expect(authSecret.Data).To(HaveKey("port"))
			Expect(authSecret.Data).To(HaveKey("authString"))
			Expect(authSecret.Data).To(HaveKeyWithValue("discovery", []byte(kcpDiscoveryEndpoint)))
		})

		// DELETE

		By("When AlicloudRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisCluster redisTier is changed", func() {

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

		alicloudRedisClusterName := uuid.NewString()
		alicloudRedisCluster := &cloudresourcesv1beta1.AlicloudRedisCluster{}
		const updateAuthSecretName = "alicloud-cluster-update-auth"

		By("When AlicloudRedisCluster is created with tier C3 and shardCount=2", func() {
			Eventually(CreateAlicloudRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					WithName(alicloudRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisClusterRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
					WithAlicloudRedisClusterShardCount(2),
					WithAlicloudRedisClusterEngineVersion("7.0"),
					WithAlicloudRedisClusterAuthSecretName(updateAuthSecretName),
				).Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created with C3 spec", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(WithName(alicloudRedisCluster.Status.Id)),
				).Should(Succeed())

			Expect(kcpRedisCluster.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.logic.sharding.4g.2db.0rodb.4proxy.default"))
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.ShardCount).To(Equal(int32(2)))
		})

		By("And Given KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("r-modify-cluster-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("And Given SKR AlicloudRedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		By("When redisTier is changed to C4", func() {
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(alicloudRedisCluster), alicloudRedisCluster); err != nil {
					return err
				}
				alicloudRedisCluster.Spec.RedisTier = cloudresourcesv1beta1.AlicloudRedisClusterTierC4
				return infra.SKR().Client().Update(infra.Ctx(), alicloudRedisCluster)
			}).Should(Succeed())
		})

		By("Then KCP RedisCluster spec is updated with C4 instanceClass", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(),
					HavingFieldValue("redis.logic.sharding.8g.2db.0rodb.4proxy.default", "spec", "instance", "alicloud", "instanceClass"),
					HavingFieldValue(int32(2), "spec", "instance", "alicloud", "shardCount"),
				).Should(Succeed())
		})

		authSecretAfterUpdate := &corev1.Secret{}
		By("And Then auth Secret still exists with valid content after tier change", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecretAfterUpdate,
					NewObjActions(
						WithName("alicloud-cluster-update-auth"),
						WithNamespace(alicloudRedisCluster.Namespace),
					),
				).Should(Succeed())
			Expect(authSecretAfterUpdate.Data).To(HaveKey("discoveryEndpoint"))
			Expect(authSecretAfterUpdate.Data).To(HaveKey("host"))
			Expect(authSecretAfterUpdate.Data).To(HaveKey("port"))
			Expect(authSecretAfterUpdate.Data).To(HaveKey("authString"))
		})

		// DELETE

		By("When AlicloudRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisCluster parameters are propagated and cleared", func() {

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

		alicloudRedisClusterName := uuid.NewString()
		alicloudRedisCluster := &cloudresourcesv1beta1.AlicloudRedisCluster{}
		initialParams := map[string]string{"maxmemory-policy": "allkeys-lru"}

		By("When AlicloudRedisCluster is created with parameters", func() {
			Eventually(CreateAlicloudRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					WithName(alicloudRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisClusterRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
					WithAlicloudRedisClusterShardCount(2),
					WithAlicloudRedisClusterEngineVersion("7.0"),
					WithAlicloudRedisClusterParameters(initialParams),
				).Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created and parameters are propagated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(WithName(alicloudRedisCluster.Status.Id)),
				).Should(Succeed())

			Expect(kcpRedisCluster.Spec.Instance.Alicloud).NotTo(BeNil())
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.Parameters).To(HaveKeyWithValue("maxmemory-policy", "allkeys-lru"))
		})

		By("And Given KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("r-params-cluster-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("And Given SKR AlicloudRedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		By("When parameters are cleared", func() {
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(alicloudRedisCluster), alicloudRedisCluster); err != nil {
					return err
				}
				alicloudRedisCluster.Spec.Parameters = nil
				return infra.SKR().Client().Update(infra.Ctx(), alicloudRedisCluster)
			}).Should(Succeed())
		})

		By("Then KCP RedisCluster parameters are cleared", func() {
			Eventually(func() bool {
				if err := infra.KCP().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(kcpRedisCluster), kcpRedisCluster); err != nil {
					return false
				}
				return len(kcpRedisCluster.Spec.Instance.Alicloud.Parameters) == 0
			}).Should(BeTrue(), "expected KCP RedisCluster parameters to be cleared")
		})

		// DELETE

		By("When AlicloudRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisCluster reflects Updating condition from KCP", func() {

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

		alicloudRedisClusterName := uuid.NewString()
		alicloudRedisCluster := &cloudresourcesv1beta1.AlicloudRedisCluster{}

		By("When AlicloudRedisCluster is created", func() {
			Eventually(CreateAlicloudRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					WithName(alicloudRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisClusterRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
					WithAlicloudRedisClusterShardCount(2),
					WithAlicloudRedisClusterEngineVersion("7.0"),
				).Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(WithName(alicloudRedisCluster.Status.Id)),
				).Should(Succeed())
		})

		By("When KCP RedisCluster has Updating condition (alongside Ready)", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithConditions(
						KcpReadyCondition(),
						metav1.Condition{
							Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
							Status:  metav1.ConditionTrue,
							Reason:  cloudcontrolv1beta1.ConditionTypeUpdating,
							Message: "Cluster is updating",
						},
					),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster reflects StateUpdating", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeUpdating),
					HavingFieldValue(cloudresourcesv1beta1.StateUpdating, "status", "state"),
				).Should(Succeed())
		})

		By("When KCP RedisCluster transitions to Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("r-updating-cluster-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster transitions to Ready and Updating condition is removed", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingFieldValue(cloudresourcesv1beta1.StateReady, "status", "state"),
				).Should(Succeed())
		})

		// DELETE

		By("When AlicloudRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AlicloudRedisCluster shardCount is changed", func() {

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

		alicloudRedisClusterName := uuid.NewString()
		alicloudRedisCluster := &cloudresourcesv1beta1.AlicloudRedisCluster{}

		By("When AlicloudRedisCluster is created with tier C3 and shardCount=2", func() {
			Eventually(CreateAlicloudRedisCluster).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					WithName(alicloudRedisClusterName),
					WithIpRange(skrIpRange.Name),
					WithAlicloudRedisClusterRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
					WithAlicloudRedisClusterShardCount(2),
					WithAlicloudRedisClusterEngineVersion("7.0"),
				).Should(Succeed())
		})

		kcpRedisCluster := &cloudcontrolv1beta1.RedisCluster{}

		By("Then KCP RedisCluster is created with shardCount=2", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingFieldSet("status", "id"),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(WithName(alicloudRedisCluster.Status.Id)),
				).Should(Succeed())

			Expect(kcpRedisCluster.Spec.Instance.Alicloud.InstanceClass).To(Equal("redis.logic.sharding.4g.2db.0rodb.4proxy.default"))
			Expect(kcpRedisCluster.Spec.Instance.Alicloud.ShardCount).To(Equal(int32(2)))
		})

		By("And Given KCP RedisCluster has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("r-shardcount-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(uuid.NewString()),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("And Given SKR AlicloudRedisCluster is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		By("When shardCount is changed to 4", func() {
			Eventually(func() error {
				if err := infra.SKR().Client().Get(infra.Ctx(),
					client.ObjectKeyFromObject(alicloudRedisCluster), alicloudRedisCluster); err != nil {
					return err
				}
				alicloudRedisCluster.Spec.ShardCount = 4
				return infra.SKR().Client().Update(infra.Ctx(), alicloudRedisCluster)
			}).Should(Succeed())
		})

		By("Then KCP RedisCluster spec is updated with shardCount=4 and matching instanceClass", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					NewObjActions(),
					HavingFieldValue("redis.logic.sharding.4g.4db.0rodb.4proxy.default", "spec", "instance", "alicloud", "instanceClass"),
					HavingFieldValue(int32(4), "spec", "instance", "alicloud", "shardCount"),
				).Should(Succeed())
		})

		By("And Given KCP RedisCluster is Ready after shardCount change", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisCluster,
					WithRedisInstanceDiscoveryEndpoint("r-shardcount-test.redis.rds.aliyuncs.com:6379"),
					WithRedisInstanceAuthString(kcpRedisCluster.Status.AuthString),
					WithConditions(KcpReadyCondition()),
				).Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster returns to Ready after shardCount change", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed())
		})

		// DELETE

		By("When AlicloudRedisCluster is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})

		By("Then SKR AlicloudRedisCluster does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), alicloudRedisCluster).
				Should(Succeed())
		})
	})
})
