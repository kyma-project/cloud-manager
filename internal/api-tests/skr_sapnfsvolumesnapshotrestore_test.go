package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testSapNfsVolumeSnapshotRestoreBuilder struct {
	instance cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore
}

func newTestSapNfsVolumeSnapshotRestoreExistingVolumeBuilder() *testSapNfsVolumeSnapshotRestoreBuilder {
	return &testSapNfsVolumeSnapshotRestoreBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreSpec{
				SourceSnapshot: cloudresourcesv1beta1.SapNfsVolumeSnapshotRef{
					Name: "test-snapshot",
				},
				Destination: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{
					ExistingVolume: &cloudresourcesv1beta1.SapNfsVolumeRef{
						Name: "test-volume",
					},
				},
			},
		},
	}
}

func newTestSapNfsVolumeSnapshotRestoreNewVolumeBuilder() *testSapNfsVolumeSnapshotRestoreBuilder {
	return &testSapNfsVolumeSnapshotRestoreBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreSpec{
				SourceSnapshot: cloudresourcesv1beta1.SapNfsVolumeSnapshotRef{
					Name: "test-snapshot",
				},
				Destination: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{
					NewVolume: &cloudresourcesv1beta1.SapNfsVolumeSnapshotNewVolume{
						Name:       "new-volume",
						CapacityGb: 100,
					},
				},
			},
		},
	}
}

func newTestSapNfsVolumeSnapshotRestoreBothBuilder() *testSapNfsVolumeSnapshotRestoreBuilder {
	return &testSapNfsVolumeSnapshotRestoreBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreSpec{
				SourceSnapshot: cloudresourcesv1beta1.SapNfsVolumeSnapshotRef{
					Name: "test-snapshot",
				},
				Destination: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{
					ExistingVolume: &cloudresourcesv1beta1.SapNfsVolumeRef{
						Name: "test-volume",
					},
					NewVolume: &cloudresourcesv1beta1.SapNfsVolumeSnapshotNewVolume{
						Name:       "new-volume",
						CapacityGb: 100,
					},
				},
			},
		},
	}
}

func newTestSapNfsVolumeSnapshotRestoreNeitherBuilder() *testSapNfsVolumeSnapshotRestoreBuilder {
	return &testSapNfsVolumeSnapshotRestoreBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreSpec{
				SourceSnapshot: cloudresourcesv1beta1.SapNfsVolumeSnapshotRef{
					Name: "test-snapshot",
				},
				Destination: cloudresourcesv1beta1.SapNfsVolumeSnapshotRestoreDestination{},
			},
		},
	}
}

func (b *testSapNfsVolumeSnapshotRestoreBuilder) Build() *cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore {
	return &b.instance
}

func (b *testSapNfsVolumeSnapshotRestoreBuilder) WithSourceSnapshot(name string) *testSapNfsVolumeSnapshotRestoreBuilder {
	b.instance.Spec.SourceSnapshot.Name = name
	return b
}

func (b *testSapNfsVolumeSnapshotRestoreBuilder) WithExistingVolume(name string) *testSapNfsVolumeSnapshotRestoreBuilder {
	b.instance.Spec.Destination.ExistingVolume = &cloudresourcesv1beta1.SapNfsVolumeRef{Name: name}
	b.instance.Spec.Destination.NewVolume = nil
	return b
}

var _ = Describe("Feature: SKR SapNfsVolumeSnapshotRestore", Ordered, func() {

	Context("Scenario: Destination validation", func() {

		canCreateSkr(
			"SapNfsVolumeSnapshotRestore with existingVolume",
			newTestSapNfsVolumeSnapshotRestoreExistingVolumeBuilder(),
		)

		canCreateSkr(
			"SapNfsVolumeSnapshotRestore with newVolume",
			newTestSapNfsVolumeSnapshotRestoreNewVolumeBuilder(),
		)

		canNotCreateSkr(
			"SapNfsVolumeSnapshotRestore with both existingVolume and newVolume",
			newTestSapNfsVolumeSnapshotRestoreBothBuilder(),
			"must have at most 1 item",
		)

		canNotCreateSkr(
			"SapNfsVolumeSnapshotRestore with neither existingVolume nor newVolume",
			newTestSapNfsVolumeSnapshotRestoreNeitherBuilder(),
			"should have at least 1 properties",
		)
	})

	Context("Scenario: Immutability", func() {

		canNotChangeSkr(
			"SapNfsVolumeSnapshotRestore SourceSnapshot is immutable",
			newTestSapNfsVolumeSnapshotRestoreExistingVolumeBuilder(),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore]) {
				b.(*testSapNfsVolumeSnapshotRestoreBuilder).WithSourceSnapshot("other-snapshot")
			},
			"SourceSnapshot is immutable",
		)

		canNotChangeSkr(
			"SapNfsVolumeSnapshotRestore Destination is immutable",
			newTestSapNfsVolumeSnapshotRestoreExistingVolumeBuilder(),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolumeSnapshotRestore]) {
				b.(*testSapNfsVolumeSnapshotRestoreBuilder).WithExistingVolume("other-volume")
			},
			"Destination is immutable",
		)
	})
})
