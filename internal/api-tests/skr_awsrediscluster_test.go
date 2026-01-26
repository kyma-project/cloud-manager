package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAwsRedisClusterBuilder struct {
	*cloudresourcesv1beta1.AwsRedisClusterBuilder
}

func newTestAwsRedisClusterBuilder() *testAwsRedisClusterBuilder {
	return &testAwsRedisClusterBuilder{
		AwsRedisClusterBuilder: cloudresourcesv1beta1.NewAwsRedisClusterBuilder().
			WithIpRange(uuid.NewString()).
			WithRedisTier(cloudresourcesv1beta1.AwsRedisTierC1).
			WithShardCount(2).
			WithEngineVersion("7.0"),
	}
}

func (b *testAwsRedisClusterBuilder) Build() *cloudresourcesv1beta1.AwsRedisCluster {
	return &b.AwsRedisCluster
}

func (b *testAwsRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AwsRedisClusterTier) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testAwsRedisClusterBuilder) WithEngineVersion(engineVersion string) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithEngineVersion(engineVersion)
	return b
}

func (b *testAwsRedisClusterBuilder) WithShardCount(shardCount int32) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithShardCount(shardCount)
	return b
}

func (b *testAwsRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithReplicasPerShard(replicasPerShard)
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretName(name string) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithAuthSecretName(name)
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAwsRedisClusterBuilder {
	b.AwsRedisClusterBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR AwsRedisCluster", Ordered, func() {

	Context("Scenario: authSecret mutability", func() {

		canNotChangeSkr(
			"AwsRedisCluster authSecret.name cannot be changed",
			newTestAwsRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AwsRedisCluster authSecret.labels can be changed",
			newTestAwsRedisClusterBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AwsRedisCluster authSecret.annotations can be changed",
			newTestAwsRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"AwsRedisCluster authSecret.extraData can be changed",
			newTestAwsRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AwsRedisCluster authSecret can be added",
			newTestAwsRedisClusterBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
