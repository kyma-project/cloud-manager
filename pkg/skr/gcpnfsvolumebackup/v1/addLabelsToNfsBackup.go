package v1

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addLabelsToNfsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsGcpNfsVolumeBackup()

	//If the backup not already exists, return
	if state.fileBackup == nil {
		return nil, nil
	}

	//If deleting, return.
	if composed.IsMarkedForDeletion(state.Obj()) {
		return nil, nil
	}

	//If a backup is not ready, return as it might be unsafe to patch it
	if backup.Status.State != cloudresourcesv1beta1.GcpNfsBackupReady {
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

	logger.WithValues("NfsBackup name", backup.Name, "NfsBackup namespace", backup.Namespace).Info("Adding missing labels to GCP File Backup")

	//Get GCP details.
	project := state.Scope.Spec.Scope.Gcp.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	state.SetFilestoreLabels()

	_, err := state.fileBackupClient.PatchFileBackup(ctx, project, location, name, "Labels", state.fileBackup)

	if err != nil {
		// Log and retry. This is not an essential action as it only impacts nuke on non prod envs that might already
		// have backups without necessary labels.
		logger.Error(err, "Error adding missing labels to File backup object in GCP")
	}

	return composed.PatchStatus(backup).
		SuccessLogMsg("Updated accessibleFrom in status of GcpNfsVolumeBackup").
		SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpOperationWaitTime)).
		Run(ctx, state)
}
