package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2/openstack/sharedfilesystems/v2/snapshots"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func snapshotCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If snapshot already exists in Manila, skip
	if state.snapshot != nil {
		return nil, ctx
	}

	// ShareId must be resolved at this point
	if snapshot.Status.ShareId == "" {
		return composed.LogErrorAndReturn(
			fmt.Errorf("missing shareId in snapshot status"),
			"Logical error: shareId not set before creating snapshot",
			composed.StopWithRequeue,
			ctx,
		)
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Creating Manila snapshot", "name", snapshot.Status.Id, "shareId", snapshot.Status.ShareId)

	manilaSnapshot, err := state.snapshotClient.CreateSnapshot(ctx, snapshots.CreateOpts{
		ShareID: snapshot.Status.ShareId,
		Name:    state.OpenStackSnapshotName(),
	})
	if err != nil {
		logger.Error(err, "Error creating Manila snapshot")
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: fmt.Sprintf("Error creating Manila snapshot: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Store the openstackId
	snapshot.Status.OpenstackId = manilaSnapshot.ID
	snapshot.Status.State = cloudresourcesv1beta1.StateCreating
	err = composed.PatchObjStatus(ctx, snapshot, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error patching snapshot with openstackId", composed.StopWithRequeue, ctx)
	}

	state.snapshot = manilaSnapshot
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
