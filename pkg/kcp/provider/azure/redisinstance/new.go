package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		state, err := stateFactory.NewState(ctx, st.(types.State), logger)
		if err != nil {
			redisInstance := st.Obj().(*v1beta1.RedisInstance)
			return composed.UpdateStatus(redisInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new Azure RedisInstance state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"azureRedisInstance",
			actions.AddCommonFinalizer(),
			loadPrivateEndPoint,
			loadPrivateDnsZoneGroup,
			loadRedis,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"azure-redisInstance-create",
					createRedis,
					updateStatusId,
					waitRedisAvailable,
					createPrivateEndPoint,
					waitPrivateEndPointAvailable,
					createPrivateDnsZoneGroup,
					modifyRedis,
					updateStatus,
				),
				composed.ComposeActions(
					"azure-redisInstance-delete",
					deleteRedis,
					waitRedisDeleted,
					deletePrivateDnsZoneGroup,
					waitPrivateDnsZoneGroupDeleted,
					deletePrivateEndPoint,
					waitPrivateEndPointDeleted,
					actions.RemoveCommonFinalizer(),
					composed.StopAndForgetAction,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
