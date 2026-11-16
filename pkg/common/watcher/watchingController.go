package watcher

import (
	"context"
	"strings"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WatchingController[T client.Object] struct {
	client  client.Client
	handler ObjectHandler
	obj     T
	gvk     schema.GroupVersionKind
}

func (r *WatchingController[T]) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	fetchedObj := r.obj.DeepCopyObject().(T)
	err := r.client.Get(ctx, req.NamespacedName, fetchedObj)
	if k8serrors.IsNotFound(err) {
		return reconcile.Result{}, r.handler.OnObject(ctx, r.gvk, req, nil)
	}
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, r.handler.OnObject(ctx, r.gvk, req, fetchedObj)
}

func (r *WatchingController[T]) SetupWithManager(mgr ctrl.Manager, opts ...builder.ForOption) error {
	var gvkString string
	if u, ok := any(r.obj).(*unstructured.Unstructured); ok {
		gvkString = u.GetKind()
		r.gvk = schema.FromAPIVersionAndKind(u.GetAPIVersion(), u.GetKind())
	} else {
		gvks, _, err := mgr.GetScheme().ObjectKinds(r.obj)
		if err != nil {
			return err
		}
		gvkString = gvks[0].Kind
		r.gvk = gvks[0]
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(r.obj, opts...).
		Named("watcher-" + strings.ToLower(gvkString)).
		Complete(r)
}
