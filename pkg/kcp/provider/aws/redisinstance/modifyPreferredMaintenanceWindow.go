package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyPreferredMaintenanceWindow(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisInstance()

	if !state.IsReplicationGroupAvailable() {
		return nil, ctx
	}

	currentPreferredMaintenanceWindow := ptr.Deref(state.memberClusters[0].PreferredMaintenanceWindow, "")
	desiredPreferredMaintenanceWindow := ptr.Deref(redisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow, "")

	if currentPreferredMaintenanceWindow == desiredPreferredMaintenanceWindow {
		return nil, ctx
	}
	if desiredPreferredMaintenanceWindow == "" {
		return nil, ctx
	}

	state.UpdatePreferredMaintenanceWindow(desiredPreferredMaintenanceWindow)

	return nil, ctx
}
