package iprange

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func subnetCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, ctx
	}

	logger.Info("Creating Azure KCP IpRange subnet")

	err := state.azureClient.CreateSubnet(
		ctx,
		state.resourceGroupName, state.virtualNetworkName, state.subnetName,
		state.ObjAsIpRange().Status.Cidr,
		ptr.Deref(state.securityGroup.ID, ""),
	)

	if azuremeta.IsTooManyRequests(err) {
		return azuremeta.LogErrorAndReturn(err, "Azure KCP IpRange too many requests on create subnet", ctx)
	}

	if err != nil {
		logger.Error(err, "Error creating Azure Subnet for KCP IpRange")

		state.ObjAsIpRange().Status.State = cloudcontrol1beta1.ErrorState

		return composed.PatchStatus(state.ObjAsIpRange()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrol1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrol1beta1.ConditionTypeError,
				Message: "Error creating Azure subnet",
			}).
			ErrorLogMessage("Error patching Azure KCP IpRange status after failed creating subnet").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
