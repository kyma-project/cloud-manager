package backupschedule

import (
	"context"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func calculateOnetimeSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)
	now := time.Now()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If not one-time schedule, continue
	if schedule.GetSchedule() != "" {
		return nil, nil
	}

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("Evaluating one-time schedule")

	//If the nextRunTime is already set, continue
	if len(schedule.GetNextRunTimes()) > 0 {
		logger.WithValues("BackupSchedule", schedule.GetName()).Info("Next RunTime is already set, continuing.")
		return nil, nil
	}

	logger.WithValues("BackupSchedule", schedule.GetName()).Info("BackupSchedule is empty and scheduling it to run.")

	//Set the next run time to the start time if it is set
	var nextRunTime time.Time
	if schedule.GetStartTime() != nil && !schedule.GetStartTime().IsZero() {
		nextRunTime = schedule.GetStartTime().Time
	} else {
		nextRunTime = now
	}

	schedule.SetState(cloudresourcesv1beta1.JobStateActive)
	schedule.SetNextRunTimes([]string{nextRunTime.UTC().Format(time.RFC3339)})

	return composed.PatchStatus(schedule.(composed.ObjWithConditions)).
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
