package cloudresources

import (
	"github.com/kyma-project/cloud-manager/api"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrazureredisinstance "github.com/kyma-project/cloud-manager/pkg/skr/azureredisinstance"
	skriprange "github.com/kyma-project/cloud-manager/pkg/skr/iprange"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Feature: SKR AzureRedisInstance", func() {

	It("Scenario: SKR AzureRedisInstance is created", func() {

		azureRedisInstanceName := "custom-redis-instance"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rt"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierP4
		azureRedisInstanceRedisConfigs := cloudresourcesv1beta1.RedisInstanceAzureConfigs{}
		azureRedisInstanceRedisConfigs.MaxClients = "5"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-custom-auth-secretname"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}
		extraData := map[string]string{
			"foo":    "bar",
			"parsed": "{{.host}}:{{.port}}",
		}

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("When AzureRedisInstance is created", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceRedisVersion(redisVersion),
					WithAzureRedisInstanceRedisTier(tier),
					WithAzureRedisInstanceRedisConfigs(azureRedisInstanceRedisConfigs),
					WithAzureRedisInstanceAuthSecretName(authSecretName),
					WithAzureRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAzureRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
					WithAzureRedisInstanceAuthSecretExtraData(extraData),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("Then KCP RedisInstance is created", func() {
			// load SKR AzureRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingAzureRedisInstanceStatusId(),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(azureRedisInstance.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureRedisInstance.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureRedisInstance.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisInstance.Spec.RemoteRef.Namespace).To(Equal(azureRedisInstance.Namespace))
			Expect(kcpRedisInstance.Spec.RemoteRef.Name).To(Equal(azureRedisInstance.Name))

			By("And has spec.instance.azure equal to SKR AzureRedisInstance.spec values")
			redisSKUFamily, redisSKUCapacity, _ := skrazureredisinstance.RedisTierToSKUCapacityConverter(azureRedisInstance.Spec.RedisTier)
			Expect(kcpRedisInstance.Spec.Instance.Azure.SKU.Capacity).To(Equal(redisSKUCapacity))
			Expect(kcpRedisInstance.Spec.Instance.Azure.SKU.Family).To(Equal(redisSKUFamily))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisVersion).To(Equal(azureRedisInstance.Spec.RedisVersion))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisConfiguration.MaxClients).To(Equal(azureRedisInstance.Spec.RedisConfiguration.MaxClients))
		})

		kcpRedisInstancePrimaryEndpoint := "192.168.0.1:6576"
		kcpRedisInstanceAuthString := "a9461793-2449-48d2-8618-0881bbe61d06"

		By("When KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpRedisInstancePrimaryEndpoint),
					WithRedisInstanceAuthString(kcpRedisInstanceAuthString),

					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("Then SKR AzureRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Then SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(authSecretName),
						WithNamespace(azureRedisInstance.Namespace),
					),
					HavingLabelKeys(
						util.WellKnownK8sLabelComponent,
						util.WellKnownK8sLabelPartOf,
						util.WellKnownK8sLabelManagedBy,
					),
					HavingLabel(cloudresourcesv1beta1.LabelRedisInstanceStatusId, azureRedisInstance.Status.Id),
					HavingLabels(authSecretLabels),
					HavingAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
			Expect(authSecret.Data).To(HaveKeyWithValue("parsed", []byte(kcpRedisInstancePrimaryEndpoint)), "expected auth secret data to have parsed=host:port")

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(api.CommonFinalizerDeletionHook))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisInstance is modified", func() {

		azureRedisInstanceName := "modified-redis-instance"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rt"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierP2
		azureRedisInstanceRedisConfigs := cloudresourcesv1beta1.RedisInstanceAzureConfigs{}
		azureRedisInstanceRedisConfigs.MaxClients = "5"
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-custom-auth-secretname"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("Given AzureRedisInstance exists", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceRedisVersion(redisVersion),
					WithAzureRedisInstanceRedisTier(tier),
					WithAzureRedisInstanceRedisConfigs(azureRedisInstanceRedisConfigs),
					WithAzureRedisInstanceAuthSecretName(authSecretName),
					WithAzureRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAzureRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And RedisInstance exists", func() {
			// load SKR AzureRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingAzureRedisInstanceStatusId(),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(azureRedisInstance.Status.Id),
					),
				).
				Should(Succeed())

			By("And has annotaton cloud-manager.kyma-project.io/kymaName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteName")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureRedisInstance.Name))

			By("And has annotaton cloud-manager.kyma-project.io/remoteNamespace")
			Expect(kcpRedisInstance.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureRedisInstance.Namespace))

			By("And has spec.scope.name equal to SKR Cluster kyma name")
			Expect(kcpRedisInstance.Spec.Scope.Name).To(Equal(infra.SkrKymaRef().Name))

			By("And has spec.remoteRef matching to to SKR IpRange")
			Expect(kcpRedisInstance.Spec.RemoteRef.Namespace).To(Equal(azureRedisInstance.Namespace))
			Expect(kcpRedisInstance.Spec.RemoteRef.Name).To(Equal(azureRedisInstance.Name))

			By("And has spec.instance.azure equal to SKR AzureRedisInstance.spec values")
			redisSKUFamily, redisSKUCapacity, _ := skrazureredisinstance.RedisTierToSKUCapacityConverter(azureRedisInstance.Spec.RedisTier)
			Expect(kcpRedisInstance.Spec.Instance.Azure.SKU.Capacity).To(Equal(redisSKUCapacity))
			Expect(kcpRedisInstance.Spec.Instance.Azure.SKU.Family).To(Equal(redisSKUFamily))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisVersion).To(Equal(azureRedisInstance.Spec.RedisVersion))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisConfiguration.MaxClients).To(Equal(azureRedisInstance.Spec.RedisConfiguration.MaxClients))
		})

		tier = cloudresourcesv1beta1.AzureRedisTierP1

		By("When AzureRedisInstance is modified", func() {
			Eventually(UpdateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithAzureRedisInstanceRedisTier(tier),
				).
				Should(Succeed())
		})

		By("And AzureRedsiInstance SKU.Capacity has modified value")
		Expect(azureRedisInstance.Spec.RedisTier).To(Equal(tier))

		By("Then KCP RedisInstance SKU.Capacity is modified", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(azureRedisInstance.Status.Id),
					),
				).
				Should(Succeed())

			By("And KCP RedisInstance SKU.Capacity has modified value")
			Expect(azureRedisInstance.Spec.RedisTier).To(Equal(tier))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisInstance is deleted", func() {

		azureRedisInstanceName := "another-azure-redis-instance"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7rcr"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		tier := cloudresourcesv1beta1.AzureRedisTierP4
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("Given AzureRedisInstance is created", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceRedisTier(tier),
					WithAzureRedisInstanceRedisVersion("6.0"),
				).
				Should(Succeed())
		})

		By("Then default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("When default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given KCP RedisInstance is created", func() {
			// load SKR AzureRedisInstance to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingAzureRedisInstanceStatusId(),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed(), "expected SKR AzureRedisInstance to get status.id")

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(azureRedisInstance.Status.Id),
					),
				).
				Should(Succeed(), "expected KCP RedisInstance to be created, but it was not")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, AddFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed adding finalizer on KCP RedisInstance")
		})

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed setting KCP RedisInstance Ready condition")
		})

		By("And Given SKR AzureRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed(), "expected AzureRedisInstance to exist and have Ready condition")
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(azureRedisInstance.Name),
						WithNamespace(azureRedisInstance.Namespace),
					),
				).
				Should(Succeed(), "failed creating auth Secret")
		})

		// DELETE START HERE

		By("When AzureRedisInstance is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
				Should(Succeed(), "failed deleting AzureRedisInstance")
		})

		By("Then SKR AzureRedisInstance has Deleting state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.StateDeleting),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateDeleting),
				).
				Should(Succeed(), "expected AzureRedisInstance to have Deleting state")
		})

		By("And Then SKR auth Secret is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), authSecret).
				Should(Succeed(), "expected authSecret not to exist")
		})

		By("And Then KCP RedisInstance is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP RedisInstance to be marked for deletion")
		})

		By("When KCP RedisInstance finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, RemoveFinalizer(api.CommonFinalizerDeletionHook)).
				Should(Succeed(), "failed removing finalizer on KCP RedisInstance")
		})

		By("Then SKR AzureRedisInstance is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
				Should(Succeed(), "expected AzureRedisInstance not to exist")
		})

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

	It("Scenario: SKR AzureRedisInstance authSecret is modified", func() {
		azureRedisInstanceName := "auth-secret-modified-redis"
		skrIpRangeId := "5c70629f-a13f-4b04-af47-1ab274c1c7as"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		redisVersion := "6.0"
		tier := cloudresourcesv1beta1.AzureRedisTierP1
		skrIpRange := &cloudresourcesv1beta1.IpRange{}

		skriprange.Ignore.AddName("default")

		const (
			authSecretName = "azure-auth-secret-test"
		)
		authSecretLabels := map[string]string{
			"env": "test",
		}
		authSecretAnnotations := map[string]string{
			"purpose": "testing",
		}

		By("Given default SKR IpRange does not exist", func() {
			Consistently(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				ShouldNot(Succeed())
		})

		By("And Given AzureRedisInstance is created with initial authSecret config", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceRedisVersion(redisVersion),
					WithAzureRedisInstanceRedisTier(tier),
					WithAzureRedisInstanceAuthSecretName(authSecretName),
					WithAzureRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAzureRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
				).
				Should(Succeed())
		})

		By("And Given default SKR IpRange is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange,
					NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Given default SKR IpRange has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrIpRange,
					WithSkrIpRangeStatusId(skrIpRangeId),
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		kcpRedisInstance := &cloudcontrolv1beta1.RedisInstance{}

		By("And Given KCP RedisInstance is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingAzureRedisInstanceStatusId(),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateCreating),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					NewObjActions(
						WithName(azureRedisInstance.Status.Id),
					),
				).
				Should(Succeed())
		})

		kcpRedisInstancePrimaryEndpoint := "10.0.0.1:6379"
		kcpRedisInstanceAuthString := "initial-auth-string-12345"

		By("And Given KCP RedisInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpRedisInstance,
					WithRedisInstancePrimaryEndpoint(kcpRedisInstancePrimaryEndpoint),
					WithRedisInstanceAuthString(kcpRedisInstanceAuthString),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AzureRedisInstance has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					HavingAzureRedisInstanceStatusState(cloudresourcesv1beta1.StateReady),
				).
				Should(Succeed())
		})

		authSecret := &corev1.Secret{}
		By("And Given SKR auth Secret is created with initial values", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					authSecret,
					NewObjActions(
						WithName(authSecretName),
						WithNamespace(azureRedisInstance.Namespace),
					),
					HavingLabel("env", "test"),
					HavingAnnotation("purpose", "testing"),
				).
				Should(Succeed())
		})

		newLabels := map[string]string{
			"env":  "production",
			"team": "platform",
		}
		newAnnotations := map[string]string{
			"purpose":     "production-testing",
			"cost-center": "12345",
		}
		newExtraData := map[string]string{
			"custom-key": "custom-value",
			"endpoint":   "{{.host}}:{{.port}}",
		}

		By("When AzureRedisInstance authSecret config is modified with new labels, annotations, and extraData", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureRedisInstance,
					NewObjActions(),
				).
				Should(Succeed())

			azureRedisInstance.Spec.AuthSecret.Labels = newLabels
			azureRedisInstance.Spec.AuthSecret.Annotations = newAnnotations

			Eventually(UpdateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithAzureRedisInstanceAuthSecretExtraData(newExtraData),
				).
				Should(Succeed())
		})

		By("Then SKR auth Secret is updated with new labels (REPLACEMENT semantics)", func() {
			Eventually(func() map[string]string {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisInstance.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				userLabels := map[string]string{}
				for k, v := range authSecret.Labels {
					if k == "env" || k == "team" {
						userLabels[k] = v
					}
				}
				return userLabels
			}).Should(And(
				HaveKeyWithValue("env", "production"),
				HaveKeyWithValue("team", "platform"),
				HaveLen(2),
			), "expected auth Secret to have updated labels with replacement semantics")

			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelRedisInstanceStatusId))
			Expect(authSecret.Labels).To(HaveKey(cloudresourcesv1beta1.LabelCloudManaged))
		})

		By("And Then SKR auth Secret has new annotations (REPLACEMENT semantics)", func() {
			Eventually(func() map[string]string {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisInstance.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				return authSecret.Annotations
			}).Should(And(
				HaveKeyWithValue("purpose", "production-testing"),
				HaveKeyWithValue("cost-center", "12345"),
				HaveLen(2),
			), "expected auth Secret to have updated annotations with replacement semantics")
		})

		By("And Then auth Secret data includes extraData fields with proper templating", func() {
			Eventually(func() map[string][]byte {
				secretKey := types.NamespacedName{Name: authSecretName, Namespace: azureRedisInstance.Namespace}
				err := infra.SKR().Client().Get(infra.Ctx(), secretKey, authSecret)
				if err != nil {
					return nil
				}
				return authSecret.Data
			}).Should(And(
				HaveKeyWithValue("custom-key", []byte("custom-value")),
				HaveKeyWithValue("endpoint", []byte(kcpRedisInstancePrimaryEndpoint)),
				HaveKey("host"),
				HaveKey("port"),
				HaveKey("authString"),
			), "expected auth Secret to have extraData fields with proper Go templating")
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
			Should(Succeed())

		By("// cleanup: delete default SKR IpRange", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrIpRange).
				Should(Succeed())
		})
	})

})
