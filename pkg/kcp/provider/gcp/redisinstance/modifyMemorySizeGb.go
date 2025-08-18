package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyMemorySizeGb(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	currentMemorySizeGb := state.gcpRedisInstance.MemorySizeGb
	desiredMemorySizeGb := redisInstance.Spec.Instance.Gcp.MemorySizeGb

	if currentMemorySizeGb == desiredMemorySizeGb {
		return nil, ctx
	}

	state.UpdateMemorySizeGb(desiredMemorySizeGb)

	return nil, ctx
}
