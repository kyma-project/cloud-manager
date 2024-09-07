package gcpnfsvolumerestore

import (
	"context"
	"fmt"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runNfsRestore(ctx context.Context, st composed.State) (error, context.Context) {
	//If deleting, continue with next steps.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsGcpNfsVolumeRestore()

	//If the restore operation already exists, skip
	if restore.Status.OpIdentifier != "" {
		return nil, nil
	}

	// If the status is FAILED or DONE, i.e. completed, skip
	if restore.Status.State == cloudresourcesv1beta1.JobStateFailed ||
		restore.Status.State == cloudresourcesv1beta1.JobStateDone {
		return nil, nil
	}

	logger.WithValues("NfsRestore :", restore.Name).Info("Creating GCP File Restore")

	//Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	srcLocation := state.GcpNfsVolumeBackup.Status.Location
	dstLocation := state.GcpNfsVolume.Status.Location
	backupName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolumeBackup.Status.Id)

	nfsInstanceName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolume.Status.Id)
	dstFullPath := client.GetFilestoreInstancePath(project, dstLocation, nfsInstanceName)
	dstFileShare := state.GcpNfsVolume.Spec.FileShareName
	srcFullPath := client.GetFileBackupPath(project, srcLocation, backupName)

	operation, err := state.fileRestoreClient.RestoreFile(ctx, project, dstFullPath, dstFileShare, srcFullPath)

	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error submitting Filestore restore request in GCP :%s. Retrying", err)).
			Run(ctx, state)
	}
	if operation != nil {
		restore.Status.OpIdentifier = operation.Name
		restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
		return composed.PatchStatus(restore).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return nil, nil
}
