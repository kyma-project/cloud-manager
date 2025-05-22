package azure

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: SKR AzureRwxBackupSchedule", func() {

	const (
		interval = time.Millisecond * 50
	)
	var (
		timeout = time.Second * 20
	)
	now := time.Now().UTC()
	skrRwxVolumeName := "azure-rwx-1-pv"
	pv := &corev1.PersistentVolume{}
	skrRwxVolumeClaimName := "azure-rws-1-pvc"
	pvc := &corev1.PersistentVolumeClaim{}
	scope := &cloudcontrolv1beta1.Scope{}

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {

			// Given Scope exists
			Eventually(GivenScopeAzureExists).
				WithArguments(
					infra.Ctx(), infra, scope,
					WithName(infra.SkrKymaRef().Name),
				).
				Should(Succeed())
		})
		By("And Given SKR namespace exists", func() {
			//Create namespace if it doesn't exist.
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
				Should(Succeed())
		})

		By("And Given SKR PV exists", func() {
			//skrazurerwxvol.Ignore.AddName(skrRwxVolumeName)
			//Create SKR AzureRwxVolume if it doesn't exist.
			Eventually(GivenPvExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pv,
					WithName(skrRwxVolumeName),
					WithPvCapacity("1Gi"),
					WithPvAccessMode(corev1.ReadWriteMany),
					WithPvCsiSource(&corev1.CSIPersistentVolumeSource{
						Driver:       "file.csi.azure.com",
						VolumeHandle: "test-file-share-01",
					}),
					WithPvLabel(cloudresourcesv1beta1.LabelCloudManaged, "true"),
				).
				Should(Succeed())
		})
		By("And Given SKR PVC exists", func() {
			Eventually(GivenPvcExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), pvc,
					WithName(skrRwxVolumeClaimName),
					WithPVName(skrRwxVolumeName),
					WithPvCapacity("1Gi"),
					WithPvAccessMode(corev1.ReadWriteMany),
					WithPvLabel(cloudresourcesv1beta1.LabelCloudManaged, "true"),
				).
				Should(Succeed(), "failed creating PVC")
		})
		By("And Given PVC is in Ready state", func() {

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), pvc,
					NewObjActions(),
					HavePvcPhase(corev1.ClaimBound),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR Recurring AzureRwxBackupSchedule - Create", func() {

		//Define variables.
		backupschedule.ToleranceInterval = 120 * time.Second
		rwxBackupSchedule := &cloudresourcesv1beta1.AzureRwxBackupSchedule{}
		rwxBackupScheduleName := "azure-rwx-backup-schedule-1"
		rwxBackupMinutelySchedule := "* * * * *"
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()

		rwxBackup1Name := "azure-rwx-backup-1-bs"

		expectedTimes := []time.Time{
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+1, 0, 0, now.Location()).UTC(),
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+2, 0, 0, now.Location()).UTC(),
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+3, 0, 0, now.Location()).UTC(),
		}

		rwxBackup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}

		skrRwxBackup1 := &cloudresourcesv1beta1.AzureRwxVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rwxBackup1Name,
				Namespace: DefaultSkrNamespace,
			},
			Spec: cloudresourcesv1beta1.AzureRwxVolumeBackupSpec{
				Source: cloudresourcesv1beta1.PvcSource{
					Pvc: cloudresourcesv1beta1.PvcRef{
						Name:      skrRwxVolumeName,
						Namespace: DefaultSkrNamespace,
					},
				},
			},
		}

		BeforeEach(func() {

			By("And Given SKR RwxVolumeBackups exists", func() {
				//Update SKR SourceRef status to Ready
				Eventually(CreateAzureRwxVolumeBackup).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrRwxBackup1,
					).
					Should(Succeed())

				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrRwxBackup1,
						NewObjActions(),
					).
					Should(Succeed())
			})
		})

		It("When AzureRwxBackupSchedule Create is called", func() {

			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxBackupSchedule,
					WithName(rwxBackupScheduleName),
					WithSchedule(rwxBackupMinutelySchedule),
					WithStartTime(start),
					WithNfsVolumeRef(skrRwxVolumeClaimName),
					WithRetentionDays(0),
				).
				Should(Succeed())
			By("Then AzureRwxBackupSchedule is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxBackupSchedule,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then AzureRwxBackupSchedule will get NextRun time(s)", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxBackupSchedule,
						NewObjActions(),
						HaveNextRunTimes(expectedTimes),
					).
					Should(Succeed())
			})
			By("And Then AzureRwxBackupSchedule has Active state", func() {
				Expect(rwxBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("And Then the RwxVolumeBackup is created", func() {
				expected, err := time.Parse(time.RFC3339, rwxBackupSchedule.Status.NextRunTimes[0])
				Expect(err).ShouldNot(HaveOccurred())
				rwxBackupName := fmt.Sprintf("%s-%d-%s", rwxBackupScheduleName, 1, expected.Format("20060102-150405"))
				//Load and check whether the RwxVolumeBackup object got created.
				Eventually(LoadAndCheck, timeout*6, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxBackup,
						NewObjActions(WithName(rwxBackupName)),
					).
					Should(Succeed())
			})

			By("And Then previous RwxVolumeBackup(s) associated with the backup schedule exists", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrRwxBackup1,
						NewObjActions(),
					).
					Should(Succeed())
			})

		})
	})

	Describe("Scenario: SKR Onetime AzureRwxBackupSchedule - Create", func() {
		//Define variables.
		rwxBackupSchedule := &cloudresourcesv1beta1.AzureRwxBackupSchedule{}
		rwxBackupScheduleName := "azure-rwx-backup-schedule-2"

		now := time.Now().UTC()
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()
		expectedTimes := []time.Time{start}

		rwxBackupName := fmt.Sprintf("%s-%d-%s", rwxBackupScheduleName, 1, start.Format("20060102-150405"))
		rwxBackup := &cloudresourcesv1beta1.AzureRwxVolumeBackup{}

		It("When AzureRwxBackupSchedule Create is called", func() {

			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), rwxBackupSchedule,
					WithName(rwxBackupScheduleName),
					WithStartTime(start),
					WithNfsVolumeRef(skrRwxVolumeClaimName),
				).
				Should(Succeed())

			By("Then AzureRwxBackupSchedule will be created in SKR and will have NextRun time", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxBackupSchedule,
						NewObjActions(),
						HaveLastCreateRun(expectedTimes[0]),
					).
					Should(Succeed())
			})
			By("And Then AzureRwxBackupSchedule has Active state", func() {
				Expect(rwxBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("Then the RwxVolumeBackup is created", func() {
				//Load and check whether the RwxVolumeBackup object got created.
				//Waiting little longer as the scheduled backup creation might take about 2 minutes.
				Eventually(LoadAndCheck, timeout*6, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), rwxBackup,
						NewObjActions(WithName(rwxBackupName)),
					).
					Should(Succeed())
			})
		})
	})

})
