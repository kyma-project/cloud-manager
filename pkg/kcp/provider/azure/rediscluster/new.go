package rediscluster

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		logger := composed.LoggerFromCtx(ctx)
		state, err := stateFactory.NewState(ctx, st.(types.State), logger)
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap Azure RedisCluster state")
			redisCluster := st.Obj().(*v1beta1.RedisCluster)
			redisCluster.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(redisCluster).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonCloudProviderError,
					Message: "Failed to create RedisCluster state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new Azure RedisCluster state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"azureRedisCluster",
			actions.AddCommonFinalizer(),
			loadPrivateEndPoint,
			loadPrivateDnsZoneGroup,
			loadRedisCluster,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"azure-redisCluster-create",
					createRedisCluster,
					updateStatusId,
					waitRedisClusterAvailable,
					createPrivateEndPoint,
					waitPrivateEndPointAvailable,
					createPrivateDnsZoneGroup,
					modifyRedisCluster,
					updateStatus,
				),
				composed.ComposeActions(
					"azure-redisCluster-delete",
					deleteRedisCluster,
					waitRedisClusterDeleted,
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
