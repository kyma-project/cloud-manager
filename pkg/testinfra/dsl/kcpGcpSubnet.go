package dsl

import (
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKcpGcpSubnet(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.GcpSubnet, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.GcpSubnet{}
	}
	acts := NewObjActions(opts...).
		Append(WithNamespace(DefaultKcpNamespace))

	acts.ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP GcpSubnet must have name set")
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

func WithKcpGcpSubnetSpecCidr(cidr string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.GcpSubnet)
			if x.Spec.Cidr == "" {
				x.Spec.Cidr = cidr
			}
		},
	}
}

func WithKcpGcpSubnetPurposePrivate() ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.GcpSubnet)
			x.Spec.Purpose = cloudcontrolv1beta1.GcpSubnetPurpose_PRIVATE
		},
	}
}

func WithKcpGcpSubnetStatusCidr(cidr string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.GcpSubnet)
			if x.Status.Cidr == "" {
				x.Status.Cidr = cidr
			}
		},
	}
}

func WithKcpGcpSubnetRemoteRef(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.GcpSubnet)
			if x.Spec.RemoteRef.Name == "" {
				x.Spec.RemoteRef.Name = name
			}
		},
	}
}

func WithKcpGcpSubnetNetwork(network string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			x := obj.(*cloudcontrolv1beta1.GcpSubnet)
			if x.Spec.Network == nil {
				x.Spec.Network = &klog.ObjectRef{}
			}
			if x.Spec.Network.Name == "" {
				x.Spec.Network.Name = network
			}
		},
	}
}
