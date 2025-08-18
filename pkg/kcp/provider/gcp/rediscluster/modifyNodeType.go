package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyNodeType(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsGcpRedisCluster()

	if state.gcpRedisCluster == nil {
		return composed.StopWithRequeue, nil
	}

	currentNodeType := state.gcpRedisCluster.NodeType.String()
	desiredNodeType := redisCluster.Spec.NodeType

	if currentNodeType == desiredNodeType {
		return nil, ctx
	}

	state.UpdateNodeType(desiredNodeType)

	return nil, ctx
}
