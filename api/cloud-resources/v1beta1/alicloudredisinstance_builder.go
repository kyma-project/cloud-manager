package v1beta1

func NewAlicloudRedisInstanceBuilder() *AlicloudRedisInstanceBuilder {
	return &AlicloudRedisInstanceBuilder{
		AlicloudRedisInstance: AlicloudRedisInstance{
			Spec: AlicloudRedisInstanceSpec{},
		},
	}
}

// +kubebuilder:object:generate=false

type AlicloudRedisInstanceBuilder struct {
	AlicloudRedisInstance AlicloudRedisInstance
}

func (b *AlicloudRedisInstanceBuilder) WithIpRange(ipRangeName string) *AlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstance.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithRedisTier(redisTier AlicloudRedisTier) *AlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstance.Spec.RedisTier = redisTier
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithEngineVersion(engineVersion string) *AlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstance.Spec.EngineVersion = engineVersion
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithParameters(parameters map[string]string) *AlicloudRedisInstanceBuilder {
	b.AlicloudRedisInstance.Spec.Parameters = parameters
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithAuthSecretName(name string) *AlicloudRedisInstanceBuilder {
	if b.AlicloudRedisInstance.Spec.AuthSecret == nil {
		b.AlicloudRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisInstance.Spec.AuthSecret.Name = name
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *AlicloudRedisInstanceBuilder {
	if b.AlicloudRedisInstance.Spec.AuthSecret == nil {
		b.AlicloudRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisInstance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *AlicloudRedisInstanceBuilder {
	if b.AlicloudRedisInstance.Spec.AuthSecret == nil {
		b.AlicloudRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisInstance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *AlicloudRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *AlicloudRedisInstanceBuilder {
	if b.AlicloudRedisInstance.Spec.AuthSecret == nil {
		b.AlicloudRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AlicloudRedisInstance.Spec.AuthSecret.ExtraData = extraData
	return b
}
