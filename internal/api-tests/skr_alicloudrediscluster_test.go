package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAlicloudRedisClusterBuilder struct {
	*cloudresourcesv1beta1.AlicloudRedisClusterBuilder
}

func newTestAlicloudRedisClusterBuilder() *testAlicloudRedisClusterBuilder {
	return &testAlicloudRedisClusterBuilder{
		AlicloudRedisClusterBuilder: cloudresourcesv1beta1.NewAlicloudRedisClusterBuilder().
			WithIpRange(uuid.NewString()).
			WithRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3).
			WithShardCount(2).
			WithEngineVersion("7.0"),
	}
}

func (b *testAlicloudRedisClusterBuilder) Build() *cloudresourcesv1beta1.AlicloudRedisCluster {
	return &b.AlicloudRedisCluster
}

func (b *testAlicloudRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisClusterTier) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithShardCount(shardCount int32) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithShardCount(shardCount)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithReplicasPerShard(replicasPerShard)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithEngineVersion(engineVersion string) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithEngineVersion(engineVersion)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithAuthSecretName(name string) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithAuthSecretName(name)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAlicloudRedisClusterBuilder {
	b.AlicloudRedisClusterBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR AlicloudRedisCluster", Ordered, func() {

	Context("Scenario: redisTier mutability", func() {

		canChangeSkr(
			"AlicloudRedisCluster redisTier can be changed",
			newTestAlicloudRedisClusterBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC4)
			},
		)
	})

	Context("Scenario: shardCount mutability", func() {

		canChangeSkr(
			"AlicloudRedisCluster shardCount can be increased",
			newTestAlicloudRedisClusterBuilder().WithShardCount(2),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithShardCount(4)
			},
		)

		canChangeSkr(
			"AlicloudRedisCluster shardCount can be decreased",
			newTestAlicloudRedisClusterBuilder().WithShardCount(4),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithShardCount(2)
			},
		)
	})

	Context("Scenario: replicasPerShard validation", func() {

		canNotCreateSkr(
			"AlicloudRedisCluster cannot be created with replicasPerShard=1",
			newTestAlicloudRedisClusterBuilder().WithReplicasPerShard(1),
			"replicasPerShard must be 0",
		)

		canNotChangeSkr(
			"AlicloudRedisCluster replicasPerShard cannot be changed to 1",
			newTestAlicloudRedisClusterBuilder().WithReplicasPerShard(0),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithReplicasPerShard(1)
			},
			"replicasPerShard must be 0",
		)
	})

	Context("Scenario: engineVersion validation", func() {

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 5.0 to 7.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("5.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("7.0")
			},
			"engineVersion is immutable.",
		)

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 7.0 to 6.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("7.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable.",
		)

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 6.0 to 5.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("6.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("5.0")
			},
			"engineVersion is immutable.",
		)
	})

	Context("Scenario: authSecret mutability", func() {

		canNotChangeSkr(
			"AlicloudRedisCluster authSecret.name cannot be changed",
			newTestAlicloudRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AlicloudRedisCluster authSecret.labels can be changed",
			newTestAlicloudRedisClusterBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithAuthSecretLabels(map[string]string{"env": "prod"})
			},
		)

		canChangeSkr(
			"AlicloudRedisCluster authSecret.annotations can be changed",
			newTestAlicloudRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b"})
			},
		)

		canChangeSkr(
			"AlicloudRedisCluster authSecret.extraData can be changed",
			newTestAlicloudRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{"key": "v1"}),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithAuthSecretExtraData(map[string]string{"key": "v2"})
			},
		)
	})
})
