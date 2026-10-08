package alicloudrediscluster

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
		return nil, ctx
	}

	if state.ObjAsAlicloudRedisCluster().Status.Id != "" {
		return nil, ctx
	}

	id := uuid.NewString()

	state.ObjAsAlicloudRedisCluster().Status.Id = id
	state.ObjAsAlicloudRedisCluster().Status.State = cloudresourcesv1beta1.StateProcessing
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudRedisCluster status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AlicloudRedisCluster updated with ID status")

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), ctx
}
