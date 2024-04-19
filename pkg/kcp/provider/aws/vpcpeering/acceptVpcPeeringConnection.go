package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func acceptVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	con, err := state.client.AcceptVpcPeeringConnection(
		ctx,
		state.vpcPeeringConnection.VpcPeeringConnectionId,
	)

	if err != nil {
		logger.Error(err, "Error accepting VPC Peering")

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedAcceptingVpcPeeringConnection,
				Message: fmt.Sprintf("Failed accepting VpcPeerings %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed accepting vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	logger.Info(fmt.Sprintf("AWS VPC Peering Connection %s accepted", *con.VpcPeeringConnectionId))

	return nil, ctx
}
