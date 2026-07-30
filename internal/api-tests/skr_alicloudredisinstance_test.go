package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAlicloudRedisInstanceBuilder struct {
	*cloudresourcesv1beta1.AlicloudRedisInstanceBuilder
}

func newTestAlicloudRedisInstanceBuilder() *testAlicloudRedisInstanceBuilder {
	return &testAlicloudRedisInstanceBuilder{
		AlicloudRedisInstanceBuilder: cloudresourcesv1beta1.NewAlicloudRedisInstanceBuilder().
			WithIpRange(uuid.NewString()).
			WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1).
			WithEngineVersion("7.0"),
	}
}

func (b *testAlicloudRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AlicloudRedisInstance {
	return &b.AlicloudRedisInstance
}

func (b *testAlicloudRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisTier) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithRedisTier(redisTier)
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

func (b *testAlicloudRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testAlicloudRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testAlicloudRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstanceBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR AlicloudRedisInstance", Ordered, func() {

	Context("Scenario: redisTier mutability", func() {

		canChangeSkr(
			"AlicloudRedisInstance redisTier can be changed within S tiers",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS2)
			},
		)

		canChangeSkr(
			"AlicloudRedisInstance redisTier can be changed within P tiers",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierP1),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierP2)
			},
		)

		canChangeSkr(
			"AlicloudRedisInstance redisTier can be changed from S to P (S↔P mutable per design decision 4)",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS1),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierP1)
			},
		)

		canChangeSkr(
			"AlicloudRedisInstance redisTier can be changed from P to S",
			newTestAlicloudRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierP3),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AlicloudRedisTierS3)
			},
		)
	})

	Context("Scenario: engineVersion validation", func() {

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 5.0 to 6.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("5.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable.",
		)

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 7.0 to 6.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("7.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable.",
		)

		canNotChangeSkr(
			"AlicloudRedisInstance engineVersion cannot be changed from 6.0 to 5.0",
			newTestAlicloudRedisInstanceBuilder().WithEngineVersion("6.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithEngineVersion("5.0")
			},
			"engineVersion is immutable.",
		)
	})

	Context("Scenario: authSecret mutability", func() {

		canNotChangeSkr(
			"AlicloudRedisInstance authSecret.name cannot be changed",
			newTestAlicloudRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AlicloudRedisInstance authSecret.labels can be changed",
			newTestAlicloudRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithAuthSecretLabels(map[string]string{"env": "prod"})
			},
		)

		canChangeSkr(
			"AlicloudRedisInstance authSecret.annotations can be changed",
			newTestAlicloudRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b"})
			},
		)

		canChangeSkr(
			"AlicloudRedisInstance authSecret.extraData can be changed",
			newTestAlicloudRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{"key": "v1"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisInstance]) {
				b.(*testAlicloudRedisInstanceBuilder).WithAuthSecretExtraData(map[string]string{"key": "v2"})
			},
		)
	})
})
