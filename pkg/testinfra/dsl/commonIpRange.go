package dsl

import (
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithIpRange(ipRangeName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			// SKR
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
			if x, ok := obj.(*cloudresourcesv1beta1.AwsRedisInstance); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudresourcesv1beta1.AzureRedisInstance); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}

			// KCP
			if x, ok := obj.(*cloudcontrolv1beta1.NfsInstance); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.RedisInstance); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}
			if x, ok := obj.(*cloudcontrolv1beta1.RedisCluster); ok {
				if x.Spec.IpRange.Name == "" {
					x.Spec.IpRange.Name = ipRangeName
				}
				return
			}

			panic(fmt.Errorf("unhandled type %T in WithIpRange", obj))
		},
	}
}
