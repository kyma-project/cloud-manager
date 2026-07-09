package dsl

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateKcpAzureManagedRedis(ctx context.Context, clnt client.Client, obj *cloudcontrolv1beta1.AzureManagedRedis, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudcontrolv1beta1.AzureManagedRedis{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultKcpNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the KCP AzureManagedRedis must have name set")
	}

	err := clnt.Create(ctx, obj)
	return err
}

func WithKcpAzureManagedRedisSKU(sku armredisenterprise.SKUName) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.SKU = string(sku)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureManagedRedisSKU", obj))
		},
	}
}

func WithKcpAzureManagedRedisClusteringPolicy(policy armredisenterprise.ClusteringPolicy) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.ClusteringPolicy = string(policy)
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureManagedRedisClusteringPolicy", obj))
		},
	}
}

func WithKcpAzureManagedRedisHighAvailability(ha bool) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.HighAvailability = ha
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureManagedRedisHighAvailability", obj))
		},
	}
}

func WithKcpAzureManagedRedisVpcNetwork(vpcNetworkName string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.VpcNetwork = corev1.LocalObjectReference{Name: vpcNetworkName}
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureManagedRedisVpcNetwork", obj))
		},
	}
}
