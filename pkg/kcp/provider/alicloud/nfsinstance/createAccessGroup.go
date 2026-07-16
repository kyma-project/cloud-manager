package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// createAccessGroup creates the NAS permission group if it does not already exist.
func createAccessGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.accessGroup != nil {
		return nil, ctx
	}

	logger.Info("Creating AliCloud NAS access group", "accessGroupName", state.accessGroupName)

	// AliCloud NAS rejects spaces in the access group description (IllegalCharacters),
	// so the description must be space-free.
	err := state.client.CreateAccessGroup(ctx, state.accessGroupName, "CloudManagerManaged")
	if err != nil {
		logger.Error(err, "Error creating AliCloud NAS access group")
		return composed.StopWithRequeue, ctx
	}

	// Requeue to reload the access group into state
	return composed.StopWithRequeue, ctx
}
