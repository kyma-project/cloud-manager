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

		skrGcpNfsVolumeName := "a1b2c3d4-1111-2222-3333-444455556601"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "a1b2c3d4-1111-2222-3333-444455556602"
		skrGcpNfsBackupName := "a1b2c3d4-1111-2222-3333-444455556603"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "a1b2c3d4-1111-2222-3333-444455556604"

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

		skrGcpNfsVolumeName := "b2c3d4e5-1111-2222-3333-444455556601"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "b2c3d4e5-1111-2222-3333-444455556602"
		skrGcpNfsBackupName := "b2c3d4e5-1111-2222-3333-444455556603"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "b2c3d4e5-1111-2222-3333-444455556604"

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

		skrGcpNfsVolumeName := "c3d4e5f6-1111-2222-3333-444455556601"
		skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		skrIpRangeName := "c3d4e5f6-1111-2222-3333-444455556602"
		skrGcpNfsBackupName := "c3d4e5f6-1111-2222-3333-444455556603"
		skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		scope := &cloudcontrolv1beta1.Scope{}
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "c3d4e5f6-1111-2222-3333-444455556604"

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
