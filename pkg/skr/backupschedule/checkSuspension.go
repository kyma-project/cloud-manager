package backupschedule

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func CheckSuspension(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(ScheduleState)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	//If the schedule is not suspended, continue
	if !schedule.GetSuspend() {
		return nil, nil
	}

	//BackupSchedule is suspended, stop reconciliation
	logger.Info("BackupSchedule is suspended. Stopping reconciliation.")
	schedule.SetState(cloudresourcesv1beta1.JobStateSuspended)
	schedule.SetNextRunTimes(nil)
	schedule.SetNextDeleteTimes(nil)
	return composed.PatchStatus(schedule).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
