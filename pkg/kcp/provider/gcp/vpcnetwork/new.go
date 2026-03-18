package vpcnetwork

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		vpcState := st.(vpcnetworktypes.State)
		cctx, state, err := sf.NewState(ctx, vpcState)
		if cctx != nil {
			ctx = cctx
		}
		if err != nil {
			return err, ctx
		}

		return composed.ComposeActionsNoName(
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
		)(ctx, state)
	}
}
