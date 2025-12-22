package gcpvpcpeering

import (
	"context"

	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.ObjAsGcpVpcPeering().Status.Id != "" {
		return nil, nil
	}

	logger.Info("SKR GcpVpcPeering generating ID for VpcPeering")
	id := uuid.NewString()

	state.ObjAsGcpVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated
	state.ObjAsGcpVpcPeering().Status.Id = id
	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating status with ID "+err.Error(), composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	logger.Info("SKR GcpVpcPeering updated with ID status ", "id", id)

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), nil
}
