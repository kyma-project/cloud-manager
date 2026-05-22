package cloudcontrol

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	awssecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/security"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuresecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/security"
	azuresecurityclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/security/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpsecurity "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/security"
	gcpsecurityclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/security/client"
	kcpruntime "github.com/kyma-project/cloud-manager/pkg/kcp/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

func SetupRuntimeReconciler(
	ctx context.Context,
	mgr ctrl.Manager,
	azureClientProvider azureclient.ClientProvider[azuresecurityclient.Client],
	gcpSecurityClientProvider gcpclient.GcpClientProvider[gcpsecurityclient.Client],
) error {
	return NewRuntimeController(
		kcpruntime.NewRuntimeReconciler(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
			awssecurity.NewStateFactory(),
			azuresecurity.NewStateFactory(azureClientProvider),
			gcpsecurity.NewStateFactory(gcpSecurityClientProvider),
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
func (r *RuntimeReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// index runtimes by binding name
	if err := mgr.GetFieldIndexer().IndexField(
		ctx, &infrastructuremanagerv1.Runtime{},
		cloudcontrolv1beta1.RuntimeFiledBindingName,
		func(rawObj client.Object) []string {
			x := rawObj.(*infrastructuremanagerv1.Runtime)
			return []string{x.Spec.Shoot.SecretBindingName}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructuremanagerv1.Runtime{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 10,
		}).
		Complete(r)
}
