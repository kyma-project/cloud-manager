package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyRedisConfigs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	currentConfig := state.gcpRedisInstance.RedisConfigs
	desiredConfig := redisInstance.Spec.Instance.Gcp.RedisConfigs

	if !AreConfigsMissmatched(currentConfig, desiredConfig) {
		return nil, nil
	}

	state.UpdateRedisConfigs(desiredConfig)

	return nil, nil
}
