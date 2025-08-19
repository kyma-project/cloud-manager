package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureVpcPeering(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureVpcPeering, opts ...ObjAction) error {
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}

func WithAzureRemotePeeringName(remotePeeringName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
			x.Spec.RemotePeeringName = remotePeeringName
		},
	}
}

func WithAzureRemoteVnet(remoteVnet string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
			x.Spec.RemoteVnet = remoteVnet
		},
	}
}

func AssertAzureVpcPeeringHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureVpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not  AzureVpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the AzureVpcPeering ID not set")
		}
		return nil
	}
}
