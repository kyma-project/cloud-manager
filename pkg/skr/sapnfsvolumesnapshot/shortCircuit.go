package sapnfsvolumesnapshot

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuit(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// If deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If failed, stop
	if snapshot.Status.State == cloudresourcesv1beta1.StateFailed {
		return composed.StopAndForget, ctx
	}

	// If ready and not being deleted, stop
	if snapshot.Status.State == cloudresourcesv1beta1.StateReady {
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
