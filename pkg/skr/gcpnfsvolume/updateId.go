package gcpnfsvolume

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.ObjAsGcpNfsVolume().Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR GcpNfsVolume with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR GcpNfsVolume updated with ID label")

	state.ObjAsGcpNfsVolume().Status.Id = id
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR GcpNfsVolume status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR GcpNfsVolume updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), nil
}
