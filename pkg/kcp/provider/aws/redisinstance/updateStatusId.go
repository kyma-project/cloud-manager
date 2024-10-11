package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	redisInstance := state.ObjAsRedisInstance()

	if redisInstance.Status.Id != "" { // already set
		return nil, nil
	}

	redisInstance.Status.Id = ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, "")

	err := state.PatchObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating RedisInstance .status.id", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
