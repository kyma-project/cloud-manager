package api_tests

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/onsi/ginkgo/v2"
)

var _ = ginkgo.Describe("Feature: KCP Subscription", func() {

	b := func() *cloudcontrolv1beta1.SubscriptionBuilder {
		return cloudcontrolv1beta1.NewSubscriptionBuilder()
	}
	bb := func(b Builder[*cloudcontrolv1beta1.Subscription]) *cloudcontrolv1beta1.SubscriptionBuilder {
		return b.(*cloudcontrolv1beta1.SubscriptionBuilder)
	}

	var _ Builder[*cloudcontrolv1beta1.Subscription] = &cloudcontrolv1beta1.SubscriptionBuilder{}

	canNotCreateKcp(
		"Empty Subscription can not be created",
		b(),
		"Exactly one of garden, aws, azure, gcp or openstack must be specified",
	)

	canCreateKcp(
		"Subscription with BindingName can be created",
		b().WithBindingName("my-secret-binding-name"),
	)

	canCreateKcp(
		"Subscription with AWS details can be created",
		b().WithAws("12345678"),
	)

	canCreateKcp(
		"Subscription with Azure details can be created",
		b().WithAzure("tenant", "subscription"),
	)

	canCreateKcp(
		"Subscription with GCP details can be created",
		b().WithGcp("gcp-project-id"),
	)

	canCreateKcp(
		"Subscription with Openstack details can be created",
		b().WithOpenstack("domain", "project"),
	)

	// can not change ==============================

	canNotChangeKcp(
		"Subscription BindingName can not be changed",
		b().WithBindingName("name1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithBindingName("name2")
		},
		"Subscription spec is immutable",
	)

	canNotChangeKcp(
		"Subscription AWS details can not be changed",
		b().WithAws("name1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithAws("name2")
		},
		"Subscription spec is immutable",
	)

	canNotChangeKcp(
		"Subscription Azure details can not be changed",
		b().WithAzure("tenant1", "sub1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithAzure("tenant2", "sub2")
		},
		"Subscription spec is immutable",
	)

	canNotChangeKcp(
		"Subscription GCP details can not be changed",
		b().WithGcp("name1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithGcp("name2")
		},
		"Subscription spec is immutable",
	)

	canNotChangeKcp(
		"Subscription Openstack details can not be changed",
		b().WithOpenstack("domain1", "project1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithOpenstack("domain2", "project2")
		},
		"Subscription spec is immutable",
	)

})
