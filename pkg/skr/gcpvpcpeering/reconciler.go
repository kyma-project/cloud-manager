package gcpvpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct {
}

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

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crGcpVpcPeeringMain",
		composed.LoadObj,
		updateId,
		loadKcpRemoteNetwork,
		loadKcpGcpVpcPeering,
		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"gcpVpcPeering-create",
				actions.AddCommonFinalizer(),
				createKcpRemoteNetwork,
				waitRemoteNetworkCreation,
				createKcpVpcPeering,
				waitKcpStatusUpdate,
				updateStatus,
				waitSkrStatusReady,
			),
			composed.ComposeActions(
				"gcpVpcPeering-delete",
				deleteKcpVpcPeering,
				deleteKcpRemoteNetwork,
				actions.RemoveCommonFinalizer(),
				composed.StopAndForgetAction,
			),
		),
		composed.StopAndForgetAction,
	)
}
