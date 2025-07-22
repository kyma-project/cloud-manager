package v1beta1

type SubscriptionBuilder struct {
	Obj Subscription
}

func NewSubscriptionBuilder() *SubscriptionBuilder {
	return &SubscriptionBuilder{Obj: Subscription{}}
}

func (b *SubscriptionBuilder) WithName(v string) *SubscriptionBuilder {
	b.Obj.Name = v
	return b
}

func (b *SubscriptionBuilder) WithNamespace(v string) *SubscriptionBuilder {
	b.Obj.Namespace = v
	return b
}

func (b *SubscriptionBuilder) WithSecretBindingName(v string) *SubscriptionBuilder {
	b.Obj.Spec.SecretBindingName = v
	return b
}

func (b *SubscriptionBuilder) Build() *Subscription {
	return &b.Obj
}
