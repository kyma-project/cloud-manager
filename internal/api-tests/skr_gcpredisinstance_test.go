package api_tests

import (
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

type testGcpRedisInstanceBuilder struct {
	instance cloudresourcesv1beta1.GcpRedisInstance
}

func newTestGcpRedisInstanceBuilder() *testGcpRedisInstanceBuilder {
	return &testGcpRedisInstanceBuilder{
		instance: cloudresourcesv1beta1.GcpRedisInstance{
			Spec: cloudresourcesv1beta1.GcpRedisInstanceSpec{
				IpRange: cloudresourcesv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				RedisTier:    "S1",
				RedisVersion: "REDIS_7_0",
				AuthEnabled:  true,
				RedisConfigs: map[string]string{
					"maxmemory-policy": "allkeys-lru",
				},
				MaintenancePolicy: &cloudresourcesv1beta1.MaintenancePolicy{
					DayOfWeek: &cloudresourcesv1beta1.DayOfWeekPolicy{
						Day: "MONDAY",
						StartTime: cloudresourcesv1beta1.TimeOfDay{
							Hours:   11,
							Minutes: 0,
						},
					},
				},
			},
		},
	}
}

func (b *testGcpRedisInstanceBuilder) Build() *cloudresourcesv1beta1.GcpRedisInstance {
	return &b.instance
}

func (b *testGcpRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.GcpRedisTier) *testGcpRedisInstanceBuilder {
	b.instance.Spec.RedisTier = redisTier
	return b
}

func (b *testGcpRedisInstanceBuilder) WithRedisVersion(redisVersion string) *testGcpRedisInstanceBuilder {
	b.instance.Spec.RedisVersion = redisVersion
	return b
}

var _ = Describe("Feature: SKR GcpRedisInstance", Ordered, func() {

	It("Given SKR default namespace exists", func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

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
})
