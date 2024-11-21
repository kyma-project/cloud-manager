package nfsinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/kcp/nfsinstance/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		logger.Info("Creating state")

		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			logger.Error(err, "Error creating state")
			nfsInstance := st.Obj().(*v1beta1.NfsInstance)
			return composed.UpdateStatus(nfsInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP NfsInstance state: %s", err)).
				Run(ctx, st)
		}
		return composed.ComposeActions(
			"gcsNfsInstance",
			validateAlways,
			actions.AddFinalizer,
			checkGcpOperation,
			loadNfsInstance,
			validatePostCreate,
			checkNUpdateState,
			checkUpdateMask,
			syncNfsInstance,
			composed.BuildBranchingAction("RunFinalizer", StatePredicate(client.Deleted, ctx, state),
				actions.RemoveFinalizer, nil),
		)(ctx, state)
	}
}

func StatePredicate(status v1beta1.StatusState, ctx context.Context, state *State) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		return status == state.curState
	}
}
