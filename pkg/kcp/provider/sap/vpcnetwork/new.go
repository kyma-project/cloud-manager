package vpcnetwork

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		vpcState := st.(vpcnetworktypes.State)
		cctx, state, err := stateFactory.NewState(ctx, vpcState)
		if cctx != nil {
			ctx = cctx
		}
		if err != nil {
			logger := composed.LoggerFromCtx(ctx)
			logger.Error(err, "error creating new sap vpcnetwork state")
			return err, ctx
		}

		return composed.ComposeActionsNoName(
			composed.If(
				state.IsKymaTypePredicate,
				composed.IfElse(
					composed.MarkedForDeletionPredicate,
					composed.ComposeActionsNoName(
						// Delete
						infraDelete,
					),
					composed.ComposeActionsNoName(
						// Create/Update
						infraCreateUpdate,
					),
				),
			),
			composed.If(
				state.IsGardenerTypePredicate,
				composed.If(
					composed.NotMarkedForDeletionPredicate,
					infraObserve,
				),
			),
		)(ctx, state)
	}
}
