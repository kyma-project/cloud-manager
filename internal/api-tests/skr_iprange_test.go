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

	// Test reserved / invalid CIDR ranges are rejected at admission
	canNotCreateSkr(
		"IpRange cannot be created with default route 0.0.0.0/0",
		newTestSkrIpRangeBuilder().WithCidr("0.0.0.0/0"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 loopback 127.0.0.0/7",
		newTestSkrIpRangeBuilder().WithCidr("127.0.0.0/7"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 loopback 127.0.0.0/8",
		newTestSkrIpRangeBuilder().WithCidr("127.0.0.0/8"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 loopback 127.0.0.0/9",
		newTestSkrIpRangeBuilder().WithCidr("127.0.0.0/9"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 link-local 169.254.0.0/15",
		newTestSkrIpRangeBuilder().WithCidr("169.254.0.0/15"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 link-local 169.254.0.0/16",
		newTestSkrIpRangeBuilder().WithCidr("169.254.0.0/16"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv4 link-local 169.254.0.0/17",
		newTestSkrIpRangeBuilder().WithCidr("169.254.0.0/16"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv6 default route ::/0",
		newTestSkrIpRangeBuilder().WithCidr("::/0"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv6 loopback ::1/128",
		newTestSkrIpRangeBuilder().WithCidr("::1/128"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)

	canNotCreateSkr(
		"IpRange cannot be created with IPv6 link-local fe80::/10",
		newTestSkrIpRangeBuilder().WithCidr("fe80::/10"),
		"Must be a valid IPv4 CIDR and cannot overlap",
	)
})
