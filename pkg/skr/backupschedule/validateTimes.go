package backupschedule

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func validateTimes(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	ctx = composed.LoggerIntoCtx(ctx, logger)
	logger.Info("Validating Start and End Times")

	start := schedule.GetStartTime()
	end := schedule.GetEndTime()

	//If no start and end time specified, continue
	if (start == nil || start.IsZero()) &&
		(end == nil || end.IsZero()) {
		return nil, nil
	}

	refTime := time.Now().UTC()
	createTime := schedule.GetCreationTimestamp()
	if !(&createTime).IsZero() {
		refTime = schedule.GetCreationTimestamp().Time
	}

	if start != nil && !start.IsZero() &&
		start.Time.Before(refTime) {
		logger.Info(fmt.Sprintf("Invalid start time : %s before %s", start, refTime))

		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonInvalidStartTime,
				Message: "Start time cannot be before creation time.",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if start != nil && !start.IsZero() {
		refTime = start.Time
	}

	if end != nil && !end.IsZero() && end.Time.Before(refTime) {
		logger.Info("Invalid end time")

		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonInvalidEndTime,
				Message: "End time cannot be before start/creation time.",
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.Info("Validated Start and End Times")

	return nil, ctx
}
