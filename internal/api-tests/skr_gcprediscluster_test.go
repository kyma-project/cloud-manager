package api_tests

import (
	"fmt"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testGcpRedisClusterBuilder struct {
	*cloudresourcesv1beta1.GcpRedisClusterBuilder
}

func newTestGcpRedisClusterBuilder() *testGcpRedisClusterBuilder {
	return &testGcpRedisClusterBuilder{
		GcpRedisClusterBuilder: cloudresourcesv1beta1.NewGcpRedisClusterBuilder().
			WithSubnet(uuid.NewString()),
	}
}

func (b *testGcpRedisClusterBuilder) Build() *cloudresourcesv1beta1.GcpRedisCluster {
	return &b.GcpRedisCluster
}

func (b *testGcpRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.GcpRedisClusterTier) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithRedisTier(redisTier)
	return b
}

func (b *testGcpRedisClusterBuilder) WithShardCount(shardCount int32) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithShardCount(shardCount)
	return b
}

func (b *testGcpRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithReplicasPerShard(replicasPerShard)
	return b
}

func (b *testGcpRedisClusterBuilder) WithAuthSecretName(name string) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithAuthSecretName(name)
	return b
}

func (b *testGcpRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithAuthSecretLabels(labels)
	return b
}

func (b *testGcpRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithAuthSecretAnnotations(annotations)
	return b
}

func (b *testGcpRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testGcpRedisClusterBuilder {
	b.GcpRedisClusterBuilder.WithAuthSecretExtraData(extraData)
	return b
}

var _ = Describe("Feature: SKR GcpRedisCluster", Ordered, func() {

	canChangeSkr(
		"GcpRedisCluster shardCount can be increased",
		newTestGcpRedisClusterBuilder().WithShardCount(2).WithReplicasPerShard(1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
			b.(*testGcpRedisClusterBuilder).WithShardCount(3)
		},
	)

	canChangeSkr(
		"GcpRedisCluster shardCount can be decreased",
		newTestGcpRedisClusterBuilder().WithShardCount(2).WithReplicasPerShard(1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
			b.(*testGcpRedisClusterBuilder).WithShardCount(1)
		},
	)

	canNotCreateSkr(
		"GcpRedisCluster shardCount can not be less than 1",
		newTestGcpRedisClusterBuilder().WithShardCount(0).WithReplicasPerShard(1),
		"spec.shardCount: Invalid value: 0",
	)

	canChangeSkr(
		"GcpRedisCluster replicasPerShard can be increased",
		newTestGcpRedisClusterBuilder().WithShardCount(2).WithReplicasPerShard(1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
			b.(*testGcpRedisClusterBuilder).WithReplicasPerShard(2)
		},
	)

	canChangeSkr(
		"GcpRedisCluster replicasPerShard can be decreased",
		newTestGcpRedisClusterBuilder().WithShardCount(2).WithReplicasPerShard(1),
		func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
			b.(*testGcpRedisClusterBuilder).WithReplicasPerShard(0)
		},
	)

	canNotCreateSkr(
		"GcpRedisCluster replicasPerShard can not be more than 2",
		newTestGcpRedisClusterBuilder().WithShardCount(0).WithReplicasPerShard(3),
		"spec.replicasPerShard: Invalid value: 3",
	)

	shardToReplicaLimits := [][]int32{
		{250, 0},
		{125, 1},
		{83, 2},
	}

	for _, shardReplicaPair := range shardToReplicaLimits {
		shards := shardReplicaPair[0]
		replicasPerShard := shardReplicaPair[1]
		canCreateSkr(
			fmt.Sprintf("GcpRedisCluster can be created with %d shards if replicasPerShard is %d", shards, replicasPerShard),
			newTestGcpRedisClusterBuilder().WithShardCount(shards).WithReplicasPerShard(replicasPerShard),
		)

		canNotCreateSkr(
			fmt.Sprintf("GcpRedisCluster can not be created with %d shards (1 over limit) if replicasPerShard is %d", shards+1, replicasPerShard),
			newTestGcpRedisClusterBuilder().WithShardCount(shards+1).WithReplicasPerShard(replicasPerShard),
			fmt.Sprintf("shardCount must be %d or less when replicasPerShard is %d", shards, replicasPerShard),
		)
	}

	Context("Scenario: authSecret mutability", func() {

		canChangeSkr(
			"GcpRedisCluster authSecret.name can be changed",
			newTestGcpRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
				b.(*testGcpRedisClusterBuilder).WithAuthSecretName("new-name")
			},
		)

		canChangeSkr(
			"GcpRedisCluster authSecret.labels can be changed",
			newTestGcpRedisClusterBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
				b.(*testGcpRedisClusterBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"GcpRedisCluster authSecret.annotations can be changed",
			newTestGcpRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
				b.(*testGcpRedisClusterBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"GcpRedisCluster authSecret.extraData can be changed",
			newTestGcpRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
				b.(*testGcpRedisClusterBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"GcpRedisCluster authSecret can be added",
			newTestGcpRedisClusterBuilder(),
			func(b Builder[*cloudresourcesv1beta1.GcpRedisCluster]) {
				b.(*testGcpRedisClusterBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
