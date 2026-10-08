package alicloudnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func requeueWaitKcpStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// if no conditions, then we're waiting for the KCP condition to appear
	if len(state.ObjAsAlicloudNfsVolume().Status.Conditions) == 0 {
		return composed.StopWithRequeueDelay(2 * util.Timing.T100ms()), nil
	}

	return nil, nil
}
