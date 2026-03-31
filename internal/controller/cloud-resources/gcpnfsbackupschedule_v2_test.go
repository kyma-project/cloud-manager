package cloudresources

import (
	"context"
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrgcpnfsvol "github.com/kyma-project/cloud-manager/pkg/skr/gcpnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR GcpNfsBackupSchedule V2", func() {

	BeforeEach(func() {
		if !feature.BackupScheduleV2.Value(context.Background()) {
			Skip("Skipping v2 GcpNfsBackupSchedule tests because backupScheduleV2 feature flag is disabled")
		}
	})

	It("Scenario: Recurring schedule full lifecycle", func() {
		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-recur-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-recur-1"
			scheduleName     = "gcp-nfs-bs-v2-recurring-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsBackupSchedule is created with recurring cron", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithSchedule("* * * * *"),
					WithGcpLocation("us-west1"),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
				).Should(Succeed())
		})

		By("Then GcpNfsBackupSchedule becomes Active with NextRunTimes", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingStatusActive(),
				).Should(Succeed())
			Expect(len(schedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		var nfsBackupName string
		By("And When fake clock advances past the first scheduled run", func() {
			expected, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			nfsBackupName = fmt.Sprintf("%s-%d-%s", scheduleName, 1, expected.Format("20060102-150405"))
			testFakeClock.Step(2 * time.Minute)
		})

		nfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		By("Then a GcpNfsVolumeBackup is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackup,
					NewObjActions(WithName(nfsBackupName)),
				).Should(Succeed())
		})

		By("And Then the backup has correct labels and spec", func() {
			Expect(nfsBackup.Labels[cloudresourcesv1beta1.LabelScheduleName]).To(Equal(scheduleName))
			Expect(nfsBackup.Labels[cloudresourcesv1beta1.LabelScheduleNamespace]).To(Equal(DefaultSkrNamespace))
			Expect(nfsBackup.Spec.Location).To(Equal("us-west1"))
			Expect(nfsBackup.Spec.Source.Volume.Name).To(Equal(skrNfsVolumeName))
		})

		By("And Then schedule status is updated", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				return schedule.Status.BackupIndex, nil
			}).Should(Equal(1))
			Expect(schedule.Status.BackupCount).To(Equal(1))
			Expect(schedule.Status.LastCreatedBackup.Name).To(Equal(nfsBackupName))
		})

		// DELETE
		By("When GcpNfsBackupSchedule is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})

		By("Then the schedule is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})
	})

	It("Scenario: One-time schedule", func() {
		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-once-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-once-1"
			scheduleName     = "gcp-nfs-bs-v2-onetime-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		fakeNow := testFakeClock.Now().UTC()
		// Set start time 2 minutes ahead of the fake clock to avoid validateTimes rejection
		startTime := time.Date(fakeNow.Year(), fakeNow.Month(), fakeNow.Day(),
			fakeNow.Hour(), fakeNow.Minute()+2, 0, 0, time.UTC)

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsBackupSchedule is created as one-time", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithGcpLocation("us-west1"),
					WithStartTime(startTime),
					WithNfsVolumeRef(skrNfsVolumeName),
				).Should(Succeed())
		})

		By("Then schedule becomes Active", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingStatusActive(),
				).Should(Succeed())
		})

		By("And When fake clock advances past start time", func() {
			testFakeClock.Step(3 * time.Minute)
		})

		nfsBackupName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, startTime.Format("20060102-150405"))
		nfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}

		By("Then a single GcpNfsVolumeBackup is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackup,
					NewObjActions(WithName(nfsBackupName)),
				).Should(Succeed())
		})

		By("And Then schedule has LastCreateRun set", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HaveLastCreateRun(startTime),
				).Should(Succeed())
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
			Should(Succeed())
	})

	It("Scenario: Schedule with AccessibleFrom", func() {
		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-af-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-af-1"
			scheduleName     = "gcp-nfs-bs-v2-accessfrom-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		accessibleFrom := []string{"shoot-1", "shoot-2"}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsBackupSchedule is created with AccessibleFrom", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithSchedule("* * * * *"),
					WithGcpLocation("us-west1"),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
					WithGcpNfsBackupScheduleAccessibleFrom(accessibleFrom),
				).Should(Succeed())
		})

		By("Then GcpNfsBackupSchedule becomes Active", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingStatusActive(),
				).Should(Succeed())
			Expect(len(schedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		By("And When fake clock advances past the first scheduled run", func() {
			testFakeClock.Step(2 * time.Minute)
		})

		nfsBackup := &cloudresourcesv1beta1.GcpNfsVolumeBackup{}
		By("Then a GcpNfsVolumeBackup is created with AccessibleFrom", func() {
			expected, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			nfsBackupName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, expected.Format("20060102-150405"))

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), nfsBackup,
					NewObjActions(WithName(nfsBackupName)),
				).Should(Succeed())

			Expect(nfsBackup.Spec.AccessibleFrom).To(ContainElements(accessibleFrom))
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
			Should(Succeed())
	})

	It("Scenario: Suspension", func() {
		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-suspend-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-suspend-1"
			scheduleName     = "gcp-nfs-bs-v2-suspend-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("And Given an active GcpNfsBackupSchedule", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithSchedule("* * * * *"),
					WithGcpLocation("us-west1"),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
				).Should(Succeed())
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingStatusActive(),
				).Should(Succeed())
			Expect(len(schedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		By("When Suspend is set to true", func() {
			Eventually(Update).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithBackupScheduleSuspend(true),
				).Should(Succeed())
		})

		By("Then GcpNfsBackupSchedule becomes Suspended", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateSuspended),
				).Should(Succeed())
		})

		By("And Then NextRunTimes is empty", func() {
			Expect(schedule.Status.NextRunTimes).To(BeEmpty())
		})

		By("And Then NextDeleteTimes is empty", func() {
			Expect(schedule.Status.NextDeleteTimes).To(BeEmpty())
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
			Should(Succeed())
	})

	It("Scenario: Invalid cron expression", func() {
		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-invalid-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-invalid-1"
			scheduleName     = "gcp-nfs-bs-v2-invalid-cron-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume in Ready state", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When GcpNfsBackupSchedule is created with invalid cron", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithSchedule("invalid-cron-expr"),
					WithGcpLocation("us-west1"),
					WithNfsVolumeRef(skrNfsVolumeName),
				).Should(Succeed())
		})

		By("Then GcpNfsBackupSchedule has Error state", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateError),
				).Should(Succeed())
		})

		By("And Then error condition is set", func() {
			cond := findCondition(schedule.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(cloudresourcesv1beta1.ReasonInvalidCronExpression))
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
			Should(Succeed())
	})

	It("Scenario: Source not ready", func() {

		Skip("I'm flaky, please fix me!")

		const (
			skrNfsVolumeName = "gcp-nfs-v2-bs-noready-1"
			skrIpRangeName   = "gcp-iprange-v2-bs-noready-1"
			scheduleName     = "gcp-nfs-bs-v2-noready-1"
		)
		schedule := &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
		skrNfsVolume := &cloudresourcesv1beta1.GcpNfsVolume{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given KCP Scope exists", func() {
			Expect(infra.GivenScopeGcpExists(infra.SkrKymaRef().Name)).NotTo(HaveOccurred())
			Eventually(func() (bool, error) {
				err := infra.KCP().Client().Get(infra.Ctx(), infra.KCP().ObjKey(infra.SkrKymaRef().Name), scope)
				return err == nil, client.IgnoreNotFound(err)
			}).Should(BeTrue(), "expected Scope to get created")
		})

		By("And Given SKR GcpNfsVolume exists but is NOT ready", func() {
			skrgcpnfsvol.Ignore.AddName(skrNfsVolumeName)
			Eventually(CreateGcpNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), skrNfsVolume,
					WithName(skrNfsVolumeName),
					WithGcpNfsVolumeIpRange(skrIpRangeName),
				).Should(Succeed())
			// Intentionally NOT setting Ready condition
		})

		By("When GcpNfsBackupSchedule is created", func() {
			Eventually(CreateBackupSchedule).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					WithName(scheduleName),
					WithSchedule("* * * * *"),
					WithGcpLocation("us-west1"),
					WithNfsVolumeRef(skrNfsVolumeName),
					WithRetentionDays(0),
				).Should(Succeed())
		})

		By("And When fake clock advances past the first scheduled run", func() {
			// Wait for schedule to become Active with NextRunTimes first
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingStatusActive(),
				).Should(Succeed())

			testFakeClock.Step(2 * time.Minute)
		})

		By("Then GcpNfsBackupSchedule has Error state with NfsVolumeNotReady reason", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateError),
				).Should(Succeed())

			cond := findCondition(schedule.Status.Conditions, cloudresourcesv1beta1.ConditionTypeError)
			Expect(cond).NotTo(BeNil())
			Expect(cond.Reason).To(Equal(cloudresourcesv1beta1.ReasonNfsVolumeNotReady))
		})

		// Cleanup
		Eventually(Delete).
			WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
			Should(Succeed())
	})
})

// findCondition finds a condition by type from a list of conditions.
func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}
