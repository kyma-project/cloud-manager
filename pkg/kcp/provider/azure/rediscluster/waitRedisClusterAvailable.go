package rediscluster

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func waitRedisClusterAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisCluster == nil {
		errorMsg := "Error: azure redis Cluster is not loaded"
		redisCluster := st.Obj().(*v1beta1.RedisCluster)
		return composed.UpdateStatus(redisCluster).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionTypeError,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	if ptr.Deref(state.azureRedisCluster.Properties.ProvisioningState, "") == armredis.ProvisioningStateSucceeded {
		return nil, ctx
	}

	logger.Info("Azure Redis Cluster is not ready yet, requeuing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
