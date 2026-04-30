package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func logAnalyticsWorkspaceLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.resourceGroupData == nil {
		return nil, ctx
	}

	resp, err := state.azureClient.GetLogAnalyticsWorkspace(ctx,
		state.resourceGroupDataName(),
		state.logAnalyticsWorkspaceName(),
		nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading log analytics workspace", ctx)
	}
	if err == nil {
		state.logAnalyticsWorkspace = &resp.Workspace
	}

	return nil, ctx
}
