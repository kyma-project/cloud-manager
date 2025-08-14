package awsnfsvolumebackup

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setIdempotencyToken(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	backup := state.ObjAsAwsNfsVolumeBackup()
	logger := composed.LoggerFromCtx(ctx)

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, nil
	}

	//If the idempotency token is not empty, continue.
	if len(strings.TrimSpace(backup.Status.IdempotencyToken)) > 0 {
		return nil, nil
	}

	logger.WithValues("AwsBackup", backup.Name).Info("Setting the Idempotency Token")
	backup.Status.IdempotencyToken = uuid.NewString()
	return composed.PatchStatus(state.ObjAsAwsNfsVolumeBackup()).
		ErrorLogMessage("Failed to set Idempotency Token").
		SuccessLogMsg("Set the idempotency token").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
