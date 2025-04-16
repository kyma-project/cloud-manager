package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateSkrGcpSubnet(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.GcpSubnet, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.GcpSubnet{}
	}
	NewObjActions(opts...).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR GcpSubnet must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithSkrGcpSubnetCidr(cidr string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if gcpSubnet, ok := obj.(*cloudresourcesv1beta1.GcpSubnet); ok {
				gcpSubnet.Spec.Cidr = cidr
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrGcpSubnetCidr", obj))
		},
	}
}

func HavingGcpSubnetStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpSubnet)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpSubnet", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR GcpSubnet ID not set")
		}
		return nil
	}
}

func HavingGcpSubnetStatusState(state string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.GcpSubnet)
		if !ok {
			return fmt.Errorf("the object %T is not SKR GcpSubnet", obj)
		}
		if x.Status.State != state {
			return fmt.Errorf("the SKR GcpSubnet State does not match. expected: %s, got: %s", state, x.Status.State)
		}
		return nil
	}
}
