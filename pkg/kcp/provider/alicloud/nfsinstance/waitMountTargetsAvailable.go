package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitMountTargetsAvailable polls the NAS mount targets until they are all Active.
func waitMountTargetsAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	mts, err := state.client.DescribeMountTargets(ctx, state.fileSystemId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud NAS mount targets while waiting for availability")
		return composed.StopWithRequeue, ctx
	}

	state.mountTargets = mts

	if len(mts) == 0 {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	for _, mt := range mts {
		if mt.Status != "Active" {
			logger.Info("Waiting for AliCloud NAS mount target to become active", "mountTargetDomain", mt.MountTargetDomain, "status", mt.Status)
			return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
		}
	}

	return nil, ctx
}
