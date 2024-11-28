package dsl

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithNetworkStatusNetwork(networkReference *cloudcontrolv1beta1.NetworkReference) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if network, ok := obj.(*cloudcontrolv1beta1.Network); ok {
				network.Status.Network = networkReference
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithNetworkStatusNetwork", obj))
		},
	}
}
