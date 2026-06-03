package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAzureManagedRedisBuilder struct {
	instance cloudresourcesv1beta1.AzureManagedRedis
}

func newTestAzureManagedRedisBuilder() *testAzureManagedRedisBuilder {
	return &testAzureManagedRedisBuilder{
		instance: cloudresourcesv1beta1.AzureManagedRedis{
			Spec: cloudresourcesv1beta1.AzureManagedRedisSpec{
				RedisTier: cloudresourcesv1beta1.AzureManagedRedisTierP1,
			},
		},
	}
}

func (b *testAzureManagedRedisBuilder) Build() *cloudresourcesv1beta1.AzureManagedRedis {
	return &b.instance
}

func (b *testAzureManagedRedisBuilder) WithTier(tier cloudresourcesv1beta1.AzureManagedRedisTier) *testAzureManagedRedisBuilder {
	b.instance.Spec.RedisTier = tier
	return b
}

func (b *testAzureManagedRedisBuilder) WithRawTier(tier string) *testAzureManagedRedisBuilder {
	b.instance.Spec.RedisTier = cloudresourcesv1beta1.AzureManagedRedisTier(tier)
	return b
}

func (b *testAzureManagedRedisBuilder) WithAuthSecretName(name string) *testAzureManagedRedisBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

func (b *testAzureManagedRedisBuilder) WithAuthSecretLabels(labels map[string]string) *testAzureManagedRedisBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAzureManagedRedisBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAzureManagedRedisBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAzureManagedRedisBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAzureManagedRedisBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AzureManagedRedis", Ordered, func() {

	Context("Scenario: tier enum + immutability", func() {

		// Spot-check one tier from each family rather than all 15 — the
		// CRD enum is generated from a single Go const block, so a single
		// representative per family is enough to exercise the enum admission
		// path. Unit tests in pkg/skr/azuremanagedredis/util_test.go cover
		// the full mapping.
		canCreateSkr(
			"AzureManagedRedis with S1 tier (Balanced/non-HA/EnterpriseCluster)",
			newTestAzureManagedRedisBuilder().WithTier(cloudresourcesv1beta1.AzureManagedRedisTierS1),
		)

		canCreateSkr(
			"AzureManagedRedis with P3 tier (ComputeOptimized/HA/EnterpriseCluster)",
			newTestAzureManagedRedisBuilder().WithTier(cloudresourcesv1beta1.AzureManagedRedisTierP3),
		)

		canCreateSkr(
			"AzureManagedRedis with C5 tier (ComputeOptimized/HA/OSSCluster)",
			newTestAzureManagedRedisBuilder().WithTier(cloudresourcesv1beta1.AzureManagedRedisTierC5),
		)

		canNotCreateSkr(
			"AzureManagedRedis with unknown tier value is rejected",
			newTestAzureManagedRedisBuilder().WithRawTier("Z9"),
			"spec.redisTier",
		)

		canNotCreateSkr(
			"AzureManagedRedis with C-tier below C3 is rejected (CRD enum starts at C3)",
			newTestAzureManagedRedisBuilder().WithRawTier("C1"),
			"spec.redisTier",
		)

		canNotChangeSkr(
			"AzureManagedRedis tier is immutable",
			newTestAzureManagedRedisBuilder().WithTier(cloudresourcesv1beta1.AzureManagedRedisTierP1),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithTier(cloudresourcesv1beta1.AzureManagedRedisTierP2)
			},
			"redisTier is immutable",
		)
	})

	Context("Scenario: authSecret create + mutability", func() {

		canCreateSkr(
			"AzureManagedRedis with no authSecret",
			newTestAzureManagedRedisBuilder(),
		)

		canCreateSkr(
			"AzureManagedRedis with authSecret name",
			newTestAzureManagedRedisBuilder().WithAuthSecretName("custom-secret"),
		)

		canCreateSkr(
			"AzureManagedRedis with authSecret labels",
			newTestAzureManagedRedisBuilder().WithAuthSecretLabels(map[string]string{
				"custom-label": "value1",
				"team":         "platform",
			}),
		)

		canCreateSkr(
			"AzureManagedRedis with authSecret annotations",
			newTestAzureManagedRedisBuilder().WithAuthSecretAnnotations(map[string]string{
				"custom-annotation": "value1",
				"owner":             "team-platform",
			}),
		)

		canCreateSkr(
			"AzureManagedRedis with authSecret extraData",
			newTestAzureManagedRedisBuilder().WithAuthSecretExtraData(map[string]string{
				"customKey":     "customValue",
				"connectionUrl": "redis://{{.host}}:{{.port}}",
			}),
		)

		canCreateSkr(
			"AzureManagedRedis with all authSecret fields",
			newTestAzureManagedRedisBuilder().
				WithAuthSecretName("full-custom-secret").
				WithAuthSecretLabels(map[string]string{"app": "myapp"}).
				WithAuthSecretAnnotations(map[string]string{"env": "prod"}).
				WithAuthSecretExtraData(map[string]string{"url": "redis://{{.host}}:{{.port}}"}),
		)

		canNotChangeSkr(
			"AzureManagedRedis authSecret.name cannot be changed",
			newTestAzureManagedRedisBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AzureManagedRedis authSecret.labels can be changed",
			newTestAzureManagedRedisBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AzureManagedRedis authSecret.annotations can be changed",
			newTestAzureManagedRedisBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"AzureManagedRedis authSecret.extraData can be changed",
			newTestAzureManagedRedisBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AzureManagedRedis authSecret can be added",
			newTestAzureManagedRedisBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				b.(*testAzureManagedRedisBuilder).WithAuthSecretName("added-secret")
			},
		)

		canChangeSkr(
			"AzureManagedRedis authSecret fields can be removed",
			newTestAzureManagedRedisBuilder().
				WithAuthSecretName("to-remove").
				WithAuthSecretLabels(map[string]string{"key": "value"}),
			func(b Builder[*cloudresourcesv1beta1.AzureManagedRedis]) {
				builder := b.(*testAzureManagedRedisBuilder)
				builder.WithAuthSecretName("")
				builder.instance.Spec.AuthSecret.Labels = nil
			},
		)
	})
})
