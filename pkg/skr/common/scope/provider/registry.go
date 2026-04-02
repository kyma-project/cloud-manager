package provider

import (
	"context"
	"errors"
	"slices"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

var ErrScopeNoMatch = errors.New("no matching scope found")

type scopeNoMatchError string

func (e scopeNoMatchError) Error() string {
	return "no matching scope found"
}

const (
	ErrNoMatchingScope = scopeNoMatchError("no matching scope found")
)

type ScopeProviderRegistry interface {
	ScopeProvider
	Add(provider ScopeProvider) ScopeProviderRegistry
}

var _ ScopeProviderRegistry = (*compositeProvider)(nil)

type compositeProvider struct {
	providers []ScopeProvider
}

func (c *compositeProvider) Add(provider ScopeProvider) ScopeProviderRegistry {
	c.providers = append(c.providers, provider)
	return c
}

func (c *compositeProvider) GetScope(ctx context.Context, objName types.NamespacedName) (klog.ObjectRef, error) {
	for _, provider := range slices.Backward(c.providers) {
		result, err := provider.GetScope(ctx, objName)
		if err == nil || errors.Is(err, ErrNoMatchingScope) {
			return result, nil
		}
	}
	return klog.ObjectRef{}, ErrNoMatchingScope
}

func New() ScopeProviderRegistry {
	return &compositeProvider{}
}
