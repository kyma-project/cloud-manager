package vpcpeering

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"time"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering != nil {
		return nil, nil
	}

	con, err := state.client.CreateVpcPeeringConnection(
		ctx,
		state.vpc.VpcId,
		state.remoteVpc.VpcId,
		ptr.To(state.remoteRegion),
		state.remoteVpc.OwnerId)

	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
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

	logger = logger.WithValues("id", ptr.Deref(con.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("AWS VPC Peering Connection created")

	state.vpcPeering = con

	state.ObjAsVpcPeering().Status.Id = ptr.Deref(con.VpcPeeringConnectionId, "")

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status with connection id", composed.StopWithRequeue, ctx)
	}
	return nil, ctx
}
