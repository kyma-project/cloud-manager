package cloudresources

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AzureRedisInstance", func() {

	It("Scenario: SKR AzureRedisInstance is created", func() {

		azureRedisInstanceName := "custom-redis-instance"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		enableNonSslPort := false
		redisVersion := "6.0"
		sku := cloudresourcesv1beta1.AzureRedisSKU{}
		sku.Capacity = 1
		azureRedisInstanceRedisConfigs := cloudresourcesv1beta1.RedisInstanceAzureConfigs{}
		azureRedisInstanceRedisConfigs.MaxClients = "5"

		const (
			authSecretName = "azure-custom-auth-secretname"
		)
		authSecretLabels := map[string]string{
			"foo": "1",
		}
		authSecretAnnotations := map[string]string{
			"bar": "2",
		}

		By("When AzureRedisInstance is created", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceEnableNonSslPort(enableNonSslPort),
					WithAzureRedisInstanceRedisVersion(redisVersion),
					WithAzureRedisInstanceSKUCapacity(sku),
					WithAzureRedisInstanceRedisConfigs(azureRedisInstanceRedisConfigs),
					WithAzureRedisInstanceAuthSecretName(authSecretName),
					WithAzureRedisInstanceAuthSecretLabels(authSecretLabels),
					WithAzureRedisInstanceAuthSecretAnnotations(authSecretAnnotations),
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
			Expect(kcpRedisInstance.Spec.Instance.Azure.EnableNonSslPort).To(Equal(azureRedisInstance.Spec.EnableNonSslPort))
			Expect(kcpRedisInstance.Spec.Instance.Azure.SKU.Capacity).To(Equal(azureRedisInstance.Spec.SKU.Capacity))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisVersion).To(Equal(azureRedisInstance.Spec.RedisVersion))
			Expect(kcpRedisInstance.Spec.Instance.Azure.RedisConfiguration.MaxClients).To(Equal(azureRedisInstance.Spec.RedisConfiguration.MaxClients))
		})

		kcpRedisInstancePrimaryEndpoint := "192.168.0.1:6576"
		kcpRedisInstanceAuthString := "a9461793-2449-48d2-8618-0881bbe61d05"

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
				).
				Should(Succeed())

			By("And it has defined cloud-manager default labels")
			Expect(authSecret.Labels[util.WellKnownK8sLabelComponent]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelPartOf]).ToNot(BeNil())
			Expect(authSecret.Labels[util.WellKnownK8sLabelManagedBy]).ToNot(BeNil())

			By("And it has defined ownmership label")
			Expect(authSecret.Labels[cloudresourcesv1beta1.LabelRedisInstanceStatusId]).To(Equal(azureRedisInstance.Status.Id))

			By("And it has user defined custom labels")
			for k, v := range authSecretLabels {
				Expect(authSecret.Labels).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected auth Secret to have label %s=%s", k, v))
			}
			By("And it has user defined custom annotations")
			for k, v := range authSecretAnnotations {
				Expect(authSecret.Annotations).To(HaveKeyWithValue(k, v), fmt.Sprintf("expected auth Secret to have annotation %s=%s", k, v))
			}

			By("And it has defined cloud-manager finalizer")
			Expect(authSecret.Finalizers).To(ContainElement(cloudresourcesv1beta1.Finalizer))
		})

		// CleanUp
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
			Should(Succeed())
	})

	It("Scenario: SKR AzureRedisInstance is deleted", func() {

		azureRedisInstanceName := "another-azure-redis-instance"
		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}
		sku := cloudresourcesv1beta1.AzureRedisSKU{}
		sku.Capacity = 1

		By("Given AzureRedisInstance is created", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(azureRedisInstanceName),
					WithAzureRedisInstanceSKUCapacity(sku),
					WithAzureRedisInstanceEnableNonSslPort(false),
					WithAzureRedisInstanceRedisVersion("6.0"),
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, AddFinalizer(cloudcontrolv1beta1.FinalizerName)).
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpRedisInstance, RemoveFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed removing finalizer on KCP RedisInstance")
		})

		By("Then SKR AzureRedisInstance is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureRedisInstance).
				Should(Succeed(), "expected AzureRedisInstance not to exist")
		})
	})

})
