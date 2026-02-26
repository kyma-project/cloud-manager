package aws

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
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

	It("Scenario: Deletes AwsNfsVolumeBackup", func() {
		suffix := "3e5f7a9b-2c4d-48e6-9a1b-7f3e5c8d2a4b"
		skrAwsNfsVolumeName := suffix
		nfsInstanceName := suffix
		backupName := suffix
		var recoveryPointArn string

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

		By("And Given SKR AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(CreateAwsNfsVolumeBackup).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					WithName(backupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertAwsNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
					AssertAwsNfsVolumeBackupHasLocation(scope.Spec.Region),
				).
				Should(Succeed())

			recoveryPointArn = awsNfsVolumeBackup.Status.Id
		})

		By("When SKR AwsNfsVolumeBackup Delete is called", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())
		})

		By("Then DeletionTimestamp is set in AwsNfsVolumeBackup", func() {
			Expect(awsNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
		})

		By("And Then the AwsNfsVolumeBackup in SKR is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())
		})

		By("And Then the backup is deleted from AWS cloud provider (mock)", func() {
			// Verify the recovery point no longer exists in the mock
			mockClient, err := awsnfsvolumebackupclient.NewMockClient()(
				infra.Ctx(),
				awsAccountId,
				scope.Spec.Region,
				"", "", "", // credentials not needed for mock
			)
			Expect(err).NotTo(HaveOccurred())

			vaultName := fmt.Sprintf("cm-%s", skrAwsNfsVolumeName)

			_, err = mockClient.DescribeRecoveryPoint(
				infra.Ctx(),
				awsAccountId,
				vaultName,
				recoveryPointArn,
			)
			Expect(err).To(HaveOccurred(), "Expected recovery point to be deleted from cloud provider")
			Expect(mockClient.IsNotFound(err)).To(BeTrue(), "Expected ResourceNotFoundException, got: %v", err)
		})
	})

	It("Scenario: Creates AwsNfsVolumeBackup with custom location", func() {
		suffix := "9d2f4e6a-8b3c-47d5-9e1a-6f4b7c9d3e5a"
		skrAwsNfsVolumeName := suffix
		nfsInstanceName := suffix
		backupName := suffix
		awsNfsVolumeBackupLocation := "us-west-1"

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

		By("When AwsNfsVolumeBackup Create is called with custom location", func() {
			Eventually(CreateAwsNfsVolumeBackup).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					WithName(backupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
					WithAwsNfsVolumeBackupLocation(awsNfsVolumeBackupLocation),
				).
				Should(Succeed())
		})

		By("Then AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertAwsNfsVolumeBackupHasLocation(awsNfsVolumeBackupLocation),
				).
				Should(Succeed())
		})

		By("// cleanup: Delete test resources", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())
		})
	})

	It("Scenario: Deletes AwsNfsVolumeBackup with custom location", func() {
		suffix := "6f8e2b4c-9d3a-45e7-8c1f-2a5b7d9e4f6c"
		skrAwsNfsVolumeName := suffix
		nfsInstanceName := suffix
		backupName := suffix
		awsNfsVolumeBackupLocation := "us-west-1"
		var recoveryPointArn string

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

		By("And Given SKR AwsNfsVolumeBackup has Ready condition", func() {
			Eventually(CreateAwsNfsVolumeBackup).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					WithName(backupName),
					WithAwsNfsVolume(skrAwsNfsVolumeName),
					WithAwsNfsVolumeBackupLocation(awsNfsVolumeBackupLocation),
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertAwsNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
					AssertAwsNfsVolumeBackupHasLocation(awsNfsVolumeBackupLocation),
				).
				Should(Succeed())

			recoveryPointArn = awsNfsVolumeBackup.Status.Id
		})

		By("When SKR AwsNfsVolumeBackup Delete is called", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())
		})

		By("Then DeletionTimestamp is set in AwsNfsVolumeBackup", func() {
			Expect(awsNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
		})

		By("And Then the AwsNfsVolumeBackup in SKR is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsNfsVolumeBackup).
				Should(Succeed())
		})

		By("And Then the backup is deleted from AWS cloud provider (mock)", func() {
			mockClient, err := awsnfsvolumebackupclient.NewMockClient()(
				infra.Ctx(), awsAccountId, awsNfsVolumeBackupLocation, "", "", "")
			Expect(err).NotTo(HaveOccurred())

			vaultName := fmt.Sprintf("cm-%s", skrAwsNfsVolumeName)

			_, err = mockClient.DescribeRecoveryPoint(
				infra.Ctx(), awsAccountId,
				vaultName,
				recoveryPointArn)
			Expect(err).To(HaveOccurred())
			Expect(mockClient.IsNotFound(err)).To(BeTrue())
		})
	})
})
