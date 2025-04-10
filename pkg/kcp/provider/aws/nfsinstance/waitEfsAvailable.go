package nfsinstance

import (
	"context"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func waitEfsAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.efs.LifeCycleState == efstypes.LifeCycleStateAvailable {
		logger.Info("EFS state is Available")
		return nil, nil
	}

	logger.
		WithValues("efsState", state.efs.LifeCycleState).
		Info("Waiting EFS state to become Available")

	return composed.StopWithRequeueDelay(time.Second), nil
}
