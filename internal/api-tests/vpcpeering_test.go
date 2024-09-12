package api_tests

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Feature: KCP VpcPeering", func() {

	nb := func() *cloudcontrolv1beta1.VpcPeeringBuilder {
		return &cloudcontrolv1beta1.VpcPeeringBuilder{}
	}
	bb := func(b Builder[*cloudcontrolv1beta1.VpcPeering]) *cloudcontrolv1beta1.VpcPeeringBuilder {
		return b.(*cloudcontrolv1beta1.VpcPeeringBuilder)
	}

	var _ Builder[*cloudcontrolv1beta1.VpcPeering] = &cloudcontrolv1beta1.VpcPeeringBuilder{}

	canCreate(
		"VpcPeering with GCP info can be created",
		nb().WithScope("s").WithRemoteRef("ns", "n").WithGcpPeering("peering", "project", "vpc", true),
	)

	canNotCreate(
		"VpcPeering with both Networks and GPC info can not be created",
		nb().WithScope("s").WithRemoteRef("ns", "n").
			WithGcpPeering("peering", "project", "vpc", true).
			WithNetworks("loc", "loc-ns", "rem", "rem-ns"),
		"Only one of networks or vpcPeering can be specified",
	)

	canNotChange(
		"VpcPeering GCP info can not change",
		nb().WithScope("s").WithRemoteRef("ns", "n").WithGcpPeering("peering", "project", "vpc", true),
		func(b Builder[*cloudcontrolv1beta1.VpcPeering]) {
			bb(b).WithGcpPeering("peering2", "project2", "vpc2", false)
		},
		"Peering info is immutable",
	)
})
