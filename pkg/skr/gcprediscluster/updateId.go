package gcprediscluster

import (
	"context"

	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func updateId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.ObjAsGcpRedisCluster().Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	state.ObjAsGcpRedisCluster().Status.Id = id
	state.ObjAsGcpRedisCluster().Status.State = cloudresourcesv1beta1.StateProcessing
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR GcpRedisCluster status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR GcpRedisCluster updated with ID status")

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), nil
}
