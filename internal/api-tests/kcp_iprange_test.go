package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testKcpIpRangeBuilder struct {
	instance cloudcontrolv1beta1.IpRange
}

func newTestKcpIpRangeBuilder() *testKcpIpRangeBuilder {
	return &testKcpIpRangeBuilder{
		instance: cloudcontrolv1beta1.IpRange{
			Spec: cloudcontrolv1beta1.IpRangeSpec{
				RemoteRef: cloudcontrolv1beta1.RemoteRef{
					Name:      uuid.NewString(),
					Namespace: "default",
				},
				Scope: cloudcontrolv1beta1.ScopeRef{
					Name: uuid.NewString(),
				},
			},
		},
	}
}

func (b *testKcpIpRangeBuilder) Build() *cloudcontrolv1beta1.IpRange {
	return &b.instance
}

func (b *testKcpIpRangeBuilder) WithCidr(cidr string) *testKcpIpRangeBuilder {
	b.instance.Spec.Cidr = cidr
	return b
}

var _ = Describe("Feature: KCP IpRange", Ordered, func() {

	// Test CIDR is optional
	canCreateKcp(
		"IpRange can be created without CIDR",
		newTestKcpIpRangeBuilder(),
	)

	canCreateKcp(
		"IpRange can be created with CIDR",
		newTestKcpIpRangeBuilder().WithCidr("10.0.0.0/16"),
	)

	// Test CIDR immutability
	canNotChangeKcp(
		"IpRange CIDR cannot be changed once created with CIDR",
		newTestKcpIpRangeBuilder().WithCidr("10.0.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.IpRange]) {
			b.(*testKcpIpRangeBuilder).WithCidr("10.1.0.0/16")
		},
		"CIDR is immutable",
	)

	canNotChangeKcp(
		"IpRange CIDR cannot be set after creation without CIDR",
		newTestKcpIpRangeBuilder(),
		func(b Builder[*cloudcontrolv1beta1.IpRange]) {
			b.(*testKcpIpRangeBuilder).WithCidr("10.0.0.0/16")
		},
		"CIDR is immutable",
	)

	canNotChangeKcp(
		"IpRange CIDR cannot be unset after creation with CIDR",
		newTestKcpIpRangeBuilder().WithCidr("10.0.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.IpRange]) {
			b.(*testKcpIpRangeBuilder).WithCidr("")
		},
		"CIDR is immutable",
	)
})
