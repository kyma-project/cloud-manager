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
		state, err := stateFactory.NewState(st.(types.State))

		if err != nil {
			err = fmt.Errorf("error creating new gcp vpcpeering state %w", err)
			logger.Error(err, "Error")
			return composed.StopAndForget, nil
		}

		return composed.ComposeActions(
			"gcpVpcPeering",
			actions.AddCommonFinalizer(),
			loadKymaNetwork,
			loadRemoteNetwork,
			loadRemoteVpcPeering,
			loadKymaVpcPeering,
			setPeeringStatusIds,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"gcpVpcPeering-create",
					checkIfRemoteVpcIsTagged,
					createRemoteVpcPeering,
					waitRemoteVpcPeeringAvailable,
					createKymaVpcPeering,
					waitVpcPeeringActive,
					updateStatus,
				),
				composed.ComposeActions(
					"gcpVpcPeering-delete",
					removeReadyCondition,
					deleteVpcPeering,
					waitKymaVpcPeeringDeletion,
					deleteRemoteVpcPeering,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
