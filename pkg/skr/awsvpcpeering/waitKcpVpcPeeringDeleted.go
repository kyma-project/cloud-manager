package awsvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitKcpVpcPeeringDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpVpcPeering == nil {
		logger.Info("VpcPeering is deleted")
		return nil, nil
	}

	logger.Info("Waiting for VpcPeering to be deleted")

	// wait until VpcPeering does not exist / gets deleted
	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
