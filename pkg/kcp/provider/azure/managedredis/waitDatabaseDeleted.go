package managedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitDatabaseDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	_, err := state.client.GetDatabase(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error polling Azure Managed Redis database deletion", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
