package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// deleteMountTargets deletes all NAS mount targets for the file system.
func deleteMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystemId == "" || len(state.mountTargets) == 0 {
		return nil, ctx
	}

	deleted := false
	for _, mt := range state.mountTargets {
		if mt.MountTargetDomain == "" {
			continue
		}
		// A mount target that is still creating cannot be deleted
		// (VolumeStatusForbidOperation); requeue until it settles to Active.
		if mt.Status != "" && mt.Status != "Active" {
			logger.Info("Waiting for AliCloud NAS mount target to become active before deletion", "mountTargetDomain", mt.MountTargetDomain, "status", mt.Status)
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}
		logger.Info("Deleting AliCloud NAS mount target", "mountTargetDomain", mt.MountTargetDomain)
		if err := state.client.DeleteMountTarget(ctx, state.fileSystemId, mt.MountTargetDomain); err != nil {
			logger.Error(err, "Error deleting AliCloud NAS mount target", "mountTargetDomain", mt.MountTargetDomain)
			return composed.StopWithRequeue, ctx
		}
		deleted = true
	}

	if deleted {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	return nil, ctx
}
