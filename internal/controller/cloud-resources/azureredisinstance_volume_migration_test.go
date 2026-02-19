package cloudresources

import (
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureRedisInstance Volume→AuthSecret Migration", func() {

	It("Scenario: AzureRedisInstance with deprecated volume field is migrated to authSecret", func() {

		const (
			redisInstanceName = "test-migration-instance"
			secretName        = "test-migration-secret"
		)

		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}

		By("When AzureRedisInstance is created with deprecated 'volume' field", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(redisInstanceName),
					WithAzureRedisInstanceRedisTier(cloudresourcesv1beta1.AzureRedisTierP1),
					WithAzureRedisInstanceVolume(&cloudresourcesv1beta1.RedisAuthSecretSpec{
						Name: secretName,
						Labels: map[string]string{
							"migrated": "true",
						},
						Annotations: map[string]string{
							"test": "migration",
						},
						ExtraData: map[string]string{
							"hostname": "{{ .host }}",
							"password": "{{ .authString }}",
							"tls":      "true",
						},
					}),
				).
				Should(Succeed())
		})

		By("Then the reconciler migrates 'volume' to 'authSecret'", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisInstanceName), azureRedisInstance)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			Eventually(func() bool {
				_ = infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisInstanceName), azureRedisInstance)
				return azureRedisInstance.Spec.AuthSecret != nil && azureRedisInstance.Spec.Volume == nil
			}).
				WithTimeout(5*time.Second).
				WithPolling(200*time.Millisecond).
				Should(BeTrue(), "Expected Volume to be migrated to AuthSecret")

			Expect(azureRedisInstance.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisInstance.Spec.AuthSecret.Name).To(Equal(secretName))
			Expect(azureRedisInstance.Spec.AuthSecret.Labels).To(HaveKeyWithValue("migrated", "true"))
			Expect(azureRedisInstance.Spec.AuthSecret.Annotations).To(HaveKeyWithValue("test", "migration"))
			Expect(azureRedisInstance.Spec.AuthSecret.ExtraData).To(HaveKeyWithValue("hostname", "{{ .host }}"))
			Expect(azureRedisInstance.Spec.AuthSecret.ExtraData).To(HaveKeyWithValue("password", "{{ .authString }}"))
			Expect(azureRedisInstance.Spec.AuthSecret.ExtraData).To(HaveKeyWithValue("tls", "true"))
			Expect(azureRedisInstance.Spec.Volume).To(BeNil())
		})
	})

	It("Scenario: AzureRedisInstance with both volume and authSecret prefers authSecret", func() {

		const (
			redisInstanceName = "test-conflict-instance"
			correctSecret     = "correct-auth-secret"
			wrongSecret       = "wrong-volume-secret"
		)

		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}

		By("When AzureRedisInstance is created with both 'volume' and 'authSecret' fields", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(redisInstanceName),
					WithAzureRedisInstanceRedisTier(cloudresourcesv1beta1.AzureRedisTierP1),
					WithAzureRedisInstanceAuthSecretName(correctSecret),
					WithAzureRedisInstanceVolume(&cloudresourcesv1beta1.RedisAuthSecretSpec{
						Name: wrongSecret,
					}),
				).
				Should(Succeed())
		})

		By("Then the reconciler prefers 'authSecret' and clears 'volume'", func() {
			// Wait for conflict resolution to happen
			Eventually(func() bool {
				_ = infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisInstanceName), azureRedisInstance)
				return azureRedisInstance.Spec.Volume == nil
			}).
				WithTimeout(5*time.Second).
				WithPolling(200*time.Millisecond).
				Should(BeTrue(), "Expected Volume to be cleared when AuthSecret is present")

			Expect(azureRedisInstance.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisInstance.Spec.AuthSecret.Name).To(Equal(correctSecret))
			Expect(azureRedisInstance.Spec.Volume).To(BeNil())
		})
	})
})
