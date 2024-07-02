package nfsbackupschedule

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func calculateOnetimeSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsNfsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If not one-time schedule, continue
	if schedule.Spec.Schedule != "" {
		return nil, nil
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Evaluating one-time schedule")

	//If the nextRunTime is already set, continue
	if len(schedule.Status.NextRunTimes) > 0 {
		logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Schedule is empty and scheduling it to run.")

	//Set the next run time to the start time if it is set
	var nextRunTime time.Time
	if schedule.Spec.StartTime != nil && !schedule.Spec.StartTime.IsZero() {
		nextRunTime = schedule.Spec.StartTime.Time
	} else {
		nextRunTime = now
	}

	schedule.Status.State = cloudresourcesv1beta1.JobStateActive
	schedule.Status.NextRunTimes = []string{nextRunTime.Format(time.RFC3339)}

	return composed.UpdateStatus(schedule).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
