package backupschedule

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func CheckCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(ScheduleState)
	schedule := state.ObjAsBackupSchedule()
	calc := state.GetScheduleCalculator()
	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Checking the State")

	//If the schedule is in Done state, stop reconciliation
	if schedule.State() == cloudresourcesv1beta1.JobStateDone {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info("Schedule already completed, stopping reconciliation.")
		return composed.PatchStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	now := calc.Now()
	endTime := schedule.GetEndTime()
	//If the current time is after the end time, stop reconciliation
	if endTime != nil && !endTime.IsZero() && endTime.Time.Before(now) {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info("Current Time is after the EndTime. Stopping reconciliation.")
		schedule.SetState(cloudresourcesv1beta1.JobStateDone)
		schedule.SetNextRunTimes(nil)
		return composed.PatchStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	//If it is a one-time schedule,
	//the job already ran, and the backups are already deleted.
	//stop reconciliation
	if schedule.GetSchedule() == "" &&
		schedule.GetLastCreateRun() != nil &&
		!schedule.GetLastCreateRun().IsZero() &&
		schedule.GetBackupCount() == 0 {

		logger.WithValues("BackupSchedule", schedule.GetName()).Info("One-time schedule already ran. Stopping reconciliation.")
		schedule.SetState(cloudresourcesv1beta1.JobStateDone)
		schedule.SetNextRunTimes(nil)
		return composed.PatchStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	//Continue otherwise
	return nil, nil
}
