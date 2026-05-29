package managedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitPrivateDnsZoneGroupDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	dzg, err := state.client.GetPrivateDnsZoneGroup(ctx, state.resourceGroupName, obj.Name+"-pe", obj.Name+"-dzg")
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error polling Azure Managed Redis Private DNS Zone Group deletion", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if dzg == nil {
		return nil, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
