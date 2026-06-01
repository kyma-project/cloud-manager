package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	runtimetypes "github.com/kyma-project/cloud-manager/pkg/kcp/runtime/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		runtimeState := st.(runtimetypes.State)
		cctx, state, err := stateFactory.NewState(ctx, runtimeState)
		if cctx != nil {
			ctx = cctx
		}
		if err != nil {
			return err, ctx
		}

		return composed.ComposeActionsNoName(
			sccServicesLoad,
			composed.IfElse(
				runtimeState.SecurityServiceEnabledOnSubscriptionPredicate,
				sccServicesEnable,
				sccServicesDisable,
			),
		)(ctx, state)
	}
}
