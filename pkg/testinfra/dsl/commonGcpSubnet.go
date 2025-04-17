package dsl

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithGcpSubnet(subnetName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			// KCP
			if x, ok := obj.(*cloudcontrolv1beta1.GcpRedisCluster); ok {
				if x.Spec.Subnet.Name == "" {
					x.Spec.Subnet.Name = subnetName
				}
				return
			}

			// SKR
			if x, ok := obj.(*cloudresourcesv1beta1.GcpRedisCluster); ok {
				if x.Spec.Subnet.Name == "" {
					x.Spec.Subnet.Name = subnetName
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithGcpSubnet", obj))
		},
	}
}
