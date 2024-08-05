package dsl

import (
	"context"
	"errors"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKcpVpcPeering(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.VpcPeering, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.VpcPeering{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP VpcPeering must have name set")
	}
	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err == nil {
		// already exists
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = clnt.Create(ctx, obj)
	return err
}

func WithKcpVpcPeeringRemoteRef(ns, name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			if x.Spec.RemoteRef.Namespace == "" {
				x.Spec.RemoteRef.Namespace = ns
			}
			if x.Spec.RemoteRef.Name == "" {
				x.Spec.RemoteRef.Name = name
			}
		},
	}
}

func WithKcpVpcPeeringSpecScope(scopeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			if x.Spec.Scope.Name == "" {
				x.Spec.Scope.Name = scopeName
			}
		},
	}
}

func WithKcpVpcPeeringSpecAws(remoteVpcId, remoteAccountId, remoteRegion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			if x.Spec.VpcPeering.Aws == nil {
				x.Spec.VpcPeering.Aws = &cloudcontrolv1beta1.AwsVpcPeering{
					RemoteVpcId:     remoteVpcId,
					RemoteAccountId: remoteAccountId,
					RemoteRegion:    remoteRegion,
				}
			}
		},
	}
}

func WithKcpVpcPeeringSpecAzure(allowVnetAccess bool, remotePeeringName, remoteVnet, remoteResourceGroup string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			if x.Spec.VpcPeering.Azure == nil {
				x.Spec.VpcPeering.Azure = &cloudcontrolv1beta1.AzureVpcPeering{
					AllowVnetAccess:     allowVnetAccess,
					RemotePeeringName:   remotePeeringName,
					RemoteVnet:          remoteVnet,
					RemoteResourceGroup: remoteResourceGroup,
				}
			}
		},
	}
}

func WithKcpVpcPeeringSpecGCP(remoteVpc, remoteProject, peeringName string, importCustomRoutes bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.VpcPeering)
			if x.Spec.VpcPeering.Gcp == nil {
				x.Spec.VpcPeering.Gcp = &cloudcontrolv1beta1.GcpVpcPeering{
					RemoteVpc:          remoteVpc,
					RemoteProject:      remoteProject,
					PeeringName:        peeringName,
					ImportCustomRoutes: importCustomRoutes,
				}
			}
		},
	}
}
