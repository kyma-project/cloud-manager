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
		suffix := "6278871d-a536-4702-87a1-ffc21dfc38ad"
		skrNfsVolumeName := suffix
		nfsInstanceName := suffix
		scheduleName := suffix
		existingBackupName := fmt.Sprintf("%s-existing", suffix)
		skrNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		scope := &cloudcontrolv1beta1.Scope{}
		backupSchedule := &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
		existingBackup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      existingBackupName,
				Namespace: DefaultSkrNamespace,
			},
		}

		DeferCleanup(func() {
			By("Cleanup: Delete test resources", func() {
				_ = Delete(infra.Ctx(), infra.SKR().Client(), backupSchedule)
				_ = Delete(infra.Ctx(), infra.SKR().Client(), existingBackup)
				_ = Delete(infra.Ctx(), infra.SKR().Client(), skrNfsVolume)
				_ = Delete(infra.Ctx(), infra.KCP().Client(), nfsInstance)
				_ = Delete(infra.Ctx(), infra.KCP().Client(), scope)
			})
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
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("And Given KCP NfsInstance exists", func() {
			Eventually(GivenNfsInstanceExists).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(nfsInstanceName),
					WithRemoteRef(skrNfsVolumeName),
					WithScope(infra.SkrKymaRef().Name),
					WithIpRange(nfsInstanceName),
					WithNfsInstanceAws(),
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
					WithAwsNfsVolumeStatusId(nfsInstanceName),
				).
				Should(Succeed())
		})

		By("And Given existing AwsNfsVolumeBackup exists", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), existingBackup).
				Should(Succeed())
		})

		By("When AwsNfsBackupSchedule is created", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					WithName(scheduleName),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithSchedule("* * * * *"), // Every minute
				).
				Should(Succeed())

			// Verify it was actually created
			err := infra.SKR().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(backupSchedule), backupSchedule)
			Expect(err).NotTo(HaveOccurred())
			Expect(backupSchedule.Name).To(Equal(scheduleName))
		})

		By("Then AwsNfsBackupSchedule has Active state", func() {
			Eventually(LoadAndCheck, 15*time.Second, 500*time.Millisecond).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					NewObjActions(),
					HavingState(string(cloudresourcesv1beta1.JobStateActive)),
				).
				Should(Succeed())
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
				// Filter backups for THIS volume only (to avoid counting backups from parallel tests)
				var volumeBackups []cloudresourcesv1beta1.AwsNfsVolumeBackup
				var allBackupNames []string
				for _, backup := range list.Items {
					allBackupNames = append(allBackupNames, fmt.Sprintf("%s(vol:%s)", backup.Name, backup.Spec.Source.Volume.Name))
					if backup.Spec.Source.Volume.Name == skrNfsVolumeName {
						volumeBackups = append(volumeBackups, backup)
					}
				}
				// Should have 2 backups: existing + scheduled
				if len(volumeBackups) != 2 {
					return fmt.Errorf("expected 2 backups for volume %s, got %d; all backups: %v",
						skrNfsVolumeName, len(volumeBackups), allBackupNames)
				}
				return nil
			}, timeout*3).Should(Succeed()) // Increased to 60s for parallel execution
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
		suffix := "a4e8b2f1-7d93-4c6a-9e5f-3b1c8d4a7e9b"
		skrNfsVolumeName := suffix
		nfsInstanceName := suffix
		scheduleName := suffix
		scope := &cloudcontrolv1beta1.Scope{}
		skrNfsVolume := &cloudresourcesv1beta1.AwsNfsVolume{}
		nfsInstance := &cloudcontrolv1beta1.NfsInstance{}
		backupSchedule := &cloudresourcesv1beta1.AwsNfsBackupSchedule{}

		DeferCleanup(func() {
			By("Cleanup: Delete test resources", func() {
				_ = Delete(infra.Ctx(), infra.SKR().Client(), backupSchedule)
				_ = Delete(infra.Ctx(), infra.SKR().Client(), skrNfsVolume)
				_ = Delete(infra.Ctx(), infra.KCP().Client(), scope)
			})
		})

		// Set tolerance and stop reconciliation
		backupschedule.ToleranceInterval = toleranceWindow
		skrawsnfsvol.Ignore.AddName(skrNfsVolumeName)
		nfsinstance.Ignore.AddName(nfsInstanceName)

		By("Given KCP Scope exists", func() {
			Expect(client.IgnoreAlreadyExists(
				CreateScopeAws(infra.Ctx(), infra, scope, awsAccountId, WithName(infra.SkrKymaRef().Name)))).
				To(Succeed())
		})

		By("And Given SKR AwsNfsVolume exists", func() {
			Eventually(GivenAwsNfsVolumeExists).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("And Given KCP NfsInstance exists", func() {
			Eventually(GivenNfsInstanceExists).
				WithArguments(infra.Ctx(), infra.KCP().Client(), nfsInstance,
					WithName(nfsInstanceName),
					WithRemoteRef(skrNfsVolumeName),
					WithScope(infra.SkrKymaRef().Name),
					WithIpRange(nfsInstanceName),
					WithNfsInstanceAws(),
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
					WithAwsNfsVolumeStatusId(nfsInstanceName),
				).
				Should(Succeed())
		})

		now := time.Now().UTC()
		startTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()

		By("When AwsNfsBackupSchedule is created with one-time schedule", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					WithName(scheduleName),
					WithStartTime(startTime),
					WithNfsVolumeRef(skrNfsVolumeName),
				).
				Should(Succeed())
		})

		By("Then AwsNfsBackupSchedule has Active state", func() {
			Eventually(LoadAndCheck, 15*time.Second, 500*time.Millisecond).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					NewObjActions(),
					HavingState(string(cloudresourcesv1beta1.JobStateActive)),
				).
				Should(Succeed())
		})

		By("And Then AwsNfsBackupSchedule has NextRunTimes with single entry", func() {
			Expect(len(backupSchedule.Status.NextRunTimes)).To(Equal(1))
			nextRunTime, err := time.Parse(time.RFC3339, backupSchedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(nextRunTime).To(BeTemporally("~", startTime, time.Second))
		})

		By("And Then scheduled AwsNfsVolumeBackup is created", func() {
			expectedBackupName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, startTime.Format("20060102-150405"))
			backup := &cloudresourcesv1beta1.AwsNfsVolumeBackup{}

			Eventually(LoadAndCheck, timeout*6, interval).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backup,
					NewObjActions(WithName(expectedBackupName)),
				).
				Should(Succeed())
		})
	})
})
