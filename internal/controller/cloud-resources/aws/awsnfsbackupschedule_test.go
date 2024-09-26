package aws

import (
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

	BeforeEach(func() {
		By("Given KCP Scope exists", func() {

			// Given Scope exists
			Eventually(GivenScopeAwsExists).
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
		nfsBackupSchedule := &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
		nfsBackupScheduleName := "nfs-backup-schedule-1"
		nfsBackupHourlySchedule := "0 * * * *"

		nfsBackup1Name := "aws-nfs-backup-1-bs"

		expectedTimes := []time.Time{
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()).UTC(),
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+2, 0, 0, 0, now.Location()).UTC(),
			time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+3, 0, 0, 0, now.Location()).UTC(),
		}

		nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, now.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

		skrNfsBackup1 := &cloudresourcesv1beta1.AwsNfsVolumeBackup{
			ObjectMeta: v1.ObjectMeta{
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
			Eventually(CreateNfsBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackupSchedule,
					WithName(nfsBackupScheduleName),
					WithSchedule(nfsBackupHourlySchedule),
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

			By("When it is time for the Next Run", func() {
				//Update SKR SourceRef status to Ready
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

			By("And Then previous NfsVolumeBackup(s) associated with the backupschedule exists", func() {
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
		nfsBackupScheduleName := "nfs-backup-schedule-2"

		now := time.Now().UTC()
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+2, 0, 0, now.Location()).UTC()
		expectedTimes := []time.Time{start}

		nfsBackupName := fmt.Sprintf("%s-%d-%s", nfsBackupScheduleName, 1, start.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

		It("When AwsNfsBackupSchedule Create is called", func() {
			Eventually(CreateNfsBackupSchedule).
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
						HaveNextRunTimes(expectedTimes),
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
