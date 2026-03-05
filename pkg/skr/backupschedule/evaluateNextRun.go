package backupschedule

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EvaluateNextRun(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(ScheduleState)
	schedule := state.ObjAsBackupSchedule()
	calc := state.GetScheduleCalculator()
	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Evaluating next run time")

	if len(schedule.GetNextRunTimes()) == 0 {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonScheduleError,
				Message: "BackupSchedule has no next run time",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg("BackupSchedule has no next run times calculated.").
			Run(ctx, state)
	}

	//Get the next run time
	nextRunTime, err := time.Parse(time.RFC3339, schedule.GetNextRunTimes()[0])
	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.PatchStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonTimeParseError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(fmt.Sprintf("Error parsing next run time :%s", err)).
			Run(ctx, state)
	}

	//If it still not time to run, reconcile with delay
	if timeLeft := calc.GetRemainingTime(nextRunTime); timeLeft > 0 {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Next Run in : %s", timeLeft))
		return composed.StopWithRequeueDelay(timeLeft), nil
	}

	//Set the state attributes
	state.SetNextRunTime(nextRunTime)
	state.SetCreateRunCompleted(schedule.GetLastCreateRun() != nil && nextRunTime.Equal(schedule.GetLastCreateRun().Time))
	state.SetDeleteRunCompleted(schedule.GetLastDeleteRun() != nil && nextRunTime.Equal(schedule.GetLastDeleteRun().Time))

	//Mark createRunCompleted to true always after first run for one-time schedules.
	if schedule.GetSchedule() == "" && schedule.GetLastCreateRun() != nil && !schedule.GetLastCreateRun().IsZero() {
		state.SetCreateRunCompleted(true)
	}

	//If create and delete tasks already completed for currentRun, reset the next run times
	if state.IsCreateRunCompleted() && state.IsDeleteRunCompleted() {

		//If we still have some time to reach the actual nextRunTime, reconcile with delay.
		//It may happen if we used tolerance when comparing.
		if timeLeft := calc.GetRemainingTimeWithTolerance(nextRunTime, 0); timeLeft > 0 {
			logger.WithValues("BackupSchedule", schedule.GetName()).Info(
				fmt.Sprintf("Run already completed for %s. Requeueing with delay : %s", nextRunTime, timeLeft))
			return composed.StopWithRequeueDelay(timeLeft), nil
		}

		schedule.SetNextRunTimes(nil)
		return composed.PatchStatus(schedule).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//log details
	logger.WithValues("BackupSchedule", schedule.GetName()).Info(
		fmt.Sprintf("NextRunTimes: %s. CreateRunCompleted : %t, DeleteRunCompleted : %t",
			schedule.GetNextRunTimes(), state.IsCreateRunCompleted(), state.IsDeleteRunCompleted()))

	//continue to next task
	return nil, nil
}
