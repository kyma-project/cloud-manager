package vpcpeering

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(types.State), logger)

		if err != nil {
			err = fmt.Errorf("error creating new aws vpcpeering state %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"awsVpcPeering",
			kcpNetworkLocalLoad,
			kcpNetworkRemoteLoad,
			statusInitiated,
			loadVpcPeeringConnection,
			loadVpc,
			loadRouteTables,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"awsVpcPeering-delete",
					removeReadyCondition,
					deleteRoutes,
					deleteVpcPeering,
					actions.PatchRemoveFinalizer,
				),
				composed.ComposeActions(
					"awsVpcPeering-non-delete",
					actions.PatchAddFinalizer,
					createRemoteClient,
					loadRemoteVpcPeeringConnection,
					loadRemoteVpc,
					loadRemoteRouteTables,
					checkNetworkTag,
					createVpcPeeringConnection,
					acceptVpcPeeringConnection,
					waitVpcPeeringActive,
					createRoutes,
					createRemoteRoutes,
					updateSuccessStatus,
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
