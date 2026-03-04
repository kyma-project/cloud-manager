package cloudresources

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	skrgcpnfsvolbackupv1 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v1"
	skrgcpnfsvolbackupv2 "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup/v2"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeRestore V2", func() {

	It("Scenario: SKR GcpNfsVolumeRestore V2 is created with backup ref and completed", func() {
		if !feature.GcpNfsRestoreV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
		}

		skrGcpNfsVolumeName := "3ec6e249-de2f-42fc-9c2f-5334114a1537"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "da5f0c69-6e3b-4b81-a9a9-4152869f2611"
		skrGcpNfsBackupName := "3e9ae34a-b225-4dd7-8d88-ba4527d816e2"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "9a63bc2b-055c-45c9-9128-37863cd2f00a"

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

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			// Set backup status fields needed by populateBackupUrl
			skrGcpNfsBackup.Status.Location = scope.Spec.Region
			skrGcpNfsBackup.Status.Id = "test-backup-id"
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeRestore is created", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed(), "expected GcpNfsVolumeRestore with Ready condition")
		})

		By("And Then GcpNfsVolumeRestore has Done state", func() {
			Expect(gcpNfsVolumeRestore.Status.State).To(Equal(cloudresourcesv1beta1.JobStateDone))
		})

		// DELETE

		By("When GcpNfsVolumeRestore is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed(), "expected GcpNfsVolumeRestore to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeRestore V2 is deleted when in Done state", func() {
		if !feature.GcpNfsRestoreV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
		}

		skrGcpNfsVolumeName := "6e854f96-d730-4333-8263-a752346b4c89"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "920ea8af-c458-4c55-9c6b-6112dfe0ae20"
		skrGcpNfsBackupName := "5ab6d98d-77a0-4747-a30f-ac8d716ffd08"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "9f65425d-c7b4-4139-b916-4c7e091f28c0"

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

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			skrGcpNfsBackup.Status.Location = scope.Spec.Region
			skrGcpNfsBackup.Status.Id = "test-backup-id-del"
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("And Given GcpNfsVolumeRestore is created and reaches Done state", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).Should(Succeed(), "expected GcpNfsVolumeRestore to reach Ready/Done state")
		})

		By("When GcpNfsVolumeRestore is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore).
				Should(Succeed(), "expected GcpNfsVolumeRestore to be deleted")
		})
	})

	It("Scenario: SKR GcpNfsVolumeRestore V2 fails when GcpNfsVolume is not ready", func() {
		if !feature.GcpNfsRestoreV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsVolumeRestore tests because gcpNfsRestoreV2 feature flag is disabled")
		}

		skrGcpNfsVolumeName := "c3d310f8-b26b-4852-b2f1-b46294fdaae0"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "dd5b1196-7419-4b19-a8d8-4b373c755c1d"
		skrGcpNfsBackupName := "3a314877-9924-4977-b91e-297a4851a1cc"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "6e5efcc7-692e-4a1c-9638-9e8a879a3544"

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume exists but is NOT in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			// NFS Volume is NOT set to Ready state
		})

		By("And Given SKR GcpNfsVolumeBackup exists in Ready state", func() {
			skrgcpnfsvolbackupv1.Ignore.AddName(skrGcpNfsBackupName)
			skrgcpnfsvolbackupv2.Ignore.AddName(skrGcpNfsBackupName)
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsVolumeRestore is created", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Error condition with NfsVolumeNotReady reason", func() {
			Eventually(func() (bool, error) {
				err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeRestore), gcpNfsVolumeRestore)
				if err != nil {
					return false, err
				}
				errCond := meta.FindStatusCondition(gcpNfsVolumeRestore.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
				if errCond == nil {
					return false, nil
				}
				return errCond.Reason == cloudresourcesv1beta1.ConditionReasonNfsVolumeNotReady, nil
			}).Should(BeTrue(), "expected GcpNfsVolumeRestore to have Error condition with NfsVolumeNotReady reason")
		})
	})
})
