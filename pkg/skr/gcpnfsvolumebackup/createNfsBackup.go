package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/file/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
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

	// Setting the uuid as id to prevent duplicate backups if updateStatus fails.
	if backup.Status.Id == "" {
		location, err := getLocation(state, logger)
		if err != nil {
			return composed.PatchStatus(backup).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonLocationInvalid,
					Message: fmt.Sprintf("Could not automatically populate the location for the backup: %s", err),
				}).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
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

	fileBackup := &file.Backup{
		SourceFileShare:    state.GcpNfsVolume.Spec.FileShareName,
		SourceInstance:     client.GetFilestoreInstancePath(project, state.GcpNfsVolume.Status.Location, nfsInstanceName),
		SourceInstanceTier: string(state.GcpNfsVolume.Spec.Tier),
	}
	op, err := state.fileBackupClient.CreateFileBackup(ctx, project, backup.Status.Location, name, fileBackup)

	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error creating Filestore backup in GCP :%s", err)).
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupCreating
	backup.Status.OpIdentifier = op.Name
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		// Give some time for backup to get created.
		SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
		Run(ctx, state)
}

func getLocation(state *State, logger logr.Logger) (string, error) {
	location := state.ObjAsGcpNfsVolumeBackup().Spec.Location
	// location is optional. So if empty, using region from source volume.
	if len(location) != 0 {
		return location, nil
	}
	if (state.GcpNfsVolume == nil) || (state.GcpNfsVolume.Status.Location == "") {
		logger.Error(nil, "Source GcpNfsVolume location is empty")
		return "", fmt.Errorf("source GcpNfsVolume location is empty")
	}
	switch state.GcpNfsVolume.Spec.Tier {
	case cloudresourcesv1beta1.ENTERPRISE, cloudresourcesv1beta1.REGIONAL:
		return state.GcpNfsVolume.Status.Location, nil
	default:
		return getRegion(state.GcpNfsVolume.Status.Location), nil
	}
}

func getRegion(zone string) string {
	//slit zone by the last '-'
	return zone[:strings.LastIndex(zone, "-")]
}
