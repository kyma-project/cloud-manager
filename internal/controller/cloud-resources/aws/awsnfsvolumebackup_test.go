package aws

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR AwsNfsVolumeBackup", func() {

	const (
		awsAccountId = "485392126"
	)

	It("Scenario: Creates AwsNfsVolumeBackup", func() {
		suffix := "b7c9e4a1-8f5d-42a3-9c6e-1d7f3a2b8e4c"
		skrAwsNfsVolumeName := suffix
		nfsInstanceName := suffix
		backupName := suffix

		scope := &cloudcontrolv1beta1.Scope{}
		skrAwsNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		awsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

		// Stop reconciliation to prevent interference
		skrawsnfsvol.Ignore.AddName(skrAwsNfsVolumeName)
		nfsinstance.Ignore.AddName(nfsInstanceName)

		By("Given KCP Scope exists", func() {
			Expect(client.IgnoreAlreadyExists(
				CreateScopeAws(infra.Ctx(), infra, scope, awsAccountId, WithName(infra.SkrKymaRef().Name)))).
				To(Succeed())
		})

		By("And Given Scope has Ready condition", func() {
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

		By("And Given KCP NfsInstance exists", func() {
			Eventually(GivenNfsInstanceExists).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(nfsInstanceName),
					WithRemoteRef(skrAwsNfsVolumeName),
					WithScope(infra.SkrKymaRef().Name),
					WithIpRange(nfsInstanceName),
					WithNfsInstanceAws(),
				).
				Should(Succeed())
		})

		By("And Given NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithConditions(KcpReadyCondition()),
					WithNfsInstanceStatusId(nfsInstance.Name),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolume has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrAwsNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithAwsNfsVolumeStatusId(nfsInstanceName),
				).
				Should(Succeed())
		})

		By("When AwsNfsVolumeBackup Create is called", func() {
			Eventually(CreateAwsNfsVolumeBackup).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					WithName(backupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
				).
				Should(Succeed())
		})

		By("Then AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertAwsNfsVolumeBackupHasLocation(scope.Spec.Region),
				).
				Should(Succeed())
		})

		By("// cleanup: Delete test resources", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR AwsNfsVolumeBackup is deleted", Ordered, func() {
		//Define variables.
		awsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
		awsNfsVolumeBackupName := "aws-nfs-volume-backup-02"

		BeforeEach(func() {
			By("And Given SKR AwsNfsVolumeBackup has Ready condition", func() {

				//Create AwsNfsVolume
				Eventually(CreateAwsNfsVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
						WithName(awsNfsVolumeBackupName),
						WithAwsNfsVolume(skrAwsNfsVolumeName),
					).
					Should(Succeed())

				//Load SKR AwsNfsVolumeBackup and check for Ready condition
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertAwsNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
						AssertAwsNfsVolumeBackupHasLocation(scope.Spec.Region),
					).
					Should(Succeed())
			})
		})
		It("When SKR AwsNfsVolumeBackup Delete is called ", func() {

			//Delete SKR AwsNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in AwsNfsVolumeBackup", func() {
				Expect(awsNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the AwsNfsVolumeBackup in SKR is deleted.", func() {
				Eventually(IsDeleted).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR AwsNfsVolumeBackup with location is created", func() {
		//Define variables.
		awsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
		awsNfsVolumeBackupName := "aws-nfs-volume-backup-02"
		awsNfsVolumeBackupLocation := "us-west-1"

		It("When AwsNfsVolumeBackup Create is called", func() {
			Eventually(CreateAwsNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					WithName(awsNfsVolumeBackupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
					WithAwsNfsVolumeBackupLocation(awsNfsVolumeBackupLocation),
				).
				Should(Succeed())

			By("And Then AwsNfsVolumeBackup has Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertAwsNfsVolumeBackupHasLocation(scope.Spec.Region),
						AssertAwsNfsVolumeBackupHasLocation(awsNfsVolumeBackupLocation),
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR AwsNfsVolumeBackup with location is deleted", Ordered, func() {
		//Define variables.
		awsNfsVolumeBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}
		awsNfsVolumeBackupName := "aws-nfs-volume-backup-02"
		awsNfsVolumeBackupLocation := "us-west-1"

		BeforeEach(func() {
			By("And Given SKR AwsNfsVolumeBackup has Ready condition", func() {

				//Create AwsNfsVolume
				Eventually(CreateAwsNfsVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
						WithName(awsNfsVolumeBackupName),
						WithAwsNfsVolume(skrAwsNfsVolumeName),
						WithAwsNfsVolumeBackupLocation(awsNfsVolumeBackupLocation),
					).
					Should(Succeed())

				//Load SKR AwsNfsVolumeBackup and check for Ready condition
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertAwsNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
						AssertAwsNfsVolumeBackupHasLocation(scope.Spec.Region),
						AssertAwsNfsVolumeBackupHasLocation(awsNfsVolumeBackupLocation),
					).
					Should(Succeed())
			})
		})
		It("When SKR AwsNfsVolumeBackup Delete is called ", func() {

			//Delete SKR AwsNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in AwsNfsVolumeBackup", func() {
				Expect(awsNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the AwsNfsVolumeBackup in SKR is deleted.", func() {
				Eventually(IsDeleted).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					).
					Should(Succeed())
			})
		})
	})
})
