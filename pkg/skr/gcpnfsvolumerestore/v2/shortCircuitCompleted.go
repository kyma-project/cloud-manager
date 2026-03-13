package v2

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// If deletion, continue to let delete branch handle cleanup.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	restore := state.ObjAsGcpNfsVolumeRestore()
	restoreState := restore.Status.State

	if restoreState == cloudresourcesv1beta1.JobStateDone || restoreState == cloudresourcesv1beta1.JobStateFailed {
		composed.LoggerFromCtx(ctx).Info("NfsVolumeRestore is completed, short-circuiting into StopAndForget")
		return composed.StopAndForget, nil
	}

	return nil, ctx
}
