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
			Eventually(CreateAwsNfsVolume).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithAwsNfsVolumeCapacity("100G"),
				).
				Should(Succeed())
		})

		By("And Given KCP NfsInstance exists", func() {
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

		// WHEN
		now := time.Now().UTC()
		startTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location()).UTC()

		By("When AwsNfsBackupSchedule is created with one-time schedule", func() {
			Eventually(CreateAwsNfsBackupSchedule).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					WithName(scheduleName),
					WithStartTime(startTime),
					WithNfsVolumeRef(skrNfsVolumeName),
				).
				Should(Succeed())
		})

		// THEN
		By("Then AwsNfsBackupSchedule has Active state", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), backupSchedule,
					NewObjActions(),
					HavingState(string(cloudresourcesv1beta1.JobStateActive)),
				).
				Should(Succeed())
		})

		By("And Then AwsNfsBackupSchedule has NextRunTimes with single entry", func() {
			Expect(len(backupSchedule.Status.NextRunTimes)).To(Equal(1))
			Expect(backupSchedule.Status.NextRunTimes[0]).To(BeTemporally("~", startTime, time.Second))
		})

		By("And Then AwsNfsBackupSchedule has Ready condition", func() {
			Expect(backupSchedule.Status.Conditions).To(ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(cloudresourcesv1beta1.ConditionTypeReady),
				"Status": Equal(metav1.ConditionTrue),
			})))
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
