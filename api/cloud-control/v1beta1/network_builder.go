package v1beta1

func NewNetworkBuilder(in ...*Network) *NetworkBuilder {
	var obj *Network
	if len(in) == 0 {
		obj = &Network{}
	} else {
		obj = in[0]
	}
	b := &NetworkBuilder{
		CommonObjBuilder[*NetworkBuilder, *Network]{
			Obj: obj,
		},
	}
	b.builder = b
	return b
}

// +kubebuilder:object:generate=false

type NetworkBuilder struct {
	CommonObjBuilder[*NetworkBuilder, *Network]
}

func (b *NetworkBuilder) WithScope(s string) *NetworkBuilder {
	b.Obj.Spec.Scope.Name = s
	return b
}

func (b *NetworkBuilder) WithType(t NetworkType) *NetworkBuilder {
	b.Obj.Spec.Type = t
	return b
}

func (b *NetworkBuilder) WithCidr(cidr string) *NetworkBuilder {
	if cidr == "" {
		b.Obj.Spec.Network.Managed = nil
		return b
	}

	if b.Obj.Spec.Network.Managed == nil {
		b.Obj.Spec.Network.Managed = &ManagedNetwork{}
	}
	b.Obj.Spec.Network.Managed.Cidr = cidr
	return b
}

func (b *NetworkBuilder) WithLocation(location string) *NetworkBuilder {
	if location == "" {
		b.Obj.Spec.Network.Managed = nil
		return b
	}

	if b.Obj.Spec.Network.Managed == nil {
		b.Obj.Spec.Network.Managed = &ManagedNetwork{}
	}
	b.Obj.Spec.Network.Managed.Location = location
	return b
}

func (b *NetworkBuilder) WithManagedNetwork() *NetworkBuilder {
	b.Obj.Spec.Network.Reference = nil
	b.Obj.Spec.Network.Managed = &ManagedNetwork{}
	return b
}

func (b *NetworkBuilder) WithGcpRef(project, name string) *NetworkBuilder {
	if project == "" && name == "" {
		if b.Obj.Spec.Network.Reference == nil {
			return b
		}
		b.Obj.Spec.Network.Reference.Gcp = nil
		return b
	}
	if b.Obj.Spec.Network.Reference == nil {
		b.Obj.Spec.Network.Reference = &NetworkReference{}
	}
	b.Obj.Spec.Network.Reference.Gcp = &GcpNetworkReference{
		GcpProject:  project,
		NetworkName: name,
	}
	return b
}

func (b *NetworkBuilder) WithAzureRef(tenant, sub, rg, name string) *NetworkBuilder {
	if tenant == "" && sub == "" && rg == "" && name == "" {
		if b.Obj.Spec.Network.Reference == nil {
			return b
		}
		b.Obj.Spec.Network.Reference.Azure = nil
		return b
	}
	if b.Obj.Spec.Network.Reference == nil {
		b.Obj.Spec.Network.Reference = &NetworkReference{}
	}
	b.Obj.Spec.Network.Reference.Azure = &AzureNetworkReference{
		TenantId:       tenant,
		SubscriptionId: sub,
		ResourceGroup:  rg,
		NetworkName:    name,
	}
	return b
}

func (b *NetworkBuilder) WithAwsRef(account, region, id, name string) *NetworkBuilder {
	if account == "" && region == "" && id == "" && name == "" {
		if b.Obj.Spec.Network.Reference == nil {
			return b
		}
		b.Obj.Spec.Network.Reference.Aws = nil
		return b
	}
	if b.Obj.Spec.Network.Reference == nil {
		b.Obj.Spec.Network.Reference = &NetworkReference{}
	}
	b.Obj.Spec.Network.Reference.Aws = &AwsNetworkReference{
		AwsAccountId: account,
		Region:       region,
		VpcId:        id,
		NetworkName:  name,
	}
	return b
}

func (b *NetworkBuilder) WithOpenStackRef(domain, project, id, name string) *NetworkBuilder {
	if domain == "" && project == "" && id == "" && name == "" {
		if b.Obj.Spec.Network.Reference == nil {
			return b
		}
		b.Obj.Spec.Network.Reference.OpenStack = nil
		return b
	}
	if b.Obj.Spec.Network.Reference == nil {
		b.Obj.Spec.Network.Reference = &NetworkReference{}
	}
	b.Obj.Spec.Network.Reference.OpenStack = &OpenStackNetworkReference{
		Domain:      domain,
		Project:     project,
		NetworkId:   id,
		NetworkName: name,
	}
	return b
}

func (b *NetworkBuilder) WithAlicloudRef(accountId, region, vpcId, networkName string) *NetworkBuilder {
	if accountId == "" && region == "" && vpcId == "" && networkName == "" {
		if b.Obj.Spec.Network.Reference == nil {
			return b
		}
		b.Obj.Spec.Network.Reference.Alicloud = nil
		return b
	}
	if b.Obj.Spec.Network.Reference == nil {
		b.Obj.Spec.Network.Reference = &NetworkReference{}
	}
	b.Obj.Spec.Network.Reference.Alicloud = &AlicloudNetworkReference{
		AccountId:   accountId,
		Region:      region,
		VpcId:       vpcId,
		NetworkName: networkName,
	}
	return b
}
