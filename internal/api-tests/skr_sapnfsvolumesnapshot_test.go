package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
)

type testSapNfsVolumeSnapshotBuilder struct {
	instance cloudresourcesv1beta1.SapNfsVolumeSnapshot
}

func newTestSapNfsVolumeSnapshotBuilder() *testSapNfsVolumeSnapshotBuilder {
	return &testSapNfsVolumeSnapshotBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolumeSnapshot{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSnapshotSpec{
				SourceVolume: corev1.ObjectReference{
					Name: "test-volume",
				},
			},
		},
	}
}

func (b *testSapNfsVolumeSnapshotBuilder) Build() *cloudresourcesv1beta1.SapNfsVolumeSnapshot {
	return &b.instance
}

func (b *testSapNfsVolumeSnapshotBuilder) WithSourceVolume(name string) *testSapNfsVolumeSnapshotBuilder {
	b.instance.Spec.SourceVolume.Name = name
	return b
}

var _ = Describe("Feature: SKR SapNfsVolumeSnapshot", Ordered, func() {

	canCreateSkr(
		"SapNfsVolumeSnapshot with sourceVolume",
		newTestSapNfsVolumeSnapshotBuilder().WithSourceVolume("my-volume"),
	)

	canNotChangeSkr(
		"SapNfsVolumeSnapshot SourceVolume is immutable",
		newTestSapNfsVolumeSnapshotBuilder().WithSourceVolume("my-volume"),
		func(b Builder[*cloudresourcesv1beta1.SapNfsVolumeSnapshot]) {
			b.(*testSapNfsVolumeSnapshotBuilder).WithSourceVolume("other-volume")
		},
		"SourceVolume is immutable",
	)
})
