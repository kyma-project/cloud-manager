package api_tests

import (
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

	It("Given SKR default namespace exists", func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

	for _, validCapacity := range []int{1024, 1280, 9984, 10240, 12800, 102400} {
		canCreateSkr(
			fmt.Sprintf("GcpNfsVolume REGIONAL tier instance can be created with valid capacity: %d", validCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(validCapacity).WithValidFileShareName(),
		)
	}
	for _, invalidCapacity := range []int{0, 1, 1023, 1025, 10496, 102401, 104960} {
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume REGIONAL tier instance can not be created with invalid capacity: %d", invalidCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(invalidCapacity).WithValidFileShareName(),
			"REGIONAL tier capacityGb must be between 1024 and 9984, and it must be divisble by 256, or between 10240 and 102400, and divisible by 2560",
		)
	}
	fileShareName65char := "tcteafkhhfhxkocrtvbvgrzqvysxpfxeeauvgwqnbassacgejobhcuvjvdlrgbkypkuxteaztzjxrdfipqfxdpercpogqdslhm"
	canNotCreateSkr(
		"GcpNfsVolume REGIONAL tier instance can not be created with invalid fileShareName length",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.REGIONAL).WithCapacityGb(1024).WithFileShareName(fileShareName65char),
		"REGIONAL tier fileShareName length must be 64 or less characters",
	)

	for _, validCapacity := range []int{2560, 2561, 65399, 65400} {
		canCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_SSD tier instance can be created with valid capacity: %d", validCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(validCapacity).WithValidFileShareName(),
		)
	}
	for _, invalidCapacity := range []int{0, 1, 2559, 65401} {
		canNotCreateSkr(
			fmt.Sprintf("GcpNfsVolume BASIC_SSD tier instance can not be created with invalid capacity: %d", invalidCapacity),
			newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(invalidCapacity).WithValidFileShareName(),
			"BASIC_SSD tier capacityGb must be between 2560 and 65400",
		)
	}
	fileShareName17char := "bwjfjlecorewsakjikpj"
	canNotCreateSkr(
		"GcpNfsVolume REGIONAL tier instance can not be created with invalid fileShareName length",
		newTestGcpNfsVolumeBuilder().WithTier(cloudresourcesv1beta1.BASIC_SSD).WithCapacityGb(1024).WithFileShareName(fileShareName17char),
		"BASIC_SSD tier fileShareName length must be 16 or less characters",
	)

})
