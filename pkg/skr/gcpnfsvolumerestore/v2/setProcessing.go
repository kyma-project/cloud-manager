package v2

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsGcpNfsVolumeRestore()

	if restore.Status.State == cloudresourcesv1beta1.JobStateFailed ||
		restore.Status.State == cloudresourcesv1beta1.JobStateDone {
		return composed.StopAndForget, nil
	}

	if restore.Status.State == "" {
		logger.Info("Setting processing")
		restore.Status.State = cloudresourcesv1beta1.JobStateProcessing
		return composed.PatchStatus(restore).
			SetExclusiveConditions().
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, nil
}
