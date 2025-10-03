package v1beta1

type AzureVNetLinkBuilder struct {
	Obj AzureVNetLink
}

func (b *AzureVNetLinkBuilder) WithScope(s string) *AzureVNetLinkBuilder {
	b.Obj.Spec.Scope.Name = s
	return b
}

func (b *AzureVNetLinkBuilder) WithRemoteVirtualPrivateLinkName(name string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemoteVNetLinkName = name
	return b
}

func (b *AzureVNetLinkBuilder) WithRemotePrivateDnsZone(remotePrivateDnsZone string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemotePrivateDnsZone = remotePrivateDnsZone
	return b
}

func (b *AzureVNetLinkBuilder) WithRemoteDnsForwardingRuleset(remoteDnsForwardingRuleset string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemoteDnsForwardingRuleset = remoteDnsForwardingRuleset
	return b
}

func (b *AzureVNetLinkBuilder) WithRemoteTenant(remoteTenant string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemoteTenant = remoteTenant
	return b
}

func (b *AzureVNetLinkBuilder) WithName(name string) *AzureVNetLinkBuilder {
	b.Obj.Name = name
	return b
}

func (b *AzureVNetLinkBuilder) WithNamespace(namespace string) *AzureVNetLinkBuilder {
	b.Obj.Namespace = namespace
	return b
}

func (b *AzureVNetLinkBuilder) WithAnnotations(annotations map[string]string) *AzureVNetLinkBuilder {
	b.Obj.Annotations = annotations
	return b
}

func (b *AzureVNetLinkBuilder) Build() *AzureVNetLink {
	return &b.Obj
}
