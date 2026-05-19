package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func flowLogsLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.watcher == nil {
		return nil, ctx
	}

	resp, err := state.azureClient.GetFlowLog(ctx,
		state.resourceGroupWatcherName(),
		state.networkWatcherName(),
		state.flowLogName(),
		nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error loading network flow logs: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error loading flow log", ctx)
	}
	if err == nil {
		state.flowLog = &resp.FlowLog
	}

	return nil, ctx
}
