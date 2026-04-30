package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func logAnalyticsWorkspaceDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.logAnalyticsWorkspace == nil {
		return nil, ctx
	}

	workspaceName := state.logAnalyticsWorkspaceName()
	logger.Info("Deleting log analytics workspace", "name", workspaceName)

	_, err := azureclient.PollUntilDone(state.azureClient.DeleteLogAnalyticsWorkspace(ctx,
		state.resourceGroupDataName(),
		workspaceName,
		nil))(ctx, nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting log analytics workspace", ctx)
	}

	state.logAnalyticsWorkspace = nil
	return nil, ctx
}
