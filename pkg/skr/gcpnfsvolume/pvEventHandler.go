package gcpnfsvolume

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var PVEventHandler handler.Funcs = handler.Funcs{

	CreateFunc: func(ctx context.Context, e event.CreateEvent, q workqueue.RateLimitingInterface) {
		if match, key := isMatchingPV(e.Object); match {
			q.Add(reconcile.Request{NamespacedName: *key})
		}
	},
	UpdateFunc: func(ctx context.Context, e event.UpdateEvent, q workqueue.RateLimitingInterface) {
		if match, key := isMatchingPV(e.ObjectNew); match {
			q.Add(reconcile.Request{NamespacedName: *key})
		}
	},
	DeleteFunc: func(ctx context.Context, e event.DeleteEvent, q workqueue.RateLimitingInterface) {
		if match, key := isMatchingPV(e.Object); match {
			q.Add(reconcile.Request{NamespacedName: *key})
		}
	},
}

func isMatchingPV(obj client.Object) (bool, *types.NamespacedName) {
	pv := v1.PersistentVolume{}
	if obj.GetObjectKind() != pv.GetObjectKind() {
		return false, nil
	}

	if name, ok := obj.GetLabels()[labelNfsVolName]; ok {
		return true, &types.NamespacedName{
			Name:      name,
			Namespace: obj.GetLabels()[labelNfsVolNS],
		}
	}

	return false, nil
}
