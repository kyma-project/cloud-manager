package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// deleteAccessGroup deletes the NAS permission group after the file system is gone.
func deleteAccessGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.accessGroup == nil {
		return nil, ctx
	}

	logger.Info("Deleting AliCloud NAS access group", "accessGroupName", state.accessGroupName)

	if err := state.client.DeleteAccessGroup(ctx, state.accessGroupName); err != nil {
		logger.Error(err, "Error deleting AliCloud NAS access group", "accessGroupName", state.accessGroupName)
		return composed.StopWithRequeue, ctx
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
