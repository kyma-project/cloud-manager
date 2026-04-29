package sapnfsvolumesnapshot

import (
	"context"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func snapshotLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If openstackId is set, fetch by ID (fast path)
	if snapshot.Status.OpenstackId != "" {
		manilaSnapshot, err := state.snapshotClient.GetSnapshot(ctx, snapshot.Status.OpenstackId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error getting Manila snapshot by ID", composed.StopWithRequeue, ctx)
		}
		// GetSnapshot returns nil if not found (404)
		state.snapshot = manilaSnapshot
		return nil, ctx
	}

	// Fallback: if we have an internal ID (status.id), search by name+shareId
	if snapshot.Status.Id != "" && snapshot.Status.ShareId != "" {
		list, err := state.snapshotClient.ListSnapshots(ctx, snapshots.ListOpts{
			Name:    state.OpenStackSnapshotName(),
			ShareID: snapshot.Status.ShareId,
		})
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error listing Manila snapshots by name", composed.StopWithRequeue, ctx)
		}

		if len(list) > 0 {
			found := &list[0]

			// Persist the openstackId for fast path on subsequent reconciliations
			snapshot.Status.OpenstackId = found.ID
			err = composed.PatchObjStatus(ctx, snapshot, state.Cluster().K8sClient())
			if err != nil {
				return composed.LogErrorAndReturn(err, "Error patching snapshot with openstackId", composed.StopWithRequeue, ctx)
			}

			state.snapshot = found
			return nil, ctx
		}
	}

	return nil, ctx
}
