package api_tests

import (
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAwsRedisInstanceBuilder struct {
	*cloudresourcesv1beta1.AwsRedisInstanceBuilder
}

func newTestAwsRedisInstanceBuilder() *testAwsRedisInstanceBuilder {
	return &testAwsRedisInstanceBuilder{
		AwsRedisInstanceBuilder: cloudresourcesv1beta1.NewAwsRedisInstanceBuilder().
			WithIpRange(uuid.NewString()),
	}
}

func (b *testAwsRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AwsRedisInstance {
	return &b.AwsRedisInstance
}

func (b *testAwsRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AwsRedisTier) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testAwsRedisInstanceBuilder) WithEngineVersion(engineVersion string) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithEngineVersion(engineVersion)
	return b
}

func (b *testAwsRedisInstanceBuilder) WithAuthSecretName(name string) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithAuthSecretName(name)
	return b
}

func (b *testAwsRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testAwsRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testAwsRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAwsRedisInstanceBuilder {
	b.AwsRedisInstanceBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR AwsRedisInstance", Ordered, func() {

	canChangeSkr(
		"AwsRedisInstance redisTier can be changed if category stays the same (standard->standard)",
		newTestAwsRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AwsRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
			b.(*testAwsRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AwsRedisTierS2)
		},
	)
	canChangeSkr(
		"AwsRedisInstance redisTier can be changed if category stays the same (premium->premium)",
		newTestAwsRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AwsRedisTierP1),
		func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
			b.(*testAwsRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AwsRedisTierP2)
		},
	)

	canNotChangeSkr(
		"AwsRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestAwsRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AwsRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
			b.(*testAwsRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AwsRedisTierP2)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)
	canNotChangeSkr(
		"AwsRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestAwsRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AwsRedisTierP1),
		func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
			b.(*testAwsRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AwsRedisTierS2)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)

	canNotCreateSkr(
		"AwsRedisInstance cannot be created with unknown tier",
		newTestAwsRedisInstanceBuilder().WithRedisTier("unknown"),
		"",
	)

	allowedVersionUpgrades := [][]string{
		{"6.x", "7.0"},
		{"6.x", "7.1"},
		{"7.0", "7.1"},
	}
	for _, upgradePair := range allowedVersionUpgrades {
		fromVersion := upgradePair[0]
		toVersion := upgradePair[1]
		canChangeSkr(
			fmt.Sprintf("AwsRedisInstance engineVersion can be upgraded (%s to %s)", fromVersion, toVersion),
			newTestAwsRedisInstanceBuilder().WithEngineVersion(fromVersion),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithEngineVersion(toVersion)
			},
		)
	}

	disallowedVersionUpgrades := [][]string{
		{"7.1", "7.0"},
		{"7.1", "6.x"},
		{"7.0", "6.x"},
	}
	for _, upgradePair := range disallowedVersionUpgrades {
		fromVersion := upgradePair[0]
		toVersion := upgradePair[1]
		canNotChangeSkr(
			fmt.Sprintf("AwsRedisInstance engineVersion can not be downgraded (%s to %s)", fromVersion, toVersion),
			newTestAwsRedisInstanceBuilder().WithEngineVersion(fromVersion),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithEngineVersion(toVersion)
			},
			"engineVersion cannot be downgraded",
		)
	}

	Context("Scenario: authSecret mutability", func() {

		canNotChangeSkr(
			"AwsRedisInstance authSecret.name cannot be changed",
			newTestAwsRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AwsRedisInstance authSecret.labels can be changed",
			newTestAwsRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AwsRedisInstance authSecret.annotations can be changed",
			newTestAwsRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"AwsRedisInstance authSecret.extraData can be changed",
			newTestAwsRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AwsRedisInstance authSecret can be added",
			newTestAwsRedisInstanceBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisInstance]) {
				b.(*testAwsRedisInstanceBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
