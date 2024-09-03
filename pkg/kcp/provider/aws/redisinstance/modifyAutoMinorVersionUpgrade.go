package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyAutoMinorVersionUpgrade(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheCluster == nil {
		return composed.StopWithRequeue, nil
	}

	currentAutoMinorVersionUpgrade := ptr.Deref(state.elastiCacheCluster.AutoMinorVersionUpgrade, false)
	desiredAutoMinorVersionUpgrade := redisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade

	if currentAutoMinorVersionUpgrade == desiredAutoMinorVersionUpgrade {
		return nil, nil
	}

	state.UpdateAutoMinorVersionUpgrade(desiredAutoMinorVersionUpgrade)

	return nil, nil
}
