package cloudcontrol

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	kcpruntime "github.com/kyma-project/cloud-manager/pkg/kcp/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupRuntimeReconciler(
	ctx context.Context,
	mgr ctrl.Manager,
) error {
	return NewRuntimeController(
		kcpruntime.NewRuntimeReconciler(composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr))),
	).SetupWithManager(ctx, mgr)
}

func NewRuntimeController(r kcpruntime.RuntimeReconciler) *RuntimeReconciler {
	return &RuntimeReconciler{
		reconciler: r,
	}
}

type RuntimeReconciler struct {
	reconciler kcpruntime.RuntimeReconciler
}

// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=runtimes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=runtimes/status,verbs=get

func (r *RuntimeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RuntimeReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.Runtime{}).
		Complete(r)
}
