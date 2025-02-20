package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyCacheNodeType(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisCluster()

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentCacheNodeType := ptr.Deref(state.elastiCacheReplicationGroup.CacheNodeType, "")
	desiredCacheNodeType := redisInstance.Spec.Instance.Aws.CacheNodeType

	if currentCacheNodeType == desiredCacheNodeType {
		return nil, nil
	}

	state.UpdateCacheNodeType(desiredCacheNodeType)

	return nil, nil
}
