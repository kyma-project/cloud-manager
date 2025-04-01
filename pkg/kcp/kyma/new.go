package kyma

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type KymaReconciler interface {
	reconcile.Reconciler
}

func New(
	baseStateFactory composed.StateFactory,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) KymaReconciler {
	return &kymaReconciler{
		stateFactory: NewStateFactory(baseStateFactory, activeSkrCollection),
	}
}

type kymaReconciler struct {
	stateFactory StateFactory
}

func (r *kymaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.stateFactory.NewState(req)
	action := r.newAction()

	return composed.Handling().
		WithMetrics("kyma", util.RequestObjToString(req)).
		Handle(action(ctx, state))
}

func (r *kymaReconciler) newAction() composed.Action {
	return composed.ComposeActionsNoName(
		composed.LoadObj,
		kymaFindModuleState,
		scopeLoad,

		composed.If(
			shouldDisable,
			skrDeactivate,
			composed.StopAndForgetAction,
		),

		composed.If(
			shouldEnable,
			skrActivate,
			composed.StopAndForgetAction,
		),
	)
}

var _ composed.Predicate = shouldDisable

func shouldDisable(_ context.Context, st composed.State) bool {
	state := st.(*State)
	if state.scope == nil {
		return true
	}
	if state.moduleState == util.KymaModuleStateNotPresent {
		return true
	}
	return false
}

var _ composed.Predicate = shouldEnable

func shouldEnable(_ context.Context, st composed.State) bool {
	state := st.(*State)
	if state.scope == nil {
		return false
	}
	if state.moduleState != util.KymaModuleStateNotPresent {
		return true
	}
	return false
}
