package dsl

import (
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudresourcesv1beta1.GcpRedisInstance); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithIpRange", obj))
		},
	}
}
