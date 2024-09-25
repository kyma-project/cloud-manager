package iprange

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func securityGroupLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	sg, err := state.azureClient.GetSecurityGroup(ctx, state.resourceGroupName, state.securityGroupName)
	if azuremeta.IsNotFound(err) {
		logger.Info("Azure KCP IpRange security group not found")
		return nil, nil
	}
	if azuremeta.IsTooManyRequests(err) {
		return azuremeta.LogErrorAndReturn(err, "Azure KCP IpRange too many requests on security group load", ctx)
	}
	if err != nil {
		logger.Error(err, "Error loading Azure KCP IpRange subnet")

		state.ObjAsIpRange().Status.State = cloudcontrolv1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error loading security group",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after load security group error").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure KCP IpRange security group loaded")

	state.securityGroup = sg

	return nil, nil
}
