package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func requeueWaitKcpStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// if no conditions, then we're waiting for the KCP condition to appear
	if len(state.ObjAsAwsNfsVolume().Status.Conditions) == 0 {
		return composed.StopWithRequeueDelay(200 * time.Millisecond), nil
	}

	return nil, nil
}
