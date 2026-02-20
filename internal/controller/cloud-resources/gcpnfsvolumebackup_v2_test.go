package cloudresources

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackup V2", func() {

	It("Scenario: SKR GcpNfsVolumeBackup V2 is created and deleted", func() {
		if !feature.GcpBackupV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeBackup tests because gcpBackupV2 feature flag is disabled")
		}

		skrGcpNfsVolumeName := "b7e6f9b1-11de-40aa-8c24-1207791fc0b9"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "3d5eef8e-b871-4147-b6a2-7a49753c8bf8"
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "8469ce3a-6529-474d-8e83-d5b8aef13362"

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume exists in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(scope.Spec.Region),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeBackup is created", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		var gcpBackupPath string
		By("Then GCP Backup is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingGcpNfsVolumeBackupStatusId(),
				).Should(Succeed(), "expected GcpNfsVolumeBackup to get status.id")

			gcpBackupPath = GcpNfsVolumeBackupPath(scope, gcpNfsVolumeBackup)
		})

		By("When GCP Backup is Ready", func() {
			infra.GcpMock().SetNfsBackupV2State(gcpBackupPath, filestorepb.Backup_READY)
		})

		By("Then GcpNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackup has .status.location set", func() {
			Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(gcpNfsVolumeBackup.Spec.Location))
			Expect(len(gcpNfsVolumeBackup.Status.Location)).To(BeNumerically(">", 0))
		})

		// DELETE

		By("When GcpNfsVolumeBackup is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed())
		})

		By("Then GCP Backup has Deleting state", func() {
			Eventually(func() filestorepb.Backup_State {
				backup := infra.GcpMock().GetNfsBackupV2ByName(gcpBackupPath)
				if backup == nil {
					return filestorepb.Backup_STATE_UNSPECIFIED
				}
				return backup.State
			}).Should(Equal(filestorepb.Backup_DELETING))
		})

		By("And When GCP Backup is deleted", func() {
			infra.GcpMock().DeleteNfsBackupV2ByName(gcpBackupPath)
		})

		By("Then GcpNfsVolumeBackup does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed(), "expected GcpNfsVolumeBackup to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeBackup V2 is created with empty location", func() {
		if !feature.GcpBackupV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeBackup tests because gcpBackupV2 feature flag is disabled")
		}

		skrGcpNfsVolumeName := "c4d4c1f0-fdbe-4673-b84b-276e735b7cc6"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "adc5accc-eb4b-4668-b760-e06b601a3893"
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "e8bc9c85-3132-4f0b-bb60-66ca233d02d2"

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume exists in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(scope.Spec.Region),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeBackup is created with empty location", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupLocation(""),
				).Should(Succeed())
		})

		var gcpBackupPath string
		By("Then GCP Backup is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingGcpNfsVolumeBackupStatusId(),
				).Should(Succeed(), "expected GcpNfsVolumeBackup to get status.id")

			gcpBackupPath = GcpNfsVolumeBackupPath(scope, gcpNfsVolumeBackup)
		})

		By("When GCP Backup is Ready", func() {
			infra.GcpMock().SetNfsBackupV2State(gcpBackupPath, filestorepb.Backup_READY)
		})

		By("Then GcpNfsVolumeBackup has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		By("And Then GcpNfsVolumeBackup has .status.location set from Scope", func() {
			Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(scope.Spec.Region))
		})

		// DELETE (cleanup)

		By("When GcpNfsVolumeBackup is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed())
		})

		By("And When GCP Backup is deleted", func() {
			infra.GcpMock().DeleteNfsBackupV2ByName(gcpBackupPath)
		})

		By("Then GcpNfsVolumeBackup does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed(), "expected GcpNfsVolumeBackup to be deleted")
		})
	})
})
