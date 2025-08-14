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

	canCreateKcp(
		"Subscription with SecretBindingName can be created",
		b().WithSecretBindingName("my-secret-binding-name"),
	)

	canNotCreateKcp(
		"Subscription with empty SecretBindingName can not be created",
		b(),
		"SecretBindingName is required.",
	)

	canNotChangeKcp(
		"Subscription SecretBindingName can not be changed",
		b().WithSecretBindingName("name1"),
		func(b Builder[*cloudcontrolv1beta1.Subscription]) {
			bb(b).WithSecretBindingName("name2")
		},
		"SecretBindingName is immutable.",
	)
})
