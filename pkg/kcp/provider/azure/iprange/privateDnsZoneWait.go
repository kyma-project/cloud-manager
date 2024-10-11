package iprange

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func privateDnsZoneWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.privateDnsZone != nil && state.privateDnsZone.Properties != nil &&
		ptr.Deref(state.privateDnsZone.Properties.ProvisioningState, "") == armprivatedns.ProvisioningStateSucceeded {
		return nil, ctx
	}

	logger.Info("Waiting for Azure KCP IpRange private DNS zone to become available")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
