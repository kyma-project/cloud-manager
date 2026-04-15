package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	gcpnfsrestoreclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runNfsRestore(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsGcpNfsVolumeRestore()

	if restore.Status.OpIdentifier != "" {
		return nil, nil
	}

	if restore.Status.State == cloudresourcesv1beta1.JobStateFailed ||
		restore.Status.State == cloudresourcesv1beta1.JobStateDone {
		return nil, nil
	}

	logger.WithValues("NfsRestore", restore.Name).Info("Creating GCP File Restore")

	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	dstLocation := state.GcpNfsVolume.Status.Location

	nfsInstanceName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolume.Status.Id)
	dstFullPath := gcpnfsrestoreclientv2.GetFilestoreInstancePath(project, dstLocation, nfsInstanceName)
	dstFileShare := state.GcpNfsVolume.Spec.FileShareName

	op, err := state.fileRestoreClient.RestoreFilestoreInstance(ctx, &filestorepb.RestoreInstanceRequest{
		Name:      dstFullPath,
		FileShare: dstFileShare,
		Source: &filestorepb.RestoreInstanceRequest_SourceBackup{
			SourceBackup: state.SrcBackupFullPath,
		},
	})

	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error submitting Filestore restore request in GCP :%s. Retrying", err)).
			Run(ctx, state)
	}

	if op != nil {
		restore.Status.OpIdentifier = op.Name()
		restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
		return composed.PatchStatus(restore).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
			Run(ctx, state)
	}
	return nil, nil
}
