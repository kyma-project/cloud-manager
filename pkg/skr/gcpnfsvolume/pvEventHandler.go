package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var PVEventHandler handler.Funcs = handler.Funcs{

	CreateFunc: func(ctx context.Context, e event.TypedCreateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
		if match, key := isMatchingPV(e.Object); match {
			w.Add(reconcile.Request{NamespacedName: *key})
		}
	},
	UpdateFunc: func(ctx context.Context, e event.TypedUpdateEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
		if match, key := isMatchingPV(e.ObjectNew); match {
			w.Add(reconcile.Request{NamespacedName: *key})
		}
	},
	DeleteFunc: func(ctx context.Context, e event.TypedDeleteEvent[client.Object], w workqueue.TypedRateLimitingInterface[reconcile.Request]) {
		if match, key := isMatchingPV(e.Object); match {
			w.Add(reconcile.Request{NamespacedName: *key})
		}
	},
}

func isMatchingPV(obj client.Object) (bool, *types.NamespacedName) {

	if _, ok := obj.(*v1.PersistentVolume); !ok {
		return false, nil
	}

	if name, ok := obj.GetLabels()[v1beta1.LabelNfsVolName]; ok {
		return true, &types.NamespacedName{
			Name:      name,
			Namespace: obj.GetLabels()[v1beta1.LabelNfsVolNS],
		}
	}

	return false, nil
}
