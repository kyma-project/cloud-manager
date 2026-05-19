package security

import (
	"context"
	"fmt"

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
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error deleting network flow logs: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error deleting flow log", ctx)
	}

	state.flowLog = nil

	return nil, ctx
}
