package gcpnfsvolumebackupdiscovery

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func setProcessing(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	if backupDiscovery.Status.State == cloudresourcesv1beta1.JobStateProcessing {
		return nil, ctx
	}

	if backupDiscovery.Status.State != "" {
		return nil, nil
	}

	backupDiscovery.Status.State = cloudresourcesv1beta1.JobStateProcessing
	return composed.UpdateStatus(backupDiscovery).
		ErrorLogMessage("Error: failed to set Processing status on GcpNfsVolumeBackupDiscovery").
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
