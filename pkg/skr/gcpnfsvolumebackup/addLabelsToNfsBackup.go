package gcpnfsvolumebackup

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
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

	if state.fileBackup.Labels != nil && state.fileBackup.Labels[gcpclient.ManagedByKey] == gcpclient.ManagedByValue && state.fileBackup.Labels[gcpclient.ScopeNameKey] == state.Scope.Name {
		// Labels have been already set, return
		return nil, nil
	}

	logger.WithValues("NfsBackup name", backup.Name, "NfsBackup namespace", backup.Namespace).Info("Adding missing labels to GCP File Backup")

	//Get GCP details.
	project := state.Scope.Spec.Scope.Gcp.Project
	location := backup.Status.Location
	name := fmt.Sprintf("cm-%.60s", backup.Status.Id)

	if state.fileBackup.Labels == nil {
		state.fileBackup.Labels = make(map[string]string)
	}
	state.fileBackup.Labels[gcpclient.ManagedByKey] = gcpclient.ManagedByValue
	state.fileBackup.Labels[gcpclient.ScopeNameKey] = state.Scope.Name

	_, err := state.fileBackupClient.PatchFileBackup(ctx, project, location, name, "Labels", state.fileBackup)

	if err != nil {
		// Log and retry. This is not an essential action as it only impacts nuke on non prod envs that might already
		// have backups without necessary labels.
		logger.Error(err, "Error adding missing labels to File backup object in GCP")
	}

	return composed.StopWithRequeueDelay(gcpclient.GcpConfig.GcpOperationWaitTime), nil
}
