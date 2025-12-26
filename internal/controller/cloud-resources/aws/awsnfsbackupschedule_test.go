package aws

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance"
	skrawsnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolume"
	"github.com/kyma-project/cloud-manager/pkg/skr/backupschedule"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR AwsNfsBackupSchedule", func() {

	const (
		interval        = time.Millisecond * 50
		timeout         = time.Second * 20
		awsAccountId    = "974658265573"
		toleranceWindow = 120 * time.Second
	)

	It("Scenario: Creates recurring backup schedule with existing backups", func() {
		var (
			suffix             string
			skrNfsVolumeName   string
			nfsInstanceName    string
			awsNfsId           string
			scheduleName       string
			existingBackupName string
			skrNfsVolume       *cloudresourcesv1beta1.AwsNfsVolume
			nfsInstance        *cloudcontrolv1beta1.NfsInstance
			scope              *cloudcontrolv1beta1.Scope
			backupSchedule     *cloudresourcesv1beta1.AwsNfsBackupSchedule
			existingBackup     *cloudresourcesv1beta1.AwsNfsVolumeBackup
		)

		// This ensures each parallel test run has different resource names
		suffix = fmt.Sprintf("recurring-%d", GinkgoParallelProcess())
		skrNfsVolumeName = fmt.Sprintf("aws-vol-%s", suffix)
		nfsInstanceName = fmt.Sprintf("aws-inst-%s", suffix)
		awsNfsId = fmt.Sprintf("fs-%s", suffix)
		scheduleName = fmt.Sprintf("aws-schedule-%s", suffix)
		existingBackupName = fmt.Sprintf("aws-backup-existing-%s", suffix)

		// Initialize objects
		skrNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{}
		nfsInstance = &cloudcontrolv1beta1.NfsInstance{}
		scope = &cloudcontrolv1beta1.Scope{}
		backupSchedule = &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
		existingBackup = &cloudresourcesv1beta1.AwsNfsVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      existingBackupName,
				Namespace: DefaultSkrNamespace,
			},
		}

		// Setup cleanup - prevents resource leaks
		DeferCleanup(func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), existingBackup).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance).
				Should(Succeed())
		})

		// Set tolerance for backup schedule timing
		backupschedule.ToleranceInterval = toleranceWindow

		// Stop reconciliation to prevent interference
		skrawsnfsvol.Ignore.AddName(skrNfsVolumeName)
		nfsinstance.Ignore.AddName(nfsInstanceName)

		By("Given KCP Scope exists", func() {
			Expect(client.IgnoreAlreadyExists(
				CreateScopeAws(infra.Ctx(), infra, scope, awsAccountId, WithName(infra.SkrKymaRef().Name)))).
				To(Succeed())
		})

		By("And Given SKR AwsNfsVolume exists", func() {
			Eventually(CreateAwsNfsVolume).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("And Given KCP NfsInstance exists", func() {
			// NfsInstance is created automatically by AwsNfsVolume controller
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					NewObjActions(),
				).
				Should(Succeed())
		})

		By("And Given NfsInstance has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithNfsInstanceStatusId(nfsInstance.Name),
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given AwsNfsVolume has Ready condition", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).
				Should(Succeed())
		})

		By("And Given existing AwsNfsVolumeBackup exists", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), existingBackup).
				Should(Succeed())
		})

		By("When AwsNfsBackupSchedule is created", func() {
			Eventually(CreateAwsNfsBackupSchedule).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					WithName(scheduleName),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithSchedule("* * * * *"), // Every minute
				).
				Should(Succeed())
		})

		By("Then AwsNfsBackupSchedule has Active state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					NewObjActions(),
					HavingState(string(cloudresourcesv1beta1.JobStateActive)),
				).
				Should(Succeed())
		})

		By("And Then AwsNfsBackupSchedule has Ready condition", func() {
			Expect(backupSchedule.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(cloudresourcesv1beta1.ConditionTypeReady),
				"Status": Equal(metav1.ConditionTrue),
			})))
		})

		By("And Then AwsNfsBackupSchedule has NextRun times populated", func() {
			Expect(len(backupSchedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		By("And Then scheduled AwsNfsVolumeBackup is created", func() {
			Eventually(func() error {
				list := &cloudresourcesv1beta1.AwsNfsVolumeBackupList{}
				err := infra.SKR().Client().List(infra.Ctx(), list)
				if err != nil {
					return err
				}
				// Should have 2 backups: existing + scheduled
				if len(list.Items) != 2 {
					return fmt.Errorf("expected 2 backups, got %d", len(list.Items))
				}
				return nil
			}, timeout).Should(Succeed())
		})

		By("And Then existing AwsNfsVolumeBackup still exists", func() {
			err := infra.SKR().Client().Get(infra.Ctx(), types.NamespacedName{
				Namespace: existingBackup.Namespace,
				Name:      existingBackup.Name,
			}, existingBackup)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	It("Scenario: Creates one-time backup schedule", func() {
		var (
			suffix           string
			skrNfsVolumeName string
			nfsInstanceName  string
			awsNfsId         string
			scheduleName     string
			scope            *cloudcontrolv1beta1.Scope
			skrNfsVolume     *cloudresourcesv1beta1.AwsNfsVolume
			nfsInstance      *cloudcontrolv1beta1.NfsInstance
			backupSchedule   *cloudresourcesv1beta1.AwsNfsBackupSchedule
		)

		// Unique naming for parallel test runs
		suffix = fmt.Sprintf("onetime-%d", GinkgoParallelProcess())
		skrNfsVolumeName = fmt.Sprintf("aws-vol-%s", suffix)
		nfsInstanceName = fmt.Sprintf("aws-inst-%s", suffix)
		awsNfsId = fmt.Sprintf("fs-%s", suffix)
		scheduleName = fmt.Sprintf("aws-schedule-%s", suffix)

		scope = &cloudcontrolv1beta1.Scope{}
		skrNfsVolume = &cloudresourcesv1beta1.AwsNfsVolume{}
		nfsInstance = &cloudcontrolv1beta1.NfsInstance{}
		backupSchedule = &cloudresourcesv1beta1.AwsNfsBackupSchedule{}

		DeferCleanup(func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume).
				Should(Succeed())
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

		// TODO: Add test steps
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
