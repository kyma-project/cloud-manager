package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func requeueWaitKcpStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	isBeingDeleted := composed.MarkedForDeletionPredicate(ctx, st)

	if !isBeingDeleted && len(state.ObjAsAwsNfsVolume().Status.Conditions) > 0 {
		return nil, nil
	}

	return composed.StopWithRequeueDelay(200 * time.Millisecond), nil
}
