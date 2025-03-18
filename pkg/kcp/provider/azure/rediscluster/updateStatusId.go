package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsRedisCluster()

	if redisCluster.Status.Id != "" { // already set
		return nil, nil
	}

	redisCluster.Status.Id = *(state.azureRedisCluster.Name)

	err := state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Azure RedisCluster success .status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
