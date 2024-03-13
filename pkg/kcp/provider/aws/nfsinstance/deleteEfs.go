package nfsinstance

import (
	"context"
	efsTypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
	"time"
)

func deleteEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.efs != nil {
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

	_, shouldRequeueDelayed := stateRequeueDelayed[state.efs.LifeCycleState]
	if shouldRequeueDelayed {
		logger.WithValues("lifeCycleState", state.efs.LifeCycleState).
			Info("Waiting for EFS LifeCycleState")
		return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
	}

	_, okToDelete := stateOkToDelete[state.efs.LifeCycleState]
	if !okToDelete {
		return nil, nil
	}

	err := state.awsClient.DeleteFileSystem(ctx, pointer.StringDeref(state.efs.FileSystemId, ""))
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting EFS", composed.StopWithRequeueDelay(300*time.Millisecond), ctx)
	}

	return composed.StopWithRequeue, nil
}
