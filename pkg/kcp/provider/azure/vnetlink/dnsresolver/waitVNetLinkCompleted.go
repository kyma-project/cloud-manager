package dnsresolver

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVNetLinkCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink == nil ||
		state.vnetLink.Properties == nil ||
		state.vnetLink.Properties.ProvisioningState == nil ||
		*state.vnetLink.Properties.ProvisioningState != armdnsresolver.ProvisioningStateSucceeded {

		logger.Info("Waiting for DNS resolver VirtualNetworkLink state Succeeded")

		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx

	}

	logger.Info("DNS resolver VirtualNetworkLink state Succeeded")
	return nil, ctx
}
