package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyReplicaCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsGcpRedisCluster()

	if state.gcpRedisCluster == nil {
		return composed.StopWithRequeue, nil
	}

	currentReplicaCount := ptr.Deref(state.gcpRedisCluster.ReplicaCount, redisCluster.Spec.ReplicasPerShard)
	desiredReplicaCount := redisCluster.Spec.ReplicasPerShard

	if currentReplicaCount == desiredReplicaCount {
		return nil, ctx
	}

	state.UpdateReplicaCount(desiredReplicaCount)

	return nil, ctx
}
