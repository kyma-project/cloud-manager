package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deletePrivateEndPoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateEndPoint == nil {
		return nil, ctx
	}
	if *state.privateEndPoint.Properties.ProvisioningState == "Deleting" {
		return nil, ctx
	}
	logger.Info("Deleting Azure PrivateEndPoint")
	redisInstanceName := state.ObjAsRedisInstance().Name
	resourceGroupName := state.resourceGroupName
	err := state.client.DeletePrivateEndPoint(ctx, resourceGroupName, redisInstanceName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		logger.Error(err, "Error deleting Azure PrivateEndPoint")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed deleting AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure PrivateEndPoint deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
