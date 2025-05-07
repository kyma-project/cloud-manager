package exposedData

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		scopeState := st.(scopetypes.State)
		if !composed.IsObjLoaded(ctx, scopeState) {
			return composed.LogErrorAndReturn(
				common.ErrLogical,
				"AWS ExposeData flow called w/out loaded Scope",
				composed.StopAndForget,
				ctx,
			)
		}
		if composed.IsMarkedForDeletion(scopeState.Obj()) {
			return composed.LogErrorAndReturn(
				common.ErrLogical,
				"AWS ExposeData flow called with Scope with deleteTimestamp",
				composed.StopAndForget,
				ctx,
			)
		}

		cctx, state, err := sf.NewState(ctx, scopeState)
		if cctx != nil {
			ctx = cctx
		}
		if err != nil {
			return err, ctx
		}

		return composed.ComposeActionsNoName(
			kcpNetworkVerify,
			vpcLoad,
			natGatewayLoad,
			exposedDataSetToScope,
			// todo: add more actions here
			composed.Noop,
		)(ctx, state)
	}
}
