package azurevnetlink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
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
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, reqest reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(reqest)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("skrazurevnetlink", util.RequestObjToString(reqest)).
		WithNoLog().
		Handle(action(ctx, state))

}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crAzureVNetLinkMain",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AzureVNetLink{}),
		composed.LoadObj,
		actions.UpdateIdAndInitState(cloudcontrolv1beta1.VirtualNetworkLinkStateInProgress),
		loadKcpAzureVNetLink,
		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"skrAzureVNetLink-create",
				actions.AddCommonFinalizer(),
				createKcpAzureVNetLink,
				updateStatus,
				actions.WaitStatusReadyAndState(cloudcontrolv1beta1.VirtualNetworkLinkStateCompleted),
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
