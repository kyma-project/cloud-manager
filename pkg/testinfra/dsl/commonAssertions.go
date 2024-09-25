package dsl

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingDeletionTimestamp() ObjAssertion {
	return func(obj client.Object) error {
		if obj.GetDeletionTimestamp().IsZero() {
			return fmt.Errorf(
				"Expected object %T %s/%s to have deletion timestamp set, but it doesnt have it",
				obj,
				obj.GetNamespace(), obj.GetName(),
			)
		}
		return nil
	}
}

func WithRemoteRef(name string) ObjAction {
	remoteRef := &cloudcontrolv1beta1.RemoteRef{
		Name:      name,
		Namespace: DefaultSkrNamespace,
	}
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				x.Spec.RemoteRef = *remoteRef
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				x.Spec.RemoteRef = *remoteRef
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.VpcPeering); ok {
				x.Spec.RemoteRef = *remoteRef
				return
			}
			panic("unhandled type in WithRemoteRef")
		},
	}
}
