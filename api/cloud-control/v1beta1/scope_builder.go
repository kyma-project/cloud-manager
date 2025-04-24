package v1beta1

func NewScopeBuilder() *ScopeBuilder {
	return &ScopeBuilder{}
}

type ScopeBuilder struct {
	Scope Scope
}

func (b *ScopeBuilder) WithProvider(provider ProviderType) *ScopeBuilder {
	b.Scope.Spec.Provider = provider
	return b
}

func (b *ScopeBuilder) WithBrokerPlan(brokerPlan string) *ScopeBuilder {
	if b.Scope.Labels == nil {
		b.Scope.Labels = map[string]string{}
	}
	b.Scope.Labels[LabelScopeBrokerPlanName] = brokerPlan
	return b
}

func (b *ScopeBuilder) Build() *Scope {
	return &b.Scope
}
