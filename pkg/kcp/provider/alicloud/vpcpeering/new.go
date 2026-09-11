package vpcpeering

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(vpcpeeringtypes.State), logger)
		if err != nil {
			err = fmt.Errorf("error creating new alicloud vpcpeering state %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, ctx
		}

		return composed.ComposeActions(
			"alicloudVpcPeering",
			statusInitiated,
			loadVpcPeerConnection,
			loadRouteTables,
			createRemoteClient,
			loadRemoteRouteTables,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"alicloudVpcPeering-delete",
					removeReadyCondition,
					remoteRoutesDelete,
					deleteRoutes,
					deleteVpcPeering,
					actions.PatchRemoveCommonFinalizer(),
				),
				composed.ComposeActions(
					"alicloudVpcPeering-non-delete",
					actions.PatchAddCommonFinalizer(),
					createVpcPeerConnection,
					waitVpcPeeringActive,
					setBandwidth,
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
