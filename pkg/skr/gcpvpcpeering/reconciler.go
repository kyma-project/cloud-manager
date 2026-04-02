package gcpvpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
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
			args.ScopeProvider,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state, err := r.factory.NewState(ctx, request)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating GcpVpcPeering state: %w", err)
	}
	action := r.newAction()

	return composed.Handling().
		WithMetrics("gcpvpcpeering", util.RequestObjToString(request)).
		WithNoLog().
		Handle(action(ctx, state))
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
				waitNetworkReady,
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
