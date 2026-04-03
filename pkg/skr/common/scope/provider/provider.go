package provider

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScopeProvider interface {
	GetScope(ctx context.Context, req types.NamespacedName) (klog.ObjectRef, error)
}

type ScopeProviderFunc func(ctx context.Context, req types.NamespacedName) (klog.ObjectRef, error)

func (f ScopeProviderFunc) GetScope(ctx context.Context, req types.NamespacedName) (klog.ObjectRef, error) {
	return f(ctx, req)
}

func Always(scopeNamespace, scopeName string) ScopeProvider {
	return ScopeProviderFunc(func(_ context.Context, _ types.NamespacedName) (klog.ObjectRef, error) {
		return klog.ObjectRef{Namespace: scopeNamespace, Name: scopeName}, nil
	})
}

func MatchingObjName(objName, scopeNamespace, scopeName string) ScopeProvider {
	return ScopeProviderFunc(func(ctx context.Context, req types.NamespacedName) (klog.ObjectRef, error) {
		if req.Name == objName {
			return klog.ObjectRef{Namespace: scopeNamespace, Name: scopeName}, nil
		}
		return klog.ObjectRef{}, ErrScopeNoMatch
	})
}

func MatchingObj(objName string, scope client.Object) ScopeProvider {
	return ScopeProviderFunc(func(ctx context.Context, req types.NamespacedName) (klog.ObjectRef, error) {
		if req.Name == objName {
			return klog.ObjectRef{Namespace: scope.GetNamespace(), Name: scope.GetName()}, nil
		}
		return klog.ObjectRef{}, ErrScopeNoMatch
	})
}
