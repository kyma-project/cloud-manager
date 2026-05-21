package v1beta1

import "k8s.io/utils/ptr"

// +kubebuilder:object:generate=false

type VpcNetworkBuilder struct {
	CommonObjBuilder[*VpcNetworkBuilder, *VpcNetwork]
}

func NewVpcNetworkBuilder(in ...*VpcNetwork) *VpcNetworkBuilder {
	var obj *VpcNetwork
	if len(in) > 0 {
		obj = in[0]
	} else {
		obj = &VpcNetwork{}
	}
	b := &VpcNetworkBuilder{
		CommonObjBuilder[*VpcNetworkBuilder, *VpcNetwork]{
			Obj: obj,
		},
	}
	b.builder = b
	return b
}

func (b *VpcNetworkBuilder) Build() *VpcNetwork {
	return b.Obj
}

func (b *VpcNetworkBuilder) WithObj(obj *VpcNetwork) *VpcNetworkBuilder {
	b.Obj = obj
	return b
}

func (b *VpcNetworkBuilder) WithName(v string) *VpcNetworkBuilder {
	b.Obj.Name = v
	return b
}

func (b *VpcNetworkBuilder) WithNamespace(v string) *VpcNetworkBuilder {
	b.Obj.Namespace = v
	return b
}

func (b *VpcNetworkBuilder) WithAnnotation(k, v string) *VpcNetworkBuilder {
	if b.Obj.Annotations == nil {
		b.Obj.Annotations = map[string]string{}
	}
	b.Obj.Annotations[k] = v
	return b
}

func (b *VpcNetworkBuilder) WithLabel(k, v string) *VpcNetworkBuilder {
	if b.Obj.Labels == nil {
		b.Obj.Labels = map[string]string{}
	}
	b.Obj.Labels[k] = v
	return b
}

func (b *VpcNetworkBuilder) WithType(t VpcNetworkType) *VpcNetworkBuilder {
	b.Obj.Spec.Type = t
	return b
}

func (b *VpcNetworkBuilder) WithSubscription(v string) *VpcNetworkBuilder {
	b.Obj.Spec.Subscription = v
	return b
}

func (b *VpcNetworkBuilder) WithRegion(v string) *VpcNetworkBuilder {
	b.Obj.Spec.Region = v
	return b
}

func (b *VpcNetworkBuilder) WithCidrBlocks(cidrs ...string) *VpcNetworkBuilder {
	b.Obj.Spec.CidrBlocks = append(b.Obj.Spec.CidrBlocks, cidrs...)
	return b
}

func (b *VpcNetworkBuilder) WithVpcNetworkName(v *string) *VpcNetworkBuilder {
	if v == nil {
		b.Obj.Spec.VpcNetworkName = nil
	} else {
		vv := ptr.Deref(v, "")
		if vv == "" {
			b.Obj.Spec.VpcNetworkName = nil
		} else {
			b.Obj.Spec.VpcNetworkName = &vv
		}
	}
	return b
}
