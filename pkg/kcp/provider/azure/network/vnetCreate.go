package network

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vnetCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.network != nil {
		return nil, ctx
	}

	logger.Info("Creating Azure VNet for KCP Network")

	_, err := azureclient.PollUntilDone(state.azureClient.CreateOrUpdateNetwork(
		ctx,
		state.resourceGroupName,
		state.vnetName,
		azureclient.NewVirtualNetwork(state.location, state.cidr, state.tags),
		nil,
	))(ctx, nil)

	if err != nil {
		logger.Error(err, "Failed to create Azure VNet for KCP Network")

		return composed.PatchStatus(state.ObjAsNetwork()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Failed to create Azure VNet",
			}).
			ErrorLogMessage("Error patching KCP Network status with error condition after Azure VNet creation failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	logger.Info("Azure VNet created successfully")

	return nil, nil
}
