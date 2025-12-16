package api_tests

import (
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testGcpRedisInstanceBuilder struct {
	*cloudresourcesv1beta1.GcpRedisInstanceBuilder
}

func newTestGcpRedisInstanceBuilder() *testGcpRedisInstanceBuilder {
	return &testGcpRedisInstanceBuilder{
		GcpRedisInstanceBuilder: cloudresourcesv1beta1.NewGcpRedisInstanceBuilder().
			WithIpRange(uuid.NewString()).
			WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS1).
			WithRedisVersion("REDIS_7_0"),
	}
}

func (b *testGcpRedisInstanceBuilder) Build() *cloudresourcesv1beta1.GcpRedisInstance {
	return &b.GcpRedisInstance
}

func (b *testGcpRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.GcpRedisTier) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testGcpRedisInstanceBuilder) WithRedisVersion(redisVersion string) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithRedisVersion(redisVersion)
	return b
}

func (b *testGcpRedisInstanceBuilder) WithAuthSecretName(name string) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithAuthSecretName(name)
	return b
}

func (b *testGcpRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testGcpRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testGcpRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *testGcpRedisInstanceBuilder {
	b.GcpRedisInstanceBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR GcpRedisInstance", Ordered, func() {

	canChangeSkr(
		"GcpRedisInstance redisTier can be changed if category stays the same (standard->standard)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS2)
		},
	)
	canChangeSkr(
		"GcpRedisInstance redisTier can be changed if category stays the same (premium->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2)
		},
	)

	canNotChangeSkr(
		"GcpRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)
	canNotChangeSkr(
		"GcpRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS2)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)

	allowedRedisTiers := []cloudresourcesv1beta1.GcpRedisTier{
		cloudresourcesv1beta1.GcpRedisTierS1,
		cloudresourcesv1beta1.GcpRedisTierS2,
		cloudresourcesv1beta1.GcpRedisTierS3,
		cloudresourcesv1beta1.GcpRedisTierS4,
		cloudresourcesv1beta1.GcpRedisTierS5,
		cloudresourcesv1beta1.GcpRedisTierS6,
		cloudresourcesv1beta1.GcpRedisTierS7,
		cloudresourcesv1beta1.GcpRedisTierS8,
		cloudresourcesv1beta1.GcpRedisTierP1,
		cloudresourcesv1beta1.GcpRedisTierP2,
		cloudresourcesv1beta1.GcpRedisTierP3,
		cloudresourcesv1beta1.GcpRedisTierP4,
		cloudresourcesv1beta1.GcpRedisTierP5,
		cloudresourcesv1beta1.GcpRedisTierP6,
	}

	for _, testCaseTier := range allowedRedisTiers {
		canCreateSkr(
			fmt.Sprintf("GcpRedisInstance can be created with tier %s", testCaseTier),
			newTestGcpRedisInstanceBuilder().WithRedisTier(testCaseTier),
		)
	}

	canNotCreateSkr(
		"GcpRedisInstance cannot be created with unknown tier",
		newTestGcpRedisInstanceBuilder().WithRedisTier("unknown"),
		"",
	)

	allowedVersionUpgrades := [][]string{
		{"REDIS_6_X", "REDIS_7_0"},
		{"REDIS_6_X", "REDIS_7_2"},
		{"REDIS_7_0", "REDIS_7_2"},
	}
	for _, upgradePair := range allowedVersionUpgrades {
		fromVersion := upgradePair[0]
		toVersion := upgradePair[1]
		canChangeSkr(
			fmt.Sprintf("GcpRedisInstance redisVersion can be upgraded (%s to %s)", fromVersion, toVersion),
			newTestGcpRedisInstanceBuilder().WithRedisVersion(fromVersion),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithRedisVersion(toVersion)
			},
		)
	}

	disallowedVersionUpgrades := [][]string{
		{"REDIS_7_2", "REDIS_7_0"},
		{"REDIS_7_2", "REDIS_6_X"},
		{"REDIS_7_0", "REDIS_6_X"},
	}
	for _, upgradePair := range disallowedVersionUpgrades {
		fromVersion := upgradePair[0]
		toVersion := upgradePair[1]
		canNotChangeSkr(
			fmt.Sprintf("GcpRedisInstance redisVersion can not be downgraded (%s to %s)", fromVersion, toVersion),
			newTestGcpRedisInstanceBuilder().WithRedisVersion(fromVersion),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithRedisVersion(toVersion)
			},
			"redisVersion cannot be downgraded",
		)
	}

	Context("Scenario: authSecret mutability", func() {

		canNotChangeSkr(
			"GcpRedisInstance authSecret.name cannot be changed",
			newTestGcpRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"GcpRedisInstance authSecret.labels can be changed",
			newTestGcpRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"GcpRedisInstance authSecret.annotations can be changed",
			newTestGcpRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"GcpRedisInstance authSecret.extraData can be changed",
			newTestGcpRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"GcpRedisInstance authSecret can be added",
			newTestGcpRedisInstanceBuilder(),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
				b.(*testGcpRedisInstanceBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
