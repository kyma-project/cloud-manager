package cloudresources

import (
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureRedisCluster Volume→AuthSecret Migration", func() {

	It("Scenario: AzureRedisCluster with deprecated volume field is migrated to authSecret", func() {

		const (
			redisClusterName = "test-migration-cluster"
			secretName       = "test-migration-secret"
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

		By("Then the reconciler migrates 'volume' to 'authSecret'", func() {
			Eventually(func() error {
				return infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisClusterName), azureRedisCluster)
			}).
				WithTimeout(5 * time.Second).
				WithPolling(200 * time.Millisecond).
				Should(Succeed())

			Eventually(func() bool {
				_ = infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisClusterName), azureRedisCluster)
				return azureRedisCluster.Spec.AuthSecret != nil && azureRedisCluster.Spec.Volume == nil
			}).
				WithTimeout(5*time.Second).
				WithPolling(200*time.Millisecond).
				Should(BeTrue(), "Expected Volume to be migrated to AuthSecret")

			Expect(azureRedisCluster.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisCluster.Spec.AuthSecret.Name).To(Equal(secretName))
			Expect(azureRedisCluster.Spec.AuthSecret.Labels).To(HaveKeyWithValue("migrated", "true"))
			Expect(azureRedisCluster.Spec.AuthSecret.Annotations).To(HaveKeyWithValue("test", "migration"))
			Expect(azureRedisCluster.Spec.AuthSecret.ExtraData).To(HaveKeyWithValue("hostname", "{{ .host }}"))
			Expect(azureRedisCluster.Spec.AuthSecret.ExtraData).To(HaveKeyWithValue("password", "{{ .authString }}"))
			Expect(azureRedisCluster.Spec.Volume).To(BeNil())
		})
	})

	It("Scenario: AzureRedisCluster with both volume and authSecret prefers authSecret", func() {

		const (
			redisClusterName = "test-conflict-cluster"
			correctSecret    = "correct-auth-secret"
			wrongSecret      = "wrong-volume-secret"
		)

		azureRedisCluster := &cloudresourcesv1beta1.AzureRedisCluster{}

		By("When AzureRedisCluster is created with both 'volume' and 'authSecret' fields", func() {
			Eventually(CreateAzureRedisCluster).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureRedisCluster,
					WithName(redisClusterName),
					WithAzureRedisClusterRedisTier(cloudresourcesv1beta1.AzureRedisTierC3),
					WithAzureRedisClusterAuthSecretName(correctSecret),
					WithAzureRedisClusterVolume(&cloudresourcesv1beta1.RedisAuthSecretSpec{
						Name: wrongSecret,
					}),
				).
				Should(Succeed())
		})

		By("Then the reconciler prefers 'authSecret' and clears 'volume'", func() {
			Eventually(func() bool {
				_ = infra.SKR().Client().Get(infra.Ctx(), infra.SKR().ObjKey(redisClusterName), azureRedisCluster)
				return azureRedisCluster.Spec.Volume == nil
			}).
				WithTimeout(5*time.Second).
				WithPolling(200*time.Millisecond).
				Should(BeTrue(), "Expected Volume to be cleared when AuthSecret is present")

			Expect(azureRedisCluster.Spec.AuthSecret).NotTo(BeNil())
			Expect(azureRedisCluster.Spec.AuthSecret.Name).To(Equal(correctSecret))
			Expect(azureRedisCluster.Spec.Volume).To(BeNil())
		})
	})
})
