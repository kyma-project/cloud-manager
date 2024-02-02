package reload

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reloader interface {
	ReloadObjKind(ctx context.Context, obj client.Object) error
	ReloadObjKindOneKey(ctx context.Context, obj client.Object, key types.NamespacedName) error
	ReloadGvk(ctx context.Context, gvk schema.GroupVersionKind) error
	ReloadGvkOneKey(ctx context.Context, gvk schema.GroupVersionKind, key types.NamespacedName) error
}
