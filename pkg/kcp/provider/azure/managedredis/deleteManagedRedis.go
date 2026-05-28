package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis == nil {
		return nil, ctx
	}

	composed.LoggerFromCtx(ctx).Info("Deleting Azure Managed Redis", "name", obj.Name)

	err := state.client.DeleteCluster(ctx, state.resourceGroupName, obj.Name)
	if err != nil && !azuremeta.IsNotFound(err) {
		composed.LoggerFromCtx(ctx).Error(err, "Error deleting Azure Managed Redis cluster")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, ctx
}
