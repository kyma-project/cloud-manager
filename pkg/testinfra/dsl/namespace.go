package dsl

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultSkrNamespace = "test"
	DefaultKcpNamespace = "kcp-system"
)

func CreateNamespace(ctx context.Context, clnt client.Client, obj *corev1.Namespace, opts ...ObjAction) error {
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
