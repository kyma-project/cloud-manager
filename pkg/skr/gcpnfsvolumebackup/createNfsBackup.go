package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	//If the backup already exists, return
	if state.fileBackup != nil {
		return nil, nil
	}

	//If a backup operation already exists, skip
	if backup.Status.OpIdentifier != "" {
		return nil, nil
	}

	//If deleting, return.
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()
	if deleting {
		return nil, nil
	}

	logger.WithValues("NfsBackup :", backup.Name).Info("Creating GCP File Backup")

	//Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	location := backup.Spec.Location
	id := uuid.NewString()
	name := fmt.Sprintf("cm-%.60s", id)
	nfsInstanceName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolume.Status.Id)

	fileBackup := &file.Backup{
		SourceFileShare:    state.GcpNfsVolume.Spec.FileShareName,
		SourceInstance:     client.GetFilestoreInstancePath(project, state.GcpNfsVolume.Spec.Location, nfsInstanceName),
		SourceInstanceTier: string(state.GcpNfsVolume.Spec.Tier),
	}
	op, err := state.fileBackupClient.CreateFileBackup(ctx, project, location, name, fileBackup)

	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.UpdateStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating Filestore backup in GCP :%s", err)).
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupCreating
	backup.Status.OpIdentifier = op.Name
	backup.Status.Id = id
	return composed.UpdateStatus(backup).
		SetExclusiveConditions().
		// Give some time for backup to get created.
		SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
		Run(ctx, state)
}
