package cloudresources

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrsapnfsvolume "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolume"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: SKR SapNfsVolumeSnapshotSchedule", func() {

	It("Scenario: Recurring schedule with cascade delete", func() {
		sapNfsVolumeName := "66da7e12-fe42-4b43-b738-b4319fae08d1"
		scheduleName := "f5d8b057-3eea-43c5-869b-290b11543291"
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		schedule := &cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "072cfdff-6c78-47b8-90c7-9dfce8a974ce"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given SapNfsVolume exists in Ready state", func() {
			Eventually(CreateSapNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithName(sapNfsVolumeName),
					WithSapNfsVolumeCapacity(100),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshotSchedule is created with recurring cron", func() {
			schedule.Name = scheduleName
			schedule.Namespace = DefaultSkrNamespace
			schedule.Spec = cloudresourcesv1beta1.SapNfsVolumeSnapshotScheduleSpec{
				Schedule:           "* * * * *",
				DeleteCascade:      true,
				MaxRetentionDays:   30,
				MaxReadySnapshots:  50,
				MaxFailedSnapshots: 5,
				Template: cloudresourcesv1beta1.SapNfsVolumeSnapshotTemplate{
					Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
						SourceVolume: corev1.ObjectReference{
							Name:      sapNfsVolumeName,
							Namespace: DefaultSkrNamespace,
						},
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), schedule)
			}).Should(Succeed())
		})

		By("Then SapNfsVolumeSnapshotSchedule becomes Active with NextRunTimes", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateActive),
				).Should(Succeed())
			Expect(len(schedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		var snapshotName string
		By("And When fake clock advances past the first scheduled run", func() {
			expected, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			snapshotName = fmt.Sprintf("%s-%d-%s", scheduleName, 1, expected.Format("20060102-150405"))
			testFakeClock.Step(2 * time.Minute)
		})

		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		By("Then a SapNfsVolumeSnapshot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(WithName(snapshotName)),
				).Should(Succeed())
		})

		By("And Then the snapshot has correct labels and spec", func() {
			Expect(snapshot.Labels[cloudresourcesv1beta1.LabelScheduleName]).To(Equal(scheduleName))
			Expect(snapshot.Labels[cloudresourcesv1beta1.LabelScheduleNamespace]).To(Equal(DefaultSkrNamespace))
			Expect(snapshot.Spec.SourceVolume.Name).To(Equal(sapNfsVolumeName))
			Expect(snapshot.Spec.DeleteAfterDays).To(Equal(30))
		})

		By("And Then schedule status is updated", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				return schedule.Status.SnapshotIndex, nil
			}).Should(Equal(1))
			Expect(schedule.Status.BackupCount).To(Equal(1))
			Expect(schedule.Status.LastCreatedBackup.Name).To(Equal(snapshotName))
		})

		// DELETE with cascade
		By("When SapNfsVolumeSnapshotSchedule is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})

		By("Then the created snapshot is deleted by cascade", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), snapshot).
				Should(Succeed(), "expected snapshot to be cascade-deleted")
		})

		By("And Then the schedule is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})
	})

	It("Scenario: One-time schedule", func() {
		sapNfsVolumeName := "70f36f4a-9c4e-4c58-87e7-28eb6dcf1cb2"
		scheduleName := "c5455655-a6ec-4bef-afe1-6d10a6dfa6d6"
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		schedule := &cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		fakeNow := testFakeClock.Now().UTC()
		startTime := time.Date(fakeNow.Year(), fakeNow.Month(), fakeNow.Day(),
			fakeNow.Hour(), fakeNow.Minute()+2, 0, 0, time.UTC)

		kymaName := "0674c653-0016-4236-86c1-d53c6a4527d5"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given SapNfsVolume exists in Ready state", func() {
			Eventually(CreateSapNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithName(sapNfsVolumeName),
					WithSapNfsVolumeCapacity(100),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshotSchedule is created as one-time", func() {
			schedule.Name = scheduleName
			schedule.Namespace = DefaultSkrNamespace
			schedule.Spec = cloudresourcesv1beta1.SapNfsVolumeSnapshotScheduleSpec{
				StartTime:        &metav1.Time{Time: startTime},
				MaxRetentionDays: 30,
				Template: cloudresourcesv1beta1.SapNfsVolumeSnapshotTemplate{
					Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
						SourceVolume: corev1.ObjectReference{
							Name:      sapNfsVolumeName,
							Namespace: DefaultSkrNamespace,
						},
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), schedule)
			}).Should(Succeed())
		})

		By("Then schedule becomes Active", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateActive),
				).Should(Succeed())
		})

		By("And When fake clock advances past start time", func() {
			testFakeClock.Step(3 * time.Minute)
		})

		snapshotName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, startTime.Format("20060102-150405"))
		snapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}

		By("Then a single SapNfsVolumeSnapshot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), snapshot,
					NewObjActions(WithName(snapshotName)),
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
		By("// Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})
	})

	It("Scenario: Retention count-based eviction", func() {
		sapNfsVolumeName := "67185c99-7382-40fb-8ad2-26aaf4d75427"
		scheduleName := "63b177f9-4000-4d65-999c-9112cbb45535"
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		schedule := &cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "a9e593e1-a57f-4b83-85b2-d6e684b99d6f"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		By("Given KCP Scope exists", func() {
			Eventually(func() error {
				return GivenScopeOpenStackExists(infra.Ctx(), infra, scope, sapMock.ProviderParams(), WithName(skrKymaRef.Name))
			}).Should(Succeed())
		})

		By("And Given SapNfsVolume exists in Ready state", func() {
			Eventually(CreateSapNfsVolume).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithName(sapNfsVolumeName),
					WithSapNfsVolumeCapacity(100),
				).Should(Succeed())

			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), sapNfsVolume,
					WithConditions(SkrReadyCondition()),
				).Should(Succeed())
		})

		By("When SapNfsVolumeSnapshotSchedule is created with maxReadySnapshots=1", func() {
			schedule.Name = scheduleName
			schedule.Namespace = DefaultSkrNamespace
			schedule.Spec = cloudresourcesv1beta1.SapNfsVolumeSnapshotScheduleSpec{
				Schedule:           "* * * * *",
				MaxRetentionDays:   30,
				MaxReadySnapshots:  1,
				MaxFailedSnapshots: 5,
				Template: cloudresourcesv1beta1.SapNfsVolumeSnapshotTemplate{
					Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
						SourceVolume: corev1.ObjectReference{
							Name:      sapNfsVolumeName,
							Namespace: DefaultSkrNamespace,
						},
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), schedule)
			}).Should(Succeed())
		})

		By("Then schedule becomes Active", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), schedule,
					NewObjActions(),
					HavingState(cloudresourcesv1beta1.JobStateActive),
				).Should(Succeed())
			Expect(len(schedule.Status.NextRunTimes)).To(BeNumerically(">", 0))
		})

		// Create first snapshot
		var firstSnapshotName string
		By("And When fake clock advances to create the first snapshot", func() {
			expected, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			firstSnapshotName = fmt.Sprintf("%s-%d-%s", scheduleName, 1, expected.Format("20060102-150405"))
			testFakeClock.Step(2 * time.Minute)
		})

		firstSnapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		By("Then the first snapshot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), firstSnapshot,
					NewObjActions(WithName(firstSnapshotName)),
				).Should(Succeed())
		})

		// Mark first snapshot as Ready (so it counts toward maxReadySnapshots)
		By("And Given the first snapshot is marked Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), firstSnapshot,
					WithConditions(SkrReadyCondition()),
					WithState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		// Wait for schedule to pick up the second run time and advance
		By("And When fake clock advances to create the second snapshot", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				return schedule.Status.SnapshotIndex, nil
			}).Should(Equal(1))
			testFakeClock.Step(2 * time.Minute)
		})

		By("Then the second snapshot is created", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				return schedule.Status.SnapshotIndex, nil
			}).Should(Equal(2))
		})

		secondSnapshot := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		By("And Then the second snapshot exists", func() {
			// Find second snapshot by label
			Eventually(func() error {
				list := &cloudresourcesv1beta1.SapNfsVolumeSnapshotList{}
				err := infra.SKR().Client().List(infra.Ctx(), list,
					client.MatchingLabels{
						cloudresourcesv1beta1.LabelScheduleName:      scheduleName,
						cloudresourcesv1beta1.LabelScheduleNamespace: DefaultSkrNamespace,
					},
					client.InNamespace(DefaultSkrNamespace),
				)
				if err != nil {
					return err
				}
				for i := range list.Items {
					if list.Items[i].Name != firstSnapshotName {
						secondSnapshot = &list.Items[i]
						return nil
					}
				}
				return fmt.Errorf("second snapshot not found")
			}).Should(Succeed())
		})

		// Mark second snapshot as Ready
		By("And Given the second snapshot is marked Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), secondSnapshot,
					WithConditions(SkrReadyCondition()),
					WithState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		// Create third snapshot (should trigger eviction of oldest)
		By("And When fake clock advances to create the third snapshot", func() {
			// Wait for second cycle to complete and NextRunTimes to be recalculated
			Eventually(func() (bool, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return false, err
				}
				return schedule.Status.SnapshotIndex == 2 && len(schedule.Status.NextRunTimes) > 0, nil
			}).Should(BeTrue())
			testFakeClock.Step(2 * time.Minute)
		})

		By("Then the third snapshot is created", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				return schedule.Status.SnapshotIndex, nil
			}).Should(Equal(3))
		})

		By("And Then the oldest (first) snapshot is evicted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), firstSnapshot).
				Should(Succeed(), "expected oldest snapshot to be evicted by retention policy")
		})

		// Cleanup
		By("// Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), schedule).
				Should(Succeed())
		})
	})
})
