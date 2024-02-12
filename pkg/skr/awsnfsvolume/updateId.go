package awsnfsvolume

import (
	"context"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.ObjAsAwsNfsVolume().Status.Id != "" {
		logger.Info("SKR AwsNfsVolume already has ID set")
		return nil, nil
	}

	id := uuid.NewString()

	if state.ObjAsAwsNfsVolume().Labels == nil {
		state.ObjAsAwsNfsVolume().Labels = map[string]string{}
	}
	state.ObjAsAwsNfsVolume().Labels[cloudresourcesv1beta1.LabelId] = id

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AwsNfsVolume updated with ID label")

	state.ObjAsAwsNfsVolume().Status.Id = id
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AwsNfsVolume updated with ID status")

	return composed.StopWithRequeue, nil
}
