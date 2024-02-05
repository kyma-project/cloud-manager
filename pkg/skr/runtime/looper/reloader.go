package looper

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/registry"
	"github.com/kyma-project/cloud-manager/pkg/skr/runtime/reload"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reload.Reloader = &reloader{}

type reloader struct {
	skrScheme   *runtime.Scheme
	descriptors []*registry.Descriptor
}

func (r *reloader) ReloadObjKind(ctx context.Context, obj client.Object) error {
	kinds, _, err := r.skrScheme.ObjectKinds(obj)
	if err != nil || len(kinds) == 0 {
		return fmt.Errorf("object %T not in SKR scheme: %w", obj, err)
	}
	gvk := kinds[0]
	return r.ReloadGvk(ctx, gvk)
}

func (r *reloader) ReloadObjKindOneKey(ctx context.Context, obj client.Object, key types.NamespacedName) error {
	kinds, _, err := r.skrScheme.ObjectKinds(obj)
	if err != nil || len(kinds) == 0 {
		return fmt.Errorf("object %T not in SKR scheme: %w", obj, err)
	}
	gvk := kinds[0]
	return r.ReloadGvkOneKey(ctx, gvk, key)
}

func (r *reloader) ReloadGvk(ctx context.Context, gvk schema.GroupVersionKind) error {
	for _, descr := range r.descriptors {
		for _, w := range descr.Watches {
			if w.Src.ObjGVK() == gvk {
				if err := w.Src.LoadAll(ctx); err != nil {
					return fmt.Errorf("error reloading source %s in descriptor %s: %w", w.Src, descr.Name, err)
				}
			}
		}
	}
	return nil
}

func (r *reloader) ReloadGvkOneKey(ctx context.Context, gvk schema.GroupVersionKind, key types.NamespacedName) error {
	for _, descr := range r.descriptors {
		for _, w := range descr.Watches {
			if w.Src.ObjGVK() == gvk {
				if err := w.Src.LoadOne(ctx, key); err != nil {
					return fmt.Errorf("error reloading key %s on source %s in descriptor %s: %w", key, w.Src, descr.Name, err)
				}
			}
		}
	}
	return nil
}
