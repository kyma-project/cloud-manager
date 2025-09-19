package dnszone

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVNetLinkCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink == nil ||
		state.vnetLink.Properties == nil ||
		state.vnetLink.Properties.VirtualNetworkLinkState == nil ||
		*state.vnetLink.Properties.VirtualNetworkLinkState != armprivatedns.VirtualNetworkLinkStateCompleted {

		logger.Info("Waiting for VirtualNetworkLink state Completed")

		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx

	}

	logger.Info("VirtualNetworkLink state Completed")
	return nil, ctx
}
