package awsnfsvolume

import (
	"context"
	"time"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.ObjAsAwsNfsVolume().Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	if state.ObjAsAwsNfsVolume().Labels == nil {
		state.ObjAsAwsNfsVolume().Labels = map[string]string{}
	}

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AwsNfsVolume updated with ID label")

	state.ObjAsAwsNfsVolume().Status.Id = id
	state.ObjAsAwsNfsVolume().Status.State = cloudresourcesv1beta1.StateProcessing
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AwsNfsVolume updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), nil
}
