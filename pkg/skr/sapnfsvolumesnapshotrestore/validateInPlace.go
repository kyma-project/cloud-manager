package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateInPlace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	// Validate the snapshot belongs to the destination volume's share
	if state.SourceSnapshot.Status.ShareId != state.shareId {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(
			fmt.Errorf("snapshot share ID %s does not match destination share ID %s", state.SourceSnapshot.Status.ShareId, state.shareId),
			"Snapshot does not belong to destination volume",
		)
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
				Message: "Snapshot does not belong to the destination volume",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Validate the snapshot is the most recent for this share
	allSnapshots, err := state.snapshotClient.ListSnapshots(ctx, snapshots.ListOpts{
		ShareID: state.shareId,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing snapshots for validation", composed.StopWithRequeue, ctx)
	}

	// Find the most recent snapshot by created_at
	var mostRecentTime time.Time
	var mostRecentID string
	for _, snap := range allSnapshots {
		snapTime := time.Time(snap.CreatedAt)
		if snapTime.After(mostRecentTime) {
			mostRecentTime = snapTime
			mostRecentID = snap.ID
		}
	}

	if mostRecentID != state.SourceSnapshot.Status.OpenstackId {
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(
			fmt.Errorf("snapshot %s is not the most recent snapshot for share %s", state.SourceSnapshot.Status.OpenstackId, state.shareId),
			"Snapshot is not the most recent",
		)
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
				Message: "Snapshot is not the most recent snapshot of the volume. Only the most recent snapshot can be used for in-place revert.",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	return nil, ctx
}
