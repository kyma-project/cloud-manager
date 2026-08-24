package alicloudnfsvolume

import (
	"context"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.ObjAsAlicloudNfsVolume().Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	if state.ObjAsAlicloudNfsVolume().Labels == nil {
		state.ObjAsAlicloudNfsVolume().Labels = map[string]string{}
	}
	state.ObjAsAlicloudNfsVolume().Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudNfsVolume with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AlicloudNfsVolume updated with ID label")

	state.ObjAsAlicloudNfsVolume().Status.Id = id
	state.ObjAsAlicloudNfsVolume().Status.State = cloudresourcesv1beta1.StateProcessing
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudNfsVolume status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AlicloudNfsVolume updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), nil
}
