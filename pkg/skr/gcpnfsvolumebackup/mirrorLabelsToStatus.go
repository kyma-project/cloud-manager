package gcpnfsvolumebackup

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func mirrorLabelsToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
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

	if state.HasAllStatusLabels() {
		return nil, nil
	}

	backup.Status.FileStoreBackupLabels = state.fileBackup.Labels

	return composed.UpdateStatus(backup).
		SuccessLogMsg("Mirrored labels to status of GcpNfsVolumeBackup").
		ErrorLogMessage("Failed to mirror labels to status of GcpNfsVolumeBackup").
		Run(ctx, state)
}
