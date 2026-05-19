package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func logAnalyticsWorkspaceCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.logAnalyticsWorkspace != nil {
		return nil, ctx
	}
	if state.resourceGroupData == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("resourceGroupData must exist before creating log analytics workspace"),
			"Cannot create log analytics workspace",
			composed.StopWithRequeue, ctx)
	}

	workspaceName := state.logAnalyticsWorkspaceName()
	logger.Info("Creating log analytics workspace", "name", workspaceName)

	params := armoperationalinsights.Workspace{
		Location: new(state.location()),
		Tags: map[string]*string{
			tagKymaRuntimeId: new(state.ObjAsRuntime().Name),
			tagKymaShootName: new(state.shootName()),
		},
		Properties: &armoperationalinsights.WorkspaceProperties{
			SKU: &armoperationalinsights.WorkspaceSKU{
				Name: new(armoperationalinsights.WorkspaceSKUNameEnumPerGB2018),
			},
		},
	}

	resp, err := azureclient.PollUntilDone(state.azureClient.CreateOrUpdateLogAnalyticsWorkspace(ctx,
		state.resourceGroupDataName(),
		workspaceName,
		params,
		nil))(ctx, nil)
	if err != nil {
		logger.Error(err, "Failed to create log analytics workspace", "workspaceParams", params)
		return azuremeta.LogErrorAndReturn(err, "Error creating log analytics workspace", ctx)
	}

	state.logAnalyticsWorkspace = &resp.Workspace

	return nil, ctx
}
