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
	schedule := state.ObjAsNfsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If one-time schedule, continue
	if schedule.Spec.Schedule == "" {
		return nil, nil
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Evaluating next run time")

	//If cron expression has not changed, and the nextRunTime is already set, continue
	if schedule.Spec.Schedule == schedule.Status.Schedule && len(schedule.Status.NextRunTimes) > 0 {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	//Evaluate next run times.
	var nextRunTimes []time.Time
	if schedule.Spec.StartTime != nil && !schedule.Spec.StartTime.IsZero() && schedule.Spec.StartTime.Time.After(now) {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("StateTime is in future.")
		nextRunTimes = state.cronExpression.NextN(schedule.Spec.StartTime.Time, MaxSchedules)
	} else {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info(fmt.Sprintf("Using current time  %s", now))
		nextRunTimes = state.cronExpression.NextN(now, MaxSchedules)
	}
	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info(fmt.Sprintf("Next RunTime is %v", nextRunTimes[0]))

	//If the next run time is after the end time, stop reconciliation
	if schedule.Spec.EndTime != nil && !schedule.Spec.EndTime.IsZero() && nextRunTimes[0].After(schedule.Spec.EndTime.Time) {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Next RunTime is after the EndTime. Stopping reconciliation.")
		schedule.Status.State = cloudresourcesv1beta1.JobStateDone
		schedule.Status.NextRunTimes = nil
		return composed.UpdateStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	//Update the status of the schedule with the next run times
	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Next RunTime is set. Updating status.")
	schedule.Status.State = cloudresourcesv1beta1.JobStateActive
	schedule.Status.Schedule = schedule.Spec.Schedule
	schedule.Status.NextRunTimes = []string{}
	for _, t := range nextRunTimes {
		schedule.Status.NextRunTimes = append(schedule.Status.NextRunTimes, t.UTC().Format(time.RFC3339))
	}
	return composed.UpdateStatus(schedule).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
