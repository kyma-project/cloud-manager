package vpcpeering

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func peeringLocalWaitReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.localPeering != nil &&
		ptr.Deref(state.localPeering.Properties.PeeringState, "") == armnetwork.VirtualNetworkPeeringStateConnected {
		return nil, nil
	}

	logger.Info("Waiting for peering Connected state")

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}
