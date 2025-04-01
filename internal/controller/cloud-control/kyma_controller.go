package cloudcontrol

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpkyma "github.com/kyma-project/cloud-manager/pkg/kcp/kyma"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func SetupKymaReconciler(
	kcpManager manager.Manager,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) error {
	return NewKymaReconciler(
		kcpkyma.New(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(kcpManager)),
			activeSkrCollection,
		),
	).
		SetupWithManager(kcpManager)
}

func NewKymaReconciler(reconciler kcpkyma.KymaReconciler) *KymaReconciler {
	return &KymaReconciler{
		Reconciler: reconciler,
	}
}

type KymaReconciler struct {
	Reconciler kcpkyma.KymaReconciler
}

// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/status,verbs=get
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/finalizers,verbs=update

func (r *KymaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

func (r *KymaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Watches(
			util.NewKymaUnstructured(),
			handler.EnqueueRequestsFromMapFunc(r.mapRequestsFromKymaCR),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Named("kyma").
		Complete(r)
}

func (r *KymaReconciler) mapRequestsFromKymaCR(_ context.Context, kymaObj client.Object) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: kymaObj.GetNamespace(),
				Name:      kymaObj.GetName(),
			},
		},
	}
}
