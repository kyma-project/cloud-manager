package nfsinstance

import (
	"context"
	"fmt"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
	"time"
)

func deleteEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.efs == nil {
		return nil, nil
	}

	logger.Info("Deciding if EFS should be deleted")

	stateRequeueDelayed := map[efstypes.LifeCycleState]struct{}{
		efstypes.LifeCycleStateCreating: {},
		efstypes.LifeCycleStateUpdating: {},
		efstypes.LifeCycleStateDeleting: {},
	}
	stateOkToDelete := map[efstypes.LifeCycleState]struct{}{
		efstypes.LifeCycleStateAvailable: {},
	}

	_, shouldRequeueDelayed := stateRequeueDelayed[state.efs.LifeCycleState]
	if shouldRequeueDelayed {
		logger.
			WithValues("waitStates", fmt.Sprintf("%v", stateRequeueDelayed)).
			Info("Waiting for EFS LifeCycleState")
		return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
	}

	_, okToDelete := stateOkToDelete[state.efs.LifeCycleState]
	if !okToDelete {
		logger.
			WithValues("deleteStates", fmt.Sprintf("%v", stateOkToDelete)).
			Info("The EFS should not be deleted")
		return nil, nil
	}

	logger.Info("Deleting EFS")
	err := state.awsClient.DeleteFileSystem(ctx, ptr.Deref(state.efs.FileSystemId, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting EFS", ctx)
	}

	return composed.StopWithRequeue, nil
}
