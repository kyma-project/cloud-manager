package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAzureRedisClusterBuilder struct {
	instance cloudresourcesv1beta1.AzureRedisCluster
}

func newTestAzureRedisClusterBuilder() *testAzureRedisClusterBuilder {
	return &testAzureRedisClusterBuilder{
		instance: cloudresourcesv1beta1.AzureRedisCluster{
			Spec: cloudresourcesv1beta1.AzureRedisClusterSpec{
				RedisTier:    cloudresourcesv1beta1.AzureRedisTierC3,
				RedisVersion: "7.2",
				RedisConfiguration: cloudresourcesv1beta1.RedisClusterAzureConfigs{
					MaxMemoryPolicy: "allkeys-lru",
				},
			},
		},
	}
}

func (b *testAzureRedisClusterBuilder) Build() *cloudresourcesv1beta1.AzureRedisCluster {
	return &b.instance
}

func (b *testAzureRedisClusterBuilder) WithAuthSecretName(name string) *testAzureRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Name = name
	return b
}

func (b *testAzureRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *testAzureRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *testAzureRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *testAzureRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *testAzureRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *testAzureRedisClusterBuilder {
	if b.instance.Spec.AuthSecret == nil {
		b.instance.Spec.AuthSecret = &cloudresourcesv1beta1.RedisAuthSecretSpec{}
	}
	b.instance.Spec.AuthSecret.ExtraData = extraData
	return b
}

var _ = Describe("Feature: SKR AzureRedisCluster", Ordered, func() {

	Context("Scenario: authSecret mutability", func() {

		canCreateSkr(
			"AzureRedisCluster with no authSecret",
			newTestAzureRedisClusterBuilder(),
		)

		canCreateSkr(
			"AzureRedisCluster with authSecret name",
			newTestAzureRedisClusterBuilder().WithAuthSecretName("custom-cluster-secret"),
		)

		canCreateSkr(
			"AzureRedisCluster with authSecret labels",
			newTestAzureRedisClusterBuilder().WithAuthSecretLabels(map[string]string{
				"cluster-label": "value1",
				"team":          "data",
			}),
		)

		canCreateSkr(
			"AzureRedisCluster with authSecret annotations",
			newTestAzureRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{
				"cluster-annotation": "value1",
				"owner":              "team-data",
			}),
		)

		canCreateSkr(
			"AzureRedisCluster with authSecret extraData",
			newTestAzureRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{
				"clusterKey":    "clusterValue",
				"connectionUrl": "redis://{{.host}}:{{.port}}",
			}),
		)

		canCreateSkr(
			"AzureRedisCluster with all authSecret fields",
			newTestAzureRedisClusterBuilder().
				WithAuthSecretName("full-cluster-secret").
				WithAuthSecretLabels(map[string]string{"app": "cache"}).
				WithAuthSecretAnnotations(map[string]string{"env": "staging"}).
				WithAuthSecretExtraData(map[string]string{"endpoint": "redis://{{.host}}:{{.port}}"}),
		)

		canChangeSkr(
			"AzureRedisCluster authSecret.name can be changed",
			newTestAzureRedisClusterBuilder().WithAuthSecretName("cluster-original"),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				b.(*testAzureRedisClusterBuilder).WithAuthSecretName("cluster-new")
			},
		)

		canChangeSkr(
			"AzureRedisCluster authSecret.labels can be changed",
			newTestAzureRedisClusterBuilder().WithAuthSecretLabels(map[string]string{"env": "test"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				b.(*testAzureRedisClusterBuilder).WithAuthSecretLabels(map[string]string{"env": "prod", "cluster": "main"})
			},
		)

		canChangeSkr(
			"AzureRedisCluster authSecret.annotations can be changed",
			newTestAzureRedisClusterBuilder().WithAuthSecretAnnotations(map[string]string{"manager": "team-x"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				b.(*testAzureRedisClusterBuilder).WithAuthSecretAnnotations(map[string]string{"manager": "team-y", "budget": "5678"})
			},
		)

		canChangeSkr(
			"AzureRedisCluster authSecret.extraData can be changed",
			newTestAzureRedisClusterBuilder().WithAuthSecretExtraData(map[string]string{"data1": "val1"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				b.(*testAzureRedisClusterBuilder).WithAuthSecretExtraData(map[string]string{"data1": "updated", "data2": "val2"})
			},
		)

		canChangeSkr(
			"AzureRedisCluster authSecret can be added",
			newTestAzureRedisClusterBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				b.(*testAzureRedisClusterBuilder).WithAuthSecretName("cluster-added")
			},
		)

		canChangeSkr(
			"AzureRedisCluster authSecret fields can be removed",
			newTestAzureRedisClusterBuilder().
				WithAuthSecretName("to-remove-cluster").
				WithAuthSecretLabels(map[string]string{"remove": "me"}),
			func(b Builder[*cloudresourcesv1beta1.AzureRedisCluster]) {
				builder := b.(*testAzureRedisClusterBuilder)
				builder.WithAuthSecretName("")
				builder.instance.Spec.AuthSecret.Labels = nil
			},
		)
	})
})
