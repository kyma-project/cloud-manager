package cloudresources

import (
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureRedisCluster Volume→AuthSecret Migration", func() {

	It("Scenario: AzureRedisCluster with deprecated volume field works (backward compatibility)", func() {

		const (
			redisClusterName = "test-backward-compat-cluster"
			secretName       = "test-compat-secret"
		)

		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}

		By("When AzureRedisCluster is created with deprecated 'volume' field", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(redisClusterName),
					WithAzureRedisClusterRedisTier(cloudresourcesv1beta1.AzureRedisTierC3),
					WithAzureRedisClusterVolume(&cloudresourcesv1beta1.RedisAuthSecretSpec{
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
						},
					}),
				).
				Should(Succeed())
		})

		By("Then the spec.volume remains unchanged (ArgoCD-safe)", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisClusterName), azureRedisCluster)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			// Verify the spec is NOT modified by the controller (ArgoCD-safe)
			Expect(azureRedisCluster.Spec.Volume).NotTo(BeNil(), "spec.volume should remain set for backward compatibility")
			Expect(azureRedisCluster.Spec.Volume.Name).To(Equal(secretName))
			Expect(azureRedisCluster.Spec.Volume.Labels).To(HaveKeyWithValue("migrated", "true"))
			Expect(azureRedisCluster.Spec.Volume.Annotations).To(HaveKeyWithValue("test", "migration"))
			Expect(azureRedisCluster.Spec.Volume.ExtraData).To(HaveKeyWithValue("hostname", "{{ .host }}"))
			Expect(azureRedisCluster.Spec.Volume.ExtraData).To(HaveKeyWithValue("password", "{{ .authString }}"))

			// The controller should NOT populate authSecret from volume (non-destructive)
			Expect(azureRedisCluster.Spec.AuthSecret).To(BeNil(), "spec.authSecret should remain nil (non-destructive migration)")
		})
	})

	It("Scenario: AzureRedisCluster with authSecret field works", func() {

		const (
			redisClusterName = "test-new-field-cluster"
			secretName       = "test-new-secret"
		)

		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}

		By("When AzureRedisCluster is created with 'authSecret' field", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(redisClusterName),
					WithAzureRedisClusterRedisTier(cloudresourcesv1beta1.AzureRedisTierC3),
					WithAzureRedisClusterAuthSecretName(secretName),
				).
				Should(Succeed())
		})

		By("Then the spec.authSecret remains unchanged", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisClusterName), azureRedisCluster)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			Expect(azureRedisCluster.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisCluster.Spec.AuthSecret.Name).To(Equal(secretName))
			Expect(azureRedisCluster.Spec.Volume).To(BeNil())
		})
	})

	It("Scenario: CEL validation rejects setting both volume and authSecret", func() {

		const (
			redisClusterName = "test-validation-cluster"
			correctSecret    = "correct-auth-secret"
			wrongSecret      = "wrong-volume-secret"
		)

		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}

		By("When AzureRedisCluster is created with both 'volume' and 'authSecret' fields", func() {
			err := CreateAzureRedisCluster(
				infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
				WithName(redisClusterName),
				WithAzureRedisClusterRedisTier(cloudresourcesv1beta1.AzureRedisTierC3),
				WithAzureRedisClusterAuthSecretName(correctSecret),
				WithAzureRedisClusterVolume(&cloudresourcesv1beta1.RedisAuthSecretSpec{
					Name: wrongSecret,
				}),
			)

			By("Then creation is rejected by CEL validation", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Cannot set both 'volume' (deprecated) and 'authSecret' fields"))
			})
		})
	})
})
