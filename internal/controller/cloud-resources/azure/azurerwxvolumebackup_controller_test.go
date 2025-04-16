package azure

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	//cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	//corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AzureRwxVolumeBackup", func() {

	skrRwxVolumeName := "azure-rwx-pv-for-backup"
	pv := &corev1.PersistentVolume{}
	skrRwxVolumeClaimName := "azure-rwx-pvc-for-backup"
	pvc := &corev1.PersistentVolumeClaim{}
	scope := &cloudcontrolv1beta1.Scope{}

	BeforeEach(func() {

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
		By("And Given SKR namespace exists", func() {
			//Create namespace if it doesn't exist.
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		By("And Given SKR PV exists", func() {
			//skrazurerwxvol.Ignore.AddName(skrRwxVolumeName)
			//Create SKR AzureRwxVolume if it doesn't exist.
			Eventually(GivenPvExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pv,
					WithName(skrRwxVolumeName),
					WithPvCapacity("1Gi"),
					WithPvAccessMode(corev1.ReadWriteMany),
					WithPvClaimRef(skrRwxVolumeClaimName, DefaultSkrNamespace),
					WithPvCsiSource(&corev1.CSIPersistentVolumeSource{
						Driver:       "file.csi.azure.com",
						VolumeHandle: "test-file-share-01",
					}),
					WithPvVolumeHandle("shoot--kyma-dev--c-6ea9b9b#f21d936aa5673444a95852a#pv-shoot-kyma-dev-c-6ea9b9b-8aa269ae-f581-427b-b05c-a2a2bbfca###default"),
					WithPvLabel(cloudresourcesv1beta1.LabelCloudManaged, "true"),
				).
				Should(Succeed())
		})
		By("And Given PV is in Ready state", func() {

			Eventually(UpdatePvPhase).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pv,
					corev1.VolumeBound,
				).
				Should(Succeed())
		})
		By("And Given SKR PVC exists", func() {
			Eventually(GivenPvcExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc,
					WithName(skrRwxVolumeClaimName),
					WithPVName(skrRwxVolumeName),
					WithPvCapacity("1Gi"),
					WithPvAccessMode(corev1.ReadWriteMany),
					WithPvLabel(cloudresourcesv1beta1.LabelCloudManaged, "true"),
					WithPvAnnotation("volume.kubernetes.io/storage-provisioner", "file.csi.azure.com"),
				).
				Should(Succeed(), "failed creating PVC")
		})
		By("And Given PVC is in Ready state", func() {

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pvc,
					NewObjActions(),
					HavePvcPhase(corev1.ClaimBound),
				).
				Should(Succeed())
		})

	})

	Describe("Scenario: SKR AzureRwxVolumeBackup - Create", func() {

		// Arrange
		backup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}
		skrRwxVolumeBackupName := "azure-rwx-backup"

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
					).
					Should(Succeed())
			})

			// TODO: this fails
			By("And Then the AzureRwxVolumeBackup has Done status", func() {
				Expect(backup.Status.State).To(Equal(cloudresourcesv1beta1.AzureRwxBackupDone))
			})

		})

	})

})
