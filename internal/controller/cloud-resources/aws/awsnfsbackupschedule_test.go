package aws

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AwsNfsBackupSchedule", func() {

	const (
		interval = time.Millisecond * 50
	)
	var (
		timeout = time.Second * 20
	)
	now := time.Now().UTC()
	skrNfsVolumeName := "aws-nfs-1-bs"
	skrNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
	nfsInstanceName := "aws-nfs-instance-01"
	nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
	awsNfsId := "aws-filesystem-01"
	scope := &cloudcontrolv1beta1.Scope{}

	awsAccountId := "974658265573"

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {

			// Given Scope exists
			Eventually(CreateScopeAws).
				WithArguments(
					infra.Ctx(), infra, scope, awsAccountId,
					WithName(infra.SkrKymaRef().Name),
				).
				Should(Succeed())
		})

		By("And Given SKR AwsNfsVolume exists", func() {
			skrawsnfsvol.Ignore.AddName(skrNfsVolumeName)
			//Create SKR AwsNfsVolume if it doesn't exist.
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithAwsNfsVolumeCapacity("1Gi"),
				).
				Should(Succeed())
		})
		By("And Given KCP NfsInstance exists", func() {
			nfsinstance.Ignore.AddName(nfsInstanceName)
			Eventually(GivenNfsInstanceExists).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(nfsInstanceName),
					WithRemoteRef(skrNfsVolumeName),
					WithScope(infra.SkrKymaRef().Name),
					WithIpRange(nfsInstanceName),
					WithNfsInstanceAws(),
				).
				Should(Succeed(), "failed creating NfsInstance")
		})
		By("And Given NfsInstance is in Ready state", func() {

			//Update KCP NfsInstance status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithConditions(KcpReadyCondition()),
					WithNfsInstanceStatusId(awsNfsId),
				).
				Should(Succeed())
		})
		By("And Given AwsNfsVolume is in Ready state", func() {

			//Update KCP NfsInstance status to Ready
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
					WithAwsNfsVolumeStatusId(nfsInstanceName),
				).
				Should(Succeed())
		})
	})

	Describe("Scenario: SKR Recurring AwsNfsBackupSchedule - Create", func() {

		//Define variables.
		backupschedule.ToleranceInterval = 120 * time.Second
		nfsBackupSchedule := &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
		nfsBackupScheduleName := "aws-nfs-backup-schedule-1"
		nfsBackupMinutelySchedule := "* * * * *"
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()

		nfsBackup1Name := "aws-nfs-backup-1-bs"

		expectedTimes := []time.Time{
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+1, 0, 0, now.Location()).UTC(),
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+2, 0, 0, now.Location()).UTC(),
			time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), start.Minute()+3, 0, 0, now.Location()).UTC(),
		}

		nfsBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

		skrNfsBackup1 := &cloudresourcesv1beta1.AwsNfsVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nfsBackup1Name,
				Namespace: DefaultSkrNamespace,
			},
			Spec: cloudresourcesv1beta1.AwsNfsVolumeBackupSpec{
				Source: cloudresourcesv1beta1.AwsNfsVolumeBackupSource{
					Volume: cloudresourcesv1beta1.VolumeRef{
						Name:      skrNfsVolumeName,
						Namespace: DefaultSkrNamespace,
					},
				},
			},
		}

		BeforeEach(func() {

			By("And Given SKR NfsVolumeBackups exists", func() {
				//Update SKR SourceRef status to Ready
				Eventually(CreateAwsNfsVolumeBackup).
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

		It("When AwsNfsBackupSchedule Create is called", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
					WithName(nfsBackupScheduleName),
					WithSchedule(nfsBackupMinutelySchedule),
					WithStartTime(start),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
				).
				Should(Succeed())
			By("Then AwsNfsBackupSchedule is created in SKR", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
					).
					Should(Succeed())
			})
			By("And Then AwsNfsBackupSchedule will get NextRun time(s)", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
						HaveNextRunTimes(expectedTimes),
					).
					Should(Succeed())
			})
			By("And Then AwsNfsBackupSchedule has Active state", func() {
				Expect(nfsBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("And Then the NfsVolumeBackup is created", func() {
				expected, err := time.Parse(time.RFC3339, nfsBackupSchedule.Status.NextRunTimes[0])
				Expect(err).ShouldNot(HaveOccurred())
				nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, expected.Format("20060102-150405"))
				//Load and check whether the NfsVolumeBackup object got created.
				Eventually(LoadAndCheck, timeout*6, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackup,
						NewObjActions(WithName(nfsBackupName)),
					).
					Should(Succeed())
			})

			By("And Then previous NfsVolumeBackup(s) associated with the backup schedule exists", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), skrNfsBackup1,
						NewObjActions(),
					).
					Should(Succeed())
			})

		})
	})

	Describe("Scenario: SKR Onetime AwsNfsBackupSchedule - Create", func() {
		//Define variables.
		nfsBackupSchedule := &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
		nfsBackupScheduleName := "aws-nfs-backup-schedule-2"

		now := time.Now().UTC()
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()
		expectedTimes := []time.Time{start}

		nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, start.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

		It("When AwsNfsBackupSchedule Create is called", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
					WithName(nfsBackupScheduleName),
					WithStartTime(start),
					WithNfsVolumeRef(skrNfsVolumeName),
				).
				Should(Succeed())

			By("Then AwsNfsBackupSchedule will be created in SKR and will have NextRun time", func() {
				Eventually(LoadAndCheck).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
						NewObjActions(),
						HaveLastCreateRun(expectedTimes[0]),
					).
					Should(Succeed())
			})
			By("And Then AwsNfsBackupSchedule has Active state", func() {
				Expect(nfsBackupSchedule.Status.State).To(Equal(cloudresourcesv1beta1.JobStateActive))
			})

			By("Then the NfsVolumeBackup is created", func() {
				//Load and check whether the NfsVolumeBackup object got created.
				//Waiting little longer as the scheduled backup creation might take about 2 minutes.
				Eventually(LoadAndCheck, timeout*6, interval).
					WithArguments(
						infra.Ctx(), infra.SKR().Client(), nfsBackup,
						NewObjActions(WithName(nfsBackupName)),
					).
					Should(Succeed())
			})
		})
	})

})
