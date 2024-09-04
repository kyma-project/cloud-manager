package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		logger.Info("Azure Redis subnet already loaded")
		return nil, nil
	}

	logger.Info("Loading Azure Redis subnet")

	resourceGroupName := state.Scope().Spec.Scope.Azure.TechnicalID
	virtualNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	subnetName := azureUtil.GetSubnetName(state.Scope().Spec.Scope.Azure.TechnicalID)
	subnetClientGetResponse, error := state.client.GetSubnet(ctx, resourceGroupName, virtualNetworkName, subnetName)

	if error != nil {
		if azuremeta.IsNotFound(error) {
			return nil, nil
		}

		logger.Error(error, "Error loading Azure subnet")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureRedis subnet: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis subnet loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.subnet = &subnetClientGetResponse.Subnet

	return nil, nil
}
