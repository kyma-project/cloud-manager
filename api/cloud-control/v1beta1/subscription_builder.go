package v1beta1

// +kubebuilder:object:generate=false

type SubscriptionBuilder struct {
	CommonObjBuilder[*SubscriptionBuilder, *Subscription]
	//Obj *Subscription
}

func NewSubscriptionBuilder(in ...*Subscription) *SubscriptionBuilder {
	var obj *Subscription
	if len(in) == 0 {
		obj = &Subscription{}
	} else {
		obj = in[0]
	}
	b := &SubscriptionBuilder{
		CommonObjBuilder[*SubscriptionBuilder, *Subscription]{
			Obj: obj,
		},
	}
	b.builder = b
	return b
}

func (b *SubscriptionBuilder) WithBindingName(v string) *SubscriptionBuilder {
	b.Obj.Spec.Details.Garden = &SubscriptionGarden{
		BindingName: v,
	}
	return b
}

func (b *SubscriptionBuilder) WithAws(accountId string) *SubscriptionBuilder {
	b.Obj.Spec.Details.Aws = &SubscriptionInfoAws{
		Account: accountId,
	}
	return b
}

func (b *SubscriptionBuilder) WithoutAws() *SubscriptionBuilder {
	b.Obj.Spec.Details.Aws = nil
	return b
}

func (b *SubscriptionBuilder) WithAzure(tenantId, subscriptionId string) *SubscriptionBuilder {
	b.Obj.Spec.Details.Azure = &SubscriptionInfoAzure{
		TenantId:       tenantId,
		SubscriptionId: subscriptionId,
	}
	return b
}

func (b *SubscriptionBuilder) WithoutAzure() *SubscriptionBuilder {
	b.Obj.Spec.Details.Azure = nil
	return b
}

func (b *SubscriptionBuilder) WithGcp(projectId string) *SubscriptionBuilder {
	b.Obj.Spec.Details.Gcp = &SubscriptionInfoGcp{
		Project: projectId,
	}
	return b
}

func (b *SubscriptionBuilder) WithoutGcp() *SubscriptionBuilder {
	b.Obj.Spec.Details.Gcp = nil
	return b
}

func (b *SubscriptionBuilder) WithOpenstack(domainName, projectName string) *SubscriptionBuilder {
	b.Obj.Spec.Details.Openstack = &SubscriptionInfoOpenStack{
		DomainName: domainName,
		TenantName: projectName,
	}
	return b
}

func (b *SubscriptionBuilder) WithoutOpenstack() *SubscriptionBuilder {
	b.Obj.Spec.Details.Openstack = nil
	return b
}
