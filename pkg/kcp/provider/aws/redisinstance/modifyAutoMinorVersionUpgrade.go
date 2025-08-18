package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyAutoMinorVersionUpgrade(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentAutoMinorVersionUpgrade := ptr.Deref(state.elastiCacheReplicationGroup.AutoMinorVersionUpgrade, false)
	desiredAutoMinorVersionUpgrade := redisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade

	if currentAutoMinorVersionUpgrade == desiredAutoMinorVersionUpgrade {
		return nil, ctx
	}

	state.UpdateAutoMinorVersionUpgrade(desiredAutoMinorVersionUpgrade)

	return nil, ctx
}
