package managedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadPrivateDnsZone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	zone, err := state.client.GetPrivateDnsZone(ctx, state.resourceGroupName, state.PrivateDNSZoneName())
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading Azure Managed Redis private DNS zone", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.privateDnsZone = zone
	return nil, ctx
}
