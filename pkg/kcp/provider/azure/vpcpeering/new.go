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
			err = fmt.Errorf("error creating new azure vpcpeering state %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"azureVpcPeering",
			initLocalNetworkId,
			initRemoteNetworkId,
			statusInitiated,
			initLocalClient,
			peeringLocalLoad,
			initRemoteClient,
			peeringRemoteLoad,
			composed.IfElse(
				composed.MarkedForDeletionPredicate,
				composed.ComposeActions(
					"azureVpcPeering-delete",
					peeringRemoteDelete,
					deleteVpcPeering,
					actions.PatchRemoveCommonFinalizer(),
				),
				composed.ComposeActions(
					"azureVpcPeering-non-delete",
					actions.PatchAddCommonFinalizer(),
					peeringRemoteRequireSpecifiedName,
					composed.If(
						predicateRequireVNetShootTag,
						vpcRemoteLoad,
						waitNetworkTag,
					),
					initLocalPeeringClient,
					peeringLocalCreate,
					peeringRemoteCreate,
					peeringLocalWaitReady,
					statusReady,
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
