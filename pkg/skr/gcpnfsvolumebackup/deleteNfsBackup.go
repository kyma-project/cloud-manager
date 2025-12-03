package gcpnfsvolumebackup

import (
	"context"
	"errors"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	//If the backup not already exists, return
	if state.fileBackup == nil {
		return nil, nil
	}

	//If not deleting, return.
	deleting := !state.Obj().GetDeletionTimestamp().IsZero()
	if !deleting {
		return nil, nil
	}

	//If a backup operation already exists, skip
	if backup.Status.OpIdentifier != "" {
		return nil, nil
	}

	logger.WithValues("NfsBackup name", backup.Name, "NfsBackup namespace", backup.Namespace).Info("Deleting GCP File Backup")

	//Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	op, err := state.fileBackupClient.DeleteFileBackup(ctx, project, location, name)

	if err != nil {
		// Handle 404 (not found) and 403 (invalid location or unauthorized) as successful deletion
		// This allows users to delete CRs with invalid locations
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 || e.Code == 403 {
				logger.WithValues("code", e.Code, "message", e.Message).
					Info("Backup not found or location is invalid, treating as successfully deleted")
				return nil, nil
			}
		}

		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error deleting File backup object in GCP :%s", err)).
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupDeleting
	backup.Status.OpIdentifier = op.Name
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessErrorNil().
		Run(ctx, state)
}
