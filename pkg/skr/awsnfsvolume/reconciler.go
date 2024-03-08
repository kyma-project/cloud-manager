package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
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
		"crAwsNfsVolumeMain",
		composed.LoadObj,
		loadSkrIpRange,
		waitIpRangeReady,
		loadVolume,
		addFinalizer,
		updateId,
		loadKcpNfsInstance,
		createKcpNfsInstance,
		updateStatus,
		createVolume,
		requeueWaitKcpStatus,
		stopIfNotBeingDeleted,

		// this below executes only when marked for deletion

		deletePv,
		waitPvDeleted,

		deleteKcpNfsInstance,
		waitKcpNfsInstanceDeleted,

		removeFinalizer,

		composed.StopAndForgetAction,
	)
}
