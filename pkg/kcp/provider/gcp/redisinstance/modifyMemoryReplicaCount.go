package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyMemoryReplicaCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	currentReplicaCount := state.gcpRedisInstance.ReplicaCount
	desiredReplicaCount := redisInstance.Spec.Instance.Gcp.ReplicaCount

	if currentReplicaCount == desiredReplicaCount {
		return nil, nil
	}

	state.UpdateReplicaCount(desiredReplicaCount)

	return nil, nil
}
