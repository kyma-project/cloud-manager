package v1beta1

func NewAlicloudRedisClusterBuilder() *AlicloudRedisClusterBuilder {
	return &AlicloudRedisClusterBuilder{
		AlicloudRedisCluster: AlicloudRedisCluster{
			Spec: AlicloudRedisClusterSpec{},
		},
	}
}

// +kubebuilder:object:generate=false

type AlicloudRedisClusterBuilder struct {
	AlicloudRedisCluster AlicloudRedisCluster
}

func (b *AlicloudRedisClusterBuilder) WithIpRange(ipRangeName string) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *AlicloudRedisClusterBuilder) WithRedisTier(redisTier AlicloudRedisClusterTier) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.RedisTier = redisTier
	return b
}

func (b *AlicloudRedisClusterBuilder) WithShardCount(shardCount int32) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.ShardCount = shardCount
	return b
}

func (b *AlicloudRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.ReplicasPerShard = replicasPerShard
	return b
}

func (b *AlicloudRedisClusterBuilder) WithEngineVersion(engineVersion string) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.EngineVersion = engineVersion
	return b
}

func (b *AlicloudRedisClusterBuilder) WithParameters(parameters map[string]string) *AlicloudRedisClusterBuilder {
	b.AlicloudRedisCluster.Spec.Parameters = parameters
	return b
}

func (b *AlicloudRedisClusterBuilder) WithAuthSecretName(name string) *AlicloudRedisClusterBuilder {
	if b.AlicloudRedisCluster.Spec.AuthSecret == nil {
		b.AlicloudRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisCluster.Spec.AuthSecret.Name = name
	return b
}

func (b *AlicloudRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *AlicloudRedisClusterBuilder {
	if b.AlicloudRedisCluster.Spec.AuthSecret == nil {
		b.AlicloudRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisCluster.Spec.AuthSecret.Labels = labels
	return b
}

func (b *AlicloudRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *AlicloudRedisClusterBuilder {
	if b.AlicloudRedisCluster.Spec.AuthSecret == nil {
		b.AlicloudRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisCluster.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *AlicloudRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *AlicloudRedisClusterBuilder {
	if b.AlicloudRedisCluster.Spec.AuthSecret == nil {
		b.AlicloudRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisCluster.Spec.AuthSecret.ExtraData = extraData
	return b
}
