package cloudcontrol

import (
	"context"
	kcpnuke "github.com/kyma-project/cloud-manager/pkg/kcp/nuke"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupNukeReconciler(
	kcpManager manager.Manager,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) error {
	return NewNukeReconciler(
		kcpnuke.New(
			kcpManager,
			activeSkrCollection,
		),
	).SetupWithManager(kcpManager)
}

func NewNukeReconciler(reconciler kcpnuke.NukeReconciler) *NukeReconciler {
	return &NukeReconciler{
		Reconciler: reconciler,
	}
}

// NukeReconciler reconciles a Nuke object
type NukeReconciler struct {
	Reconciler kcpnuke.NukeReconciler
}

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nukes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nukes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=nukes/finalizers,verbs=update

func (r *NukeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NukeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.Nuke{}).
		Complete(r)
}
