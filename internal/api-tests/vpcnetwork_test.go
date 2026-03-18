package api_tests

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/onsi/ginkgo/v2"
)

var _ = ginkgo.Describe("Feature: KCP VpcNetwork", func() {

	b := func() *cloudcontrolv1beta1.VpcNetworkBuilder {
		return cloudcontrolv1beta1.NewVpcNetworkBuilder()
	}
	bb := func(b Builder[*cloudcontrolv1beta1.VpcNetwork]) *cloudcontrolv1beta1.VpcNetworkBuilder {
		return b.(*cloudcontrolv1beta1.VpcNetworkBuilder)
	}

	_ = bb

	var _ Builder[*cloudcontrolv1beta1.VpcNetwork] = &cloudcontrolv1beta1.VpcNetworkBuilder{}

	// canCreateKcp -------------------------------------------------------

	canCreateKcp(
		"VpcNetwork can be created with subscription, region, and cidrBlocks",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
	)

	canCreateKcp(
		"VpcNetwork can be created with kyma type, subscription, region, and cidrBlocks",
		b().
			WithType(cloudcontrolv1beta1.VpcNetworkTypeKyma).
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
	)

	canCreateKcp(
		"VpcNetwork can be created with garden type, subscription, region, and cidrBlocks",
		b().
			WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener).
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
	)

	// canNotCreateKcp -------------------------------------------------------

	canNotCreateKcp(
		"VpcNetwork can not be created without subscription",
		b().
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
		"The subscription cannot be empty",
	)

	canNotCreateKcp(
		"VpcNetwork can not be created without region",
		b().
			WithSubscription("my-subscription").
			WithCidrBlocks("10.250.0.0/16"),
		"The region cannot be empty",
	)

	canNotCreateKcp(
		"VpcNetwork can not be created with null cidrBlocks",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region"),
		`Invalid value: null`,
	)

	canNotCreateKcp(
		"VpcNetwork can not be created with empty array of cidrBlocks",
		func() Builder[*cloudcontrolv1beta1.VpcNetwork] {
			res := b().
				WithSubscription("my-subscription").
				WithRegion("some-region")
			res.Build().Spec.CidrBlocks = []string{}
			return res
		}(),
		`spec.cidrBlocks in body should have at least 1 items`,
	)

	canNotCreateKcp(
		"VpcNetwork can not be created with invalid cidr",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("invalid-cidr"),
		`cidrBlocks must be a list of valid CIDRs`,
	)

	// canNotChangeKcp -------------------------------------------------------

	canNotChangeKcp(
		"VpcNetwork type is not editable",
		b().
			WithType(cloudcontrolv1beta1.VpcNetworkTypeKyma).
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.VpcNetwork]) {
			bb(b).WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener)
		},
		"The type is immutable",
	)

	canNotChangeKcp(
		"VpcNetwork subscription is not editable",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.VpcNetwork]) {
			bb(b).WithSubscription("other")
		},
		"The subscription is immutable",
	)

	canNotChangeKcp(
		"VpcNetwork region is not editable",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.VpcNetwork]) {
			bb(b).WithRegion("other")
		},
		"The region is immutable",
	)

	// canChangeKcp -------------------------------------------------------

	canChangeKcp(
		"VpcNetwork cidrBlocks are editable",
		b().
			WithSubscription("my-subscription").
			WithRegion("some-region").
			WithCidrBlocks("10.250.0.0/16"),
		func(b Builder[*cloudcontrolv1beta1.VpcNetwork]) {
			bb(b).WithCidrBlocks("10.250.0.0/16", "10.251.0.0/16")
		},
	)

})
