package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateAwsVpcPeering(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AwsVpcPeering, opts ...ObjAction) error {
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)
	return err
}

func WithAwsRemoteRegion(remoteRegion string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AwsVpcPeering)
			x.Spec.RemoteRegion = remoteRegion
		},
	}
}

func WithAwsRemoteAccountId(remoteAccountId string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AwsVpcPeering)
			x.Spec.RemoteAccountId = remoteAccountId
		},
	}
}

func WithAwsRemoteVpcId(remoteVpcId string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.AwsVpcPeering)
			x.Spec.RemoteVpcId = remoteVpcId
		},
	}
}

func AssertAwsVpcPeeringHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.AwsVpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not  AwsVpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the AwsVpcPeering ID not set")
		}
		return nil
	}
}
