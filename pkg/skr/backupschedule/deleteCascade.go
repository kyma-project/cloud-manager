package backupschedule

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

	//If not marked for deletion, return
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If deleteCascade is not true return
	if !schedule.GetDeleteCascade() {
		return nil, nil
	}

	//If the list of backups is empty, continue
	if len(state.Backups) == 0 {
		return nil, nil
	}

	logger.WithValues("GcpNfsBackupSchedule :", schedule.GetName()).Info("Cascade delete of created backups.")

	for _, backup := range state.Backups {
		logger.WithValues("Backup", backup.GetName()).Info("Deleting backup object")
		err := state.Cluster().K8sClient().Delete(ctx, backup)
		if err != nil {
			logger.Error(err, "Error deleting the backup object.")
			continue
		}
	}

	schedule.SetState(cloudresourcesv1beta1.StateDeleting)
	return composed.UpdateStatus(schedule).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
