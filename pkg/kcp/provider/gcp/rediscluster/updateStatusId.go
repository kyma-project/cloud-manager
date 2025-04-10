package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	redisCluster := state.ObjAsGcpRedisCluster()

	if redisCluster.Status.Id != "" { // already set
		return nil, nil
	}

	redisCluster.Status.Id = state.gcpRedisCluster.Name

	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating GcpRedisCluster success .status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
