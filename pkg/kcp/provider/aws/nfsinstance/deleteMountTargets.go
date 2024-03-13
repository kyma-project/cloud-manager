package nfsinstance

import (
	"context"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
	"time"
)

func deleteMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.efs == nil {
		return nil, nil
	}

	stateRequeueDelayed := map[efsTypes.LifeCycleState]struct{}{
		efsTypes.LifeCycleStateCreating: {},
		efsTypes.LifeCycleStateUpdating: {},
		efsTypes.LifeCycleStateDeleting: {},
	}
	stateOkToDelete := map[efsTypes.LifeCycleState]struct{}{
		efsTypes.LifeCycleStateAvailable: {},
	}

	for _, mt := range state.mountTargets {
		lll := logger.
			WithValues(
				"mountTargetId", pointer.StringDeref(mt.MountTargetId, ""),
				"subnetId", pointer.StringDeref(mt.SubnetId, ""),
				"availabilityZone", pointer.StringDeref(mt.AvailabilityZoneName, ""),
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
		err := state.awsClient.DeleteMountTarget(ctx, pointer.StringDeref(mt.MountTargetId, ""))
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error deleting mount target", composed.StopWithRequeueDelay(300*time.Millisecond), ctx)
		}
	}

	return nil, nil
}
