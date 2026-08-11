package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAlicloudRedisClusterBuilder struct {
	instance cloudresourcesv1beta1.AlicloudRedisCluster
}

func newTestAlicloudRedisClusterBuilder() *testAlicloudRedisClusterBuilder {
	return &testAlicloudRedisClusterBuilder{
		instance: cloudresourcesv1beta1.AlicloudRedisCluster{
			Spec: cloudresourcesv1beta1.AlicloudRedisClusterSpec{
				RedisTier:     cloudresourcesv1beta1.AlicloudRedisClusterTierC3,
				ShardCount:    2,
				EngineVersion: "5.0",
			},
		},
	}
}

func (b *testAlicloudRedisClusterBuilder) Build() *cloudresourcesv1beta1.AlicloudRedisCluster {
	return &b.instance
}

func (b *testAlicloudRedisClusterBuilder) WithRedisTier(redisTier cloudresourcesv1beta1.AlicloudRedisClusterTier) *testAlicloudRedisClusterBuilder {
	b.instance.Spec.RedisTier = redisTier
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithIpRange(ipRangeName string) *testAlicloudRedisClusterBuilder {
	b.instance.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithEngineVersion(engineVersion string) *testAlicloudRedisClusterBuilder {
	b.instance.Spec.EngineVersion = engineVersion
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *testAlicloudRedisClusterBuilder {
	b.instance.Spec.ReplicasPerShard = replicasPerShard
	return b
}

func (b *testAlicloudRedisClusterBuilder) WithAuthSecretName(name string) *testAlicloudRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

var _ = Describe("Feature: SKR AlicloudRedisCluster", Ordered, func() {

	Context("Scenario: redisTier enum validation", func() {

		canCreateSkr(
			"AlicloudRedisCluster can be created with C3 tier",
			newTestAlicloudRedisClusterBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC3),
		)

		canCreateSkr(
			"AlicloudRedisCluster can be created with C7 tier",
			newTestAlicloudRedisClusterBuilder().WithRedisTier(cloudresourcesv1beta1.AlicloudRedisClusterTierC7),
		)

		canNotCreateSkr(
			"AlicloudRedisCluster cannot be created with invalid redisTier",
			newTestAlicloudRedisClusterBuilder().WithRedisTier("C1"),
			"",
		)
	})

	Context("Scenario: replicasPerShard must be 0", func() {

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

	Context("Scenario: engineVersion immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 5.0 to 7.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("5.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("7.0")
			},
			"engineVersion is immutable",
		)

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 7.0 to 6.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("7.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("6.0")
			},
			"engineVersion is immutable",
		)

		canNotChangeSkr(
			"AlicloudRedisCluster engineVersion cannot be changed from 6.0 to 5.0",
			newTestAlicloudRedisClusterBuilder().WithEngineVersion("6.0"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithEngineVersion("5.0")
			},
			"engineVersion is immutable",
		)
	})

	Context("Scenario: IpRange immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisCluster IpRange cannot be changed",
			newTestAlicloudRedisClusterBuilder().WithIpRange("original-ip-range"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithIpRange("changed-ip-range")
			},
			"IpRange is immutable",
		)
	})

	Context("Scenario: authSecret immutability", func() {

		canNotChangeSkr(
			"AlicloudRedisCluster authSecret.name cannot be changed",
			newTestAlicloudRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AlicloudRedisCluster]) {
				b.(*testAlicloudRedisClusterBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)
	})
})
