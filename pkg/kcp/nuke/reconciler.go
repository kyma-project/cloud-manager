package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NukeReconciler interface {
	reconcile.Reconciler
}

func New(
	mgr manager.Manager,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) NukeReconciler {
	return &nukeReconciler{
		stateFactory: NewStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(mgr)),
			activeSkrCollection,
		),
	}
}

type nukeReconciler struct {
	stateFactory StateFactory
}

func (r *nukeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.stateFactory.NewState(req)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *nukeReconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"nukeMain",
		composed.LoadObj,
		composed.If(
			// if Nuke not marked for deletion
			composed.Not(composed.MarkedForDeletionPredicate),
			loadResources,
			statusDiscovered,
			deleteResources,
			statusDeleting,
			statusDeleted,
			allDeleted,
		),
	)
}
