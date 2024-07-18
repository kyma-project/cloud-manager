package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

const (
	MaxSchedules = 3
)

func calculateRecurringSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If one-time schedule, continue
	if schedule.GetSchedule() == "" {
		return nil, nil
	}

	logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Evaluating next run time")

	//If cron expression has not changed, and the nextRunTime is already set, continue
	if schedule.GetSchedule() == schedule.GetActiveSchedule() && len(schedule.GetNextRunTimes()) > 0 {
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	//Evaluate next run times.
	var nextRunTimes []time.Time
	if schedule.GetStartTime() != nil && !schedule.GetStartTime().IsZero() && schedule.GetStartTime().Time.After(now) {
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("StateTime is in future.")
		nextRunTimes = state.cronExpression.NextN(schedule.GetStartTime().Time, MaxSchedules)
	} else {
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info(fmt.Sprintf("Using current time  %s", now))
		nextRunTimes = state.cronExpression.NextN(now, MaxSchedules)
	}
	logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info(fmt.Sprintf("Next RunTime is %v", nextRunTimes[0]))

	//If the next run time is after the end time, stop reconciliation
	if schedule.GetEndTime() != nil && !schedule.GetEndTime().IsZero() && nextRunTimes[0].After(schedule.GetEndTime().Time) {
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Next RunTime is after the EndTime. Stopping reconciliation.")
		schedule.SetState(cloudresourcesv1beta1.JobStateDone)
		schedule.SetNextRunTimes(nil)
		return composed.UpdateStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	//Update the status of the schedule with the next run times
	logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Next RunTime is set. Updating status.")
	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetActiveSchedule(schedule.GetSchedule())

	var runtimes []string
	for _, t := range nextRunTimes {
		runtimes = append(runtimes, t.UTC().Format(time.RFC3339))
	}
	schedule.SetNextRunTimes(runtimes)

	return composed.UpdateStatus(schedule).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
