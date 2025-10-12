package fake

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotImplemented = fmt.Errorf("not implemented")

type Cache struct {
	Runnable
	// Wait if provided will be used to block WaitForCacheSync until it is closed.
	Wait <-chan struct{}
	Started bool
	Synced  bool
}

func (f *Cache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if !f.Started {
		return &cache.ErrCacheNotStarted{}
	}
	return nil
}

func (f *Cache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if !f.Started {
		return &cache.ErrCacheNotStarted{}
	}
	return nil
}

func (f *Cache) GetInformer(ctx context.Context, obj client.Object, opts ...cache.InformerGetOption) (cache.Informer, error) {
	return nil, ErrNotImplemented
}

func (f *Cache) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...cache.InformerGetOption) (cache.Informer, error) {
	return nil, ErrNotImplemented
}

func (f *Cache) RemoveInformer(ctx context.Context, obj client.Object) error {
	return ErrNotImplemented
}

func (f *Cache) WaitForCacheSync(ctx context.Context) bool {
	if f.Wait != nil {
		<-f.Wait
	}
	return f.Synced
}

func (f *Cache) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return ErrNotImplemented
}
