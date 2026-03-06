package backupschedule

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func CalculateRecurringSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(ScheduleState)
	schedule := state.ObjAsBackupSchedule()
	calc := state.GetScheduleCalculator()
	logger := composed.LoggerFromCtx(ctx)

	//If one-time schedule, continue
	if schedule.GetSchedule() == "" {
		return nil, nil
	}

	logger.Info("Evaluating next run time")

	//If cron expression has not changed, and the nextRunTime is already set, continue
	if schedule.GetSchedule() == schedule.GetActiveSchedule() && len(schedule.GetNextRunTimes()) > 0 {
		logger.Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	//Evaluate next run times using the calculator
	expr := state.GetCronExpression()
	var startTime *time.Time
	if schedule.GetStartTime() != nil && !schedule.GetStartTime().IsZero() {
		t := schedule.GetStartTime().Time
		startTime = &t
	}
	nextRunTimes := calc.NextRunTimes(expr, startTime, MaxSchedules)

	logger.Info(fmt.Sprintf("Next RunTime is %v", nextRunTimes[0]))

	//Update the status of the schedule with the next run times
	logger.Info("Next RunTime is set. Updating status.")
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
