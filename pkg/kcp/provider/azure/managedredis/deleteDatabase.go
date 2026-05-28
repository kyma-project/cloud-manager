package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteDatabase(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedisDatabase == nil {
		return nil, ctx
	}

	err := state.client.DeleteDatabase(ctx, state.resourceGroupName, obj.Name, DefaultDatabaseName)
	if err != nil && !azuremeta.IsNotFound(err) {
		composed.LoggerFromCtx(ctx).Error(err, "Error deleting Azure Managed Redis database")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, ctx
}
