package aws

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	skrawsnfsbackup "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR AwsNfsVolumeRestore", func() {

	const (
		awsAccountId = "382492127"
	)

	It("Scenario: Creates AwsNfsVolumeRestore", func() {
		suffix := "a1b2c3d4-e5f6-47a8-9b0c-1d2e3f4a5b6c"
		skrAwsNfsVolumeName := suffix
		skrAwsNfsVolumeBackupName := suffix
		restoreName := suffix

		scope := &cloudcontrolv1beta1.Scope{}
		skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		skrAwsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
		awsNfsVolumeRestore := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}

		// Stop reconciliation to prevent interference
		skrawsnfsvol.Ignore.AddName(skrAwsNfsVolumeName)
		skrawsnfsbackup.Ignore.AddName(skrAwsNfsVolumeBackupName)

		By("Given KCP Scope exists", func() {
			Expect(client.IgnoreAlreadyExists(
				CreateScopeAws(infra.Ctx(), infra, scope, awsAccountId, WithName(infra.SkrKymaRef().Name)))).
				To(Succeed())
		})

		By("And Given Scope has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope,
					NewObjActions(),
				).
				Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolume exists", func() {
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithName(skrAwsNfsVolumeName),
					WithAwsNfsVolumeCapacity("1Gi"),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolume has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolumeBackup exists", func() {
			Eventually(GivenAwsNfsVolumeBackupExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithName(skrAwsNfsVolumeBackupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("When AwsNfsVolumeRestore Create is called", func() {
			Eventually(CreateAwsNfsVolumeRestore).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					WithName(restoreName),
					WithAwsNfsVolumeBackup(skrAwsNfsVolumeBackupName),
				).
				Should(Succeed())
		})

		By("Then AwsNfsVolumeRestore has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("// cleanup: Delete test resources", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume).
				Should(Succeed())
		})
	})

	It("Scenario: Deletes AwsNfsVolumeRestore", func() {
		suffix := "f7e8d9c0-b1a2-43d4-95e6-7f8a9b0c1d2e"
		skrAwsNfsVolumeName := suffix
		skrAwsNfsVolumeBackupName := suffix
		restoreName := suffix

		scope := &cloudcontrolv1beta1.Scope{}
		skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		skrAwsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
		awsNfsVolumeRestore := &cloudresourcesv1beta1.AwsNfsVolumeRestore{}

		// Stop reconciliation to prevent interference
		skrawsnfsvol.Ignore.AddName(skrAwsNfsVolumeName)
		skrawsnfsbackup.Ignore.AddName(skrAwsNfsVolumeBackupName)

		By("Given KCP Scope exists", func() {
			Expect(client.IgnoreAlreadyExists(
				CreateScopeAws(infra.Ctx(), infra, scope, awsAccountId, WithName(infra.SkrKymaRef().Name)))).
				To(Succeed())
		})

		By("And Given Scope has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope,
					NewObjActions(),
				).
				Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolume exists", func() {
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithName(skrAwsNfsVolumeName),
					WithAwsNfsVolumeCapacity("1Gi"),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolume has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolumeBackup exists", func() {
			Eventually(GivenAwsNfsVolumeBackupExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithName(skrAwsNfsVolumeBackupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolumeRestore has Ready condition", func() {
			Eventually(CreateAwsNfsVolumeRestore).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					WithName(restoreName),
					WithAwsNfsVolumeBackup(skrAwsNfsVolumeBackupName),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertAwsNfsVolumeRestoreHasState(cloudresourcesv1beta1.JobStateDone),
				).
				Should(Succeed())
		})

		By("When SKR AwsNfsVolumeRestore Delete is called", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore,
					NewObjActions(),
					HaveDeletionTimestamp(),
				).
				Should(SucceedIgnoreNotFound())
		})

		By("Then the AwsNfsVolumeRestore in SKR is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeRestore).
				Should(Succeed())
		})

		By("// cleanup: Delete remaining test resources", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolumeBackup).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume).
				Should(Succeed())
		})
	})
})
