package azure

import (
	//cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	//corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AzureRwxVolumeBackup", func() {

	//skrRwxVolumeName := "azure-rwx-pv-for-backup"
	//pv := &corev1.PersistentVolume{}
	skrRwxVolumeClaimName := "azure-rwx-pvc-for-backup"
	//pvc := &corev1.PersistentVolumeClaim{}
	//scope := &cloudcontrolv1beta1.Scope{}
	skrRwxVolumeBackupName := "azure-rwx-backup-for-backup"
	backup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}

	Describe("Scenario: SKR AzureRwxVolumeBackup - Create", func() {

		It("When AzureRwxVolumeBackup Create is called", func() {
			Eventually(CreateAzureRwxVolumeBackup).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					backup,
					WithName(skrRwxVolumeBackupName),
					WithSourcePvc(skrRwxVolumeClaimName, DefaultSkrNamespace),
					// TODO: check if additional options need to be passed in
				).
				Should(Succeed())

			By("Then AzureRwxVolumeBackup is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(),
						infra.SKR().Client(),
						backup,
						NewObjActions(),
						// TODO: check if additional parameters need to be passed in
					).
					Should(Succeed())
			})

			By("And Then the AzureRwxVolumeBackup has Done status", func() {
				Expect(backup.Status.State).To(Equal(cloudresourcesv1beta1.AzureRwxBackupDone))
			})

		})

	})

})
