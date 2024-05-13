package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	skrgcpnfsvolbackup "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolumebackup"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsVolumeRestore", func() {

	const (
		interval = time.Millisecond * 50
	)
	var (
		timeout = time.Second * 20
	)

	skrGcpNfsVolumeName := "gcp-nfs-1"
	skrGcpNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
	skrGcpNfsBackupName := "gcp-nfs-1-backup"
	skrGcpNfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
	skrIpRangeName := "gcp-iprange-1"
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
		By("And Given SKR GcpNfsVolumeBackup exists", func() {
			// tell skrgcpnfsvolbackup reconciler to ignore this SKR GcpNfsVolumeBackup
			skrgcpnfsvolbackup.Ignore.AddName(skrGcpNfsBackupName)
			//Create SKR GcpNfsVolumeBackup if it doesn't exist.
			Eventually(CreateGcpNfsVolumeBackup).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithName(skrGcpNfsBackupName),
					WithGcpNfsVolume(skrGcpNfsVolumeName),
					WithGcpNfsVolumeBackupValues(),
				).
				Should(Succeed())

		})
		By("And Given SKR GcpNfsVolume in Ready state", func() {
			//Update SKR GcpNfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})
		By("And Given SKR GcpNfsVolumeBackup in Ready state", func() {
			//Update SKR GcpNfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrGcpNfsBackup,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeRestore is created", func() {
		//Define variables.
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "gcp-nfs-volume-restore-1"

		It("When GcpNfsVolumeRestore Create is called", func() {
			Eventually(CreateGcpNfsVolumeRestore).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					WithName(gcpNfsVolumeRestoreName),
					WithRestoreSourceBackup(skrGcpNfsBackupName),
					WithRestoreDestinationVolume(skrGcpNfsVolumeName),
				).
				Should(Succeed())
			By("Then GcpNfsVolumeRestore is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then SKR GcpNfsVolumeRestore will get Ready condition", func() {
				Eventually(func() (exists bool, err error) {
					err = infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(gcpNfsVolumeRestore), gcpNfsVolumeRestore)
					if err != nil {
						return false, err
					}
					exists = meta.IsStatusConditionTrue(gcpNfsVolumeRestore.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
					return exists, nil
				}, timeout, interval).
					Should(BeTrue(), "expected GcpNfsVolumeRestore with Ready condition")
			})
			By("And Then SKR GcpNfsVolumeRestore has Done state", func() {
				Expect(gcpNfsVolumeRestore.Status.State).To(Equal(cloudresourcesv1beta1.JobStateDone))
			})
		})
	})

	Describe("Scenario: SKR GcpNfsVolumeRestore is deleted", Ordered, func() {
		//Define variables.
		gcpNfsVolumeRestore := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		gcpNfsVolumeRestoreName := "gcp-nfs-volume-restore-1"

		BeforeEach(func() {
			By("And Given SKR GcpNfsVolumeRestore exists", func() {

				//Create GcpNfsVolume
				Eventually(CreateGcpNfsVolumeRestore).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
						WithName(gcpNfsVolumeRestoreName),
						WithRestoreSourceBackup(skrGcpNfsBackupName),
						WithRestoreDestinationVolume(skrGcpNfsVolumeName),
					).
					Should(Succeed())

				//Update SKR GcpNfsVolumeRestore to Done state
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
						WithConditions(SkrReadyCondition()),
						WithGcpNfsVolumeRestoreState(cloudresourcesv1beta1.JobStateDone),
					).
					Should(Succeed())

			})
		})
		It("When SKR GcpNfsVolumeRestore Delete is called ", func() {

			//Delete SKR GcpNfsVolume
			Eventually(Delete).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
				).Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					NewObjActions(),
				).
				Should(Succeed())

			By("Then DeletionTimestamp is set in GcpNfsVolumeRestore", func() {
				Expect(gcpNfsVolumeRestore.DeletionTimestamp.IsZero()).NotTo(BeTrue())
			})

			By("And Then the GcpNfsVolumeRestore in SKR is deleted.", func() {
				Eventually(IsDeleted, timeout, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), gcpNfsVolumeRestore,
					).
					Should(Succeed())
			})
		})
	})

})
