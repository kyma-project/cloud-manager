package cloudresources

import (
	"strings"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		By("And Given SKR GcpNfsVolume exists", func() {
			// tell skrgcpnfsvol reconciler to ignore this SKR GcpNfsVolume
			skrgcpnfsvol.Ignore.AddName(skrGcpNfsVolumeName)
			//Create SKR GcpNfsVolume if it doesn't exist.
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithName(skrGcpNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR GcpNfsVolume in Ready state", func() {
			//Update SKR GcpNfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithGcpNfsVolumeStatusLocation(scope.Spec.Region),
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
			By("And Then GcpNfsVolumeBackup has .status location set from spec", func() {
				Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(gcpNfsVolumeBackup.Spec.Location))
				Expect(len(gcpNfsVolumeBackup.Status.Location)).To(BeNumerically(">", 0))
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
			By("And Then GcpNfsVolumeBackup has .status location set from Scope", func() {
				Expect(gcpNfsVolumeBackup.Status.Location).To(Equal(scope.Spec.Region))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeBackup is deleted with empty location", Ordered, func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "gcp-nfs-volume-backup-4"
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

	Describe("Scenario: SKR GcpNfsVolumeBackup is set to be available from multiple shoots", func() {
		//Define variables.
		gcpNfsVolumeBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		gcpNfsVolumeBackupName := "8fb2aaaa-cf54-4ee0-b5d9-065ed69ef2f9"

		additionalShoots := []string{
			"shoot-eu-1",
			"shoot-eu-2",
		}

		It("When GcpNfsVolumeBackup Create is called", func() {
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
					WithName(gcpNfsVolumeBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupAccessibleFrom(additionalShoots),
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

			By("And Then GcpNfsVolumeBackup has proper AccessibleFrom field in status", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup,
						NewObjActions(),
						HavingGcpNfsVolumeBackupAccessibleFromStatus(strings.Join(additionalShoots, ",")),
					).
					Should(Succeed())
			})

			// CleanUp
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeBackup).
				Should(Succeed())
		})
	})
})
