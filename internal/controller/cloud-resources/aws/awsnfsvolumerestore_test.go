package aws

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	skrawsnfsbackup "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AwsNfsVolumeRestore", func() {

	//Define variables
	scope := &cloudcontrolv1beta1.Scope{}
	skrAwsNfsVolumeName := "restore-aws-nfs-volume-01"
	skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
	skrAwsNfsVolumeBackupName := "restore-aws-nfs-volume-backup-01"
	skrAwsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

	awsAccountId := "382492127"

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {
			// Given Scope exists
			Eventually(CreateScopeAws).
				WithArguments(
					infra.Ctx(), infra, scope, awsAccountId,
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

		By("And Given SKR AwsNfsVolume exists", func() {
			skrawsnfsvol.Ignore.AddName(skrAwsNfsVolumeName)
			//Create SKR AwsNfsVolume if it doesn't exist.
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithName(skrAwsNfsVolumeName),
					WithAwsNfsVolumeCapacity("1Gi"),
				).
				Should(Succeed())
		})
		By("And Given AwsNfsVolume is in Ready state", func() {

			//Update SKR AwsNfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})
		By("And Given SKR AwsNfsVolumeBackup exists", func() {
			skrawsnfsbackup.Ignore.AddName(skrAwsNfsVolumeBackupName)
			//Create SKR AwsNfsVolume if it doesn't exist.
			Eventually(GivenAwsNfsVolumeBackupExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithName(skrAwsNfsVolumeBackupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
				).
				Should(Succeed())
		})
		By("And Given AwsNfsVolumeBackup is in Ready state", func() {

			//Update SKR AwsNfsVolumeBackup status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR AwsNfsVolumeRestore is created", func() {
		//Define variables.
		awsNfsVolumeRestore := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
		awsNfsVolumeRestoreName := "restore-aws-nfs-volume-restore-01"

		It("When AwsNfsVolumeRestore Create is called", func() {
			Eventually(CreateAwsNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					WithName(awsNfsVolumeRestoreName),
					WithAwsNfsVolumeBackup(skrAwsNfsVolumeBackupName),
				).
				Should(Succeed())

			By("And Then AwsNfsVolumeRestore has Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR AwsNfsVolumeRestore is deleted", Ordered, func() {
		//Define variables.
		awsNfsVolumeRestore := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}
		awsNfsVolumeRestoreName := "restore-aws-nfs-volume-restore-02"

		BeforeEach(func() {
			By("And Given SKR AwsNfsVolumeRestore has Ready condition", func() {

				//Create AwsNfsVolume
				Eventually(CreateAwsNfsVolumeRestore).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
						WithName(awsNfsVolumeRestoreName),
						WithAwsNfsVolumeBackup(skrAwsNfsVolumeBackupName),
					).
					Should(Succeed())

				//Load SKR AwsNfsVolumeRestore and check for Ready condition
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertAwsNfsVolumeRestoreHasState(cloudresourcesv1beta1.JobStateDone),
					).
					Should(Succeed())
			})
		})
		It("When SKR AwsNfsVolumeRestore Delete is called ", func() {

			//Delete SKR AwsNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					NewObjActions(),
					HaveDeletionTimestamp(),
				).
				Should(SucceedIgnoreNotFound())

			By("And Then the AwsNfsVolumeRestore in SKR is deleted.", func() {
				Eventually(IsDeleted).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					).
					Should(Succeed())
			})
		})
	})
})
