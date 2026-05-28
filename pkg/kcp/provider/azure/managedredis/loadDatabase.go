package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadDatabase(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	db, err := state.client.GetDatabase(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading Azure Managed Redis database", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.managedRedisDatabase = db
	return nil, ctx
}
