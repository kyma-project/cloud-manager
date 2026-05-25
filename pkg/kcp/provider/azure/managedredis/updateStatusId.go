package managedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if obj.Status.Id != "" {
		return nil, ctx
	}

	if state.managedRedis == nil || state.managedRedis.Name == nil {
		return nil, ctx
	}

	obj.Status.Id = *state.managedRedis.Name

	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating AzureManagedRedis status.id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
