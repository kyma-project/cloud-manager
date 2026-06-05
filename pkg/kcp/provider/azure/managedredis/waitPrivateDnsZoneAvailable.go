package managedredis

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitPrivateDnsZoneAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.privateDnsZone != nil &&
		state.privateDnsZone.Properties != nil &&
		ptr.Deref(state.privateDnsZone.Properties.ProvisioningState, "") == armprivatedns.ProvisioningStateSucceeded {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Waiting for Azure Managed Redis Private DNS Zone to become available")
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
