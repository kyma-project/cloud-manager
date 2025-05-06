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

type testGcpRedisClusterBuilder struct {
	instance cloudresourcesv1beta1.GcpRedisCluster
}

func newTestGcpRedisClusterBuilder() *testGcpRedisClusterBuilder {
	return &testGcpRedisClusterBuilder{
		instance: cloudresourcesv1beta1.GcpRedisCluster{
			Spec: cloudresourcesv1beta1.GcpRedisClusterSpec{
				Subnet: cloudresourcesv1beta1.GcpSubnetRef{
					Name: uuid.NewString(),
				},
				RedisTier: cloudresourcesv1beta1.GcpRedisClusterTierC1,
				RedisConfigs: map[string]string{
					"maxmemory-policy": "allkeys-lru",
				},
				ShardCount:       2,
				ReplicasPerShard: 1,
			},
		},
	}
}

func (b *testGcpRedisClusterBuilder) Build() *cloudresourcesv1beta1.GcpRedisCluster {
	return &b.instance
}

func (b *testGcpRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.GcpRedisClusterTier) *testGcpRedisClusterBuilder {
	b.instance.Spec.RedisTier = redisTier
	return b
}

func (b *testGcpRedisClusterBuilder) WithShardCount(shardCount int32) *testGcpRedisClusterBuilder {
	b.instance.Spec.ShardCount = shardCount
	return b
}
func (b *testGcpRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testGcpRedisClusterBuilder {
	b.instance.Spec.ReplicasPerShard = replicasPerShard
	return b
}

var _ = Describe("Feature: SKR GcpRedisCluster", Ordered, func() {

	It("Given SKR default namespace exists", func() {
		Eventually(CreateNamespace).
			WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}).
			Should(Succeed())
	})

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
})
