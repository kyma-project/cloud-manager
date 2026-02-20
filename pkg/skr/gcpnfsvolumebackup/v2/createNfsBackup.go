package v2

import (
	"context"
	"fmt"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	v2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	// If the backup already exists, return
	if state.fileBackup != nil {
		return nil, nil
	}

	// If a backup operation already exists, skip
	if backup.Status.OpIdentifier != "" {
		return nil, nil
	}

	logger.Info("Creating GCP File Backup")

	// Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project

	// Setting the uuid as id to prevent duplicate backups if updateStatus fails.
	if backup.Status.Id == "" {
		location := getLocation(state)
		backup.Status.Location = location
		backup.Status.Id = uuid.NewString()
		return composed.PatchStatus(backup).
			SetExclusiveConditions().
			// Give some time for backup to get created.
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)
	nfsInstanceName := fmt.Sprintf("cm-%.60s", state.GcpNfsVolume.Status.Id)

	fileBackup := &filestorepb.Backup{
		SourceFileShare: state.GcpNfsVolume.Spec.FileShareName,
		SourceInstance:  v2client.GetFilestoreInstancePath(project, state.GcpNfsVolume.Status.Location, nfsInstanceName),
		Labels:          map[string]string{gcpclient.ManagedByKey: gcpclient.ManagedByValue, gcpclient.ScopeNameKey: state.Scope.Name},
	}
	opName, err := state.fileBackupClient.CreateBackup(ctx, project, backup.Status.Location, name, fileBackup)

	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating Filestore backup in GCP :%s", err)).
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupCreating
	backup.Status.OpIdentifier = opName
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		// Give some time for backup to get created.
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
		Run(ctx, state)
}

func getLocation(state *State) string {
	location := state.ObjAsGcpNfsVolumeBackup().Spec.Location
	if len(location) != 0 {
		return location
	}
	return state.Scope.Spec.Region
}
