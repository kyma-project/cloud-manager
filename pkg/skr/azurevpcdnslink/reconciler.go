package azurevpcdnslink

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct{}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.ScopeProvider,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, reqest reconcile.Request) (reconcile.Result, error) {
	state, err := r.factory.NewState(ctx, reqest)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating AzureVpcDnsLink state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("azurevpcdnslink", util.RequestObjToString(reqest)).
		WithNoLog().
		Handle(action(ctx, state))

}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crAzureVNetLinkMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AzureVpcDnsLink{}),
		composed.LoadObj,
		actions.UpdateIdAndInitState(cloudcontrolv1beta1.VirtualNetworkLinkStateInProgress),
		loadKcpAzureVNetLink,
		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"skrAzureVNetLink-create",
				actions.AddCommonFinalizer(),
				createKcpAzureVNetLink,
				updateStatus,
				actions.WaitStatusReady(),
			),
			composed.ComposeActions(
				"skrAzureVNetLink-delete",
				deleteKcpVpcPeering,
				waitKcpVpcPeeringDeleted,
				actions.RemoveCommonFinalizer(),
			),
		),
		composed.StopAndForgetAction,
	)
}
