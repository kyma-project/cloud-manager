package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAzureManagedRedisClusterBuilder struct {
	cluster cloudresourcesv1beta1.AzureManagedRedisCluster
}

func newTestAzureManagedRedisClusterBuilder() *testAzureManagedRedisClusterBuilder {
	return &testAzureManagedRedisClusterBuilder{
		cluster: cloudresourcesv1beta1.AzureManagedRedisCluster{
			Spec: cloudresourcesv1beta1.AzureManagedRedisClusterSpec{
				RedisTier: cloudresourcesv1beta1.AzureManagedRedisClusterTierC3,
			},
		},
	}
}

func (b *testAzureManagedRedisClusterBuilder) Build() *cloudresourcesv1beta1.AzureManagedRedisCluster {
	return &b.cluster
}

func (b *testAzureManagedRedisClusterBuilder) WithRedisTier(tier cloudresourcesv1beta1.AzureManagedRedisClusterTier) *testAzureManagedRedisClusterBuilder {
	b.cluster.Spec.RedisTier = tier
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithRawRedisTier(tier string) *testAzureManagedRedisClusterBuilder {
	b.cluster.Spec.RedisTier = cloudresourcesv1beta1.AzureManagedRedisClusterTier(tier)
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithIpRangeName(name string) *testAzureManagedRedisClusterBuilder {
	b.cluster.Spec.IpRange.Name = name
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithAuthSecretName(name string) *testAzureManagedRedisClusterBuilder {
	if b.cluster.Spec.AuthSecret == nil {
		b.cluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.cluster.Spec.AuthSecret.Name = name
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testAzureManagedRedisClusterBuilder {
	if b.cluster.Spec.AuthSecret == nil {
		b.cluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.cluster.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAzureManagedRedisClusterBuilder {
	if b.cluster.Spec.AuthSecret == nil {
		b.cluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.cluster.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAzureManagedRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAzureManagedRedisClusterBuilder {
	if b.cluster.Spec.AuthSecret == nil {
		b.cluster.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.cluster.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AzureManagedRedisCluster", Ordered, func() {

	Context("Scenario: redisTier validation", func() {

		// All C-tier values are accepted.
		for _, tier := range []cloudresourcesv1beta1.AzureManagedRedisClusterTier{
			cloudresourcesv1beta1.AzureManagedRedisClusterTierC3,
			cloudresourcesv1beta1.AzureManagedRedisClusterTierC4,
			cloudresourcesv1beta1.AzureManagedRedisClusterTierC5,
			cloudresourcesv1beta1.AzureManagedRedisClusterTierC6,
			cloudresourcesv1beta1.AzureManagedRedisClusterTierC7,
		} {
			canCreateSkr(
				"AzureManagedRedisCluster with redisTier="+string(tier)+" is accepted",
				newTestAzureManagedRedisClusterBuilder().WithRedisTier(tier),
			)
		}

		// Instance tiers (S/P) belong on AzureManagedRedisInstance, not Cluster.
		canNotCreateSkr(
			"AzureManagedRedisCluster with instance tier S1 is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier("S1"),
			"",
		)
		canNotCreateSkr(
			"AzureManagedRedisCluster with instance tier P1 is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier("P1"),
			"",
		)

		// Out-of-range / invalid C-tier numbers.
		canNotCreateSkr(
			"AzureManagedRedisCluster with redisTier=C1 is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier("C1"),
			"",
		)
		canNotCreateSkr(
			"AzureManagedRedisCluster with redisTier=C8 is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier("C8"),
			"",
		)

		// Arbitrary garbage is rejected.
		canNotCreateSkr(
			"AzureManagedRedisCluster with unknown redisTier is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier("Z9"),
			"",
		)

		canNotCreateSkr(
			"AzureManagedRedisCluster with empty redisTier is rejected",
			newTestAzureManagedRedisClusterBuilder().WithRawRedisTier(""),
			"",
		)
	})

	Context("Scenario: redisTier immutability", func() {

		canNotChangeSkr(
			"AzureManagedRedisCluster redisTier cannot be changed",
			newTestAzureManagedRedisClusterBuilder().WithRedisTier(cloudresourcesv1beta1.AzureManagedRedisClusterTierC3),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithRedisTier(cloudresourcesv1beta1.AzureManagedRedisClusterTierC5)
			},
			"redisTier is immutable",
		)
	})

	Context("Scenario: ipRange is optional", func() {

		canCreateSkr(
			"AzureManagedRedisCluster can be created without ipRange",
			newTestAzureManagedRedisClusterBuilder(),
		)

		canCreateSkr(
			"AzureManagedRedisCluster can be created with ipRange",
			newTestAzureManagedRedisClusterBuilder().WithIpRangeName("my-iprange"),
		)
	})

	Context("Scenario: authSecret mutability", func() {

		canCreateSkr(
			"AzureManagedRedisCluster with no authSecret",
			newTestAzureManagedRedisClusterBuilder(),
		)

		canCreateSkr(
			"AzureManagedRedisCluster with authSecret name",
			newTestAzureManagedRedisClusterBuilder().WithAuthSecretName("custom-secret"),
		)

		canCreateSkr(
			"AzureManagedRedisCluster with all authSecret fields",
			newTestAzureManagedRedisClusterBuilder().
				WithAuthSecretName("full-custom-secret").
				WithAuthSecretLabels(map[string]string{"app": "myapp"}).
				WithAuthSecretAnnotations(map[string]string{"env": "prod"}).
				WithAuthSecretExtraData(map[string]string{"url": "redis://{{.host}}:{{.port}}"}),
		)

		canNotChangeSkr(
			"AzureManagedRedisCluster authSecret.name cannot be changed",
			newTestAzureManagedRedisClusterBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AzureManagedRedisCluster authSecret.labels can be changed",
			newTestAzureManagedRedisClusterBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisCluster authSecret.annotations can be changed",
			newTestAzureManagedRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisCluster authSecret.extraData can be changed",
			newTestAzureManagedRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisCluster authSecret can be added after creation",
			newTestAzureManagedRedisClusterBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisCluster]) {
				b.(*testAzureManagedRedisClusterBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
