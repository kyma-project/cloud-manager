package v1beta1

func NewScopeBuilder(in ...*Scope) *ScopeBuilder {
	var obj *Scope
	if len(in) > 0 {
		obj = in[0]
	} else {
		obj = &Scope{}
	}
	b := &ScopeBuilder{
		CommonObjBuilder[*ScopeBuilder, *Scope]{
			Obj: obj,
		},
	}
	b.builder = b
	return b
}

// +kubebuilder:object:generate=false

type ScopeBuilder struct {
	CommonObjBuilder[*ScopeBuilder, *Scope]
}

func (b *ScopeBuilder) WithProvider(provider ProviderType) *ScopeBuilder {
	b.Obj.Spec.Provider = provider
	return b
}

func (b *ScopeBuilder) WithBrokerPlan(brokerPlan string) *ScopeBuilder {
	if b.Obj.Labels == nil {
		b.Obj.Labels = map[string]string{}
	}
	b.Obj.Labels[LabelScopeBrokerPlanName] = brokerPlan
	return b
}
