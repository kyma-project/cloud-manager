package sapnfsvolumesnapshotrestore

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func restoreInPlace(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	// If revert was already initiated (persisted in status), skip calling revert again
	if restore.Status.RevertInitiated {
		return nil, ctx
	}

	// Guard against the case where revert succeeded but RevertInitiated was not persisted:
	// check actual share status before issuing revert.
	share, err := state.shareClient.GetShare(ctx, state.shareId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting share status before revert", composed.StopWithRequeue, ctx)
	}
	if share != nil && (share.Status == "reverting") {
		// Revert already in progress — persist the marker and skip
		restore.Status.RevertInitiated = true
		err = composed.PatchObjStatus(ctx, restore, state.Cluster().K8sClient())
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error patching restore status with revertInitiated", composed.StopWithRequeue, ctx)
		}
		return nil, ctx
	}

	err = state.snapshotClient.RevertShareToSnapshot(ctx, state.shareId, state.SourceSnapshot.Status.OpenstackId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error reverting share to snapshot", composed.StopWithRequeue, ctx)
	}

	// Persist the marker so we never call revert again on requeue
	restore.Status.RevertInitiated = true
	err = composed.PatchObjStatus(ctx, restore, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching restore status with revertInitiated", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
