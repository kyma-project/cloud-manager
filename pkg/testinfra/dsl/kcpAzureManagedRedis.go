package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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

func WithKcpAzureManagedRedisSKU(sku string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.SKU = sku
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithKcpAzureManagedRedisSKU", obj))
		},
	}
}

func WithKcpAzureManagedRedisClusteringPolicy(policy string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Spec.ClusteringPolicy = policy
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
