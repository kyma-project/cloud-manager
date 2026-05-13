package cloudresources

import (
	"fmt"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	skrsapnfsvolume "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolume"
	skrsapnfsvolumesnapshot "github.com/kyma-project/cloud-manager/pkg/skr/sapnfsvolumesnapshot"
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
			skrsapnfsvolumesnapshot.Ignore.AddName(snapshotName)
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
		skrsapnfsvolumesnapshot.Ignore.AddName(snapshotName)
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
		olderSnapName := scheduleName + "-pre-older"
		newerSnapName := scheduleName + "-pre-newer"
		sapNfsVolume := &cloudresourcesv1beta1.SapNfsVolume{}
		schedule := &cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{}
		scope := &cloudcontrolv1beta1.Scope{}

		sapMock := infra.SapMock().NewProject()

		kymaName := "a9e593e1-a57f-4b83-85b2-d6e684b99d6f"
		skrKymaRef := util.Must(infra.ScopeProvider().GetScope(infra.Ctx(), types.NamespacedName{Name: kymaName}))

		skrsapnfsvolume.Ignore.AddName(sapNfsVolumeName)
		defer skrsapnfsvolume.Ignore.RemoveName(sapNfsVolumeName)

		// Ignore pre-created snapshots so the snapshot reconciler doesn't interfere
		skrsapnfsvolumesnapshot.Ignore.AddName(olderSnapName)
		skrsapnfsvolumesnapshot.Ignore.AddName(newerSnapName)

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

		// Pre-create two Ready snapshots with schedule labels so deleteSnapshots
		// can evaluate retention without multi-cycle clock races.
		olderSnap := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		By("And Given an older pre-existing Ready snapshot", func() {
			olderSnap = &cloudresourcesv1beta1.SapNfsVolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      olderSnapName,
					Namespace: DefaultSkrNamespace,
					Labels: map[string]string{
						cloudresourcesv1beta1.LabelScheduleName:      scheduleName,
						cloudresourcesv1beta1.LabelScheduleNamespace: DefaultSkrNamespace,
					},
				},
				Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
					SourceVolume: corev1.ObjectReference{
						Name:      sapNfsVolumeName,
						Namespace: DefaultSkrNamespace,
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), olderSnap)
			}).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), olderSnap,
					WithConditions(SkrReadyCondition()),
					WithState(cloudresourcesv1beta1.StateReady),
				).Should(Succeed())
		})

		newerSnap := &cloudresourcesv1beta1.SapNfsVolumeSnapshot{}
		By("And Given a newer pre-existing Ready snapshot", func() {
			newerSnap = &cloudresourcesv1beta1.SapNfsVolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newerSnapName,
					Namespace: DefaultSkrNamespace,
					Labels: map[string]string{
						cloudresourcesv1beta1.LabelScheduleName:      scheduleName,
						cloudresourcesv1beta1.LabelScheduleNamespace: DefaultSkrNamespace,
					},
				},
				Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
					SourceVolume: corev1.ObjectReference{
						Name:      sapNfsVolumeName,
						Namespace: DefaultSkrNamespace,
					},
				},
			}
			Eventually(func() error {
				return infra.SKR().Client().Create(infra.Ctx(), newerSnap)
			}).Should(Succeed())
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), newerSnap,
					WithConditions(SkrReadyCondition()),
					WithState(cloudresourcesv1beta1.StateReady),
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

		By("And When fake clock advances to trigger a run", func() {
			// Compute expected snapshot name and add to Ignore before stepping
			expected, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
			Expect(err).NotTo(HaveOccurred())
			newSnapName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, expected.Format("20060102-150405"))
			skrsapnfsvolumesnapshot.Ignore.AddName(newSnapName)
			testFakeClock.Step(2 * time.Minute)
		})

		By("Then the schedule creates a new snapshot", func() {
			Eventually(func() (int, error) {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), schedule, NewObjActions()); err != nil {
					return 0, err
				}
				// If cache race caused NextRunTimes recalculation past our clock, step again
				if schedule.Status.SnapshotIndex == 0 && len(schedule.Status.NextRunTimes) > 0 {
					nextRun, err := time.Parse(time.RFC3339, schedule.Status.NextRunTimes[0])
					if err == nil && nextRun.After(testFakeClock.Now()) {
						newSnapName := fmt.Sprintf("%s-%d-%s", scheduleName, 1, nextRun.UTC().Format("20060102-150405"))
						skrsapnfsvolumesnapshot.Ignore.AddName(newSnapName)
						testFakeClock.Step(2 * time.Minute)
					}
				}
				return schedule.Status.SnapshotIndex, nil
			}).Should(Equal(1))
		})

		By("And Then a pre-existing Ready snapshot is evicted by retention", func() {
			// With maxReadySnapshots=1, deleteSnapshots iterates newest-first:
			// - new snapshot (not Ready) → not counted
			// - one pre-existing (Ready) → readyCount=1, kept
			// - other pre-existing (Ready) → readyCount>=1, EVICTED
			// Check that the total snapshot count drops to 2
			Eventually(func() (int, error) {
				list := &cloudresourcesv1beta1.SapNfsVolumeSnapshotList{}
				err := infra.SKR().Client().List(infra.Ctx(), list,
					client.MatchingLabels{
						cloudresourcesv1beta1.LabelScheduleName:      scheduleName,
						cloudresourcesv1beta1.LabelScheduleNamespace: DefaultSkrNamespace,
					},
					client.InNamespace(DefaultSkrNamespace),
				)
				if err != nil {
					return 0, err
				}
				return len(list.Items), nil
			}).Should(Equal(2), "expected one snapshot to be evicted by retention policy")
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
