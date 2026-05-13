package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func flowLogsDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.flowLog == nil {
		return nil, ctx
	}

	flowLogName := state.flowLogName()
	logger.Info("Deleting flow log", "name", flowLogName)

	_, err := azureclient.PollUntilDone(state.azureClient.DeleteFlowLog(ctx,
		state.resourceGroupWatcherName(),
		state.networkWatcherName(),
		flowLogName,
		nil))(ctx, nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting flow log", ctx)
	}

	state.flowLog = nil

	return nil, ctx
}
