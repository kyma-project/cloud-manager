package v1beta1

func NewAwsRedisClusterBuilder() *AwsRedisClusterBuilder {
	return &AwsRedisClusterBuilder{
		AwsRedisCluster: AwsRedisCluster{
			Spec: AwsRedisClusterSpec{},
		},
	}
}

type AwsRedisClusterBuilder struct {
	AwsRedisCluster AwsRedisCluster
}

func (b *AwsRedisClusterBuilder) Reset() *AwsRedisClusterBuilder {
	b.AwsRedisCluster = AwsRedisCluster{}
	return b
}

func (b *AwsRedisClusterBuilder) WithIpRange(ipRangeName string) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *AwsRedisClusterBuilder) WithRedisTier(redisTier AwsRedisClusterTier) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.RedisTier = redisTier
	return b
}

func (b *AwsRedisClusterBuilder) WithEngineVersion(engineVersion string) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.EngineVersion = engineVersion
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthEnabled(authEnabled bool) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.AuthEnabled = authEnabled
	return b
}

func (b *AwsRedisClusterBuilder) WithShardCount(shardCount int32) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.ShardCount = shardCount
	return b
}

func (b *AwsRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.ReplicasPerShard = replicasPerShard
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthSecret(name string, labels, annotations, extraData map[string]string) *AwsRedisClusterBuilder {
	b.AwsRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
		ExtraData:   extraData,
	}
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthSecretName(name string) *AwsRedisClusterBuilder {
	if b.AwsRedisCluster.Spec.AuthSecret == nil {
		b.AwsRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisCluster.Spec.AuthSecret.Name = name
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *AwsRedisClusterBuilder {
	if b.AwsRedisCluster.Spec.AuthSecret == nil {
		b.AwsRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisCluster.Spec.AuthSecret.Labels = labels
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *AwsRedisClusterBuilder {
	if b.AwsRedisCluster.Spec.AuthSecret == nil {
		b.AwsRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisCluster.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *AwsRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *AwsRedisClusterBuilder {
	if b.AwsRedisCluster.Spec.AuthSecret == nil {
		b.AwsRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisCluster.Spec.AuthSecret.ExtraData = extraData
	return b
}
