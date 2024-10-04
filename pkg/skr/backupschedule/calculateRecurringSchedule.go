package backupschedule

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

const (
	MaxSchedules = 3
)

func calculateRecurringSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
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

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Evaluating next run time")

	//If cron expression has not changed, and the nextRunTime is already set, continue
	if schedule.GetSchedule() == schedule.GetActiveSchedule() && len(schedule.GetNextRunTimes()) > 0 {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	//Evaluate next run times.
	var nextRunTimes []time.Time
	if schedule.GetStartTime() != nil && !schedule.GetStartTime().IsZero() && schedule.GetStartTime().Time.After(now) {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info("StartTime is in future, using it.")
		nextRunTimes = state.cronExpression.NextN(schedule.GetStartTime().Time.UTC(), MaxSchedules)
	} else {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Using current time  %s", now))
		nextRunTimes = state.cronExpression.NextN(now.UTC(), MaxSchedules)
	}
	logger.WithValues("BackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Next RunTime is %v", nextRunTimes[0]))

	//Update the status of the schedule with the next run times
	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Next RunTime is set. Updating status.")
	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetActiveSchedule(schedule.GetSchedule())

	var runtimes []string
	for _, t := range nextRunTimes {
		runtimes = append(runtimes, t.UTC().Format(time.RFC3339))
	}
	schedule.SetNextRunTimes(runtimes)

	return composed.PatchStatus(schedule).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
