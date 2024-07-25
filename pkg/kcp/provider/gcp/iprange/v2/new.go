package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			ipRange := st.Obj().(*v1beta1.IpRange)
			return composed.UpdateStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(err.Error()).
				Run(ctx, st)
		}
		return composed.ComposeActions(
			"gcpIpRange",
			preventCidrEdit,
			allocateIpRange,
			copyCidrToStatus,
			validateCidr,
			actions.AddFinalizer,
			checkGcpOperation,
			loadAddress,
			fixStatusId,
			loadPsaConnection,
			compareStates,
			updateState,
			composed.BuildSwitchAction(
				"gcpIpRangeSwitch",
				nil,
				composed.NewCase(StatePredicate(client.SyncAddress, ctx, state), syncAddress),
				composed.NewCase(StatePredicate(client.SyncPsaConnection, ctx, state), syncPsaConnection),
				composed.NewCase(StatePredicate(client.DeletePsaConnection, ctx, state), syncPsaConnection),
				composed.NewCase(StatePredicate(client.DeleteAddress, ctx, state), syncAddress),
				composed.NewCase(StatePredicate(client.Deleted, ctx, state), actions.RemoveFinalizer),
			),
		)(ctx, state)
	}
}

func StatePredicate(status v1beta1.StatusState, ctx context.Context, state *State) composed.Predicate {
	return func(ctx context.Context, st composed.State) bool {
		return status == state.curState
	}
}
