package sapnfsvolumesnapshot

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
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
		// If TTL is set, requeue to check expiry later
		if snapshot.Spec.DeleteAfterDays > 0 {
			ttl := time.Duration(snapshot.Spec.DeleteAfterDays) * 24 * time.Hour
			remaining := snapshot.CreationTimestamp.Time.Add(ttl).Sub(state.clock.Now())
			if remaining > 0 {
				delay := min(remaining, util.Timing.T300000ms())
				return composed.StopWithRequeueDelay(delay), ctx
			}
			// TTL already expired, continue to ttlExpiry action
			return nil, ctx
		}
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
