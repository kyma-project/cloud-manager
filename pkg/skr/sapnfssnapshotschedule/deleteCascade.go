package sapnfssnapshotschedule

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteCascade(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	// If deleteCascade is not true, return
	if !schedule.GetDeleteCascade() {
		return nil, ctx
	}

	// If the list of snapshots is empty, continue
	if len(state.Snapshots) == 0 {
		return nil, ctx
	}

	logger.Info("Cascade delete of created snapshots.")

	for _, snapshot := range state.Snapshots {
		if composed.IsMarkedForDeletion(snapshot) {
			logger.WithValues("Snapshot", snapshot.GetName()).Info("Snapshot is already being deleted.")
			continue
		}
		logger.WithValues("Snapshot", snapshot.GetName()).Info("Deleting snapshot object")
		err := state.Cluster().K8sClient().Delete(ctx, snapshot)
		if err != nil {
			logger.Error(err, "Error deleting the snapshot object.")
			continue
		}
	}

	schedule.SetState(cloudresourcesv1beta1.StateDeleting)
	return composed.PatchStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
