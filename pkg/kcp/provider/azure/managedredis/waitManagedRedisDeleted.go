package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitManagedRedisDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	_, err := state.client.GetCluster(ctx, state.resourceGroupName, obj.Name)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
