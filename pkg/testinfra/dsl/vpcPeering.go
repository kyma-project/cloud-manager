package dsl

import (
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingVpcPeeringStatusId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.VpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not KCP VpcPeering", obj)
		}
		if x.Status.Id == "" {
			return errors.New("the KCP VpcPeering .status.id is not set")
		}
		return nil
	}
}

func HavingVpcPeeringStatusRemoteId() ObjAssertion {
	return func(obj client.Object) error {
		x, ok := obj.(*cloudcontrolv1beta1.VpcPeering)
		if !ok {
			return fmt.Errorf("the object %T is not KCP VpcPeering", obj)
		}
		if x.Status.RemoteId == "" {
			return errors.New("the KCP VpcPeering .status.remoteId not set")
		}
		return nil
	}
}
