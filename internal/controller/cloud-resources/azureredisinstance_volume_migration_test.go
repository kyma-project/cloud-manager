package cloudresources

import (
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureRedisInstance Volume→AuthSecret Migration", func() {

	It("Scenario: AzureRedisInstance with deprecated volume field works (backward compatibility)", func() {

		const (
			redisInstanceName = "test-backward-compat-instance"
			secretName        = "test-compat-secret"
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

		By("Then the spec.volume remains unchanged (ArgoCD-safe)", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisInstanceName), azureRedisInstance)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			// Verify the spec is NOT modified by the controller (ArgoCD-safe)
			Expect(azureRedisInstance.Spec.Volume).NotTo(BeNil(), "spec.volume should remain set for backward compatibility")
			Expect(azureRedisInstance.Spec.Volume.Name).To(Equal(secretName))
			Expect(azureRedisInstance.Spec.Volume.Labels).To(HaveKeyWithValue("migrated", "true"))
			Expect(azureRedisInstance.Spec.Volume.Annotations).To(HaveKeyWithValue("test", "migration"))
			Expect(azureRedisInstance.Spec.Volume.ExtraData).To(HaveKeyWithValue("hostname", "{{ .host }}"))
			Expect(azureRedisInstance.Spec.Volume.ExtraData).To(HaveKeyWithValue("password", "{{ .authString }}"))
			Expect(azureRedisInstance.Spec.Volume.ExtraData).To(HaveKeyWithValue("tls", "true"))

			// The controller should NOT populate authSecret from volume (non-destructive)
			Expect(azureRedisInstance.Spec.AuthSecret).To(BeNil(), "spec.authSecret should remain nil (non-destructive migration)")
		})
	})

	It("Scenario: AzureRedisInstance with authSecret field works", func() {

		const (
			redisInstanceName = "test-new-field-instance"
			secretName        = "test-new-secret"
		)

		azureRedisInstance := &cloudresourcesv1beta1.AzureRedisInstance{}

		By("When AzureRedisInstance is created with 'authSecret' field", func() {
			Eventually(CreateAzureRedisInstance).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisInstance,
					WithName(redisInstanceName),
					WithAzureRedisInstanceRedisTier(cloudresourcesv1beta1.AzureRedisTierP1),
					WithAzureRedisInstanceAuthSecretName(secretName),
				).
				Should(Succeed())
		})

		By("Then the spec.authSecret remains unchanged", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisInstanceName), azureRedisInstance)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			Expect(azureRedisInstance.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisInstance.Spec.AuthSecret.Name).To(Equal(secretName))
			Expect(azureRedisInstance.Spec.Volume).To(BeNil())
		})
	})
})
