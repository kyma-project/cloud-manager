package nfsinstance

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shareNetworkCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.shareNetwork != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	logger.Info("Creating CCEE shareNetwork")

	shareNetwork, err := state.cceeClient.CreateShareNetwork(ctx, state.network.ID, state.subnet.ID, state.ShareNetworkName())
	if err != nil {
		logger.Error(err, "Error creating CCEE shareNetwork")
		state.ObjAsNfsInstance().Status.State = cloudcontrolv1beta1.ErrorState
		return composed.PatchStatus(state.ObjAsNfsInstance()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error creating CCEE shareNetwork",
			}).
			ErrorLogMessage("Error patching CCEE NfsInstance status with error state after failed shareNetwork creation").
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	state.shareNetwork = shareNetwork
	state.ObjAsNfsInstance().SetStateData(StateDataShareNetworkId, shareNetwork.ID)

	return composed.PatchStatus(state.ObjAsNfsInstance()).
		ErrorLogMessage("Error updating CCEE NfsInstance state data with created shareNetworkId").
		SuccessErrorNil().
		Run(ctx, state)
}
