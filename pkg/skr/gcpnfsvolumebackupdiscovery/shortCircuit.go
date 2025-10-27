package gcpnfsvolumebackupdiscovery

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuit(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)

	backupDiscovery := state.ObjAsGcpNfsVolumeBackupDiscovery()

	if backupDiscovery.Status.State != cloudresourcesv1beta1.JobStateProcessing {
		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
