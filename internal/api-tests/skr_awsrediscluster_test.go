package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAwsRedisClusterBuilder struct {
	instance cloudresourcesv1beta1.AwsRedisCluster
}

func newTestAwsRedisClusterBuilder() *testAwsRedisClusterBuilder {
	return &testAwsRedisClusterBuilder{
		instance: cloudresourcesv1beta1.AwsRedisCluster{
			Spec: cloudresourcesv1beta1.AwsRedisClusterSpec{
				IpRange: cloudresourcesv1beta1.IpRangeRef{
					Name: uuid.NewString(),
				},
				RedisTier:        cloudresourcesv1beta1.AwsRedisTierC1,
				EngineVersion:    "7.0",
				AuthEnabled:      true,
				ShardCount:       2,
				ReplicasPerShard: 1,
			},
		},
	}
}

func (b *testAwsRedisClusterBuilder) Build() *cloudresourcesv1beta1.AwsRedisCluster {
	return &b.instance
}

func (b *testAwsRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AwsRedisClusterTier) *testAwsRedisClusterBuilder {
	b.instance.Spec.RedisTier = redisTier
	return b
}

func (b *testAwsRedisClusterBuilder) WithEngineVersion(engineVersion string) *testAwsRedisClusterBuilder {
	b.instance.Spec.EngineVersion = engineVersion
	return b
}

func (b *testAwsRedisClusterBuilder) WithShardCount(shardCount int32) *testAwsRedisClusterBuilder {
	b.instance.Spec.ShardCount = shardCount
	return b
}

func (b *testAwsRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testAwsRedisClusterBuilder {
	b.instance.Spec.ReplicasPerShard = replicasPerShard
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretName(name string) *testAwsRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testAwsRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAwsRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAwsRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAwsRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AwsRedisCluster", Ordered, func() {

	Context("Scenario: authSecret mutability", func() {

		canChangeSkr(
			"AwsRedisCluster authSecret.name can be changed",
			newTestAwsRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AwsRedisCluster]) {
				b.(*testAwsRedisClusterBuilder).WithAuthSecretName("new-name")
			},
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
