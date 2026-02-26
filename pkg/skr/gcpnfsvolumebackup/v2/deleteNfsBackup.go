package v2

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	// If the backup not already exists, return
	if state.fileBackup == nil {
		return nil, nil
	}

	// If a backup operation already exists, skip
	if backup.Status.OpIdentifier != "" {
		return nil, nil
	}

	logger.Info("Deleting GCP File Backup")

	// Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	opName, err := state.fileBackupClient.DeleteBackup(ctx, project, location, name)

	if err != nil {
		// If not found, the backup is already deleted
		if gcpmeta.IsNotFound(err) {
			state.fileBackup = nil
			return nil, nil
		}

		backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error deleting File backup object in GCP :%s", err)).
			Run(ctx, state)
	}

	backup.Status.State = cloudresourcesv1beta1.GcpNfsBackupDeleting
	backup.Status.OpIdentifier = opName
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessErrorNil().
		Run(ctx, state)
}
