package sapnfsvolumesnapshot

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func ttlExpiry(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// Only check TTL for non-deletion, ready snapshots
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if snapshot.Status.State != cloudresourcesv1beta1.StateReady {
		return nil, ctx
	}

	if snapshot.Spec.DeleteAfterDays <= 0 {
		return nil, ctx
	}

	creationTime := snapshot.CreationTimestamp.Time
	ttl := time.Duration(snapshot.Spec.DeleteAfterDays) * 24 * time.Hour
	expiryTime := creationTime.Add(ttl)

	if time.Now().Before(expiryTime) {
		return nil, ctx
	}

	// TTL expired - trigger deletion
	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Snapshot TTL expired, triggering deletion", "deleteAfterDays", snapshot.Spec.DeleteAfterDays)

	err := state.Cluster().K8sClient().Delete(ctx, snapshot)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting expired snapshot", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, ctx
}
