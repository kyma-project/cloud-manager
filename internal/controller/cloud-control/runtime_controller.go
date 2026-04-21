package cloudcontrol

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type RuntimeController struct {
	reconciler runtime.RuntimeReconciler
}

// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=runtimes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=runtimes/status,verbs=get

func (r *RuntimeController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.reconciler.Reconcile(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RuntimeController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.Runtime{}).
		Complete(r)
}
