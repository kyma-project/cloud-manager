package awsredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitKcpStatusUpdate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.KcpRedisInstance.Status.Conditions) == 0 {
		return composed.StopWithRequeueDelay(2 * util.Timing.T10000ms()), nil
	}

	return nil, ctx
}
