package managedredis

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func updateStatusId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	obj := state.ObjAsAzureManagedRedis()

	if state.managedRedis == nil || state.managedRedis.ID == nil {
		return nil, ctx
	}

	if obj.Status.Id == *state.managedRedis.ID {
		return nil, ctx
	}

	obj.Status.Id = *state.managedRedis.ID
	return composed.UpdateStatus(obj).
		ErrorLogMessage("Error updating AzureManagedRedis status id").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, st)
}
