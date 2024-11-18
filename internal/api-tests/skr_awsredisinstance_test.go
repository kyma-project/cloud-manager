package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

type testAwsRedisInstanceBuilder struct {
	instance cloudresourcesv1beta1.AwsRedisInstance
}

func newTestAwsRedisInstanceBuilder() *testAwsRedisInstanceBuilder {
	return &testAwsRedisInstanceBuilder{
		instance: cloudresourcesv1beta1.AwsRedisInstance{
			Spec: cloudresourcesv1beta1.AwsRedisInstanceSpec{
				IpRange: cloudresourcesv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				RedisTier:     "S1",
				EngineVersion: "7.0",
				AuthEnabled:   true,
				Parameters: map[string]string{
					"maxmemory-policy": "allkeys-lru",
				},
			},
		},
	}
}

func (b *testAwsRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AwsRedisInstance {
	return &b.instance
}

func (b *testAwsRedisInstanceBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AwsRedisTier) *testAwsRedisInstanceBuilder {
	b.instance.Spec.RedisTier = redisTier
	return b
}

var _ = Describe("Feature: SKR AwsRedisInstance", Ordered, func() {

	It("Given SKR default namespace exists", func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

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
})
