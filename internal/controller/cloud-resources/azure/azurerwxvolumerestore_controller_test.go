/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrazurerwxvolumebackup "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AzureRwxVolumeRestore in place", func() {

	skrRwxVolumeName := "azure-rwx-pv-for-restore"
	pv := &corev1.PersistentVolume{}
	skrRwxVolumeClaimName := "azure-rwx-pvc-for-restore"
	pvc := &corev1.PersistentVolumeClaim{}
	scope := &cloudcontrolv1beta1.Scope{}
	skrRwxVolumeBackupName := "azure-rwx-backup-for-restore"
	backup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}

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
		By("And Given SKR AzureRwxVolumeBackup exists", func() {
			skrazurerwxvolumebackup.Ignore.AddName(skrRwxVolumeBackupName)
			//Create SKR AzureRwxVolumeBackup if it doesn't exist.
			Eventually(CreateAzureRwxVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), backup,
					WithName(skrRwxVolumeBackupName),
					WithSourcePvc(skrRwxVolumeClaimName, DefaultSkrNamespace),
				).
				Should(Succeed())
		})
		By("And Given AzureRwxVolumeBackup is in Ready state", func() {
			//Update SKR AzureRwxVolumeBackup status to Ready
			subscriptionId := "3f1d2fbd-117a-4742-8bde-6edbcdee6a03"
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), backup,
					WithConditions(SkrReadyCondition()),
					WithRecoveryPointId(fmt.Sprintf("/subscriptions/%s/resourceGroups/rg-test/providers/Microsoft.RecoveryServices/vaults/v-test/backupFabrics/Azure/protectionContainers/StorageContainer;Storage;test;testsa/protectedItems/AzureFileShare;2DAC3CBDBBD863B2292F25490DC0794F35AAA4C27890D5DCA82B0A33E9596217/recoveryPoints/5639661428710522320", subscriptionId)),
					WithStorageAccountPath(fmt.Sprintf("/subscriptions/%s/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/testsa", subscriptionId)),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR AzureRwxVolumeRestore - Create", func() {

		//Define variables.
		rwxVolumeRestore := &cloudresourcesv1beta1.AzureRwxVolumeRestore{}
		rwxVolumeRestoreName := "azure-rwx-volume-restore-for-restore"

		It("When AzureRwxVolumeRestore Create is called", func() {
			Eventually(CreateAzureRwxVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
					WithName(rwxVolumeRestoreName),
					WithSourceBackupRef(skrRwxVolumeBackupName, DefaultSkrNamespace),
					WithDestinationPvcRef(skrRwxVolumeClaimName, DefaultSkrNamespace),
				).
				Should(Succeed())
			By("Then AzureRwxVolumeRestore is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then AzureRwxVolumeRestore has inProgress state", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
						NewObjActions(),
						HavingState(cloudresourcesv1beta1.JobStateInProgress),
					).
					Should(Succeed())
			})
			By("And Then AzureRwxVolumeRestore has Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())
			})
			By("And Then the AzureRwxVolumeRestore has Done status", func() {
				Expect(cloudresourcesv1beta1.JobStateDone).To(Equal(rwxVolumeRestore.Status.State))
			})

		})
	})

	Describe("Scenario: SKR AzureRwxVolumeRestore - Delete", func() {

		//Define variables.
		rwxVolumeRestore := &cloudresourcesv1beta1.AzureRwxVolumeRestore{}
		rwxVolumeRestoreName := "azure-rwx-volume-restore-for-restore"

		It("And Given AzureRwxVolumeRestore exists", func() {
			Eventually(CreateAzureRwxVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
					WithName(rwxVolumeRestoreName),
					WithSourceBackupRef(skrRwxVolumeBackupName, DefaultSkrNamespace),
					WithDestinationPvcRef(skrRwxVolumeClaimName, DefaultSkrNamespace),
				).
				Should(Succeed())
			//Update SKR AzureRwxVolumeRestore to Done state
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
					WithConditions(SkrReadyCondition()),
					WithAzureRwxVolumeRestoreState(cloudresourcesv1beta1.JobStateDone),
				).
				Should(Succeed())
		})
		It("When SKR AzureRwxVolumeRestore Delete is called ", func() {

			//Delete SKR AzureRwxVolumeRestore
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
					NewObjActions(),
					HaveDeletionTimestamp(),
				).
				Should(SucceedIgnoreNotFound())

			By("And Then the AzureRwxVolumeRestore in SKR is deleted.", func() {
				Eventually(IsDeleted).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxVolumeRestore,
					).
					Should(Succeed())
			})
		})
	})

})
