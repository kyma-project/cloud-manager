package cloudresources

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR NfsBackupSchedule", func() {

	const (
		interval = time.Millisecond * 50
	)
	var (
		timeout = time.Second * 20
	)
	now := time.Now().UTC()
	skrIpRangeName := "gcp-iprange-1"
	skrNfsVolumeName := "gcp-nfs-1"
	skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
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

		By("And Given SKR NfsVolume exists", func() {
			// tell skrgcpnfsvol reconciler to ignore this SKR GcpNfsVolume
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			//Create SKR GcpNfsVolume if it doesn't exist.
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).
				Should(Succeed())
		})
		By("And Given SKR NfsVolume in Ready state", func() {
			//Update SKR NfsVolume status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

	})

	Describe("Scenario: SKR Recurring NfsBackupSchedule - Create", func() {

		//Define variables.
		nfsBackupSchedule := &cloudresourcesv1beta1.NfsBackupSchedule{}
		nfsBackupScheduleName := "nfs-schedule-1"
		nfsBackupHourlySchedule := "0 * * * *"
		nfsBackupLocation := "us-west1"

		nfsBackup1Name := "gcp-nfs-backup-1"

		expectedTimes := []time.Time{
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()).UTC(),
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+2, 0, 0, 0, now.Location()).UTC(),
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+3, 0, 0, 0, now.Location()).UTC(),
		}

		nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, now.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}

		skrNfsBackup1 := &cloudresourcesv1beta1.GcpNfsVolumeBackup{
			ObjectMeta: v1.ObjectMeta{
				Name:      nfsBackup1Name,
				Namespace: DefaultSkrNamespace,
			},
			Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
				Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
					Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
						Name:      skrNfsVolumeName,
						Namespace: DefaultSkrNamespace,
					},
				},
				Location: "us-west1-a",
			},
		}

		BeforeEach(func() {

			By("And Given SKR NfsVolumeBackups exists", func() {
				//Update SKR NfsVolume status to Ready
				Eventually(CreateGcpNfsVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrNfsBackup1,
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrNfsBackup1,
						NewObjActions(),
					).
					Should(Succeed())
			})
		})

		It("When NfsBackupSchedule Create is called", func() {
			Eventually(CreateNfsBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
					WithName(nfsBackupScheduleName),
					WithSchedule(nfsBackupHourlySchedule),
					WithLocation(nfsBackupLocation),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
				).
				Should(Succeed())
			By("Then NfsBackupSchedule is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then NfsBackupSchedule will get NextRun time(s)", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
						HaveNextRunTimes(expectedTimes),
					).
					Should(Succeed())
			})
			By("And Then NfsBackupSchedule has Active state", func() {
				Expect(nfsBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("When it is time for the Next Run", func() {
				//Update SKR NfsVolume status to Ready
				Eventually(UpdateStatus).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						WithNextRunTime(now),
					).
					Should(Succeed())
			})

			By("Then the NfsVolumeBackup is created", func() {
				//Load and check whether the NfsVolumeBackup object got created.
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackup,
						NewObjActions(WithName(nfsBackupName)),
					).
					Should(Succeed())
			})

			By("And Then previous NfsVolumeBackup(s) associated with the schedule exists", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrNfsBackup1,
						NewObjActions(),
					).
					Should(Succeed())
			})

		})
	})

	Describe("Scenario: SKR Onetime NfsBackupSchedule - Create", func() {
		//Define variables.
		nfsBackupSchedule := &cloudresourcesv1beta1.NfsBackupSchedule{}
		nfsBackupScheduleName := "nfs-schedule-2"
		nfsBackupLocation := "us-west1"

		now := time.Now().UTC()
		expectedTimes := []time.Time{now}

		nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, now.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}

		It("When NfsBackupSchedule Create is called", func() {
			Eventually(CreateNfsBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
					WithName(nfsBackupScheduleName),
					WithLocation(nfsBackupLocation),
					WithStartTime(now),
					WithNfsVolumeRef(skrNfsVolumeName),
				).
				Should(Succeed())
			By("Then NfsBackupSchedule is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then NfsBackupSchedule will get NextRun time", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
						HaveNextRunTimes(expectedTimes),
					).
					Should(Succeed())
			})
			By("And Then NfsBackupSchedule has Active state", func() {
				Expect(nfsBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("Then the NfsVolumeBackup is created", func() {
				//Load and check whether the NfsVolumeBackup object got created.
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackup,
						NewObjActions(WithName(nfsBackupName)),
					).
					Should(Succeed())
			})
		})
	})

})
