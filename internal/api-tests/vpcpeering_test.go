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

	canCreateKcp(
		"VpcPeering with GCP info can be created",
		nb().WithScope("s").WithRemoteRef("ns", "n").
			WithGcpPeering("peering", "project", "vpc", true),
	)

	canCreateKcp(
		"VpcPeering with network details",
		nb().WithScope("s").WithRemoteRef("ns", "n").
			WithDetails("loc", "loc-ns", "rem", "rem-ns", "name", true, false),
	)

	canNotCreateKcp(
		"VpcPeering with both network details and GPC info can not be created",
		nb().WithScope("s").WithRemoteRef("ns", "n").
			WithGcpPeering("peering", "project", "vpc", true).
			WithDetails("loc", "loc-ns", "rem", "rem-ns", "name", true, false),
		"Only one of details or vpcPeering can be specified",
	)

	canNotChangeKcp(
		"VpcPeering GCP info can not change",
		nb().WithScope("s").WithRemoteRef("ns", "n").WithGcpPeering("peering", "project", "vpc", true),
		func(b Builder[*cloudcontrolv1beta1.VpcPeering]) {
			bb(b).WithGcpPeering("peering2", "project2", "vpc2", false)
		},
		"Peering info is immutable",
	)

	canNotChangeKcp(
		"VpcPeering network reference can not change",
		nb().WithScope("s").WithRemoteRef("ns", "n").
			WithDetails("loc", "loc-ns", "rem", "rem-ns", "name", true, false),
		func(b Builder[*cloudcontrolv1beta1.VpcPeering]) {
			bb(b).
				WithDetails("loc2", "loc-ns2", "rem2", "rem-ns2", "name2", false, false)
		},
		"Peering details are immutable",
	)
})
