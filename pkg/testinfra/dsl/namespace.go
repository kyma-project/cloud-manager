package dsl

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/testinfra"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraScheme"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/infraTypes"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultSkrNamespace    = testinfra.DefaultSkrNamespace
	DefaultKcpNamespace    = testinfra.DefaultKcpNamespace
	DefaultGardenNamespace = testinfra.DefaultGardenNamespace
)

func SetDefaultNamespace(obj client.Object) {
	switch infraScheme.ObjToClusterType(obj) {
	case infraTypes.ClusterTypeKcp:
		obj.SetNamespace(DefaultKcpNamespace)
	case infraTypes.ClusterTypeSkr:
		obj.SetNamespace(DefaultSkrNamespace)
	case infraTypes.ClusterTypeGarden:
		obj.SetNamespace(DefaultGardenNamespace)
	}
}

func CreateNamespace(ctx context.Context, clnt client.Client, obj *corev1.Namespace, opts ...ObjAction) error {
	if true {
		return nil
	}
	if obj == nil {
		obj = &corev1.Namespace{}
	}
	NewObjActions(opts...).
		Append(WithName(DefaultSkrNamespace)).
		ApplyOnObject(obj)

	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err == nil {
		// already exists
		return nil
	}
	if client.IgnoreNotFound(err) != nil {
		// some error
		return err
	}
	err = clnt.Create(ctx, obj)
	return client.IgnoreAlreadyExists(err)
}
