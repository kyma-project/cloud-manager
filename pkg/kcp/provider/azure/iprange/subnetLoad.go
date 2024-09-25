package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func subnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	subnet, err := state.azureClient.GetSubnet(ctx, state.resourceGroupName, state.virtualNetworkName, state.subnetName)
	if azuremeta.IsNotFound(err) {
		logger.Info("Azure KCP IpRange subnet not found")
		return nil, nil
	}
	if azuremeta.IsTooManyRequests(err) {
		return azuremeta.LogErrorAndReturn(err, "Azure KCP IpRange too many requests on subnet load", ctx)
	}
	if err != nil {
		logger.Error(err, "Error loading Azure KCP IpRange subnet")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error loading subnet",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after load subnet error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure KCP IpRange subnet loaded")

	state.subnet = subnet

	return nil, nil
}
