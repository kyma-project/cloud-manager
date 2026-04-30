package cloudcontrol

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	awsruntime "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/runtime"
	azureruntime "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/runtime"
	gcpruntime "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/runtime"
	kcpruntime "github.com/kyma-project/cloud-manager/pkg/kcp/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func SetupRuntimeReconciler(
	ctx context.Context,
	mgr ctrl.Manager,
) error {
	return NewRuntimeController(
		kcpruntime.NewRuntimeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
			awsruntime.NewStateFactory(),
			azureruntime.NewStateFactory(),
			gcpruntime.NewStateFactory(),
		),
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
func (r *RuntimeReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.Runtime{}).
		Complete(r)
}
