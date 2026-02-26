package api_tests

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/onsi/ginkgo/v2"
)

var _ = ginkgo.Describe("Feature: KCP Network", func() {

	nb := func() *cloudcontrolv1beta1.NetworkBuilder {
		return &cloudcontrolv1beta1.NetworkBuilder{}

	}
	bb := func(b Builder[*cloudcontrolv1beta1.Network]) *cloudcontrolv1beta1.NetworkBuilder {
		return b.(*cloudcontrolv1beta1.NetworkBuilder)
	}

	var _ Builder[*cloudcontrolv1beta1.Network] = &cloudcontrolv1beta1.NetworkBuilder{}

	canNotCreateKcp(
		"Managed network w/out scope can not be created",
		nb().WithCidr("10.0.0.0/24"),
		"Scope is required",
	)
	canNotCreateKcp(
		"GCP network reference w/out scope can not be created",
		nb().WithGcpRef("proj", "net"),
		"Scope is required",
	)
	canNotCreateKcp(
		"Network can not be managed and reference at the same time",
		nb().WithScope("s").WithCidr("10.0.0.0/24").WithGcpRef("proj", "net"),
		"Too many: 2: must have at most 1 item",
	)

	canNotChangeKcp(
		"Managed network can not change scope",
		nb().WithScope("s").WithCidr("10.11.0.0/25"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithScope("xxx")
		},
		"Scope is immutable",
	)
	canNotChangeKcp(
		"Managed network can not change cidr",
		nb().WithScope("s").WithCidr("10.11.0.0/25"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithCidr("10.22.0.0/24")
		},
		"Network is immutable",
	)
	canNotChangeKcp(
		"GCP network reference can not change scope",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithScope("xxx")
		},
		"Scope is immutable",
	)

	canNotChangeKcp(
		"GCP network reference can not change to AWS",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithGcpRef("", "").WithAwsRef("acc", "us-east1", "vpc123", "name")
		},
		"Network is immutable",
	)
	canNotChangeKcp(
		"Managed network can not change to GCP reference",
		nb().WithScope("s").WithCidr("10.10.0.0/25"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithCidr("").WithGcpRef("proj", "name")
		},
		"Network is immutable",
	)
	canNotChangeKcp(
		"Network reference can not change GCP attribute",
		nb().WithScope("s").WithGcpRef("proj", "net"),
		func(b Builder[*cloudcontrolv1beta1.Network]) {
			bb(b).WithGcpRef("xxx", "yyy")
		},
		"Network is immutable",
	)

})
