package awsrediscluster

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

	if state.ObjAsAwsRedisCluster().Status.Id != "" {
		return nil, nil
	}

	id := uuid.NewString()

	state.ObjAsAwsRedisCluster().Status.Id = id
	state.ObjAsAwsRedisCluster().Status.State = cloudresourcesv1beta1.StateProcessing
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsRedisCluster status with ID label", composed.StopWithRequeue, ctx)
	}
	logger.Info("SKR AwsRedisCluster updated with ID status")

	return composed.StopWithRequeueDelay(util.Timing.T100ms()), nil
}
