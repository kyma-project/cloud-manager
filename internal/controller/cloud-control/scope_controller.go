package cloudcontrol

import (
	"context"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsexposeddata "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData"
	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureexposeddata "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpexposeddata "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
	sapexposeddata "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData"
	sapexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/exposedData/client"
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
	ctx context.Context,
	kcpManager manager.Manager,
	awsStsClientProvider awsclient.GardenClientProvider[scopeclient.AwsStsClient],
	activeSkrCollection skrruntime.ActiveSkrCollection,
	gcpServiceUsageClientProvider gcpclient.ClientProvider[gcpclient.ServiceUsageClient],
	awsClientProvider awsclient.SkrClientProvider[awsexposeddataclient.Client],
	azureClientProvider azureclient.ClientProvider[azureexposeddataclient.Client],
	gcpClientProvider gcpclient.GcpClientProvider[gcpexposeddataclient.Client],
	sapClientProvider sapclient.SapClientProvider[sapexposeddataclient.Client],
) error {
	return NewScopeReconciler(
		kcpscope.New(
			kcpManager,
			awsStsClientProvider,
			activeSkrCollection,
			gcpServiceUsageClientProvider,
			awsexposeddata.NewStateFactory(awsClientProvider),
			azureexposeddata.NewStateFactory(azureClientProvider),
			gcpexposeddata.NewStateFactory(gcpClientProvider),
			sapexposeddata.NewStateFactory(sapClientProvider),
		),
	).SetupWithManager(ctx, kcpManager)
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

// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=scopes/finalizers,verbs=update
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/status,verbs=get
// +kubebuilder:rbac:groups=operator.kyma-project.io,resources=kymas/finalizers,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=skrstatuses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=skrstatuses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-control.kyma-project.io,resources=skrstatuses/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=gardenerclusters,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=gardenerclusters/status,verbs=get
// +kubebuilder:rbac:groups=infrastructuremanager.kyma-project.io,resources=gardenerclusters/finalizers,verbs=update

func (r *ScopeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return r.Reconciler.Reconcile(ctx, req)
}

func (r *ScopeReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// index networks by scope name
	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&cloudcontrolv1beta1.Network{},
		cloudcontrolv1beta1.NetworkFieldScope,
		func(obj client.Object) []string {
			net := obj.(*cloudcontrolv1beta1.Network)
			return []string{net.Spec.Scope.Name}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcontrolv1beta1.Scope{}).
		Watches(
			util.NewGardenerClusterUnstructured(),
			handler.EnqueueRequestsFromMapFunc(r.mapRequestsFromGardenerClusterCR),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *ScopeReconciler) mapRequestsFromGardenerClusterCR(_ context.Context, gcObj client.Object) []reconcile.Request {
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: gcObj.GetNamespace(),
				Name:      gcObj.GetName(),
			},
		},
	}
}
