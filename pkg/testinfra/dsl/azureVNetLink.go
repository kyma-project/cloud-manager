package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAzureVNetLink(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureVNetLink, opts ...ObjAction) error {
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}

func WithAzureRemoteVNetLinkName(remoteVNetLinkName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVNetLink)
			x.Spec.RemoteVNetLinkName = remoteVNetLinkName
		},
	}
}

func WithAzureRemotePrivateDnsZone(remotePrivateDnsZone string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AzureVNetLink)
			x.Spec.RemotePrivateDnsZone = remotePrivateDnsZone
		},
	}
}

func AssertAzureVNetPeeringHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AzureVNetLink)
		if !ok {
			return fmt.Errorf("the object %T is not  AzureVNetLink", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the AzureVNetLink ID not set")
		}
		return nil
	}
}
