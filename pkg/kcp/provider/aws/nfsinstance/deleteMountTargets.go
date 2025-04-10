package nfsinstance

import (
	"context"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"time"
)

func deleteMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.efs == nil {
		return nil, nil
	}

	stateRequeueDelayed := map[efstypes.LifeCycleState]struct{}{
		efstypes.LifeCycleStateCreating: {},
		efstypes.LifeCycleStateUpdating: {},
		efstypes.LifeCycleStateDeleting: {},
	}
	stateOkToDelete := map[efstypes.LifeCycleState]struct{}{
		efstypes.LifeCycleStateAvailable: {},
	}

	anyDeleted := false
	for _, mt := range state.mountTargets {
		lll := logger.
			WithValues(
				"mountTargetId", ptr.Deref(mt.MountTargetId, ""),
				"subnetId", ptr.Deref(mt.SubnetId, ""),
				"availabilityZone", ptr.Deref(mt.AvailabilityZoneName, ""),
				"lifeCycleState", mt.LifeCycleState,
			)
		_, shouldRequeueDelayed := stateRequeueDelayed[mt.LifeCycleState]
		if shouldRequeueDelayed {
			lll.Info("Waiting for mount target LifeCycleState")
			return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
		}
		_, okToDelete := stateOkToDelete[mt.LifeCycleState]
		if !okToDelete {
			continue
		}

		lll.Info("Deleting mount target")
		err := state.awsClient.DeleteMountTarget(ctx, ptr.Deref(mt.MountTargetId, ""))
		if err != nil {
			return awsmeta.LogErrorAndReturn(err, "Error deleting mount target", ctx)
		}

		anyDeleted = true
	}

	if anyDeleted {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, nil
}
