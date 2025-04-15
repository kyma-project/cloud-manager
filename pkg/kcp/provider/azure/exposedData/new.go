package exposedData

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func New(sf StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		scopeState := st.(scopetypes.State)
		if !composed.IsObjLoaded(ctx, scopeState) {
			return composed.LogErrorAndReturn(
				errors.New("logical error"),
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
			kcpNetworkLoad,
			vnetLoad,
			subnetsLoad,
			natGatewaysLoad,
			publicIpAddressesLoad,
			exposedDataSet,
		)(ctx, state)
	}
}
