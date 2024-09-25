package backupschedule

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func evaluateNextRun(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info("Evaluating next run time")

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
			SuccessLogMsg("BackupSchedule has not next run times calculated.").
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
	if nextRunTime.After(now) {
		timeLeft := nextRunTime.Unix() - now.Unix()
		logger.WithValues("GcpNfsBackupSchedule", schedule.GetName()).Info(fmt.Sprintf("Next Run in : %d seconds", timeLeft))
		return composed.StopWithRequeueDelay(time.Duration(timeLeft) * time.Second), nil
	}

	//If create and delete tasks already completed for currentRun, reset the next run times
	if schedule.GetLastCreateRun() != nil && nextRunTime.Equal(schedule.GetLastCreateRun().Time) &&
		schedule.GetLastDeleteRun() != nil && nextRunTime.Equal(schedule.GetLastDeleteRun().Time) {
		schedule.SetNextRunTimes(nil)
		return composed.PatchStatus(schedule).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	//Set the next run time in the state object, and continue to next task
	state.nextRunTime = nextRunTime
	return nil, nil
}
