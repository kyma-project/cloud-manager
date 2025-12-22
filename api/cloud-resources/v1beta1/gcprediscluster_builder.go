package v1beta1

func NewGcpRedisClusterBuilder() *GcpRedisClusterBuilder {
	return &GcpRedisClusterBuilder{
		GcpRedisCluster: GcpRedisCluster{
			Spec: GcpRedisClusterSpec{},
		},
	}
}

type GcpRedisClusterBuilder struct {
	GcpRedisCluster GcpRedisCluster
}

func (b *GcpRedisClusterBuilder) Reset() *GcpRedisClusterBuilder {
	b.GcpRedisCluster = GcpRedisCluster{}
	return b
}

func (b *GcpRedisClusterBuilder) WithSubnet(subnetName string) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.Subnet.Name = subnetName
	return b
}

func (b *GcpRedisClusterBuilder) WithRedisTier(redisTier GcpRedisClusterTier) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.RedisTier = redisTier
	return b
}

func (b *GcpRedisClusterBuilder) WithRedisConfigs(redisConfigs map[string]string) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.RedisConfigs = redisConfigs
	return b
}

func (b *GcpRedisClusterBuilder) WithShardCount(shardCount int32) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.ShardCount = shardCount
	return b
}

func (b *GcpRedisClusterBuilder) WithReplicasPerShard(replicasPerShard int32) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.ReplicasPerShard = replicasPerShard
	return b
}

func (b *GcpRedisClusterBuilder) WithAuthSecret(name string, labels, annotations, extraData map[string]string) *GcpRedisClusterBuilder {
	b.GcpRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
		ExtraData:   extraData,
	}
	return b
}

func (b *GcpRedisClusterBuilder) WithAuthSecretName(name string) *GcpRedisClusterBuilder {
	if b.GcpRedisCluster.Spec.AuthSecret == nil {
		b.GcpRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisCluster.Spec.AuthSecret.Name = name
	return b
}

func (b *GcpRedisClusterBuilder) WithAuthSecretLabels(labels map[string]string) *GcpRedisClusterBuilder {
	if b.GcpRedisCluster.Spec.AuthSecret == nil {
		b.GcpRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisCluster.Spec.AuthSecret.Labels = labels
	return b
}

func (b *GcpRedisClusterBuilder) WithAuthSecretAnnotations(annotations map[string]string) *GcpRedisClusterBuilder {
	if b.GcpRedisCluster.Spec.AuthSecret == nil {
		b.GcpRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisCluster.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *GcpRedisClusterBuilder) WithAuthSecretExtraData(extraData map[string]string) *GcpRedisClusterBuilder {
	if b.GcpRedisCluster.Spec.AuthSecret == nil {
		b.GcpRedisCluster.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisCluster.Spec.AuthSecret.ExtraData = extraData
	return b
}
