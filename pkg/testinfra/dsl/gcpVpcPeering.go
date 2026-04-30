package dsl

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithGcpRemoteProject(remoteProject string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpVpcPeering)
			x.Spec.RemoteProject = remoteProject
		},
	}
}

func WithGcpRemoteVpc(remoteVpc string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpVpcPeering)
			x.Spec.RemoteVpc = remoteVpc
		},
	}
}

func WithGcpPeeringName(peeringName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpVpcPeering)
			x.Spec.RemotePeeringName = peeringName
		},
	}
}

func WithImportCustomRoute(importCustomRoute bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.GcpVpcPeering)
			x.Spec.ImportCustomRoutes = importCustomRoute
		},
	}
}
