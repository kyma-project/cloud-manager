package nfsbackupschedule

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func evaluateNextRun(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If it is a one-time schedule, and the job already ran, stop reconciliation
	if schedule.GetSchedule() == "" &&
		schedule.GetLastCreateRun() != nil &&
		!schedule.GetLastCreateRun().IsZero() {

		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("One-time schedule already ran. Stopping reconciliation.")
		schedule.SetState(cloudresourcesv1beta1.JobStateDone)
		schedule.SetNextRunTimes(nil)
		return composed.UpdateStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Evaluating next run time")

	if len(schedule.GetNextRunTimes()) == 0 {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonScheduleError,
				Message: "Schedule has no next run time",
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(fmt.Sprintf("Schedule has not next run times calculated.")).
			Run(ctx, state)
	}

	//Get the next run time
	nextRunTime, err := time.Parse(time.RFC3339, schedule.GetNextRunTimes()[0])

	if err != nil {
		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
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

	//If the next run time is after the end time, stop reconciliation
	if schedule.GetEndTime() != nil && !schedule.GetEndTime().IsZero() &&
		nextRunTime.After(schedule.GetEndTime().Time) {
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info("Next RunTime is after the EndTime. Stopping reconciliation.")
		schedule.SetState(cloudresourcesv1beta1.JobStateDone)
		schedule.SetNextRunTimes(nil)
		return composed.UpdateStatus(schedule).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	//If it still not time to run, reconcile with delay
	if nextRunTime.After(now) {
		timeLeft := nextRunTime.Unix() - now.Unix()
		logger.WithValues("NfsBackupSchedule :", schedule.GetName()).Info(fmt.Sprintf("Next Run in : %d seconds", timeLeft))
		return composed.StopWithRequeueDelay(time.Duration(timeLeft) * time.Second), nil
	}

	//If create and delete tasks already completed for currentRun, reset the next run times
	if schedule.GetLastCreateRun() != nil && nextRunTime.Equal(schedule.GetLastCreateRun().Time) &&
		schedule.GetLastDeleteRun() != nil && nextRunTime.Equal(schedule.GetLastDeleteRun().Time) {
		schedule.SetNextRunTimes(nil)
		return composed.UpdateStatus(schedule).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//Set the next run time in the state object, and continue to next task
	state.nextRunTime = nextRunTime
	return nil, nil
}
