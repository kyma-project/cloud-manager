package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAlicloudRedisInstanceBuilder struct {
	*cloudresourcesv1beta1.AlicloudRedisInstanceBuilder
}

func newTestAlicloudRedisInstanceBuilder() *testAlicloudRedisInstanceBuilder {
	return &testAlicloudRedisInstanceBuilder{
		AlicloudRedisInstanceBuilder: cloudresourcesv1beta1.NewAlicloudRedisInstanceBuilder().
			WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1).
			WithEngineVersion("5.0"),
	}
}

func (b *testAlicloudRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AlicloudRedisInstance {
	return &b.AlicloudRedisInstance
}

func (b *testAlicloudRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisTier) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testAlicloudRedisInstanceBuilder) WithIpRange(ipRangeName string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithIpRange(ipRangeName)
	return b
}

func (b *testAlicloudRedisInstanceBuilder) WithEngineVersion(engineVersion string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithEngineVersion(engineVersion)
	return b
}

func (b *testAlicloudRedisInstanceBuilder) WithAuthSecretName(name string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithAuthSecretName(name)
	return b
}

var _ = Describe("Feature: SKR AlicloudRedisInstance", Ordered, func() {

	Context("Scenario: redisTier enum validation", func() {

		canCreateSkr(
			"AlicloudRedisInstance can be created with S1 tier",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
		)

		canCreateSkr(
			"AlicloudRedisInstance can be created with P5 tier",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierP5),
		)

		canNotCreateSkr(
			"AlicloudRedisInstance cannot be created with invalid redisTier",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier("X1"),
			"",
		)
	})

	Context("Scenario: engineVersion enum validation", func() {

		canNotCreateSkr(
			"AlicloudRedisInstance cannot be created with invalid engineVersion",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("8.0"),
			"",
		)
	})

	Context("Scenario: engineVersion immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 5.0 to 6.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("5.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable",
		)

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 6.0 to 5.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("6.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("5.0")
			},
			"engineVersion is immutable",
		)

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 7.0 to 6.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("7.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable",
		)
	})

	Context("Scenario: IpRange immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisInstance IpRange cannot be changed",
			newTestAlicloudRedisInstanceBuilder().WithIpRange("original-ip-range"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithIpRange("changed-ip-range")
			},
			"IpRange is immutable",
		)
	})

	Context("Scenario: authSecret immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisInstance authSecret.name cannot be changed",
			newTestAlicloudRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)
	})
})
