package watcher

import (
	"context"
	"fmt"
	"strings"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Watcher interface {
	For(obj client.Object, opts ...builder.ForOption) error
	ForGV(gv schema.GroupVersion, opts ...builder.ForOption) error
	ForGVK(gvk schema.GroupVersionKind, opts ...builder.ForOption) error
	ForSKR() error
	ForKCP() error
	Emit(ctx context.Context, request reconcile.Request, obj client.Object)
}

type defaultWatcher struct {
	mgr     ctrl.Manager
	handler ObjectHandler
}

func New(mgr ctrl.Manager, handler ObjectHandler) Watcher {
	return &defaultWatcher{mgr: mgr, handler: handler}
}

func (w *defaultWatcher) For(obj client.Object, opts ...builder.ForOption) error {
	c := &WatchingController[client.Object]{
		client:  w.mgr.GetClient(),
		handler: w.handler,
		obj:     obj,
	}
	return c.SetupWithManager(w.mgr, opts...)
}

func (w *defaultWatcher) ForGVK(gvk schema.GroupVersionKind, opts ...builder.ForOption) error {
	obj, err := w.mgr.GetScheme().New(gvk)
	if err != nil {
		return fmt.Errorf("GroupVersionKind %s not registered in scheme: %w", gvk, err)
	}
	clientObj, ok := obj.(client.Object)
	if !ok {
		return nil
	}
	return w.For(clientObj, opts...)
}

func (w *defaultWatcher) ForGV(gv schema.GroupVersion, opts ...builder.ForOption) error {
	knownTypes := w.mgr.GetScheme().KnownTypes(gv)
	if len(knownTypes) == 0 {
		return fmt.Errorf("no types registered for GroupVersion %s", gv)
	}
	found := 0
	for kind := range knownTypes {
		if strings.HasSuffix(kind, "List") || strings.HasSuffix(kind, "Options") {
			continue
		}
		if err := w.ForGVK(gv.WithKind(kind), opts...); err != nil {
			return err
		}
		found++
	}
	if found == 0 {
		return fmt.Errorf("no watchable types found for GroupVersion %s", gv)
	}
	return nil
}

func (w *defaultWatcher) ForSKR() error {
	if err := w.ForGV(schema.GroupVersion{Group: "cloud-resources.kyma-project.io", Version: "v1beta1"}); err != nil {
		return err
	}
	cloudManagedPredicate := builder.WithPredicates(predicate.NewTypedPredicateFuncs(func(obj client.Object) bool {
		return obj.GetLabels()[cloudresourcesv1beta1.LabelCloudManaged] == "true"
	}))
	if err := w.ForGVK(schema.FromAPIVersionAndKind("v1", "PersistentVolume"), cloudManagedPredicate); err != nil {
		return err
	}
	if err := w.ForGVK(schema.FromAPIVersionAndKind("v1", "PersistentVolumeClaim"), cloudManagedPredicate); err != nil {
		return err
	}
	if err := w.ForGVK(schema.FromAPIVersionAndKind("v1", "ConfigMap"), cloudManagedPredicate); err != nil {
		return err
	}
	if err := w.ForGVK(schema.FromAPIVersionAndKind("v1", "Secret"), cloudManagedPredicate); err != nil {
		return err
	}
	return nil
}

func (w *defaultWatcher) ForKCP() error {
	if err := w.ForGV(schema.GroupVersion{Group: "cloud-control.kyma-project.io", Version: "v1beta1"}); err != nil {
		return err
	}
	return nil
}

func (w *defaultWatcher) Emit(ctx context.Context, request reconcile.Request, obj client.Object) {
	var gvk schema.GroupVersionKind
	if obj != nil {
		if u, ok := obj.(*unstructured.Unstructured); ok {
			gvk = schema.FromAPIVersionAndKind(u.GetAPIVersion(), u.GetKind())
		} else {
			gvks, _, err := w.mgr.GetScheme().ObjectKinds(obj)
			if err == nil && len(gvks) > 0 {
				gvk = gvks[0]
			}
		}
	}
	_ = w.handler.OnObject(ctx, gvk, request, obj)
}
