package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
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
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
			args.Reloader,
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if Ignore.ShouldIgnoreKey(req) {
		return ctrl.Result{}, nil
	}

	state := r.factory.NewState(req)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *reconciler) newState() *State {
	return nil
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crIpRangeMain",
		composed.LoadObj,
		preventCidrChange,
		validateCidr,
		addFinalizer,
		loadKcpIpRange,
		createKcpIpRange,
		deleteKcpIpRange,
		removeFinalizer,
		updateStatus,
		composed.StopAndForgetAction,
	)
}
