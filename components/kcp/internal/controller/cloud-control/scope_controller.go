package cloudresources

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/abstractions"
	kcpscope "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/scope"
	scopeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/scope/client"
	awsclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/aws/client"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewScopeReconciler(mgr manager.Manager, awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient]) *ScopeReconciler {
	return &ScopeReconciler{
		Reconciler: kcpscope.NewScopeReconciler(kcpscope.NewStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromManager(mgr)),
			abstractions.NewFileReader(),
			awsStsClientProvider,
		)),
	}
}

type ScopeReconciler struct {
	Reconciler kcpscope.ScopeReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scope,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scope/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scope/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas,verbs=get;list;watch
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/status,verbs=get

func (r *ScopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

func (r *ScopeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.Scope{}).
		Watches(
			util.NewKymaUnstructured(),
			handler.EnqueueRequestsFromMapFunc(r.mapRequestsFromKymaCR),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *ScopeReconciler) mapRequestsFromKymaCR(ctx context.Context, kymaObj client.Object) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: kymaObj.GetNamespace(),
				Name:      kymaObj.GetName(),
			},
		},
	}
}
