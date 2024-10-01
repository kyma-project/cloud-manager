package awsnfsvolumerestore

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setIdempotencyToken(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAwsNfsVolumeRestore()
	logger := composed.LoggerFromCtx(ctx)

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(restore) {
		return nil, nil
	}

	//If the idempotency token is not empty, continue.
	if len(strings.TrimSpace(restore.Status.IdempotencyToken)) > 0 {
		return nil, nil
	}

	logger.WithValues("AwsRestore", restore.Name).Info("Setting the Idempotency Token")
	restore.Status.IdempotencyToken = uuid.NewString()
	return composed.PatchStatus(restore).
		ErrorLogMessage("Failed to set Idempotency Token").
		SuccessLogMsg("Set the idempotency token").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
