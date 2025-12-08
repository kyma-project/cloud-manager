package gcpnfsvolumebackup

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/googleapi"
)

func loadNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backup := state.ObjAsGcpNfsVolumeBackup()
	logger.WithValues("nfsBackup", backup.Name).Info("Loading GCP FileBackup")

	if backup.Status.Id == "" {
		// Backup is not created yet.
		return nil, nil
	}

	//Get GCP details.
	gcpScope := state.Scope.Spec.Scope.Gcp
	project := gcpScope.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	bkup, err := state.fileBackupClient.GetFileBackup(ctx, project, location, name)
	if err != nil {

		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				state.fileBackup = nil
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
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			SuccessLogMsg(fmt.Sprintf("Error getting GCP backup : %s", err)).
			Run(ctx, state)
	}

	//Store the file backup in state
	state.fileBackup = bkup

	return nil, nil
}
