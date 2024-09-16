package dsl

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AssertKcpIpRangeScope(expectedScopeName string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.IpRange)
		if !ok {
			return fmt.Errorf("expected *cloudcontrolv1beta1.IpRange, but got %T", obj)
		}
		if x.Spec.Scope.Name != expectedScopeName {
			return fmt.Errorf("the KCP IpRange %s/%s expected scope name is %s, but it has %s",
				x.Namespace, x.Name, expectedScopeName, x.Spec.Scope.Name)
		}
		return nil
	}
}

func AssertKcpIpRangeCidr(expectedCidr string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.IpRange)
		if !ok {
			return fmt.Errorf("expected *cloudcontrolv1beta1.IpRange, but got %T", obj)
		}
		if x.Spec.Cidr != expectedCidr {
			return fmt.Errorf("the KCP IpRange %s/%s expected cidr is %s, but it has %s",
				x.Namespace, x.Name, expectedCidr, x.Spec.Cidr)
		}
		return nil
	}
}

func AssertKcpIpRangeRemoteRef(expectedNamespace, expectedName string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.IpRange)
		if !ok {
			return fmt.Errorf("expected *cloudcontrolv1beta1.IpRange, but got %T", obj)
		}
		if x.Spec.RemoteRef.Namespace != expectedNamespace {
			return fmt.Errorf("the KCP IpRange %s/%s expected remoteRef namespace is %s, but it has %s",
				x.Namespace, x.Name, expectedNamespace, x.Spec.Cidr)
		}
		if x.Spec.RemoteRef.Name != expectedName {
			return fmt.Errorf("the KCP IpRange %s/%s expected remoteRef name is %s, but it has %s",
				x.Namespace, x.Name, expectedNamespace, x.Spec.Cidr)
		}
		return nil
	}
}

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
