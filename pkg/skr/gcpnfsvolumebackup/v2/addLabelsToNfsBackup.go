package v2

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addLabelsToNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	// If the backup not already exists, return
	if state.fileBackup == nil {
		return nil, nil
	}

	if state.HasProperLabels() {
		if backup.Status.AccessibleFrom != state.specCommaSeparatedAccessibleFrom() {
			backup.Status.AccessibleFrom = state.specCommaSeparatedAccessibleFrom()
			return composed.PatchStatus(backup).
				SuccessLogMsg("Updated accessibleFrom in status of GcpNfsVolumeBackup").
				Run(ctx, state)
		}
		return nil, nil
	}

	logger.Info("Adding missing labels to GCP File Backup")

	// Get GCP details.
	project := state.Scope.Spec.Scope.Gcp.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	state.SetFilestoreLabels()

	_, err := state.fileBackupClient.UpdateBackup(ctx, project, location, name, state.fileBackup, []string{"labels"})

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
			SuccessLogMsg(fmt.Sprintf("Error adding labels to File backup in GCP: %s", err)).
			Run(ctx, state)
	}

	return composed.PatchStatus(backup).
		SuccessLogMsg("Updated accessibleFrom in status of GcpNfsVolumeBackup").
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
		Run(ctx, state)
}
