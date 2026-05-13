package sapnfsvolumesnapshotrestore

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	if composed.IsMarkedForDeletion(restore) {
		return nil, ctx
	}

	if restore.Status.State == cloudresourcesv1beta1.JobStateDone ||
		restore.Status.State == cloudresourcesv1beta1.JobStateFailed {
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
