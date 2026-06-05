package dsl

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateSkrAzureManagedRedis creates an SKR AzureManagedRedis resource using the
// shared object-action DSL. Caller must apply WithName(...) and WithSkrAzureManagedRedisTier(...).
func CreateSkrAzureManagedRedis(ctx context.Context, clnt client.Client, obj *cloudresourcesv1beta1.AzureManagedRedis, opts ...ObjAction) error {
	if obj == nil {
		obj = &cloudresourcesv1beta1.AzureManagedRedis{}
	}
	NewObjActions(opts...).
		Append(
			WithNamespace(DefaultSkrNamespace),
		).
		ApplyOnObject(obj)

	if obj.Name == "" {
		return errors.New("the SKR AzureManagedRedis must have name set")
	}
	if obj.Spec.RedisTier == "" {
		return errors.New("the SKR AzureManagedRedis must have spec.redisTier set")
	}

	return clnt.Create(ctx, obj)
}

// WithSkrAzureManagedRedisTier sets the Kyma service tier (S1-S5, P1-P5, C3-C7).
func WithSkrAzureManagedRedisTier(tier cloudresourcesv1beta1.AzureManagedRedisTier) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudresourcesv1beta1.AzureManagedRedis); ok {
				amr.Spec.RedisTier = tier
				return
			}
			panic(fmt.Errorf("unhandled type %T in WithSkrAzureManagedRedisTier", obj))
		},
	}
}

// WithKcpAzureManagedRedisStatusPrimaryEndpoint sets KCP AzureManagedRedis Status.PrimaryEndpoint
// (used by SKR-side tests to fake KCP reconciliation completion).
func WithKcpAzureManagedRedisStatusPrimaryEndpoint(primaryEndpoint string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Status.PrimaryEndpoint = primaryEndpoint
			}
		},
	}
}

// WithKcpAzureManagedRedisStatusPort sets KCP AzureManagedRedis Status.Port.
func WithKcpAzureManagedRedisStatusPort(port int32) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Status.Port = port
			}
		},
	}
}

// WithKcpAzureManagedRedisStatusAuthString sets KCP AzureManagedRedis Status.AuthString.
func WithKcpAzureManagedRedisStatusAuthString(authString string) ObjStatusAction {
	return &objStatusAction{
		f: func(obj client.Object) {
			if amr, ok := obj.(*cloudcontrolv1beta1.AzureManagedRedis); ok {
				amr.Status.AuthString = authString
			}
		},
	}
}
