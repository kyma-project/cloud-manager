package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKcpIpRange(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.IpRange, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.IpRange{}
	}
	acts := NewObjActions(opts...).
		Append(WithNamespace(DefaultKcpNamespace))

	acts.ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP IpRange must have name set")
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

func WithKcpIpRangeSpecCidr(cidr string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.IpRange)
			if x.Spec.Cidr == "" {
				x.Spec.Cidr = cidr
			}
		},
	}
}

func WithKcpIpRangeStatusCidr(cidr string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.IpRange)
			if x.Status.Cidr == "" {
				x.Status.Cidr = cidr
			}
		},
	}
}

func WithKcpIpRangeRemoteRef(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.IpRange)
			if x.Spec.RemoteRef.Name == "" {
				x.Spec.RemoteRef.Name = name
			}
		},
	}
}

func WithKcpIpRangeNetwork(network string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.IpRange)
			if x.Spec.Network == nil {
				x.Spec.Network = &klog.ObjectRef{}
			}
			if x.Spec.Network.Name == "" {
				x.Spec.Network.Name = network
			}
		},
	}
}

func HavingKcpIpRangeStatusCidr(cidr string) ObjAssertion {
	return func(obj client.Object) error {
		if x, ok := obj.(*cloudcontrolv1beta1.IpRange); ok {
			if x.Status.Cidr == cidr {
				return nil
			}
			return fmt.Errorf("the KCP IpRange expected status cidr %s, but it has %s", cidr, x.Status.Cidr)
		}
		return fmt.Errorf("unhandled type %T", obj)
	}
}
