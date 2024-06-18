package nfsbackupschedule

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkSuspension(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	schedule := state.ObjAsNfsBackupSchedule()
	logger := composed.LoggerFromCtx(ctx)

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	//If the schedule is not suspended, continue
	if !schedule.Spec.Suspend {
		return nil, nil
	}

	//Schedule is suspended, stop reconciliation
	logger.WithValues("NfsBackupSchedule :", schedule.Name).Info("Schedule is suspended. Stopping reconciliation.")
	schedule.Status.State = cloudresourcesv1beta1.JobStateSuspended
	schedule.Status.NextRunTimes = nil
	return composed.UpdateStatus(schedule).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
