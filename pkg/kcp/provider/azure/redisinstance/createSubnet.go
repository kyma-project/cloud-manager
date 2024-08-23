package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, nil
	}

	logger.Info("Creating Azure Redis subnet")

	resourceGroupName := state.Scope().Spec.Scope.Azure.TechnicalID
	virtualNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	subnetName := azureUtil.GetSubnetName(state.Scope().Spec.Scope.Azure.TechnicalID)

	existingRanges := []string{
		state.Scope().Spec.Scope.Azure.Network.Nodes,
		state.Scope().Spec.Scope.Azure.Network.Pods,
		state.Scope().Spec.Scope.Azure.Network.Services,
	}
	cidr, error := iprangeallocate.AllocateCidr(22, existingRanges)

	if error != nil {
		logger.Error(error, "Error creating Azure subnet")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating AzureRedis subnet: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis subnet creating",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	error = state.client.CreateSubnet(ctx, resourceGroupName, virtualNetworkName, subnetName, cidr)

	if error != nil {
		logger.Error(error, "Error creating Azure subnet")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating AzureRedis subnet: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis subnet creating",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, nil
}
