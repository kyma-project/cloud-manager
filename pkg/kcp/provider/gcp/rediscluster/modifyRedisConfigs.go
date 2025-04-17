package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyRedisConfigs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsGcpRedisCluster()

	if state.gcpRedisCluster == nil {
		return composed.StopWithRequeue, nil
	}

	currentConfig := state.gcpRedisCluster.RedisConfigs
	desiredConfig := redisInstance.Spec.RedisConfigs

	if !AreConfigsMissmatched(currentConfig, desiredConfig) {
		return nil, nil
	}

	state.UpdateRedisConfigs(desiredConfig)

	return nil, nil
}
