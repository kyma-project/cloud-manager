package v1beta1

func NewGcpRedisInstanceBuilder() *GcpRedisInstanceBuilder {
	return &GcpRedisInstanceBuilder{
		GcpRedisInstance: GcpRedisInstance{
			Spec: GcpRedisInstanceSpec{
				RedisTier:    "S1",
				RedisVersion: "REDIS_7_0",
				AuthEnabled:  true,
				RedisConfigs: map[string]string{
					"maxmemory-policy": "allkeys-lru",
				},
				MaintenancePolicy: &MaintenancePolicy{
					DayOfWeek: &DayOfWeekPolicy{
						Day: "MONDAY",
						StartTime: TimeOfDay{
							Hours:   11,
							Minutes: 0,
						},
					},
				},
			},
		},
	}
}

type GcpRedisInstanceBuilder struct {
	GcpRedisInstance GcpRedisInstance
}

func (b *GcpRedisInstanceBuilder) Reset() *GcpRedisInstanceBuilder {
	b.GcpRedisInstance = GcpRedisInstance{}
	return b
}

func (b *GcpRedisInstanceBuilder) WithIpRange(ipRangeName string) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *GcpRedisInstanceBuilder) WithRedisTier(redisTier GcpRedisTier) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.RedisTier = redisTier
	return b
}

func (b *GcpRedisInstanceBuilder) WithRedisVersion(redisVersion string) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.RedisVersion = redisVersion
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthEnabled(authEnabled bool) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.AuthEnabled = authEnabled
	return b
}

func (b *GcpRedisInstanceBuilder) WithRedisConfigs(redisConfigs map[string]string) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.RedisConfigs = redisConfigs
	return b
}

func (b *GcpRedisInstanceBuilder) WithMaintenancePolicy(maintenancePolicy *MaintenancePolicy) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.MaintenancePolicy = maintenancePolicy
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthSecret(name string, labels, annotations, extraData map[string]string) *GcpRedisInstanceBuilder {
	b.GcpRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{
		Name:        name,
		Labels:      labels,
		Annotations: annotations,
		ExtraData:   extraData,
	}
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthSecretName(name string) *GcpRedisInstanceBuilder {
	if b.GcpRedisInstance.Spec.AuthSecret == nil {
		b.GcpRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisInstance.Spec.AuthSecret.Name = name
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthSecretLabels(labels map[string]string) *GcpRedisInstanceBuilder {
	if b.GcpRedisInstance.Spec.AuthSecret == nil {
		b.GcpRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisInstance.Spec.AuthSecret.Labels = labels
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthSecretAnnotations(annotations map[string]string) *GcpRedisInstanceBuilder {
	if b.GcpRedisInstance.Spec.AuthSecret == nil {
		b.GcpRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisInstance.Spec.AuthSecret.Annotations = annotations
	return b
}

func (b *GcpRedisInstanceBuilder) WithAuthSecretExtraData(extraData map[string]string) *GcpRedisInstanceBuilder {
	if b.GcpRedisInstance.Spec.AuthSecret == nil {
		b.GcpRedisInstance.Spec.AuthSecret = &RedisAuthSecretSpec{}
	}
	b.GcpRedisInstance.Spec.AuthSecret.ExtraData = extraData
	return b
}
