package azurevnetlink

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitKcpVpcPeeringDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpAzureVNetLink == nil {
		logger.Info("AzureVNetLink is deleted")
		return nil, ctx
	}

	logger.Info("Waiting for AzureVNetLink to be deleted")

	// wait until AzureVNetLink does not exist / gets deleted
	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
