package rediscluster

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

func loadPrivateEndPoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	if state.privateEndPoint != nil {
		logger.Info("Azure Private EndPoint already loaded")
		return nil, nil
	}
	logger.Info("Loading Azure Private EndPoint")
	privateEndPointName := state.ObjAsRedisCluster().Name
	resourceGroupName := state.resourceGroupName
	privateEndPointCluster, err := state.client.GetPrivateEndPoint(ctx, resourceGroupName, privateEndPointName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			logger.Info("Azure Private EndPoint Cluster not found")
			return nil, nil
		}
		logger.Error(err, "Error loading Azure Private EndPoint")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure Private EndPoint loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	logger.Info("Azure Private EndPoint Cluster loaded", "provisioningState", privateEndPointCluster.Properties.ProvisioningState)
	state.privateEndPoint = privateEndPointCluster
	return nil, nil
}
