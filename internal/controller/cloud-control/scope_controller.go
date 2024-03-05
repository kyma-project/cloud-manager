package cloudcontrol

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
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

func SetupScopeReconciler(
	kcpManager manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
) error {
	return NewScopeReconciler(
		kcpscope.New(
			kcpManager,
			awsStsClientProvider,
			activeSkrCollection,
		),
	).SetupWithManager(kcpManager)
}

func NewScopeReconciler(
	reconciler kcpscope.ScopeReconciler,
) *ScopeReconciler {
	return &ScopeReconciler{
		Reconciler: reconciler,
	}
}

type ScopeReconciler struct {
	Reconciler kcpscope.ScopeReconciler
}

//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas,verbs=get;list;watch
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/status,verbs=get
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

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
