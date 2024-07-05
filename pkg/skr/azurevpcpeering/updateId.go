package azurevpcpeering

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

	obj := state.ObjAsAzureVpcPeering()
	if obj.Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	if obj.Labels == nil {
		obj.Labels = map[string]string{}
	}

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AzureVpcPeering updated with ID label")

	obj.Status.Id = id

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AzureVpcPeering status with ID label", composed.StopWithRequeue, ctx)
	}

	logger.Info("SKR AzureVpcPeering updated with ID status")

	return composed.StopWithRequeueDelay(100 * time.Millisecond), nil
}
