package awsnfsvolumerestore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkRestoreJob(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsAwsNfsVolumeRestore()

	if restore.Status.JobId == "" {
		return nil, nil
	}

	restoreJobOutput, err := state.awsClient.DescribeRestoreJob(ctx, restore.Status.JobId)

	if err != nil && !state.awsClient.IsNotFound(err) {
		// Don't reset JobId - retry with fixed delay
		// The job exists in AWS, we just can't describe it right now
		return composed.LogErrorAndReturn(err, "Error loading AWS restore Job", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting AWS Restore Job.").
			Run(ctx, state)
	}
	switch restoreJobOutput.Status {
	case types.RestoreJobStatusCompleted:
		restore.Status.State = cloudresourcesv1beta1.JobStateDone
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: "Restore Job completed successfully",
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Restore Job completed successfully.").
			Run(ctx, state)
	case types.RestoreJobStatusFailed:
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: fmt.Sprintf("Restore Job failed: %s", *restoreJobOutput.StatusMessage),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Restore Job failed.").
			Run(ctx, state)
	case types.RestoreJobStatusAborted:
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: fmt.Sprintf("Restore Job aborted: %s", *restoreJobOutput.StatusMessage),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Restore Job aborted.").
			Run(ctx, state)
	case types.RestoreJobStatusRunning, types.RestoreJobStatusPending:
		logger.Info("Restore Job hasn't completed yet.", "status", restoreJobOutput.Status)
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	default:
		return composed.LogErrorAndReturn(fmt.Errorf("unknown Restore Job status: %s", restoreJobOutput.Status), "unknown Restore Job status", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}
}
