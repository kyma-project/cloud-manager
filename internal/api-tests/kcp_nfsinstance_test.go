package api_tests

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Feature: KCP NfsInstance", func() {

	nb := func() *cloudcontrolv1beta1.NfsInstanceBuilder {
		return cloudcontrolv1beta1.NewNfsInstanceBuilder()
	}
	bb := func(b Builder[*cloudcontrolv1beta1.NfsInstance]) *cloudcontrolv1beta1.NfsInstanceBuilder {
		return b.(*cloudcontrolv1beta1.NfsInstanceBuilder)
	}

	// CREATE =============================================================================

	// All Specified Can Create ======================================

	canCreateKcp(
		"Can create AWS NfsInstance",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
	)

	canCreateKcp(
		"Can create GCP NfsInstance",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
	)

	canCreateKcp(
		"Can create OpenStack NfsInstance",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
	)

	// w/out scope can not create =================================

	canNotCreateKcp(
		"Can not create AWS NfsInstance without Scope",
		nb().
			WithIpRange("iprange").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
		"Scope is required",
	)

	canNotCreateKcp(
		"Can not create GCP NfsInstance without Scope",
		nb().
			WithIpRange("iprange").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
		"Scope is required",
	)

	canNotCreateKcp(
		"Can not create OpenStack NfsInstance without Scope",
		nb().
			WithIpRange("iprange").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
		"Scope is required",
	)

	// w/out remoteRef can not create  ======================================

	canNotCreateKcp(
		"Can not create AWS NfsInstance without RemoteRef",
		nb().
			WithScope("scope").
			WithIpRange("iprange").
			WithAwsDummyDefaults(),
		"RemoteRef is required",
	)

	canNotCreateKcp(
		"Can not create GCP NfsInstance without RemoteRef",
		nb().
			WithScope("scope").
			WithIpRange("iprange").
			WithGcpDummyDefaults(),
		"RemoteRef is required",
	)

	canNotCreateKcp(
		"Can not create OpenStack NfsInstance without RemoteRef",
		nb().
			WithScope("scope").
			WithIpRange("iprange").
			WithOpenStackDummyDefaults(),
		"RemoteRef is required",
	)

	// w/out iprange can not create ======================================

	canNotCreateKcp(
		"Can not create AWS NfsInstance without IpRange",
		nb().
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
		"IpRange is required",
	)

	canNotCreateKcp(
		"Can not create GCP NfsInstance without IpRange",
		nb().
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
		"IpRange is required",
	)

	canNotCreateKcp(
		"Can not create OpenStack NfsInstance without IpRange",
		nb().
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
		"IpRange is required",
	)

	// UPDATE =============================================================================

	// can not change iprange ======================================

	canNotChangeKcp(
		"Can not change AWS NfsInstance IpRange",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithIpRange("other")
		},
		"IpRange is immutable",
	)

	canNotChangeKcp(
		"Can not change GCP NfsInstance IpRange",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithIpRange("other")
		},
		"IpRange is immutable",
	)

	canNotChangeKcp(
		"Can not change OpenStack NfsInstance IpRange",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithIpRange("other")
		},
		"IpRange is immutable",
	)

	// can not change scope ======================================

	canNotChangeKcp(
		"Can not change AWS NfsInstance Scope",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithScope("other")
		},
		"Scope is immutable",
	)

	canNotChangeKcp(
		"Can not change GCP NfsInstance Scope",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithScope("other")
		},
		"Scope is immutable",
	)

	canNotChangeKcp(
		"Can not change OpenStack NfsInstance Scope",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithScope("other")
		},
		"Scope is immutable",
	)

	// can not change remoteRef ======================================

	canNotChangeKcp(
		"Can not change AWS NfsInstance RemoteRef",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithAwsDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithRemoteRef("other", "other")
		},
		"RemoteRef is immutable",
	)

	canNotChangeKcp(
		"Can not change GCP NfsInstance RemoteRef",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithGcpDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithRemoteRef("other", "other")
		},
		"RemoteRef is immutable",
	)

	canNotChangeKcp(
		"Can not change OpenStack NfsInstance RemoteRef",
		nb().
			WithIpRange("iprange").
			WithScope("scope").
			WithRemoteRef("ns", "n").
			WithOpenStackDummyDefaults(),
		func(b Builder[*cloudcontrolv1beta1.NfsInstance]) {
			bb(b).WithRemoteRef("other", "other")
		},
		"RemoteRef is immutable",
	)

})
