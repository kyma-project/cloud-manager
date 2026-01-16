package api_tests

import (
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testGcpNfsVolumeBuilder struct {
	instance cloudresourcesv1beta1.GcpNfsVolume
}

func newTestGcpNfsVolumeBuilder() *testGcpNfsVolumeBuilder {
	return &testGcpNfsVolumeBuilder{
		instance: cloudresourcesv1beta1.GcpNfsVolume{
			Spec: cloudresourcesv1beta1.GcpNfsVolumeSpec{},
		},
	}
}

func (b *testGcpNfsVolumeBuilder) Build() *cloudresourcesv1beta1.GcpNfsVolume {
	return &b.instance
}

func (b *testGcpNfsVolumeBuilder) WithTier(tier cloudresourcesv1beta1.GcpFileTier) *testGcpNfsVolumeBuilder {
	b.instance.Spec.Tier = tier
	return b
}

func (b *testGcpNfsVolumeBuilder) WithCapacityGb(capacityGb int) *testGcpNfsVolumeBuilder {
	b.instance.Spec.CapacityGb = capacityGb
	return b
}

func (b *testGcpNfsVolumeBuilder) WithFileShareName(fileShareName string) *testGcpNfsVolumeBuilder {
	b.instance.Spec.FileShareName = fileShareName
	return b
}

func (b *testGcpNfsVolumeBuilder) WithValidFileShareName() *testGcpNfsVolumeBuilder {
	b.instance.Spec.FileShareName = "foo"
	return b
}

var _ = Describe("Feature: SKR GcpNfsVolume", Ordered, func() {

	fileShareName17char := "file12345678901234567"
	fileShareName64char := "file1234567890123456789012345678901234567890123456789012345678901234"

	// REGIONAL and ZONAL tiers have same constraints
	for _, tier := range []cloudresourcesv1beta1.GcpFileTier{cloudresourcesv1beta1.REGIONAL, cloudresourcesv1beta1.ZONAL} {
		for _, validCapacity := range []int{1024, 1280, 9984, 10240, 12800, 102400} {
			canCreateSkr(
				fmt.Sprintf("GcpNfsVolume %s tier instance can be created with valid capacity: %d", tier, validCapacity),
				newTestGcpNfsVolumeBuilder().WithTier(tier).WithCapacityGb(validCapacity).WithValidFileShareName(),
			)
		}
		for _, invalidCapacity := range []int{0, 1, 1023, 1025, 10496, 102401, 104960} {
			canNotCreateSkr(
				fmt.Sprintf("GcpNfsVolume %s tier instance can not be created with invalid capacity: %d", tier, invalidCapacity),
				newTestGcpNfsVolumeBuilder().WithTier(tier).WithCapacityGb(invalidCapacity).WithValidFileShareName(),
				fmt.Sprintf("%s tier capacityGb must be between 1024 and 9984, and it must be divisble by 256, or between 10240 and 102400, and divisible by 2560", tier),
			)
		}
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume %s tier instance can not be created with invalid fileShareName length", tier),
			newTestGcpNfsVolumeBuilder().WithTier(tier).WithCapacityGb(1024).WithFileShareName(fileShareName64char),
			fmt.Sprintf("%s tier fileShareName length must be 63 or less characters", tier),
		)
		canChangeSkr(
			fmt.Sprintf("GcpNfsVolume %s tier instance capacity can be increased", tier),
			newTestGcpNfsVolumeBuilder().WithTier(tier).WithCapacityGb(1024).WithValidFileShareName(),
			func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
				b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1280)
			},
		)
		canChangeSkr(
			fmt.Sprintf("GcpNfsVolume %s tier instance capacity can be reduced", tier),
			newTestGcpNfsVolumeBuilder().WithTier(tier).WithCapacityGb(1280).WithValidFileShareName(),
			func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
				b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1024)
			},
		)
	}

	for _, validCapacity := range []int{2560, 2561, 65432, 65433} {
		canCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_SSD tier instance can be created with valid capacity: %d", validCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(validCapacity).WithValidFileShareName(),
		)
	}
	for _, invalidCapacity := range []int{0, 1, 2559, 65434} {
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_SSD tier instance can not be created with invalid capacity: %d", invalidCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(invalidCapacity).WithValidFileShareName(),
			"BASIC_SSD tier capacityGb must be between 2560 and 65433",
		)
	}
	canNotCreateSkr(
		"GcpNfsVolume BASIC_SSD tier instance can not be created with invalid fileShareName length",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(1024).WithFileShareName(fileShareName17char),
		"BASIC_SSD tier fileShareName length must be 16 or less characters",
	)
	canChangeSkr(
		"GcpNfsVolume BASIC_SSD tier instance capacity can be increased",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(2560).WithValidFileShareName(),
		func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
			b.(*testGcpNfsVolumeBuilder).WithCapacityGb(2561)
		},
	)
	canNotChangeSkr(
		"GcpNfsVolume BASIC_SSD tier instance capacity can not be reduced",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(2561).WithValidFileShareName(),
		func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
			b.(*testGcpNfsVolumeBuilder).WithCapacityGb(2560)
		},
		"BASIC_SSD tier capacityGb cannot be reduced",
	)

	for _, validCapacity := range []int{1024, 1025, 65432, 65433} {
		canCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_HDD tier instance can be created with valid capacity: %d", validCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_HDD).WithCapacityGb(validCapacity).WithValidFileShareName(),
		)
	}
	for _, invalidCapacity := range []int{0, 1, 1023, 65434} {
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_HDD tier instance can not be created with invalid capacity: %d", invalidCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_HDD).WithCapacityGb(invalidCapacity).WithValidFileShareName(),
			"BASIC_HDD tier capacityGb must be between 1024 and 65433",
		)
	}
	canNotCreateSkr(
		"GcpNfsVolume BASIC_HDD tier instance can not be created with invalid fileShareName length",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_HDD).WithCapacityGb(1024).WithFileShareName(fileShareName17char),
		"BASIC_HDD tier fileShareName length must be 16 or less characters",
	)
	canChangeSkr(
		"GcpNfsVolume BASIC_HDD tier instance capacity can be increased",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_HDD).WithCapacityGb(1024).WithValidFileShareName(),
		func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
			b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1025)
		},
	)
	canNotChangeSkr(
		"GcpNfsVolume BASIC_HDD tier instance capacity can not be reduced",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_HDD).WithCapacityGb(1025).WithValidFileShareName(),
		func(b Builder[*cloudresourcesv1beta1.GcpNfsVolume]) {
			b.(*testGcpNfsVolumeBuilder).WithCapacityGb(1024)
		},
		"BASIC_HDD tier capacityGb cannot be reduced",
	)

	for _, removedTier := range []string{"STANDARD", "PREMIUM", "HIGH_SCALE_SSD", "ENTERPRISE"} {
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume cannot be created with removed tier: %s", removedTier),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.GcpFileTier(removedTier)).WithCapacityGb(1024).WithValidFileShareName(),
			fmt.Sprintf("spec.tier: Unsupported value: \"%s\"", removedTier),
		)
	}
})
