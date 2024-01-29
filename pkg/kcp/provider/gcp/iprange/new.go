package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/iprange/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/gcp/client"

	"github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			err = fmt.Errorf("error creating new gcp iprange state: %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}
		return composed.ComposeActions(
			"gcpIpRange",
			validateCidr,
			actions.AddFinalizer,
			checkGcpOperation,
			loadAddress,
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
