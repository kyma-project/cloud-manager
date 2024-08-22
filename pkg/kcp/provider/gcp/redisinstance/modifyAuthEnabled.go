package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyAuthEnabled(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	currentAuthEnabled := state.gcpRedisInstance.AuthEnabled
	desiredAuthEnabled := redisInstance.Spec.Instance.Gcp.AuthEnabled

	if currentAuthEnabled == desiredAuthEnabled {
		return nil, nil
	}

	state.UpdateAuthEnabled(desiredAuthEnabled)

	return nil, nil
}
