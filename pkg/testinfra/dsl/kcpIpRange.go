package dsl

import (
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

func AssertKcpIpRangeCidr(exectedCidr string) ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.IpRange)
		if !ok {
			return fmt.Errorf("expected *cloudcontrolv1beta1.IpRange, but got %T", obj)
		}
		if x.Spec.Cidr != exectedCidr {
			return fmt.Errorf("the KCP IpRange %s/%s expected cidr is %s, but it has %s",
				x.Namespace, x.Name, exectedCidr, x.Spec.Cidr)
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
