package v1beta1

func NewNfsInstanceBuilder() *NfsInstanceBuilder {
	return &NfsInstanceBuilder{
		NfsInstance: NfsInstance{},
	}
}

type NfsInstanceBuilder struct {
	NfsInstance NfsInstance
}

func (b *NfsInstanceBuilder) Reset() *NfsInstanceBuilder {
	b.NfsInstance = NfsInstance{}
	return b
}

func (b *NfsInstanceBuilder) WithRemoteRef(namespace, name string) *NfsInstanceBuilder {
	b.NfsInstance.Spec.RemoteRef.Namespace = namespace
	b.NfsInstance.Spec.RemoteRef.Name = name
	return b
}

func (b *NfsInstanceBuilder) WithIpRange(ipRangeName string) *NfsInstanceBuilder {
	b.NfsInstance.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *NfsInstanceBuilder) WithScope(s string) *NfsInstanceBuilder {
	b.NfsInstance.Spec.Scope.Name = s
	return b
}

func (b *NfsInstanceBuilder) WithGcp(capacityGb int, location string, tier GcpFileTier) *NfsInstanceBuilder {
	b.NfsInstance.Spec.Instance.OpenStack = nil
	b.NfsInstance.Spec.Instance.Aws = nil
	b.NfsInstance.Spec.Instance.Azure = nil
	b.NfsInstance.Spec.Instance.Gcp = &NfsInstanceGcp{
		Location:      location,
		Tier:          tier,
		CapacityGb:    capacityGb,
		FileShareName: "vol1",
		ConnectMode:   PRIVATE_SERVICE_ACCESS,
	}
	return b
}

func (b *NfsInstanceBuilder) WithGcpDummyDefaults() *NfsInstanceBuilder {
	return b.WithGcp(1024, "us-east", BASIC_HDD)
}

func (b *NfsInstanceBuilder) WithAws(performanceMode AwsPerformanceMode, throughput AwsThroughputMode) *NfsInstanceBuilder {
	b.NfsInstance.Spec.Instance.OpenStack = nil
	b.NfsInstance.Spec.Instance.Azure = nil
	b.NfsInstance.Spec.Instance.Gcp = nil
	b.NfsInstance.Spec.Instance.Aws = &NfsInstanceAws{
		PerformanceMode: performanceMode,
		Throughput:      throughput,
	}
	return b
}

func (b *NfsInstanceBuilder) WithAwsDummyDefaults() *NfsInstanceBuilder {
	return b.WithAws(AwsPerformanceModeBursting, AwsThroughputModeBursting)
}

func (b *NfsInstanceBuilder) WithOpenStack(sizeGb int) *NfsInstanceBuilder {
	b.NfsInstance.Spec.Instance.Azure = nil
	b.NfsInstance.Spec.Instance.Gcp = nil
	b.NfsInstance.Spec.Instance.Aws = nil
	b.NfsInstance.Spec.Instance.OpenStack = &NfsInstanceOpenStack{
		SizeGb: sizeGb,
	}
	return b
}

func (b *NfsInstanceBuilder) WithOpenStackDummyDefaults() *NfsInstanceBuilder {
	return b.WithOpenStack(1)
}

func (b *NfsInstanceBuilder) Build() *NfsInstance {
	return &b.NfsInstance
}
