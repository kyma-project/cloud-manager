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
				"Azure ExposeData flow called w/out loaded Scope",
				composed.StopAndForget,
				ctx,
			)
		}

		state, err := sf.NewState(ctx, scopeState)
		if err != nil {
			return err, ctx
		}

		return composed.ComposeActionsNoName(
			routersLoad,
			addressesLoad,
			exposedDataSetToScope,
		)(ctx, state)
	}
}
