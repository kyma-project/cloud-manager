package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	networktypes "github.com/kyma-project/cloud-manager/pkg/kcp/network/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		networkState := st.(networktypes.State)
		state, err := stateFactory.NewState(ctx, networkState)
		if err != nil {
			logger.Error(err, "Error creating Azure network state")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"azureNetwork",
			initState,
			resourceGroupLoad,
			vnetLoad,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureNetworkDelete",
					vnetDelete,
					resourceGroupDelete,
					actions.PatchRemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
				composed.ComposeActions(
					"azureNetworkCreate",
					actions.PatchAddCommonFinalizer(),
					resourceGroupCreate,
					vnetCreate,
					statusReady,
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
