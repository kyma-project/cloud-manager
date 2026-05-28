package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAzureManagedRedisInstanceBuilder struct {
	instance cloudresourcesv1beta1.AzureManagedRedisInstance
}

func newTestAzureManagedRedisInstanceBuilder() *testAzureManagedRedisInstanceBuilder {
	return &testAzureManagedRedisInstanceBuilder{
		instance: cloudresourcesv1beta1.AzureManagedRedisInstance{
			Spec: cloudresourcesv1beta1.AzureManagedRedisInstanceSpec{
				RedisTier: cloudresourcesv1beta1.AzureManagedRedisInstanceTierP1,
			},
		},
	}
}

func (b *testAzureManagedRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AzureManagedRedisInstance {
	return &b.instance
}

func (b *testAzureManagedRedisInstanceBuilder) WithRedisTier(tier cloudresourcesv1beta1.AzureManagedRedisInstanceTier) *testAzureManagedRedisInstanceBuilder {
	b.instance.Spec.RedisTier = tier
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithRawRedisTier(tier string) *testAzureManagedRedisInstanceBuilder {
	b.instance.Spec.RedisTier = cloudresourcesv1beta1.AzureManagedRedisInstanceTier(tier)
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithIpRangeName(name string) *testAzureManagedRedisInstanceBuilder {
	b.instance.Spec.IpRange.Name = name
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithAuthSecretName(name string) *testAzureManagedRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *testAzureManagedRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAzureManagedRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAzureManagedRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAzureManagedRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AzureManagedRedisInstance", Ordered, func() {

	Context("Scenario: redisTier validation", func() {

		// All S-tier (non-HA) values are accepted.
		for _, tier := range []cloudresourcesv1beta1.AzureManagedRedisInstanceTier{
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierS1,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierS2,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierS3,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierS4,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierS5,
		} {
			canCreateSkr(
				"AzureManagedRedisInstance with redisTier="+string(tier)+" is accepted",
				newTestAzureManagedRedisInstanceBuilder().WithRedisTier(tier),
			)
		}

		// All P-tier (HA) values are accepted.
		for _, tier := range []cloudresourcesv1beta1.AzureManagedRedisInstanceTier{
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierP1,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierP2,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierP3,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierP4,
			cloudresourcesv1beta1.AzureManagedRedisInstanceTierP5,
		} {
			canCreateSkr(
				"AzureManagedRedisInstance with redisTier="+string(tier)+" is accepted",
				newTestAzureManagedRedisInstanceBuilder().WithRedisTier(tier),
			)
		}

		// C-tiers belong on AzureManagedRedisCluster, not Instance.
		canNotCreateSkr(
			"AzureManagedRedisInstance with cluster tier C3 is rejected",
			newTestAzureManagedRedisInstanceBuilder().WithRawRedisTier("C3"),
			"",
		)

		// Arbitrary garbage is rejected.
		canNotCreateSkr(
			"AzureManagedRedisInstance with unknown redisTier is rejected",
			newTestAzureManagedRedisInstanceBuilder().WithRawRedisTier("Z9"),
			"",
		)

		canNotCreateSkr(
			"AzureManagedRedisInstance with empty redisTier is rejected",
			newTestAzureManagedRedisInstanceBuilder().WithRawRedisTier(""),
			"",
		)
	})

	Context("Scenario: redisTier immutability", func() {

		canNotChangeSkr(
			"AzureManagedRedisInstance redisTier cannot be changed",
			newTestAzureManagedRedisInstanceBuilder().WithRedisTier(cloudresourcesv1beta1.AzureManagedRedisInstanceTierP1),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithRedisTier(cloudresourcesv1beta1.AzureManagedRedisInstanceTierP3)
			},
			"redisTier is immutable",
		)
	})

	Context("Scenario: ipRange is optional", func() {

		canCreateSkr(
			"AzureManagedRedisInstance can be created without ipRange",
			newTestAzureManagedRedisInstanceBuilder(),
		)

		canCreateSkr(
			"AzureManagedRedisInstance can be created with ipRange",
			newTestAzureManagedRedisInstanceBuilder().WithIpRangeName("my-iprange"),
		)
	})

	Context("Scenario: authSecret mutability", func() {

		canCreateSkr(
			"AzureManagedRedisInstance with no authSecret",
			newTestAzureManagedRedisInstanceBuilder(),
		)

		canCreateSkr(
			"AzureManagedRedisInstance with authSecret name",
			newTestAzureManagedRedisInstanceBuilder().WithAuthSecretName("custom-secret"),
		)

		canCreateSkr(
			"AzureManagedRedisInstance with all authSecret fields",
			newTestAzureManagedRedisInstanceBuilder().
				WithAuthSecretName("full-custom-secret").
				WithAuthSecretLabels(map[string]string{"app": "myapp"}).
				WithAuthSecretAnnotations(map[string]string{"env": "prod"}).
				WithAuthSecretExtraData(map[string]string{"url": "redis://{{.host}}:{{.port}}"}),
		)

		canNotChangeSkr(
			"AzureManagedRedisInstance authSecret.name cannot be changed",
			newTestAzureManagedRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AzureManagedRedisInstance authSecret.labels can be changed",
			newTestAzureManagedRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisInstance authSecret.annotations can be changed",
			newTestAzureManagedRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisInstance authSecret.extraData can be changed",
			newTestAzureManagedRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AzureManagedRedisInstance authSecret can be added after creation",
			newTestAzureManagedRedisInstanceBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedisInstance]) {
				b.(*testAzureManagedRedisInstanceBuilder).WithAuthSecretName("added-secret")
			},
		)
	})
})
