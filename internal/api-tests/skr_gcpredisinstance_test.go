package api_tests

import (
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
				ReplicaCount: 0,
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

func (b *testGcpRedisInstanceBuilder) WithReplicaCount(replicaCount int32) *testGcpRedisInstanceBuilder {
	b.instance.Spec.ReplicaCount = replicaCount
	return b
}

var _ = Describe("Feature: SKR GcpRedisInstance", Ordered, func() {

	It("Given SKR default namespace exists", func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

	canCreateSkr(
		"GcpRedisInstance can be created with no replicas in standard category",
		newTestGcpRedisInstanceBuilder(),
	)
	canNotCreateSkr(
		"GcpRedisInstance cannot be created with replicas in standard category",
		newTestGcpRedisInstanceBuilder().WithReplicaCount(1),
		"replicaCount must be zero for Standard service tier",
	)

	canCreateSkr(
		"GcpRedisInstance can be created with replicas in premium category",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2).WithReplicaCount(1),
	)
	canNotCreateSkr(
		"GcpRedisInstance cannot be created without replicas in premium category",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2).WithReplicaCount(0),
		"replicaCount must be defined with value between 1 and 5 for Premium service tier",
	)

	canChangeSkr(
		"GcpRedisInstance redisTier can be changed if category stays the same (standard->standard)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS2)
		},
	)
	canChangeSkr(
		"GcpRedisInstance redisTier can be changed if category stays the same (premium->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP1).WithReplicaCount(int32(2)),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2)
		},
	)

	canNotChangeSkr(
		"GcpRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP2).WithReplicaCount(1)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)
	canNotChangeSkr(
		"GcpRedisInstance redisTier can not be changed if category changes (standard->premium)",
		newTestGcpRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.GcpRedisTierP1).WithReplicaCount(1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisInstance]) {
			b.(*testGcpRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.GcpRedisTierS2).WithReplicaCount(0)
		},
		"Service tier cannot be changed within redisTier. Only capacity tier can be changed.",
	)
})
