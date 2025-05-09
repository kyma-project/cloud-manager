package v1beta1

type AzureVNetLinkBuilder struct {
	Obj AzureVNetLink
}

func (b *AzureVNetLinkBuilder) WithScope(s string) *AzureVNetLinkBuilder {
	b.Obj.Spec.Scope.Name = s
	return b
}

func (b *AzureVNetLinkBuilder) WithRemoteVirtualPrivateLinkName(name string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemoteVirtualPrivateLinkName = name
	return b
}

func (b *AzureVNetLinkBuilder) WithRemotePrivateDnsZone(remotePrivateDnsZone string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemotePrivateDnsZone = remotePrivateDnsZone
	return b
}

func (b *AzureVNetLinkBuilder) WithRemoteTenant(remoteTenant string) *AzureVNetLinkBuilder {
	b.Obj.Spec.RemoteTenant = remoteTenant
	return b
}

func (b *AzureVNetLinkBuilder) Build() *AzureVNetLink {
	return &b.Obj
}
