package cloudresources

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeBackup", func() {

	const (
		interval = time.Millisecond * 50
	)
	var (
		timeout = time.Second * 20
	)

	skrGcpNfsVolumeName := "gcp-nfs-1-b"
	skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
	skrIpRangeName := "gcp-iprange-1-b"
	scope := &cloudcontrolv1beta1.Scope{}

	shouldSkipIfGcpNfsVolumeAutomaticLocationAllocationDisabled := func() (bool, string) {
		if feature.GcpNfsVolumeAutomaticLocationAllocation.Value(context.Background()) {
			return false, ""
		}
		return true, "gcpNfsVolumeAutomaticLocationAllocation is disabled"
	}

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {

			// Given Scope exists
			Expect(
				infra.GivenScopeGcpExists(infra.SkrKymaRef().Name),
			).NotTo(HaveOccurred())
			// Load created scope
			Eventually(func() (exists bool, err error) {
				err = infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				exists = err == nil
				return exists, client.IgnoreNotFound(err)
			}, timeout, interval).
				Should(BeTrue(), "expected Scope to get created")
		})
		By("And Given SKR namespace exists", func() {
			//Create namespace if it doesn't exist.
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		By("And Given SKR GcpNfsVolume exists", func() {
			// tell skrgcpnfsvol reconciler to ignore this SKR GcpNfsVolume
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			//Create SKR GcpNfsVolume if it doesn't exist.
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
					WithGcpNfsValues(),
				).
				Should(Succeed())
		})
		By("And Given SKR GcpNfsVolume in Ready state", func() {
			//Update SKR GcpNfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(skrGcpNfsVolume.Spec.Location),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeBackup is created", func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-1"

		It("When GcpNfsVolumeBackup Create is called", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
				).
				Should(Succeed())
			By("Then GcpNfsVolumeBackup is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then GcpNfsVolumeBackup will get Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())
			})
			By("And Then GcpNfsVolumeBackup has Ready state", func() {
				Expect(gcpNfsVolumeBackup.Status.State).To(Equal(cloudresourcesv1beta1.GcpNfsBackupReady))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeBackup is deleted", Ordered, func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-2"

		BeforeEach(func() {
			By("And Given SKR GcpNfsVolumeBackup has Ready condition", func() {

				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						WithName(gcpNfsVolumeBackupName),
						WithGcpNfsVolume(skrGcpNfsVolumeName),
					).
					Should(Succeed())

				//Load SKR GcpNfsVolumeBackup and check for Ready condition
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
					).
					Should(Succeed())
			})
		})
		It("When SKR GcpNfsVolumeBackup Delete is called ", func() {

			//Delete SKR GcpNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in GcpNfsVolumeBackup", func() {
				Expect(gcpNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the GcpNfsVolumeBackup in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					).
					Should(Succeed())
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeBackup is created with empty location", func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-3"
		It("When GcpNfsVolumeBackup Create is called", func() {
			shouldSkip, msg := shouldSkipIfGcpNfsVolumeAutomaticLocationAllocationDisabled()
			if shouldSkip {
				Skip(msg)
			}
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupLocation(""),
				).
				Should(Succeed())
			By("Then GcpNfsVolumeBackup is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then GcpNfsVolumeBackup will get Ready condition", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
					).
					Should(Succeed())
			})
			By("And Then GcpNfsVolumeBackup has Ready state", func() {
				Expect(gcpNfsVolumeBackup.Status.State).To(Equal(cloudresourcesv1beta1.GcpNfsBackupReady))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeBackup is deleted with empty location", Ordered, func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-4"
		BeforeEach(func() {
			shouldSkip, msg := shouldSkipIfGcpNfsVolumeAutomaticLocationAllocationDisabled()
			if shouldSkip {
				Skip(msg)
			}
			By("And Given SKR GcpNfsVolumeBackup has Ready condition", func() {
				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						WithName(gcpNfsVolumeBackupName),
						WithGcpNfsVolume(skrGcpNfsVolumeName),
					).
					Should(Succeed())

				//Load SKR GcpNfsVolumeBackup and check for Ready condition
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
						HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
						AssertGcpNfsVolumeBackupHasState(cloudresourcesv1beta1.StateReady),
					).
					Should(Succeed())
			})
		})
		It("When SKR GcpNfsVolumeBackup Delete is called ", func() {
			//Delete SKR GcpNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in GcpNfsVolumeBackup", func() {
				Expect(gcpNfsVolumeBackup.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the GcpNfsVolumeBackup in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					).
					Should(Succeed())
			})
		})
	})
})
