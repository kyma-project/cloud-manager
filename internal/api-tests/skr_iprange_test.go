package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testSkrIpRangeBuilder struct {
	instance cloudresourcesv1beta1.IpRange
}

func newTestSkrIpRangeBuilder() *testSkrIpRangeBuilder {
	return &testSkrIpRangeBuilder{
		instance: cloudresourcesv1beta1.IpRange{
			Spec: cloudresourcesv1beta1.IpRangeSpec{},
		},
	}
}

func (b *testSkrIpRangeBuilder) Build() *cloudresourcesv1beta1.IpRange {
	return &b.instance
}

func (b *testSkrIpRangeBuilder) WithCidr(cidr string) *testSkrIpRangeBuilder {
	b.instance.Spec.Cidr = cidr
	return b
}

var _ = Describe("Feature: SKR IpRange", Ordered, func() {

	// Test CIDR is optional
	canCreateSkr(
		"IpRange can be created without CIDR",
		newTestSkrIpRangeBuilder(),
	)

	canCreateSkr(
		"IpRange can be created with CIDR",
		newTestSkrIpRangeBuilder().WithCidr("10.0.0.0/16"),
	)

	// Test CIDR immutability
	canNotChangeSkr(
		"IpRange CIDR cannot be changed once created with CIDR",
		newTestSkrIpRangeBuilder().WithCidr("10.0.0.0/16"),
		func(b Builder[*cloudresourcesv1beta1.IpRange]) {
			b.(*testSkrIpRangeBuilder).WithCidr("10.1.0.0/16")
		},
		"CIDR is immutable",
	)

	canNotChangeSkr(
		"IpRange CIDR cannot be set after creation without CIDR",
		newTestSkrIpRangeBuilder(),
		func(b Builder[*cloudresourcesv1beta1.IpRange]) {
			b.(*testSkrIpRangeBuilder).WithCidr("10.0.0.0/16")
		},
		"CIDR is immutable",
	)

	canNotChangeSkr(
		"IpRange CIDR cannot be unset after creation with CIDR",
		newTestSkrIpRangeBuilder().WithCidr("10.0.0.0/16"),
		func(b Builder[*cloudresourcesv1beta1.IpRange]) {
			b.(*testSkrIpRangeBuilder).WithCidr("")
		},
		"CIDR is immutable",
	)
})
