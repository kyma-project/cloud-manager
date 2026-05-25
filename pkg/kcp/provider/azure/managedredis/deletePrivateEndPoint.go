package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deletePrivateEndPoint(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.privateEndpoint == nil {
		return nil, ctx
	}

	err := state.client.DeletePrivateEndPoint(ctx, state.resourceGroupName, obj.Name+"-pe")
	if err != nil && !azuremeta.IsNotFound(err) {
		composed.LoggerFromCtx(ctx).Error(err, "Error deleting Private Endpoint for Azure Managed Redis")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return nil, ctx
}
