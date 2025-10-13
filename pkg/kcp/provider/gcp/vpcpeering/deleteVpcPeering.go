package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsVpcPeering()
	logger := composed.LoggerFromCtx(ctx)

	if state.kymaVpcPeering == nil {
		logger.Info("VPC Peering is not loaded")
		return nil, nil
	}

	logger.Info("Deleting GCP VPC Peering " + obj.Spec.Details.PeeringName)

	err := state.client.DeleteVpcPeering(
		ctx,
		state.getKymaVpcPeeringName(),
		state.LocalNetwork().Status.Network.Gcp.GcpProject,
		state.LocalNetwork().Status.Network.Gcp.NetworkName,
	)

	if err != nil {
		logger.Error(err, "Error deleting GCP VPC Peering")
		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Error deleting local network VpcPeering",
			}).
			ErrorLogMessage("Error deleting local network VpcPeering").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
