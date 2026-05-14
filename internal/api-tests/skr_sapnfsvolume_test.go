package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
)

type testSapNfsVolumeBuilder struct {
	instance cloudresourcesv1beta1.SapNfsVolume
}

func newTestSapNfsVolumeBuilder() *testSapNfsVolumeBuilder {
	return &testSapNfsVolumeBuilder{
		instance: cloudresourcesv1beta1.SapNfsVolume{
			Spec: cloudresourcesv1beta1.SapNfsVolumeSpec{
				CapacityGb: 100,
			},
		},
	}
}

func (b *testSapNfsVolumeBuilder) Build() *cloudresourcesv1beta1.SapNfsVolume {
	return &b.instance
}

func (b *testSapNfsVolumeBuilder) WithCapacityGb(capacityGb int) *testSapNfsVolumeBuilder {
	b.instance.Spec.CapacityGb = capacityGb
	return b
}

func (b *testSapNfsVolumeBuilder) WithIpRange(name string) *testSapNfsVolumeBuilder {
	b.instance.Spec.IpRange = cloudresourcesv1beta1.IpRangeRef{Name: name}
	return b
}

func (b *testSapNfsVolumeBuilder) WithVolumeName(name string) *testSapNfsVolumeBuilder {
	if b.instance.Spec.PersistentVolume == nil {
		b.instance.Spec.PersistentVolume = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
	}
	b.instance.Spec.PersistentVolume.Name = name
	return b
}

func (b *testSapNfsVolumeBuilder) WithVolumeClaimName(name string) *testSapNfsVolumeBuilder {
	if b.instance.Spec.PersistentVolumeClaim == nil {
		b.instance.Spec.PersistentVolumeClaim = &cloudresourcesv1beta1.NameLabelsAnnotationsSpec{}
	}
	b.instance.Spec.PersistentVolumeClaim.Name = name
	return b
}

func (b *testSapNfsVolumeBuilder) WithDataSourceSnapshot(name string) *testSapNfsVolumeBuilder {
	b.instance.Spec.DataSource = &cloudresourcesv1beta1.SapNfsVolumeDataSource{
		Snapshot: &corev1.ObjectReference{Name: name},
	}
	return b
}

func (b *testSapNfsVolumeBuilder) WithEmptyDataSource() *testSapNfsVolumeBuilder {
	b.instance.Spec.DataSource = &cloudresourcesv1beta1.SapNfsVolumeDataSource{}
	return b
}

func (b *testSapNfsVolumeBuilder) WithDataSourceSnapshotChanged(name string) *testSapNfsVolumeBuilder {
	b.instance.Spec.DataSource.Snapshot.Name = name
	return b
}

var _ = Describe("Feature: SKR SapNfsVolume", Ordered, func() {

	Context("Scenario: CapacityGb validation", func() {

		canCreateSkr(
			"SapNfsVolume can be created with capacityGb = 1",
			newTestSapNfsVolumeBuilder().WithCapacityGb(1),
		)

		canNotCreateSkr(
			"SapNfsVolume cannot be created with capacityGb = 0",
			newTestSapNfsVolumeBuilder().WithCapacityGb(0),
			"The field capacityGb must be greater than zero",
		)

		canNotCreateSkr(
			"SapNfsVolume cannot be created with negative capacityGb",
			newTestSapNfsVolumeBuilder().WithCapacityGb(-1),
			"The field capacityGb must be greater than zero",
		)

		canChangeSkr(
			"SapNfsVolume capacityGb can be changed",
			newTestSapNfsVolumeBuilder().WithCapacityGb(100),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolume]) {
				b.(*testSapNfsVolumeBuilder).WithCapacityGb(200)
			},
		)
	})

	Context("Scenario: IpRange", func() {

		canCreateSkr(
			"SapNfsVolume can be created with ipRange",
			newTestSapNfsVolumeBuilder().WithIpRange("my-iprange"),
		)

		canChangeSkr(
			"SapNfsVolume ipRange can be changed",
			newTestSapNfsVolumeBuilder().WithIpRange("my-iprange"),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolume]) {
				b.(*testSapNfsVolumeBuilder).WithIpRange("other-iprange")
			},
		)
	})

	Context("Scenario: Volume and VolumeClaim metadata", func() {

		canCreateSkr(
			"SapNfsVolume can be created with volume name",
			newTestSapNfsVolumeBuilder().WithVolumeName("my-pv"),
		)

		canChangeSkr(
			"SapNfsVolume volume name can be changed",
			newTestSapNfsVolumeBuilder().WithVolumeName("my-pv"),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolume]) {
				b.(*testSapNfsVolumeBuilder).WithVolumeName("other-pv")
			},
		)

		canCreateSkr(
			"SapNfsVolume can be created with volumeClaim name",
			newTestSapNfsVolumeBuilder().WithVolumeClaimName("my-pvc"),
		)

		canChangeSkr(
			"SapNfsVolume volumeClaim name can be changed",
			newTestSapNfsVolumeBuilder().WithVolumeClaimName("my-pvc"),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolume]) {
				b.(*testSapNfsVolumeBuilder).WithVolumeClaimName("other-pvc")
			},
		)
	})

	Context("Scenario: DataSource validation", func() {

		canCreateSkr(
			"SapNfsVolume can be created without dataSource",
			newTestSapNfsVolumeBuilder(),
		)

		canCreateSkr(
			"SapNfsVolume can be created with dataSource.snapshot",
			newTestSapNfsVolumeBuilder().WithDataSourceSnapshot("test-snapshot"),
		)

		canNotCreateSkr(
			"SapNfsVolume cannot be created with empty dataSource",
			newTestSapNfsVolumeBuilder().WithEmptyDataSource(),
			"should have at least 1 properties",
		)
	})

	Context("Scenario: Immutability", func() {

		canNotChangeSkr(
			"SapNfsVolume DataSource is immutable",
			newTestSapNfsVolumeBuilder().WithDataSourceSnapshot("original-snapshot"),
			func(b Builder[*cloudresourcesv1beta1.SapNfsVolume]) {
				b.(*testSapNfsVolumeBuilder).WithDataSourceSnapshotChanged("other-snapshot")
			},
			"DataSource is immutable",
		)
	})
})
