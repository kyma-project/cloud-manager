package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func flowLogsCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.flowLog != nil {
		return nil, ctx
	}

	if state.VpcNetwork() == nil || state.VpcNetwork().Status.Identifiers.Vpc == "" {
		return nil, ctx
	}

	if state.watcher == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("network watcher must exist before creating flow log"),
			"Cannot create flow log",
			composed.StopWithRequeue, ctx)
	}
	if state.storageAccount == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("storage account must exist before creating flow log"),
			"Cannot create flow log",
			composed.StopWithRequeue, ctx)
	}
	//if state.logAnalyticsWorkspace == nil {
	//	return composed.LogErrorAndReturn(
	//		fmt.Errorf("log analytics workspace must exist before creating flow log"),
	//		"Cannot create flow log",
	//		composed.StopWithRequeue, ctx)
	//}

	flowLogName := FlowLogName(state.VpcNetwork())
	logger.Info("Creating flow log", "name", flowLogName)

	params := armnetwork.FlowLog{
		Location: new(state.ObjAsRuntime().Spec.Shoot.Region),
		Properties: &armnetwork.FlowLogPropertiesFormat{
			StorageID:        state.storageAccount.ID,
			TargetResourceID: new(state.VpcNetwork().Status.Identifiers.Vpc),
			Enabled:          new(true),
			RetentionPolicy: &armnetwork.RetentionPolicyParameters{
				Days:    new(int32(flowLogRetentionDays)),
				Enabled: new(true),
			},
			//FlowAnalyticsConfiguration: &armnetwork.TrafficAnalyticsProperties{
			//	NetworkWatcherFlowAnalyticsConfiguration: &armnetwork.TrafficAnalyticsConfigurationProperties{
			//		Enabled:                  new(true),
			//		WorkspaceID:              state.logAnalyticsWorkspace.Properties.CustomerID,
			//		WorkspaceRegion:          state.logAnalyticsWorkspace.Location,
			//		WorkspaceResourceID:      state.logAnalyticsWorkspace.ID,
			//		TrafficAnalyticsInterval: new(int32(60)),
			//	},
			//},
		},
	}

	resp, err := azureclient.PollUntilDone(state.azureClient.CreateOrUpdateFlowLog(ctx,
		ResourceGroupWatcherName(),
		NetworkWatcherName(state.ObjAsRuntime().Spec.Shoot.Region),
		flowLogName,
		params,
		nil))(ctx, nil)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error creating network flow logs: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error creating flow log", ctx)
	}

	state.flowLog = &resp.FlowLog

	return nil, ctx
}
