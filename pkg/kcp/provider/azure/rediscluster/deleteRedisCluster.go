package rediscluster

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisCluster == nil {
		return nil, nil
	}

	if *state.azureRedisCluster.Properties.ProvisioningState == "Deleting" {
		return nil, nil // delete is waited in next action
	}

	logger.Info("Deleting Azure Redis")

	redisClusterName := state.ObjAsRedisCluster().Name
	resourceGroupName := state.resourceGroupName

	err := state.client.DeleteRedisInstance(ctx, resourceGroupName, redisClusterName)
	if err != nil {
		if azureMeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error deleting Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed deleting AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure redis deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
