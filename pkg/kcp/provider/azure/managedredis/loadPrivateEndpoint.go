package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadPrivateEndpoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	pe, err := state.client.GetPrivateEndPoint(ctx, state.resourceGroupName, obj.Name+"-pe")
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading Azure Managed Redis private endpoint", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.privateEndpoint = pe
	return nil, ctx
}
