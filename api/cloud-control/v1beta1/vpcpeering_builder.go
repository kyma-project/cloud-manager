package v1beta1

// VpcPeeringBuilder each call to Build() returns the same VpcPeering instance and the builder
// allows you to create invalid objects not implementing the VpcPeering CEL validation rules
type VpcPeeringBuilder struct {
	Obj VpcPeering
}

func (b *VpcPeeringBuilder) WithScope(s string) *VpcPeeringBuilder {
	b.Obj.Spec.Scope.Name = s
	return b
}

func (b *VpcPeeringBuilder) WithRemoteRef(ns, n string) *VpcPeeringBuilder {
	b.Obj.Spec.RemoteRef.Namespace = ns
	b.Obj.Spec.RemoteRef.Name = n
	return b
}

func (b *VpcPeeringBuilder) WithGcpPeering(remotePeeringName, remoteProject, remoteVpc string, importCustomRoutes bool) *VpcPeeringBuilder {
	if remotePeeringName == "" {
		if b.Obj.Spec.VpcPeering == nil {
			return b
		}
		b.Obj.Spec.VpcPeering.Gcp = nil
		return b
	}
	if b.Obj.Spec.VpcPeering == nil {
		b.Obj.Spec.VpcPeering = &VpcPeeringInfo{}
	}
	if b.Obj.Spec.VpcPeering.Gcp == nil {
		b.Obj.Spec.VpcPeering.Gcp = &GcpVpcPeering{}
	}
	b.Obj.Spec.VpcPeering.Gcp.RemotePeeringName = remotePeeringName
	b.Obj.Spec.VpcPeering.Gcp.RemoteProject = remoteProject
	b.Obj.Spec.VpcPeering.Gcp.RemoteVpc = remoteVpc
	b.Obj.Spec.VpcPeering.Gcp.ImportCustomRoutes = importCustomRoutes
	return b
}

func (b *VpcPeeringBuilder) WithAzurePeering(remotePeeringName, remoteVNet, remoteResourceGroup string) *VpcPeeringBuilder {
	if remotePeeringName == "" {
		if b.Obj.Spec.VpcPeering == nil {
			return b
		}
		b.Obj.Spec.VpcPeering.Azure = nil
		return b
	}
	if b.Obj.Spec.VpcPeering == nil {
		b.Obj.Spec.VpcPeering = &VpcPeeringInfo{}
	}
	if b.Obj.Spec.VpcPeering.Azure == nil {
		b.Obj.Spec.VpcPeering.Azure = &AzureVpcPeering{}
	}
	b.Obj.Spec.VpcPeering.Azure.RemotePeeringName = remotePeeringName
	b.Obj.Spec.VpcPeering.Azure.RemoteVnet = remoteVNet
	b.Obj.Spec.VpcPeering.Azure.RemoteResourceGroup = remoteResourceGroup
	return b
}

func (b *VpcPeeringBuilder) WithAwsPeering(remoteVpcId, remoteRegion, remoteAccountId string) *VpcPeeringBuilder {
	if remoteVpcId == "" {
		if b.Obj.Spec.VpcPeering == nil {
			return b
		}
		b.Obj.Spec.VpcPeering.Aws = nil
		return b
	}
	if b.Obj.Spec.VpcPeering == nil {
		b.Obj.Spec.VpcPeering = &VpcPeeringInfo{}
	}
	if b.Obj.Spec.VpcPeering.Aws == nil {
		b.Obj.Spec.VpcPeering.Aws = &AwsVpcPeering{}
	}
	b.Obj.Spec.VpcPeering.Aws.RemoteVpcId = remoteVpcId
	b.Obj.Spec.VpcPeering.Aws.RemoteRegion = remoteRegion
	b.Obj.Spec.VpcPeering.Aws.RemoteAccountId = remoteAccountId
	return b
}

func (b *VpcPeeringBuilder) WithDetails(localName, localNamespace, remoteName, remoteNamespace, peeringName string, importCustomRoutes bool) *VpcPeeringBuilder {
	if localName == "" {
		if b.Obj.Spec.Details == nil {
			return b
		}
		b.Obj.Spec.Details = nil
		return b
	}
	if b.Obj.Spec.Details == nil {
		b.Obj.Spec.Details = &VpcPeeringDetails{}
	}
	b.Obj.Spec.Details.LocalNetwork.Name = localName
	b.Obj.Spec.Details.LocalNetwork.Namespace = localNamespace
	b.Obj.Spec.Details.RemoteNetwork.Name = remoteName
	b.Obj.Spec.Details.RemoteNetwork.Namespace = remoteNamespace
	b.Obj.Spec.Details.PeeringName = peeringName
	b.Obj.Spec.Details.ImportCustomRoutes = importCustomRoutes
	return b
}

func (b *VpcPeeringBuilder) Build() *VpcPeering {
	return &b.Obj
}
