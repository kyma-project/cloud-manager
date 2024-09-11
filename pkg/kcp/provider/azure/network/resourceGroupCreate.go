package network

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceGroupCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup != nil {
		return nil, nil
	}

	logger.Info("Creating CloudManager Azure resource group")

	rg, err := state.azureClient.CreateResourceGroup(
		ctx,
		state.resourceGroupName,
		state.location,
		state.tags,
	)
	if err != nil {
		logger.Error(err, "Error creating CloudManager resource group")

		return composed.PatchStatus(state.ObjAsNetwork()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Failed to create Azure Resource Group",
			}).
			ErrorLogMessage("Error patching KCP Network status with error condition after Azure create resource group failed").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			Run(ctx, state)
	}

	state.resourceGroup = rg

	return nil, nil
}
