package api_tests

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/api/resource"
)

type testKcpNfsInstanceGcpBuilder struct {
	instance cloudcontrolv1beta1.NfsInstance
}

func newTestKcpNfsInstanceGcpBuilder() *testKcpNfsInstanceGcpBuilder {
	return &testKcpNfsInstanceGcpBuilder{
		instance: cloudcontrolv1beta1.NfsInstance{
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Namespace: "default",
					Name:      "test",
				},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: "test-iprange",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: "test-scope",
				},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{},
			},
		},
	}
}

func (b *testKcpNfsInstanceGcpBuilder) Build() *cloudcontrolv1beta1.NfsInstance {
	return &b.instance
}

func (b *testKcpNfsInstanceGcpBuilder) WithTier(tier cloudcontrolv1beta1.GcpFileTier) *testKcpNfsInstanceGcpBuilder {
	if b.instance.Spec.Instance.Gcp == nil {
		b.instance.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
	}
	b.instance.Spec.Instance.Gcp.Tier = tier
	return b
}

func (b *testKcpNfsInstanceGcpBuilder) WithCapacityGb(capacityGb int) *testKcpNfsInstanceGcpBuilder {
	if b.instance.Spec.Instance.Gcp == nil {
		b.instance.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
	}
	b.instance.Spec.Instance.Gcp.CapacityGb = capacityGb
	return b
}

func (b *testKcpNfsInstanceGcpBuilder) WithLocation(location string) *testKcpNfsInstanceGcpBuilder {
	if b.instance.Spec.Instance.Gcp == nil {
		b.instance.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
	}
	b.instance.Spec.Instance.Gcp.Location = location
	return b
}

func (b *testKcpNfsInstanceGcpBuilder) WithFileShareName(name string) *testKcpNfsInstanceGcpBuilder {
	if b.instance.Spec.Instance.Gcp == nil {
		b.instance.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
	}
	b.instance.Spec.Instance.Gcp.FileShareName = name
	return b
}

func (b *testKcpNfsInstanceGcpBuilder) WithConnectMode(mode cloudcontrolv1beta1.GcpConnectMode) *testKcpNfsInstanceGcpBuilder {
	if b.instance.Spec.Instance.Gcp == nil {
		b.instance.Spec.Instance.Gcp = &cloudcontrolv1beta1.NfsInstanceGcp{}
	}
	b.instance.Spec.Instance.Gcp.ConnectMode = mode
	return b
}

func (b *testKcpNfsInstanceGcpBuilder) WithValidDefaults() *testKcpNfsInstanceGcpBuilder {
	return b.
		WithLocation("us-central1").
		WithFileShareName("vol1").
		WithConnectMode(cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS)
}

func (b *testKcpNfsInstanceGcpBuilder) WithStatusCapacityGb(capacityGb int) *testKcpNfsInstanceGcpBuilder {
	b.instance.Status.CapacityGb = capacityGb
	b.instance.Status.Capacity = *resource.NewQuantity(int64(capacityGb)*1024*1024*1024, resource.BinarySI)
	return b
}

var _ = Describe("Feature: KCP NfsInstance GCP", Ordered, func() {

	// ========================================================================
	// BASIC_HDD Tier Tests
	// ========================================================================

	Describe("BASIC_HDD tier capacity validation", func() {

		for _, validCapacity := range []int{1024, 2048, 32768, 65433} {
			canCreateKcp(
				fmt.Sprintf("NfsInstance GCP BASIC_HDD tier can be created with valid capacity: %d", validCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.BASIC_HDD).
					WithCapacityGb(validCapacity).
					WithValidDefaults(),
			)
		}

		for _, invalidCapacity := range []int{0, 1, 1023, 65434, 100000} {
			canNotCreateKcp(
				fmt.Sprintf("NfsInstance GCP BASIC_HDD tier cannot be created with invalid capacity: %d", invalidCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.BASIC_HDD).
					WithCapacityGb(invalidCapacity).
					WithValidDefaults(),
				"BASIC_HDD tier capacityGb must be between 1024 and 65433",
			)
		}

		canChangeKcp(
			"NfsInstance GCP BASIC_HDD tier capacity can be increased",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.BASIC_HDD).
				WithCapacityGb(1024).
				WithValidDefaults().
				WithStatusCapacityGb(1024),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(2048)
			},
		)

		canNotChangeKcp(
			"NfsInstance GCP BASIC_HDD tier capacity cannot be reduced",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.BASIC_HDD).
				WithCapacityGb(2048).
				WithValidDefaults().
				WithStatusCapacityGb(2048),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(1024)
			},
			"Cannot reduce capacity for BASIC_HDD tier instances",
		)
	})

	// ========================================================================
	// BASIC_SSD Tier Tests
	// ========================================================================

	Describe("BASIC_SSD tier capacity validation", func() {

		for _, validCapacity := range []int{2560, 3000, 32768, 65433} {
			canCreateKcp(
				fmt.Sprintf("NfsInstance GCP BASIC_SSD tier can be created with valid capacity: %d", validCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.BASIC_SSD).
					WithCapacityGb(validCapacity).
					WithValidDefaults(),
			)
		}

		for _, invalidCapacity := range []int{0, 1, 2559, 65434, 100000} {
			canNotCreateKcp(
				fmt.Sprintf("NfsInstance GCP BASIC_SSD tier cannot be created with invalid capacity: %d", invalidCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.BASIC_SSD).
					WithCapacityGb(invalidCapacity).
					WithValidDefaults(),
				"BASIC_SSD tier capacityGb must be between 2560 and 65433",
			)
		}

		canChangeKcp(
			"NfsInstance GCP BASIC_SSD tier capacity can be increased",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.BASIC_SSD).
				WithCapacityGb(2560).
				WithValidDefaults().
				WithStatusCapacityGb(2560),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(3000)
			},
		)

		canNotChangeKcp(
			"NfsInstance GCP BASIC_SSD tier capacity cannot be reduced",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.BASIC_SSD).
				WithCapacityGb(3000).
				WithValidDefaults().
				WithStatusCapacityGb(3000),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(2560)
			},
			"Cannot reduce capacity for BASIC_SSD tier instances",
		)
	})

	// ========================================================================
	// ZONAL Tier Tests
	// ========================================================================

	Describe("ZONAL tier capacity validation", func() {

		for _, validCapacity := range []int{1024, 1280, 1536, 5120, 10240} {
			canCreateKcp(
				fmt.Sprintf("NfsInstance GCP ZONAL tier can be created with valid capacity: %d", validCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.ZONAL).
					WithCapacityGb(validCapacity).
					WithValidDefaults(),
			)
		}

		for _, invalidCapacity := range []int{0, 1023, 1025, 1200, 10241, 20000} {
			canNotCreateKcp(
				fmt.Sprintf("NfsInstance GCP ZONAL tier cannot be created with invalid capacity: %d", invalidCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.ZONAL).
					WithCapacityGb(invalidCapacity).
					WithValidDefaults(),
				"ZONAL tier capacityGb must be between 1024 and 10240, and divisible by 256",
			)
		}

		canChangeKcp(
			"NfsInstance GCP ZONAL tier capacity can be increased",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.ZONAL).
				WithCapacityGb(1024).
				WithValidDefaults().
				WithStatusCapacityGb(1024),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(1280)
			},
		)

		canChangeKcp(
			"NfsInstance GCP ZONAL tier capacity can be reduced",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.ZONAL).
				WithCapacityGb(1280).
				WithValidDefaults().
				WithStatusCapacityGb(1280),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(1024)
			},
		)
	})

	// ========================================================================
	// REGIONAL Tier Tests
	// ========================================================================

	Describe("REGIONAL tier capacity validation", func() {

		for _, validCapacity := range []int{1024, 1280, 1536, 5120, 10240} {
			canCreateKcp(
				fmt.Sprintf("NfsInstance GCP REGIONAL tier can be created with valid capacity: %d", validCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.REGIONAL).
					WithCapacityGb(validCapacity).
					WithValidDefaults(),
			)
		}

		for _, invalidCapacity := range []int{0, 1023, 1025, 1200, 10241, 20000} {
			canNotCreateKcp(
				fmt.Sprintf("NfsInstance GCP REGIONAL tier cannot be created with invalid capacity: %d", invalidCapacity),
				newTestKcpNfsInstanceGcpBuilder().
					WithTier(cloudcontrolv1beta1.REGIONAL).
					WithCapacityGb(invalidCapacity).
					WithValidDefaults(),
				"REGIONAL tier capacityGb must be between 1024 and 10240, and divisible by 256",
			)
		}

		canChangeKcp(
			"NfsInstance GCP REGIONAL tier capacity can be increased",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.REGIONAL).
				WithCapacityGb(1024).
				WithValidDefaults().
				WithStatusCapacityGb(1024),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(1280)
			},
		)

		canChangeKcp(
			"NfsInstance GCP REGIONAL tier capacity can be reduced",
			newTestKcpNfsInstanceGcpBuilder().
				WithTier(cloudcontrolv1beta1.REGIONAL).
				WithCapacityGb(1280).
				WithValidDefaults().
				WithStatusCapacityGb(1280),
			func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
				b.(*testKcpNfsInstanceGcpBuilder).WithCapacityGb(1024)
			},
		)
	})
})
