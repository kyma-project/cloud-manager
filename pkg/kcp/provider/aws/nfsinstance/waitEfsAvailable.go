package nfsinstance

import (
	"context"
	"time"

	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func waitEfsAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.efs.LifeCycleState == efstypes.LifeCycleStateAvailable {
		return nil, ctx
	}

	logger.
		WithValues("efsState", state.efs.LifeCycleState).
		Info("Waiting EFS state to become Available")

	return composed.StopWithRequeueDelay(10 * time.Second), ctx
}
