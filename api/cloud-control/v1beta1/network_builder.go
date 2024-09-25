package v1beta1

func NewNetworkBuilder() *NetworkBuilder {
	return &NetworkBuilder{}
}

// NetworkBuilder **on purpose** if Build() called several times will always return one the same instance of the
// Network that might be mutated by calling builder methods in between builds. After the build the Network
// instance is not reset and all set attributes are still effective. If you need to build separate instances
// of Network call DeepCopy() or instantiate a new NetworkBuilder.
type NetworkBuilder struct {
	Net Network
}

func (b *NetworkBuilder) WithScope(s string) *NetworkBuilder {
	b.Net.Spec.Scope.Name = s
	return b
}

func (b *NetworkBuilder) WithType(t NetworkType) *NetworkBuilder {
	b.Net.Spec.Type = t
	return b
}

func (b *NetworkBuilder) WithCidr(cidr string) *NetworkBuilder {
	if cidr == "" {
		b.Net.Spec.Network.Managed = nil
		return b
	}

	if b.Net.Spec.Network.Managed == nil {
		b.Net.Spec.Network.Managed = &ManagedNetwork{}
	}
	b.Net.Spec.Network.Managed.Cidr = cidr
	return b
}

func (b *NetworkBuilder) WithLocation(location string) *NetworkBuilder {
	if location == "" {
		b.Net.Spec.Network.Managed = nil
		return b
	}

	if b.Net.Spec.Network.Managed == nil {
		b.Net.Spec.Network.Managed = &ManagedNetwork{}
	}
	b.Net.Spec.Network.Managed.Location = location
	return b
}

func (b *NetworkBuilder) WithManagedNetwork() *NetworkBuilder {
	b.Net.Spec.Network.Reference = nil
	b.Net.Spec.Network.Managed = &ManagedNetwork{}
	return b
}

func (b *NetworkBuilder) WithGcpRef(project, name string) *NetworkBuilder {
	if project == "" && name == "" {
		if b.Net.Spec.Network.Reference == nil {
			return b
		}
		b.Net.Spec.Network.Reference.Gcp = nil
		return b
	}
	if b.Net.Spec.Network.Reference == nil {
		b.Net.Spec.Network.Reference = &NetworkReference{}
	}
	b.Net.Spec.Network.Reference.Gcp = &GcpNetworkReference{
		GcpProject:  project,
		NetworkName: name,
	}
	return b
}

func (b *NetworkBuilder) WithAzureRef(tenant, sub, rg, name string) *NetworkBuilder {
	if tenant == "" && sub == "" && rg == "" && name == "" {
		if b.Net.Spec.Network.Reference == nil {
			return b
		}
		b.Net.Spec.Network.Reference.Azure = nil
		return b
	}
	if b.Net.Spec.Network.Reference == nil {
		b.Net.Spec.Network.Reference = &NetworkReference{}
	}
	b.Net.Spec.Network.Reference.Azure = &AzureNetworkReference{
		TenantId:       tenant,
		SubscriptionId: sub,
		ResourceGroup:  rg,
		NetworkName:    name,
	}
	return b
}

func (b *NetworkBuilder) WithAwsRef(account, region, id, name string) *NetworkBuilder {
	if account == "" && region == "" && id == "" && name == "" {
		if b.Net.Spec.Network.Reference == nil {
			return b
		}
		b.Net.Spec.Network.Reference.Aws = nil
		return b
	}
	if b.Net.Spec.Network.Reference == nil {
		b.Net.Spec.Network.Reference = &NetworkReference{}
	}
	b.Net.Spec.Network.Reference.Aws = &AwsNetworkReference{
		AwsAccountId: account,
		Region:       region,
		VpcId:        id,
		NetworkName:  name,
	}
	return b
}

func (b *NetworkBuilder) WithOpenStackRef(domain, project, id, name string) *NetworkBuilder {
	if domain == "" && project == "" && id == "" && name == "" {
		if b.Net.Spec.Network.Reference == nil {
			return b
		}
		b.Net.Spec.Network.Reference.OpenStack = nil
		return b
	}
	if b.Net.Spec.Network.Reference == nil {
		b.Net.Spec.Network.Reference = &NetworkReference{}
	}
	b.Net.Spec.Network.Reference.OpenStack = &OpenStackNetworkReference{
		Domain:      domain,
		Project:     project,
		NetworkId:   id,
		NetworkName: name,
	}
	return b
}

func (b *NetworkBuilder) WithName(name string) *NetworkBuilder {
	b.Net.Name = name
	return b
}

func (b *NetworkBuilder) WithNamespace(ns string) *NetworkBuilder {
	b.Net.Namespace = ns
	return b
}

func (b *NetworkBuilder) Build() *Network {
	return &b.Net
}
