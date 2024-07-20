package backupschedule

import (
	"context"
	"github.com/gorhill/cronexpr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateSchedule(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)
	schedule := state.ObjAsBackupSchedule()

	//If marked for deletion, continue
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	ctx = composed.LoggerIntoCtx(ctx, logger)
	logger.Info("Validating BackupSchedule - Cron Expression")

	sch := schedule.GetSchedule()

	//If schedule is empty, continue
	if sch == "" {
		return nil, nil
	}
	expr, err := cronexpr.Parse(sch)

	if err != nil {
		logger.Info("Invalid cron expression")

		schedule.SetState(cloudresourcesv1beta1.JobStateError)
		return composed.UpdateStatus(schedule).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ReasonInvalidCronExpression,
				Message: err.Error(),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.Info("Validated Cron Expression")

	state.cronExpression = expr

	return nil, ctx
}
