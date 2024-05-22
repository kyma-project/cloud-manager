package gcpnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkStates(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsGcpNfsVolumeRestore()

	//If deleting, continue with next steps.
	deleting := composed.IsMarkedForDeletion(state.Obj())
	if deleting {
		return nil, nil
	}

	logger.WithValues("NfsRestore :", restore.Name).Info("Checking States")

	// If the status is FAILED or DONE, i.e. completed, skip
	if restore.Status.State == cloudresourcesv1beta1.JobStateFailed ||
		restore.Status.State == cloudresourcesv1beta1.JobStateDone {
		return composed.StopAndForget, nil
	}

	//If state is empty or error, Reset the error conditions, and restart with processing state
	if restore.Status.State == "" || restore.Status.State == cloudresourcesv1beta1.JobStateError {
		restore.Status.State = cloudresourcesv1beta1.JobStateProcessing
		return composed.UpdateStatus(restore).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//continue to next action.
	return nil, nil
}
