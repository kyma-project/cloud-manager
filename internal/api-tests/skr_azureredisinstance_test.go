package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAzureRedisInstanceBuilder struct {
	instance cloudresourcesv1beta1.AzureRedisInstance
}

func newTestAzureRedisInstanceBuilder() *testAzureRedisInstanceBuilder {
	return &testAzureRedisInstanceBuilder{
		instance: cloudresourcesv1beta1.AzureRedisInstance{
			Spec: cloudresourcesv1beta1.AzureRedisInstanceSpec{
				RedisTier:    cloudresourcesv1beta1.AzureRedisTierP1,
				RedisVersion: "6.0",
			},
		},
	}
}

func (b *testAzureRedisInstanceBuilder) Build() *cloudresourcesv1beta1.AzureRedisInstance {
	return &b.instance
}

func (b *testAzureRedisInstanceBuilder) WithAuthSecretName(name string) *testAzureRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

func (b *testAzureRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *testAzureRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAzureRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAzureRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAzureRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAzureRedisInstanceBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AzureRedisInstance", Ordered, func() {

	Context("Scenario: authSecret mutability", func() {

		canCreateSkr(
			"AzureRedisInstance with no authSecret",
			newTestAzureRedisInstanceBuilder(),
		)

		canCreateSkr(
			"AzureRedisInstance with authSecret name",
			newTestAzureRedisInstanceBuilder().WithAuthSecretName("custom-secret"),
		)

		canCreateSkr(
			"AzureRedisInstance with authSecret labels",
			newTestAzureRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{
				"custom-label": "value1",
				"team":         "platform",
			}),
		)

		canCreateSkr(
			"AzureRedisInstance with authSecret annotations",
			newTestAzureRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{
				"custom-annotation": "value1",
				"owner":             "team-platform",
			}),
		)

		canCreateSkr(
			"AzureRedisInstance with authSecret extraData",
			newTestAzureRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{
				"customKey":     "customValue",
				"connectionUrl": "redis://{{.host}}:{{.port}}",
			}),
		)

		canCreateSkr(
			"AzureRedisInstance with all authSecret fields",
			newTestAzureRedisInstanceBuilder().
				WithAuthSecretName("full-custom-secret").
				WithAuthSecretLabels(map[string]string{"app": "myapp"}).
				WithAuthSecretAnnotations(map[string]string{"env": "prod"}).
				WithAuthSecretExtraData(map[string]string{"url": "redis://{{.host}}:{{.port}}"}),
		)

		canNotChangeSkr(
			"AzureRedisInstance authSecret.name cannot be changed",
			newTestAzureRedisInstanceBuilder().WithAuthSecretName("original-name"),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				b.(*testAzureRedisInstanceBuilder).WithAuthSecretName("new-name")
			},
			"name is immutable",
		)

		canChangeSkr(
			"AzureRedisInstance authSecret.labels can be changed",
			newTestAzureRedisInstanceBuilder().WithAuthSecretLabels(map[string]string{"env": "dev"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				b.(*testAzureRedisInstanceBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "team": "platform"})
			},
		)

		canChangeSkr(
			"AzureRedisInstance authSecret.annotations can be changed",
			newTestAzureRedisInstanceBuilder().WithAuthSecretAnnotations(map[string]string{"owner": "team-a"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				b.(*testAzureRedisInstanceBuilder).WithAuthSecretAnnotations(map[string]string{"owner": "team-b", "cost-center": "1234"})
			},
		)

		canChangeSkr(
			"AzureRedisInstance authSecret.extraData can be changed",
			newTestAzureRedisInstanceBuilder().WithAuthSecretExtraData(map[string]string{"key1": "value1"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				b.(*testAzureRedisInstanceBuilder).WithAuthSecretExtraData(map[string]string{"key1": "new-value", "key2": "value2"})
			},
		)

		canChangeSkr(
			"AzureRedisInstance authSecret can be added",
			newTestAzureRedisInstanceBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				b.(*testAzureRedisInstanceBuilder).WithAuthSecretName("added-secret")
			},
		)

		canChangeSkr(
			"AzureRedisInstance authSecret fields can be removed",
			newTestAzureRedisInstanceBuilder().
				WithAuthSecretName("to-remove").
				WithAuthSecretLabels(map[string]string{"key": "value"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisInstance]) {
				builder := b.(*testAzureRedisInstanceBuilder)
				builder.WithAuthSecretName("")
				builder.instance.Spec.AuthSecret.Labels = nil
			},
		)
	})
})
