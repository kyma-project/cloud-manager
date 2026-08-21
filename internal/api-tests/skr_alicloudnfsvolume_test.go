package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAlicloudNfsVolumeBuilder struct {
	*cloudresourcesv1beta1.AlicloudNfsVolumeBuilder
}

func newTestAlicloudNfsVolumeBuilder() *testAlicloudNfsVolumeBuilder {
	return &testAlicloudNfsVolumeBuilder{
		AlicloudNfsVolumeBuilder: cloudresourcesv1beta1.NewAlicloudNfsVolumeBuilder().
			WithCapacity("100G"),
	}
}

func (b *testAlicloudNfsVolumeBuilder) Build() *cloudresourcesv1beta1.AlicloudNfsVolume {
	return &b.AlicloudNfsVolume
}

func (b *testAlicloudNfsVolumeBuilder) WithCapacity(capacity string) *testAlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolumeBuilder.WithCapacity(capacity)
	return b
}

func (b *testAlicloudNfsVolumeBuilder) WithStorageType(storageType cloudresourcesv1beta1.AlicloudNfsStorageType) *testAlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolumeBuilder.WithStorageType(storageType)
	return b
}

func (b *testAlicloudNfsVolumeBuilder) WithIpRange(ipRangeName string) *testAlicloudNfsVolumeBuilder {
	b.AlicloudNfsVolumeBuilder.WithIpRange(ipRangeName)
	return b
}

var _ = Describe("Feature: SKR AlicloudNfsVolume", Ordered, func() {

	Context("Scenario: capacity", func() {

		canCreateSkr(
			"AlicloudNfsVolume can be created with capacity",
			newTestAlicloudNfsVolumeBuilder().WithCapacity("100G"),
		)
	})

	Context("Scenario: storageType enum validation", func() {

		canCreateSkr(
			"AlicloudNfsVolume can be created with Performance storageType",
			newTestAlicloudNfsVolumeBuilder().WithStorageType(cloudresourcesv1beta1.AlicloudNfsStorageTypePerformance),
		)

		canCreateSkr(
			"AlicloudNfsVolume can be created with Capacity storageType",
			newTestAlicloudNfsVolumeBuilder().WithStorageType(cloudresourcesv1beta1.AlicloudNfsStorageTypeCapacity),
		)

		canCreateSkr(
			"AlicloudNfsVolume can be created with Premium storageType",
			newTestAlicloudNfsVolumeBuilder().WithStorageType(cloudresourcesv1beta1.AlicloudNfsStorageTypePremium),
		)

		canCreateSkr(
			"AlicloudNfsVolume can be created without storageType (server-side default applied)",
			newTestAlicloudNfsVolumeBuilder(),
		)

		canNotCreateSkr(
			"AlicloudNfsVolume cannot be created with invalid storageType",
			newTestAlicloudNfsVolumeBuilder().WithStorageType("Ultra"),
			"",
		)
	})

})
