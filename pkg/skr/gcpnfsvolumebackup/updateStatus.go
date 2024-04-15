package gcpnfsvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	backup := state.ObjAsGcpNfsVolumeBackup()

	logger := composed.LoggerFromCtx(ctx)

	//Handle if deletion is in progress
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()
	if deleting {
		//If the backup not already exists, return
		if state.fileBackup == nil {
			return composed.StopAndForget, nil
		} else {
			return composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), nil
		}
	}

	//Check if the backup has a ready condition
	backupReady := meta.FindStatusCondition(backup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)

	// Requeue to check the status of the backup
	if state.fileBackup == nil || state.fileBackup.State != "READY" {
		return composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), nil
	}

	//If file backup is ready, update the status of the backup
	if backupReady != nil && backupReady.Status == metav1.ConditionTrue {
		// already with Ready condition
		return composed.StopAndForget, nil
	} else {
		logger.Info("Updating SKR GcpNfsVolumeBackup status with Ready condition")
		return composed.UpdateStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeReady,
				Message: "Backup is ready for use.",
			}).
			SuccessLogMsg("GcpNfsVolumeBackup status got updated with Ready condition ").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}
}
