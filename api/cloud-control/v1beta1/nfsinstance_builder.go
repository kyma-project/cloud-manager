package v1beta1

func NewNfsInstanceBuilder(in ...*NfsInstance) *NfsInstanceBuilder {
	var obj *NfsInstance
	if len(in) > 0 {
		obj = in[0]
	} else {
		obj = &NfsInstance{}
	}
	b := &NfsInstanceBuilder{
		CommonObjBuilder[*NfsInstanceBuilder, *NfsInstance]{
			Obj: obj,
		},
	}
	b.builder = b
	return b
}

// +kubebuilder:object:generate=false

type NfsInstanceBuilder struct {
	CommonObjBuilder[*NfsInstanceBuilder, *NfsInstance]
}

func (b *NfsInstanceBuilder) WithRemoteRef(namespace, name string) *NfsInstanceBuilder {
	b.Obj.Spec.RemoteRef.Namespace = namespace
	b.Obj.Spec.RemoteRef.Name = name
	return b
}

func (b *NfsInstanceBuilder) WithIpRange(ipRangeName string) *NfsInstanceBuilder {
	b.Obj.Spec.IpRange.Name = ipRangeName
	return b
}

func (b *NfsInstanceBuilder) WithScope(s string) *NfsInstanceBuilder {
	b.Obj.Spec.Scope.Name = s
	return b
}

func (b *NfsInstanceBuilder) WithGcp(capacityGb int, location string, tier GcpFileTier) *NfsInstanceBuilder {
	b.Obj.Spec.Instance.OpenStack = nil
	b.Obj.Spec.Instance.Aws = nil
	b.Obj.Spec.Instance.Azure = nil
	b.Obj.Spec.Instance.Gcp = &NfsInstanceGcp{
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
	b.Obj.Spec.Instance.OpenStack = nil
	b.Obj.Spec.Instance.Azure = nil
	b.Obj.Spec.Instance.Gcp = nil
	b.Obj.Spec.Instance.Aws = &NfsInstanceAws{
		PerformanceMode: performanceMode,
		Throughput:      throughput,
	}
	return b
}

func (b *NfsInstanceBuilder) WithAwsDummyDefaults() *NfsInstanceBuilder {
	return b.WithAws(AwsPerformanceModeBursting, AwsThroughputModeBursting)
}

func (b *NfsInstanceBuilder) WithOpenStack(sizeGb int) *NfsInstanceBuilder {
	b.Obj.Spec.Instance.Azure = nil
	b.Obj.Spec.Instance.Gcp = nil
	b.Obj.Spec.Instance.Aws = nil
	b.Obj.Spec.Instance.OpenStack = &NfsInstanceOpenStack{
		SizeGb: sizeGb,
	}
	return b
}

func (b *NfsInstanceBuilder) WithOpenStackDummyDefaults() *NfsInstanceBuilder {
	return b.WithOpenStack(1)
}
