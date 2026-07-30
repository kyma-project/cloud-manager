package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)
		shared := st.(types.State)
		state, err := stateFactory.NewState(ctx, shared)
		if err != nil {
			err = fmt.Errorf("error creating new alicloud redisinstance state: %w", err)
			logger.Error(err, "Error")
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
		}

		return composed.ComposeActionsNoName(
			actions.AddCommonFinalizer(),
			loadRedis,

			// delete ================================================================================
			composed.If(composed.MarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					removeReadyCondition,
					deleteRedis,
					waitRedisDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),

			// create/update =========================================================================
			composed.If(composed.NotMarkedForDeletionPredicate,
				composed.ComposeActionsNoName(
					createRedis,
					waitRedisAvailable,
					setSecurityIps,
					enableSsl,
					addUpdatingCondition,
					modifyInstanceClass,
					waitRedisAvailable,
					modifyParameters,
					updateStatus,
				),
			),

			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
