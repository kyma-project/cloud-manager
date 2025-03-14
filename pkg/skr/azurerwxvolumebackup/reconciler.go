package azurerwxvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)

	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

// TODO: fill out the rest of actions
func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"azureRwxVolumeBackup",
	)
}

func NewReconciler(args skrruntime.ReconcilerArguments) reconcile.Reconciler {

	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
			composed.NewStateClusterFromCluster(args.SkrCluster),
		),
	}
}
