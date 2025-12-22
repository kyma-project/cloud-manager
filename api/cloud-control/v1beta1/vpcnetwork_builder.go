package v1beta1

type VpcNetworkBuilder struct {
	obj *VpcNetwork
}

func NewVpcNetworkBuilder() *VpcNetworkBuilder {
	return &VpcNetworkBuilder{obj: &VpcNetwork{}}
}

func (b *VpcNetworkBuilder) WithObj(obj *VpcNetwork) *VpcNetworkBuilder {
	b.obj = obj
	return b
}

func (b *VpcNetworkBuilder) WithName(v string) *VpcNetworkBuilder {
	b.obj.Name = v
	return b
}

func (b *VpcNetworkBuilder) WithNamespace(v string) *VpcNetworkBuilder {
	b.obj.Namespace = v
	return b
}

func (b *VpcNetworkBuilder) WithAnnotation(k, v string) *VpcNetworkBuilder {
	if b.obj.Annotations == nil {
		b.obj.Annotations = map[string]string{}
	}
	b.obj.Annotations[k] = v
	return b
}

func (b *VpcNetworkBuilder) WithLabel(k, v string) *VpcNetworkBuilder {
	if b.obj.Labels == nil {
		b.obj.Labels = map[string]string{}
	}
	b.obj.Labels[k] = v
	return b
}

func (b *VpcNetworkBuilder) WithSubscription(v string) *VpcNetworkBuilder {
	b.obj.Spec.Subscription = v
	return b
}

func (b *VpcNetworkBuilder) WithRegion(v string) *VpcNetworkBuilder {
	b.obj.Spec.Region = v
	return b
}

func (b *VpcNetworkBuilder) WithCidrBlock(cidr string) *VpcNetworkBuilder {
	b.obj.Spec.CidrBlocks = append(b.obj.Spec.CidrBlocks, cidr)
	return b
}
