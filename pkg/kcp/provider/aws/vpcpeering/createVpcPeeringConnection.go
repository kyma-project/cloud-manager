package vpcpeering

import (
	"context"
	"errors"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"time"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeeringConnection != nil {
		return nil, nil
	}

	con, err := state.client.CreateVpcPeeringConnection(
		ctx,
		state.vpc.VpcId,
		state.remoteVpc.VpcId,
		pointer.String(state.remoteRegion),
		state.remoteVpc.OwnerId)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: fmt.Sprintf("Failed creating VpcPeerings %s", err),
			}).
			ErrorLogMessage("Error updating VpcPeering status due to failed creating vpc peering connection").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(time.Minute)).
			Run(ctx, state)
	}

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS VPC Peering Connection", composed.StopWithRequeue, ctx)
	}

	logger = logger.WithValues("connectionId", pointer.StringDeref(con.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("AWS VPC Peering Connection created")

	state.vpcPeeringConnection = con

	state.ObjAsVpcPeering().Status.ConnectionId = *state.vpcPeeringConnection.VpcPeeringConnectionId

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status with connection id", composed.StopWithRequeue, ctx)
	}

	if state.vpcPeeringConnection == nil {
		logger.Error(errors.New("unable to load just created VPC Peering Connection"), "Logical error!!!")

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonUnknown,
				Message: "Failed creating VPC Peering",
			}).
			ErrorLogMessage("Error updating KCP VPC Peering status after failed loading of just created VPC Peering Connection").
			Run(ctx, state)
	}

	return nil, ctx
}
