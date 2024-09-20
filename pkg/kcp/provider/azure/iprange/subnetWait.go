package iprange

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func subnetWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil && ptr.Deref(state.subnet.Properties.ProvisioningState, "") == armnetwork.ProvisioningStateSucceeded {
		return nil, ctx
	}

	logger.Info("Waiting for Azure KCP IpRange subnet to become available")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
