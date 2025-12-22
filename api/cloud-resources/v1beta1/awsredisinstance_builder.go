package v1beta1

func NewAwsRedisInstanceBuilder() *AwsRedisInstanceBuilder {
	return &AwsRedisInstanceBuilder{
		AwsRedisInstance: AwsRedisInstance{
			Spec: AwsRedisInstanceSpec{},
		},
	}
}

type AwsRedisInstanceBuilder struct {
	AwsRedisInstance AwsRedisInstance
}

func (b *AwsRedisInstanceBuilder) Reset() *AwsRedisInstanceBuilder {
	b.AwsRedisInstance = AwsRedisInstance{}
	return b
}

func (b *AwsRedisInstanceBuilder) WithIpRange(ipRangeName string) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *AwsRedisInstanceBuilder) WithRedisTier(redisTier AwsRedisTier) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.RedisTier = redisTier
	return b
}

func (b *AwsRedisInstanceBuilder) WithEngineVersion(engineVersion string) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.EngineVersion = engineVersion
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthEnabled(authEnabled bool) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.AuthEnabled = authEnabled
	return b
}

func (b *AwsRedisInstanceBuilder) WithParameters(parameters map[string]string) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.Parameters = parameters
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthSecret(name string, labels, annotations, extraData map[string]string) *AwsRedisInstanceBuilder {
	b.AwsRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
		ExtraData:   extraData,
	}
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthSecretName(name string) *AwsRedisInstanceBuilder {
	if b.AwsRedisInstance.Spec.AuthSecret == nil {
		b.AwsRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisInstance.Spec.AuthSecret.Name = name
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *AwsRedisInstanceBuilder {
	if b.AwsRedisInstance.Spec.AuthSecret == nil {
		b.AwsRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisInstance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *AwsRedisInstanceBuilder {
	if b.AwsRedisInstance.Spec.AuthSecret == nil {
		b.AwsRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisInstance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *AwsRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *AwsRedisInstanceBuilder {
	if b.AwsRedisInstance.Spec.AuthSecret == nil {
		b.AwsRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.AwsRedisInstance.Spec.AuthSecret.ExtraData = extraData
	return b
}
