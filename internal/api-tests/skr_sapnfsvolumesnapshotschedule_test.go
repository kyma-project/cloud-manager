package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
)

type testSapNfsVolumeSnapshotScheduleBuilder struct {
	instance cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule
}

func newTestSapNfsVolumeSnapshotScheduleBuilder() *testSapNfsVolumeSnapshotScheduleBuilder {
	return &testSapNfsVolumeSnapshotScheduleBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotScheduleSpec{
				Schedule:           "0 0 * * *",
				MaxRetentionDays:   375,
				MaxReadySnapshots:  50,
				MaxFailedSnapshots: 5,
				Template: cloudresourcesv1beta1.SapNfsVolumeSnapshotTemplate{
					Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
						SourceVolume: corev1.ObjectReference{
							Name: "test-volume",
						},
					},
				},
			},
		},
	}
}

func (b *testSapNfsVolumeSnapshotScheduleBuilder) Build() *cloudresourcesv1beta1.SapNfsVolumeSnapshotSchedule {
	return &b.instance
}

func (b *testSapNfsVolumeSnapshotScheduleBuilder) WithSchedule(schedule string) *testSapNfsVolumeSnapshotScheduleBuilder {
	b.instance.Spec.Schedule = schedule
	return b
}

var _ = Describe("Feature: SKR SapNfsVolumeSnapshotSchedule", Ordered, func() {

	canCreateSkr(
		"SapNfsVolumeSnapshotSchedule with cron schedule",
		newTestSapNfsVolumeSnapshotScheduleBuilder(),
	)

	canCreateSkr(
		"SapNfsVolumeSnapshotSchedule one-time (no schedule)",
		newTestSapNfsVolumeSnapshotScheduleBuilder().WithSchedule(""),
	)
})
