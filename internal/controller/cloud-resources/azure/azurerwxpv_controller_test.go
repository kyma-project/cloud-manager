package azure

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR Azure CSI PVC Deletion", func() {

	skrRwxVolumeName := "azure-rwx-pv-deletion-test"
	fileShareName := "file-share-01"
	volumeHandle := "azure-rg-01#azure-storage-account-01#file-share-01###default"
	pv := &corev1.PersistentVolume{}
	scope := &cloudcontrolv1beta1.Scope{}

	BeforeEach(func() {
		//Disable the test case if the feature is not enabled.
		if !feature.FFNukeBackupsAzure.Value(context.Background()) {
			Skip("PV Reconciler for Azure is disabled")
		}

		By("Given KCP Scope exists", func() {

			// Given Scope exists
			Eventually(GivenScopeAzureExists).
				WithArguments(
					infra.Ctx(), infra, scope,
					WithName(infra.SkrKymaRef().Name),
				).
				Should(Succeed())
		})
		By("And Given Scope is in Ready state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					NewObjActions(),
				).
				Should(Succeed())

			//Update KCP Scope status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR PV exists", func() {
			//Create PV if it doesn't exist.
			Eventually(GivenPvExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pv,
					WithName(skrRwxVolumeName),
					WithPvCapacity("1Gi"),
					WithPvAccessMode(corev1.ReadWriteMany),
					WithPvCsiSource(&corev1.CSIPersistentVolumeSource{
						Driver:       "file.csi.azure.com",
						VolumeHandle: fileShareName,
					}),
					WithPvVolumeHandle(volumeHandle),
					WithPVReclaimPolicy(corev1.PersistentVolumeReclaimDelete),
					WithPvLabel(cloudresourcesv1beta1.LabelCloudManaged, "true"),
				).
				Should(Succeed())
		})
		By("And Given PV is in Available state", func() {

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pv,
					NewObjActions(), HavePvPhase(corev1.VolumeAvailable),
				).
				Should(Succeed())
		})

		clientProvider := infra.AzureMock().RwxPvProvider()
		subscriptionId, tenantId := scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId
		rwxPvClient, _ := clientProvider(infra.Ctx(), "", "", subscriptionId, tenantId, "")
		By(" And Given Azure FileShare exits for the same scope", func() {

			err := rwxPvClient.CreateFileShare(infra.Ctx(), volumeHandle)
			Expect(err).
				ShouldNot(HaveOccurred(), "failed to create Azure File Share ")

			fileShare, err := rwxPvClient.GetFileShare(infra.Ctx(), volumeHandle)
			Expect(fileShare).NotTo(BeNil())
			Expect(err).ShouldNot(HaveOccurred(), "failed to get Azure File Share")
		})

	})

	Describe("Scenario: SKR Azure PVC - Delete", func() {

		It("When delete is called on Azure RWX PVC", func() {

			//Disable the test case if the feature is not enabled.
			if !feature.FFNukeBackupsAzure.Value(context.Background()) {
				Skip("PV Reconciler for Azure is disabled")
			}

			azureconfig.AzureConfig.AzureFileShareDeletionWaitDuration = 1 * time.Millisecond
			By("And Then PV in SKR goes to Released State.", func() {
				Eventually(UpdatePvPhase).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), pv,
						corev1.VolumeReleased,
					).
					Should(Succeed())
			})
			clientProvider := infra.AzureMock().RwxPvProvider()
			subscriptionId, tenantId := scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId
			rwxPvClient, _ := clientProvider(infra.Ctx(), "", "", subscriptionId, tenantId, "")
			By(" And Given Azure FileShare is deleted", func() {

				time.Sleep(time.Second)
				fileShare, err := rwxPvClient.GetFileShare(infra.Ctx(), volumeHandle)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileShare).To(BeNil())
			})
		})
	})
})
