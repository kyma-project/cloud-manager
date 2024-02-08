package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultIpRangeCidr = "10.181.0.0/16"
)

func WithSkrIpRangeSpecCidr(cidr string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.IpRange)
			if x.Spec.Cidr == "" {
				x.Spec.Cidr = cidr
			}
		},
	}
}

func WithSkrIpRangeStatusCidr(cidr string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudresourcesv1beta1.IpRange)
			if x.Status.Cidr == "" {
				x.Status.Cidr = cidr
			}
		},
	}
}

func WithSkrIpRangeStatusId(id string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.IpRange); ok {
				if x.Status.Id == "" {
					x.Status.Id = id
				}
			}
		},
	}
}

func CreateSkrIpRange(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.IpRange, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.IpRange{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
			WithSkrIpRangeSpecCidr(DefaultIpRangeCidr),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR IpRange must have name set")
	}
	err := clnt.Create(ctx, obj)
	return err
}

func AssertSkrIpRangeHasId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudresourcesv1beta1.IpRange)
		if !ok {
			return fmt.Errorf("the object %T is not SKR IpRange", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the SKR IpRange ID not set")
		}
		return nil
	}
}
