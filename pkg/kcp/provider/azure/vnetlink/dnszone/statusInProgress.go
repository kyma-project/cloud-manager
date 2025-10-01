package dnszone

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInProgress(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	changed := false

	if state.ObjAsAzureVNetLink().Status.State == "" {
		logger.Info("Updating KCP AzureVNetLink status state to InProgress")
		state.ObjAsAzureVNetLink().Status.State = cloudcontrolv1beta1.VirtualNetworkLinkStateInProgress
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsAzureVNetLink()).
		ErrorLogMessage("Error setting KCP AzureVNetLink status state").
		SuccessErrorNil().
		Run(ctx, state)
}
