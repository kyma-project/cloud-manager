package iprange

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func virtualNetworkLinkWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink != nil && state.virtualNetworkLink.Properties != nil &&
		ptr.Deref(state.virtualNetworkLink.Properties.ProvisioningState, "") == armprivatedns.ProvisioningStateSucceeded {
		return nil, ctx
	}

	logger.Info("Waiting for Azure KCP IpRange virtual network link to become available")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
